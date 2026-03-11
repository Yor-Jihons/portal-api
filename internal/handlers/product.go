package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ここでは各HTTPメソッドごとに渡すコールバック関数を規定する

// GETメソッドの場合(本来はデータを返す)
func GetProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "get all products"})
}

// POSTメソッドの場合(本来はデータを追加する)
func CreateProduct(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "product created"})
}
