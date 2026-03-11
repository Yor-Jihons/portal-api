// internal/routes/routes.go
package routes

import (
	"time"

	"github.com/Yor-Jihons/portal-api/internal/handlers"
	"github.com/Yor-Jihons/portal-api/internal/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// --- CORSの設定 ---
	r.Use(cors.New(cors.Config{
		// 許可するアクセス元（Next.jsのURLを指定）
		AllowOrigins: []string{
			"http://localhost:3000",        // ローカル開発用
			"https://your-site.vercel.app", // 本番用（後で書き換える）
		},
		// 許可するHTTPメソッド
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		// 許可するHTTPヘッダー
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"X-API-KEY", // ★APIキーを送るために必須！
		},
		// プリフライトリクエスト（OPTIONS）の結果をキャッシュする時間
		MaxAge: 12 * time.Hour,
	}))

	// ルート設定
	authGroup := r.Group("/study-histories")
	authGroup.Use(middlewares.ApiKeyAuth())
	{
		authGroup.GET("", handlers.GetStudyHistories)
		authGroup.POST("", handlers.CreateStudyHistory)
	}

	return r
}
