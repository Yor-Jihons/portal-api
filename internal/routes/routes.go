package routes

import (
	"github.com/Yor-Jihons/portal-api/internal/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Person関連のルート
	personGroup := r.Group("/persons")
	{
		personGroup.GET("", handlers.GetPersons)
		personGroup.POST("", handlers.CreatePerson)
	}

	// Product関連のルート
	productGroup := r.Group("/products")
	{
		productGroup.GET("", handlers.GetProducts)
		productGroup.POST("", handlers.CreateProduct)
	}

	return r
}
