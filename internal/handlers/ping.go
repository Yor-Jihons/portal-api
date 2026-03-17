package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// コールドスタート用
func (h *StudyHistoryHandler) GetPing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Pong!"})
}
