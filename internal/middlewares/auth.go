package middlewares

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func ApiKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ヘッダーからキーを取得
		key := c.GetHeader("X-API-KEY")
		expectedKey := os.Getenv("API_KEY")

		// デバッグ用（Renderのログに表示されます）
		// ※確認後は必ず消してください！
		if key != expectedKey {
			fmt.Printf("Debug: Received[%s] Expected[%s]\n", key, expectedKey)
		}

		// キーが一致しない場合は 401 Unauthorized を返して終了
		if key == "" || key != expectedKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// 一致すれば次の処理（ハンドラー）へ進む
		c.Next()
	}
}
