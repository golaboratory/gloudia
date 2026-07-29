package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/golaboratory/gloudia/auth"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ----- auth.go -----

func TestNewAuthProvider_HostClaimsTenantMismatch403(t *testing.T) {
	hexKey := auth.GenerateRandomKey()
	maker, err := auth.NewTokenMaker(hexKey)
	require.NoError(t, err)
	provider := NewAuthProvider(maker)

	// Token belongs to tenant-a, but host resolved tenant-b into the context.
	token, err := maker.CreateToken(7, "tenant-a", 3, time.Minute)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	// Simulate NewTenantResolution having set a host-derived tenant id.
	rCtx := context.WithValue(r.Context(), KeyTenantID, "tenant-b")
	r = r.WithContext(rCtx)
	ctx := humatest.NewContext(nil, r, w)

	nextCalled := false
	provider(ctx, func(c huma.Context) {
		nextCalled = true
	})

	assert.False(t, nextCalled, "next must not be called on tenant mismatch")
	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
}

func TestNewAuthProvider_SetsAuthoritativeTenantIDOnMatch(t *testing.T) {
	hexKey := auth.GenerateRandomKey()
	maker, err := auth.NewTokenMaker(hexKey)
	require.NoError(t, err)
	provider := NewAuthProvider(maker)

	token, err := maker.CreateToken(99, "tenant-x", 5, time.Minute)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	// Host-derived tenant MATCHES the claims tenant -> must pass through.
	rCtx := context.WithValue(r.Context(), KeyTenantID, "tenant-x")
	r = r.WithContext(rCtx)
	ctx := humatest.NewContext(nil, r, w)

	nextCalled := false
	provider(ctx, func(c huma.Context) {
		nextCalled = true
		// Authoritative tenant id must be set from verified claims.
		gotTenant, ok := c.Context().Value(KeyTenantID).(string)
		assert.True(t, ok, "KeyTenantID must be a string in context")
		assert.Equal(t, "tenant-x", gotTenant, "KeyTenantID must equal verified claims.TenantID")
		claims, ok := c.Context().Value(KeyClaims).(*auth.Claims)
		require.True(t, ok)
		assert.Equal(t, int64(99), claims.UserID)
		assert.Equal(t, int64(5), claims.RoleID)
	})

	assert.True(t, nextCalled, "next must be called when host tenant matches claims")
}

func TestNewAuthProvider_BearerSchemeParsingEdges(t *testing.T) {
	hexKey := auth.GenerateRandomKey()
	maker, err := auth.NewTokenMaker(hexKey)
	require.NoError(t, err)
	provider := NewAuthProvider(maker)

	goodToken, err := maker.CreateToken(1, "tenant1", 2, time.Minute)
	require.NoError(t, err)

	tests := []struct {
		name       string
		header     string
		wantNext   bool
		wantStatus int
	}{
		{"bearer keyword only, no token", "Bearer", false, http.StatusUnauthorized},
		{"whitespace only header treated as missing", "   ", false, http.StatusUnauthorized},
		{"lowercase scheme accepted", "bearer " + goodToken, true, 0},
		{"uppercase scheme accepted", "BEARER " + goodToken, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			r.Header.Set("Authorization", tt.header)
			ctx := humatest.NewContext(nil, r, w)

			nextCalled := false
			provider(ctx, func(c huma.Context) {
				nextCalled = true
			})

			assert.Equal(t, tt.wantNext, nextCalled)
			if !tt.wantNext {
				assert.Equal(t, tt.wantStatus, w.Result().StatusCode)
			}
		})
	}
}

// ----- rls.go -----

func TestNewRLSProvider_ClaimsTenantMismatch403(t *testing.T) {
	// db=nil is safe: the 403 mismatch path returns before db.Begin is reached.
	provider := NewRLSProvider(nil)

	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	otherUUID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	tests := []struct {
		name     string
		claims   *auth.Claims
		tenantID string
	}{
		{
			name:     "claims tenant differs from resolved tenant",
			claims:   &auth.Claims{TenantID: otherUUID, UserID: 1},
			tenantID: validUUID,
		},
		{
			name:     "claims tenant is empty",
			claims:   &auth.Claims{TenantID: "", UserID: 1},
			tenantID: validUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			rCtx := context.WithValue(r.Context(), KeyTenantID, tt.tenantID)
			rCtx = context.WithValue(rCtx, KeyClaims, tt.claims)
			r = r.WithContext(rCtx)
			ctx := humatest.NewContext(nil, r, w)

			nextCalled := false
			provider(ctx, func(c huma.Context) {
				nextCalled = true
			})

			assert.False(t, nextCalled, "next must not run on claims/tenant mismatch")
			assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
		})
	}
}

// ----- tenant.go -----

