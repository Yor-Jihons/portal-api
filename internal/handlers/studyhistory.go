package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"

	"github.com/Yor-Jihons/portal-api/internal/models"
)

type StudyHistoryHandler struct {
	DB *sql.DB
}

func NewStudyHistoryHandler(db *sql.DB) *StudyHistoryHandler {
	return &StudyHistoryHandler{DB: db}
}

// GETメソッド: 全履歴取得
func (h *StudyHistoryHandler) GetStudyHistories(c *gin.Context) {
	query := `
		SELECT s.id, s.description, s.content, strftime('%Y-%m-%d', s.date) as date, s.time, s.ref, 
		    GROUP_CONCAT(c.category_name, ',') as categories
		FROM study_logs s
		LEFT JOIN study_log_categories r ON s.id = r.study_log_id
		LEFT JOIN categories c ON c.id = r.category_id
		GROUP BY s.id
		ORDER BY s.date DESC
	`

	rows, err := h.DB.Query(query)
	if err != nil {
		slog.Error("Failed to query histories", "error", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error 1"})
		return
	}
	defer rows.Close()

	histories := []models.StudyHistory{}
	for rows.Next() {
		var h_data models.StudyHistory
		var categoryStr sql.NullString
		var ref sql.NullString

		err := rows.Scan(&h_data.ID, &h_data.Description, &h_data.Content, &h_data.Date, &h_data.Time, &ref, &categoryStr)
		if err != nil {
			slog.Error("Failed to scan history", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error 2"})
			return
		}

		if ref.Valid {
			h_data.Ref = ref.String
		} else {
			h_data.Ref = ""
		}

		if categoryStr.Valid && categoryStr.String != "" {
			h_data.Categories = strings.Split(categoryStr.String, ",")
		} else {
			h_data.Categories = []string{}
		}
		histories = append(histories, h_data)
	}

	data := models.StudyHistoryResponse{
		Status:    200,
		Histories: histories,
	}
	c.JSON(http.StatusOK, gin.H{"message": "get all histories", "data": data})
}

// POSTメソッド: 新規登録
func (h *StudyHistoryHandler) CreateStudyHistory(c *gin.Context) {
	var input models.StudyHistory
	if err := c.ShouldBindJSON(&input); err != nil {
		slog.Warn("Invalid input for creation", "error", err, "ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// サニタイズ
	p := bluemonday.StrictPolicy()
	input.Description = p.Sanitize(input.Description)
	input.Content = p.Sanitize(input.Content)
	input.Ref = p.Sanitize(input.Ref)

	tx, err := h.DB.Begin()
	if err != nil {
		slog.Error("Failed to start transaction", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error 3"})
		return
	}

	query := `INSERT INTO study_logs (description, content, date, time, ref) VALUES (?, ?, ?, ?, ?) RETURNING id`
	var logID int
	err = tx.QueryRow(query, input.Description, input.Content, input.Date, input.Time, input.Ref).Scan(&logID)
	if err != nil {
		tx.Rollback()
		slog.Error("Failed to insert log", "error", err, "description", input.Description)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error 4"})
		return
	}

	if err := h.syncCategories(tx, logID, input.Categories); err != nil {
		tx.Rollback()
		slog.Error("Failed to sync categories", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error 5"})
		return
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error 6"})
		return
	}

	slog.Info("History created successfully", "id", logID, "ip", c.ClientIP())
	c.JSON(http.StatusCreated, gin.H{"message": "history created", "id": logID})
}

// PUTメソッド: 更新
func (h *StudyHistoryHandler) UpdateStudyHistory(c *gin.Context) {
	id := c.Param("id")
	var input models.StudyHistory
	if err := c.ShouldBindJSON(&input); err != nil {
		slog.Warn("Invalid input for update", "id", id, "error", err, "ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// サニタイズ
	p := bluemonday.StrictPolicy()
	input.Description = p.Sanitize(input.Description)
	input.Content = p.Sanitize(input.Content)
	input.Ref = p.Sanitize(input.Ref)

	tx, err := h.DB.Begin()
	if err != nil {
		slog.Error("Failed to start transaction for update", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error 7"})
		return
	}

	query := `UPDATE study_logs SET description = ?, content = ?, date = ?, time = ?, ref = ? WHERE id = ?`
	result, err := tx.Exec(query, input.Description, input.Content, input.Date, input.Time, input.Ref, id)
	if err != nil {
		tx.Rollback()
		slog.Error("Failed to update log", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error 8"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "History not found"})
		return
	}

	// 中間テーブルの関連を一旦削除して再登録
	_, err = tx.Exec("DELETE FROM study_log_categories WHERE study_log_id = ?", id)
	if err != nil {
		tx.Rollback()
		slog.Error("Failed to clear old categories", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error 9"})
		return
	}

	if err := h.syncCategories(tx, id, input.Categories); err != nil {
		tx.Rollback()
		slog.Error("Failed to sync categories for update", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error 10"})
		return
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Failed to commit update", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error 11"})
		return
	}

	slog.Info("History updated successfully", "id", id, "ip", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "history updated"})
}

// DELETEメソッド: 削除
func (h *StudyHistoryHandler) DeleteStudyHistory(c *gin.Context) {
	id := c.Param("id")

	result, err := h.DB.Exec("DELETE FROM study_logs WHERE id = ?", id)
	if err != nil {
		slog.Error("Failed to delete history", "id", id, "error", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error 12"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "History not found"})
		return
	}

	slog.Info("History deleted successfully", "id", id, "ip", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "history deleted"})
}

// syncCategories はカテゴリの同期処理（新規作成と紐付け）を行います
func (h *StudyHistoryHandler) syncCategories(tx *sql.Tx, logID interface{}, categories []string) error {
	processedCategories := make(map[string]bool)
	for _, rawCatName := range categories {
		catName := strings.ToLower(strings.TrimSpace(rawCatName))
		if catName == "" || processedCategories[catName] {
			continue
		}
		processedCategories[catName] = true

		var catID int
		err := tx.QueryRow("SELECT id FROM categories WHERE category_name = ?", catName).Scan(&catID)
		if err != nil {
			if err == sql.ErrNoRows {
				err = tx.QueryRow("INSERT INTO categories (category_name) VALUES (?) RETURNING id", catName).Scan(&catID)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		_, err = tx.Exec("INSERT INTO study_log_categories (study_log_id, category_id) VALUES (?, ?)", logID, catID)
		if err != nil {
			return err
		}
	}
	return nil
}
