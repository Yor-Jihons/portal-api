package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// ここでは各HTTPメソッドごとに渡すコールバック関数を規定する


// GETメソッドの場合(本来はデータを返す)
func GetPersons(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "get all persons"})
}

// POSTメソッドの場合(本来はデータを追加する)
func CreatePerson(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "person created"})
}