func TestIsTrustedProxy_Branches(t *testing.T) {
	_, cidr10, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)
	explicit := []*net.IPNet{cidr10}

	tests := []struct {
		name       string
		remoteAddr string
		trusted    []*net.IPNet
		want       bool
	}{
		// 許可リスト未設定時はいかなる送信元も信頼しない (fail-closed)。
		// プライベート/ループバックの暗黙信頼は、同一プライベート網に第三者の
		// ワークロードが同居する環境で成りすましを許すため廃止した。
		{"no list: loopback NOT trusted", "127.0.0.1:1234", nil, false},
		{"no list: private NOT trusted", "192.168.1.5:1234", nil, false},
		{"no list: public not trusted", "203.0.113.7:1234", nil, false},
		{"no list: bare ip without port NOT trusted", "127.0.0.1", nil, false},
		{"explicit list: in range trusted", "10.1.2.3:9999", explicit, true},
		{"explicit list: bare ip in range trusted", "10.1.2.3", explicit, true},
		{"explicit list: private but NOT in list rejected", "192.168.1.5:9999", explicit, false},
		{"explicit list: loopback NOT in list rejected", "127.0.0.1:9999", explicit, false},
		{"unparseable remote addr", "not-an-ip", explicit, false},
		{"empty remote addr", "", explicit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isTrustedProxy(tt.remoteAddr, tt.trusted))
		})
	}
}

func TestParseTrustedProxyCIDRs(t *testing.T) {
	t.Run("unset returns nil", func(t *testing.T) {
		t.Setenv("TRUSTED_PROXY_CIDRS", "")
		assert.Nil(t, parseTrustedProxyCIDRs())
	})

	t.Run("valid list parsed", func(t *testing.T) {
		t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8, 192.168.0.0/16")
		nets := parseTrustedProxyCIDRs()
		require.Len(t, nets, 2)
		assert.True(t, nets[0].Contains(net.ParseIP("10.1.2.3")))
		assert.True(t, nets[1].Contains(net.ParseIP("192.168.5.5")))
	})

	t.Run("invalid and empty entries skipped, valid kept", func(t *testing.T) {
		t.Setenv("TRUSTED_PROXY_CIDRS", "garbage, ,10.0.0.0/8,not/a/cidr")
		nets := parseTrustedProxyCIDRs()
		require.Len(t, nets, 1)
		assert.True(t, nets[0].Contains(net.ParseIP("10.9.9.9")))
	})
}

func TestNewTenantResolution_SkipPathAndExplicitCIDRGating(t *testing.T) {
	t.Run("skip path bypasses dispatcher and tenant context", func(t *testing.T) {
		md := new(MockDispatcher)
		mw := NewTenantResolution(md)
		nextCalled := false
		h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			assert.Nil(t, r.Context().Value(KeyTenantID), "skip path must not set tenant id")
			assert.Nil(t, r.Context().Value(KeyTenantDomainName))
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "http://tenant-a.example.com/health", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.True(t, nextCalled)
		assert.Equal(t, http.StatusOK, w.Code)
		// Dispatcher must not have been consulted for a skip path.
		md.AssertNotCalled(t, "FindTenantIDByDomainName", mock.Anything, mock.Anything)
	})

	t.Run("explicit CIDR: X-Forwarded-Host honored only from configured proxy", func(t *testing.T) {
		// Construct AFTER Setenv so parseTrustedProxyCIDRs picks up the list.
		t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8")
		md := new(MockDispatcher)
		mw := NewTenantResolution(md)
		h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test-Tenant-ID", r.Context().Value(KeyTenantID).(string))
			w.WriteHeader(http.StatusOK)
		}))

		// In-list proxy -> forwarded host honored (tenant-fwd).
		md.On("FindTenantIDByDomainName", mock.Anything, "tenant-fwd").Return("uuid-fwd", nil).Once()
		req := httptest.NewRequest("GET", "http://real-host.example.com/api", nil)
		req.Header.Set("X-Forwarded-Host", "tenant-fwd.example.com")
		req.RemoteAddr = "10.5.5.5:55555"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "uuid-fwd", w.Header().Get("X-Test-Tenant-ID"))

		// Private IP NOT in the configured list -> forwarded host ignored, real host used.
		md.On("FindTenantIDByDomainName", mock.Anything, "real-host").Return("uuid-real", nil).Once()
		req2 := httptest.NewRequest("GET", "http://real-host.example.com/api", nil)
		req2.Header.Set("X-Forwarded-Host", "attacker.example.com")
		req2.RemoteAddr = "192.168.1.1:40000"
		w2 := httptest.NewRecorder()
		h.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
		assert.Equal(t, "uuid-real", w2.Header().Get("X-Test-Tenant-ID"))

		md.AssertExpectations(t)
	})
}

// ----- logger.go -----

