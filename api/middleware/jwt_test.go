package middleware

import (
	"fmt"
	"testing"
	"time"

	"github.com/golaboratory/gloudia/core/text"
	gjwt "github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

// TestCreateJWT は、CreateJWT関数によって生成されたJWTトークンが正しく署名され、エンコードされているかを検証します。
func TestCreateJWT(t *testing.T) {
	// テスト用のClaims（jwt.goで使用されるClaims型がある前提）
	testClaims := Claims{
		UserID: "testuser",
		// 必要に応じて他のフィールドも追加可能
	}

	// JWTトークンの生成
	tokenStr, err := CreateJWT(testClaims)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	// JWTトークンの解析
	parsedToken, err := gjwt.Parse(tokenStr, func(token *gjwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*gjwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWTSecret), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// 生成時にシリアライズされたClaimsのJSON文字列を抽出
	mapClaims, ok := parsedToken.Claims.(gjwt.MapClaims)
	assert.True(t, ok)
	authJson, ok := mapClaims["auth"].(string)
	assert.True(t, ok)

	// JSON文字列をClaimsに復元
	parsedClaims, err := text.DeserializeJson[Claims](authJson)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", parsedClaims.UserID)
}

// TestJWTExpiration は、JWTトークンの有効期限が設定されているかを確認します。
func TestJWTExpiration(t *testing.T) {
	testClaims := Claims{
		UserID: "expiretest",
	}

	tokenStr, err := CreateJWT(testClaims)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	parsedToken, err := gjwt.Parse(tokenStr, func(token *gjwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	mapClaims, ok := parsedToken.Claims.(gjwt.MapClaims)
	assert.True(t, ok)

	expValue, ok := mapClaims["exp"].(float64)
	assert.True(t, ok)

	expTime := time.Unix(int64(expValue), 0)
	// 現在時刻より後であることを確認
	assert.True(t, expTime.After(time.Now()))
}
