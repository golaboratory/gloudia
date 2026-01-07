package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDispatcher is a mock of Dispatcher interface
type MockDispatcher struct {
	mock.Mock
}

func (m *MockDispatcher) FindTenantIDByDomainName(ctx context.Context, domainName string) (string, error) {
	args := m.Called(ctx, domainName)
	return args.String(0), args.Error(1)
}

func TestNewTenantResolution(t *testing.T) {
	mockDispatcher := new(MockDispatcher)
	middleware := NewTenantResolution(mockDispatcher)

	// Dummy handler to check context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Context().Value(KeyTenantID)
		tenantName := r.Context().Value(KeyTenantDomainName)

		if tenantID == nil || tenantName == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("X-Test-Tenant-ID", tenantID.(string))
		w.Header().Set("X-Test-Tenant-Name", tenantName.(string))
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	t.Run("resolve from host", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://tenant-a.example.com/api", nil)
		w := httptest.NewRecorder()

		// Expect dispatcher call
		mockDispatcher.On("FindTenantIDByDomainName", mock.Anything, "tenant-a").Return("uuid-a", nil).Once()

		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "uuid-a", w.Header().Get("X-Test-Tenant-ID"))
		assert.Equal(t, "tenant-a", w.Header().Get("X-Test-Tenant-Name"))
		mockDispatcher.AssertExpectations(t)
	})

	t.Run("resolve from X-Forwarded-Host", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://localhost/api", nil)
		req.Header.Set("X-Forwarded-Host", "tenant-b.example.com")
		w := httptest.NewRecorder()

		// Expect dispatcher call
		mockDispatcher.On("FindTenantIDByDomainName", mock.Anything, "tenant-b").Return("uuid-b", nil).Once()

		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "uuid-b", w.Header().Get("X-Test-Tenant-ID"))
		mockDispatcher.AssertExpectations(t)
	})

	t.Run("failed resolution", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://unknown.example.com/api", nil)
		w := httptest.NewRecorder()

		// Expect dispatcher call returning error
		mockDispatcher.On("FindTenantIDByDomainName", mock.Anything, "unknown").Return("", errors.New("not found")).Once()

		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockDispatcher.AssertExpectations(t)
	})
}