func TestNewLogger_DebugRedactsBodyInLog(t *testing.T) {
	t.Setenv("IS_DEBUG", "true")
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	mw := NewLogger()
	require.True(t, IsDebug, "NewLogger must enable IsDebug from IS_DEBUG env")

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req := httptest.NewRequest(http.MethodPost, "/login?token=topsecret", bytes.NewBufferString(`{"password":"hunter2"}`))
	rec := httptest.NewRecorder()
	mw(h).ServeHTTP(rec, req)

	out := buf.String()
	assert.Contains(t, out, `body=`, "debug log must include the body field")
	assert.Contains(t, out, `***`, "redacted marker must be present")
	assert.NotContains(t, out, "hunter2", "plaintext password must never be logged")
	assert.NotContains(t, out, "topsecret", "plaintext query token must never be logged")
}

func TestNewLogger_NonDebugOmitsBodyField(t *testing.T) {
	t.Setenv("IS_DEBUG", "false")
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	mw := NewLogger()
	require.False(t, IsDebug)

	var handlerBody string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		handlerBody = string(b)
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewBufferString(`{"password":"hunter2"}`))
	rec := httptest.NewRecorder()
	mw(h).ServeHTTP(rec, req)

	out := buf.String()
	assert.NotContains(t, out, "body=", "non-debug log must not include the body field")
	assert.NotContains(t, out, "hunter2")
	assert.Equal(t, `{"password":"hunter2"}`, handlerBody, "downstream handler must still receive the full body")
}

func TestRedactSensitive_ExtraKeysAndForms(t *testing.T) {
	cases := []struct {
		name           string
		in             string
		mustNotContain []string
		mustContain    []string
	}{
		{"json authorization redacted", `{"Authorization":"Bearer abc.def","id":1}`, []string{"abc.def", "Bearer"}, []string{`"Authorization":"***"`, `"id":1`}},
		{"json secret and api_key", `{"secret":"s1","api_key":"k1"}`, []string{"s1", "k1"}, []string{`"secret":"***"`, `"api_key":"***"`}},
		{"json refresh and access token", `{"access_token":"a1","refresh_token":"r1"}`, []string{"a1", "r1"}, []string{`"access_token":"***"`, `"refresh_token":"***"`}},
		{"case insensitive uppercase key", `{"PASSWORD":"P"}`, []string{`"P"`}, []string{`"PASSWORD":"***"`}},
		{"unicode multibyte value redacted", `{"password":"パスワード"}`, []string{"パスワード"}, []string{`"password":"***"`}},
		{"empty value redacted", `{"password":""}`, nil, []string{`"password":"***"`}},
		{"query authorization and secret", `Authorization=Bearer%20zzz&secret=mysecret&keep=1`, []string{"zzz", "mysecret"}, []string{"Authorization=***", "secret=***", "keep=1"}},
		{"multiple json fields in one body", `{"password":"p","token":"t","name":"alice"}`, []string{`"p"`, `"t"`}, []string{`"password":"***"`, `"token":"***"`, `"name":"alice"`}},
		{"false positive: word in free text not redacted", `{"note":"please reset your password soon"}`, nil, []string{"please reset your password soon"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := redactSensitive(c.in)
			for _, m := range c.mustNotContain {
				assert.NotContains(t, got, m)
			}
			for _, m := range c.mustContain {
				assert.Contains(t, got, m)
			}
		})
	}
}

// ----- ratelimit.go -----

