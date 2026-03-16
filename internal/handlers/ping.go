package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GETメソッド: 全履歴取得
func (h *StudyHistoryHandler) GetPing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Pong!"})
}
