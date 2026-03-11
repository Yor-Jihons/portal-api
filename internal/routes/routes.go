package routes

import (
	"github.com/Yor-Jihons/portal-api/internal/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// study-histories関連のルート
	personGroup := r.Group("/study-histories")
	{
		personGroup.GET("", handlers.GetStudyHistories)
		personGroup.POST("", handlers.CreateStudyHistory)
	}

	return r
}