func TestNewRedisRateLimiter_429HeaderContract(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cfg := RateLimitConfig{Rate: 1, Burst: 1, Period: time.Minute, Name: "hdr_test"}
	mw := NewRedisRateLimiter(rdb, cfg)

	makeReq := func() *http.Response {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/x", nil)
		r.Header.Set("X-Real-IP", "203.0.113.55")
		ctx := humatest.NewContext(nil, r, w)
		mw(ctx, func(c huma.Context) {})
		return w.Result()
	}

	first := makeReq()
	assert.Equal(t, http.StatusOK, first.StatusCode)

	res := makeReq()
	assert.Equal(t, http.StatusTooManyRequests, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.Equal(t, strconv.Itoa(cfg.Rate), res.Header.Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", res.Header.Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, res.Header.Get("X-RateLimit-Reset"))
	ra := res.Header.Get("Retry-After")
	require.NotEmpty(t, ra)
	raInt, convErr := strconv.Atoi(ra)
	require.NoError(t, convErr)
	assert.GreaterOrEqual(t, raInt, 1, "Retry-After must be clamped to at least 1 second")

	var body map[string]any
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	assert.Equal(t, float64(429), body["status"])
	assert.Equal(t, "Too Many Requests", body["title"])
	require.IsType(t, "", body["message"])
	assert.True(t, strings.Contains(body["message"].(string), "Retry after"))
}

func TestNewRedisRateLimiter_EmptyNameDefaultsKey(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cfg := RateLimitConfig{Rate: 1, Burst: 1, Period: time.Minute, Name: ""}
	// X-Real-IP は明示的に信頼したプロキシ経由の場合のみ採用される。
	// NewRedisRateLimiter は生成時に環境変数を解析するため、事前に設定する。
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8")
	mw := NewRedisRateLimiter(rdb, cfg)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/x", nil)
	r.RemoteAddr = "10.0.0.1:9000" // explicitly trusted peer so X-Real-IP is honored
	r.Header.Set("X-Real-IP", "198.51.100.10")
	ctx := humatest.NewContext(nil, r, w)
	called := false
	mw(ctx, func(c huma.Context) { called = true })
	assert.True(t, called)

	keys := mr.Keys()
	found := false
	for _, k := range keys {
		if strings.Contains(k, "ratelimit:default:198.51.100.10") {
			found = true
		}
	}
	assert.True(t, found, "empty Name must fall back to 'default' in the redis key, got %v", keys)
}

// ----- ratelimit.go (client ip) -----

func TestRateLimitClientIP_ForwardedForAndPrecedence(t *testing.T) {
	newCtx := func(remoteAddr, xff, xRealIP string) huma.Context {
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = remoteAddr
		if xff != "" {
			r.Header.Set("X-Forwarded-For", xff)
		}
		if xRealIP != "" {
			r.Header.Set("X-Real-IP", xRealIP)
		}
		return humatest.NewContext(nil, r, httptest.NewRecorder())
	}

	_, cidr10, cidrErr := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, cidrErr)
	trusted := []*net.IPNet{cidr10}

	t.Run("trusted peer uses first X-Forwarded-For entry", func(t *testing.T) {
		ctx := newCtx("10.0.0.1:9000", "198.51.100.7, 70.41.3.18, 150.172.238.178", "")
		assert.Equal(t, "198.51.100.7", rateLimitClientIP(ctx, trusted))
	})
	t.Run("X-Real-IP wins over X-Forwarded-For when both present", func(t *testing.T) {
		ctx := newCtx("10.0.0.1:9000", "9.9.9.9", "8.8.8.8")
		assert.Equal(t, "8.8.8.8", rateLimitClientIP(ctx, trusted))
	})
	t.Run("untrusted peer ignores forwarded headers, keys on direct peer", func(t *testing.T) {
		ctx := newCtx("203.0.113.9:5000", "10.1.2.3", "10.9.9.9")
		assert.Equal(t, "203.0.113.9", rateLimitClientIP(ctx, trusted))
	})
	t.Run("no trusted CIDR configured ignores forwarded headers even from private peer", func(t *testing.T) {
		ctx := newCtx("10.0.0.1:9000", "9.9.9.9", "8.8.8.8")
		assert.Equal(t, "10.0.0.1", rateLimitClientIP(ctx, nil))
	})
	t.Run("remote without port returned as-is for untrusted", func(t *testing.T) {
		ctx := newCtx("203.0.113.9", "10.1.2.3", "")
		assert.Equal(t, "203.0.113.9", rateLimitClientIP(ctx, nil))
	})
}

// ----- cors.go -----

func TestNewCORS_OriginMatrix(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	cases := []struct {
		name        string
		env         string
		origin      string
		wantAllowed string
	}{
		{"default localhost allowed", "", "http://localhost:5173", "http://localhost:5173"},
		{"localhost different port rejected", "", "http://localhost:3000", ""},
		{"wildcard subdomain reflects exact origin", "https://*.example.com", "https://sub.example.com", "https://sub.example.com"},
		{"wildcard does not match apex", "https://*.example.com", "https://example.com", ""},
		{"wildcard does not match attacker domain", "https://*.example.com", "https://example.com.evil.com", ""},
		{"multiple origins second entry allowed", "https://a.example.com,https://b.example.com", "https://b.example.com", "https://b.example.com"},
		{"whitespace and empty entries tolerated", " , https://app.example.com , ", "https://app.example.com", "https://app.example.com"},
		{"scheme mismatch rejected", "https://app.example.com", "http://app.example.com", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.env != "" {
				t.Setenv("CORS_ALLOWED_ORIGINS", c.env)
			}
			handler := NewCORS().Handler(next)
			req := httptest.NewRequest("GET", "http://api.example.com/resource", nil)
			req.Header.Set("Origin", c.origin)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, c.wantAllowed, w.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

func TestNewCORS_PreflightWithCredentials(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight must be short-circuited by CORS, not passed to next handler")
	})
	handler := NewCORS().Handler(next)
	req := httptest.NewRequest(http.MethodOptions, "http://api.example.com/resource", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "300", w.Header().Get("Access-Control-Max-Age"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
}
