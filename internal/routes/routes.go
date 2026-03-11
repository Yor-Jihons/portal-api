package routes

import (
	"github.com/Yor-Jihons/portal-api/internal/handlers"
	"github.com/Yor-Jihons/portal-api/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS設定（VercelのURLからのアクセスのみ許可する設定をここに追加）

	// グループ全体に認証をかける
	authGroup := r.Group("/study-histories")
	authGroup.Use(middlewares.ApiKeyAuth()) // このグループ以下の全ルートに門番を設置
	{
		authGroup.GET("", handlers.GetStudyHistories)
		authGroup.POST("", handlers.CreateStudyHistory)
	}

	return r
}
