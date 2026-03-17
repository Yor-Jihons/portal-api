package routes

import (
	"database/sql"
	"time"

	"github.com/Yor-Jihons/portal-api/internal/handlers"
	"github.com/Yor-Jihons/portal-api/internal/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	// --- レート制限の設定 ---
	// IP ごとに 1秒間に 1リクエスト、最大 5リクエストのバーストを許可
	limiter := middlewares.NewIPRateLimiter(1, 5)
	r.Use(middlewares.RateLimitMiddleware(limiter))

	// ハンドラーの初期化
	studyHandler := handlers.NewStudyHistoryHandler(db)

	// --- CORSの設定 ---
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"https://*.vercel.app",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"X-API-KEY",
		},
		MaxAge: 12 * time.Hour,
	}))

	// ルート設定
	authGroup1 := r.Group("/study-histories")
	authGroup1.Use(middlewares.ApiKeyAuth())
	{
		authGroup1.GET("", studyHandler.GetStudyHistories)
		authGroup1.POST("", studyHandler.CreateStudyHistory)
		authGroup1.PUT("/:id", studyHandler.UpdateStudyHistory)
		authGroup1.DELETE("/:id", studyHandler.DeleteStudyHistory)
	}

	// ルート設定(コールドスタート対策)
	authGroup2 := r.Group("/ping")
	authGroup2.Use(middlewares.ApiKeyAuth())
	{
		authGroup2.GET("", studyHandler.GetPing)
	}

	return r
}
