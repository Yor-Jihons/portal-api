package middlewares

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestApiKeyAuth(t *testing.T) {
	// 環境変数の設定（テスト終了後に元に戻す）
	originalKey := os.Getenv("API_KEY")
	testKey := "test-secret-key"
	os.Setenv("API_KEY", testKey)
	defer os.Setenv("API_KEY", originalKey)

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		apiKey     string
		wantStatus int
	}{
		{
			name:       "Valid API Key",
			apiKey:     testKey,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Missing API Key",
			apiKey:     "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid API Key",
			apiKey:     "wrong-key",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// レコーダーとエンジンの準備
			w := httptest.NewRecorder()
			r := gin.New()
			
			// テスト対象のミドルウェアを登録
			r.Use(ApiKeyAuth())
			
			// ミドルウェアを通過したことを確認するためのダミーハンドラー
			r.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// リクエストの作成
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-KEY", tt.apiKey)
			}

			// 実行
			r.ServeHTTP(w, req)

			// 検証
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
