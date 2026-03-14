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

	// ハンドラーの初期化
	studyHandler := handlers.NewStudyHistoryHandler(db)

	// --- CORSの設定 ---
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"https://your-site.vercel.app",
		},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"X-API-KEY",
		},
		MaxAge: 12 * time.Hour,
	}))

	// ルート設定
	authGroup := r.Group("/study-histories")
	authGroup.Use(middlewares.ApiKeyAuth())
	{
		authGroup.GET("", studyHandler.GetStudyHistories)
		authGroup.POST("", studyHandler.CreateStudyHistory)
	}

	return r
}
