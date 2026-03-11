package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Yor-Jihons/portal-api/internal/models"
)

// ここでは各HTTPメソッドごとに渡すコールバック関数を規定する

// GETメソッドの場合(本来はデータを返す)
func GetStudyHistories(c *gin.Context) {
	histories := []models.StudyHistory{
		{
			Description: "Go言語の基礎",
			Content:     "ディレクトリ構造について学んだ",
			Categories:  []string{"Go言語", "フレームワーク"},
			Date:        "2026/03/11",
			Time:        models.StudyTime{Hours: 0, Min: 30},
		},
	}

	data := models.StudyHistoryResponse{
		Status:    202,
		Histories: histories,
	}
	c.JSON(http.StatusOK, gin.H{"message": "get all histories", "data": data})
}

// POSTメソッドの場合(本来はデータを追加する)
func CreateStudyHistory(c *gin.Context) {
	var input models.StudyHistory

	// JSONを構造体に流し込む
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "history created"})
}
