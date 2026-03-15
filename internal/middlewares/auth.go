package middlewares

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func ApiKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-KEY")
		expectedKey := os.Getenv("API_KEY")

		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "リクエストヘッダーが空です"})
			return
		}

		if key != expectedKey {
			// 【注意】本番では消すべきですが、テストのために不一致をログに出します
			fmt.Printf("Auth Debug: Received=[%s], Expected=[%s]\n", key, expectedKey)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":        "キーが一致しません",
				"received_len": len(key),
				"expected_len": len(expectedKey),
			})
			return
		}

		c.Next()
	}
}
