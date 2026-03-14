package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Yor-Jihons/portal-api/internal/db"
	"github.com/Yor-Jihons/portal-api/internal/models"
)

// GETメソッドの場合
func GetStudyHistories(c *gin.Context) {
	// 1. SQLクエリ（GROUP_CONCAT でカテゴリをカンマ区切りで取得）
	query := `
		SELECT s.id, s.description, s.content, s.date, s.time, 
		    GROUP_CONCAT(c.category_name, ',') as categories
		FROM study_logs s
		LEFT JOIN study_log_categories r ON s.id = r.study_log_id
		LEFT JOIN categories c ON c.id = r.category_id
		GROUP BY s.id
		ORDER BY s.date DESC
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch histories: " + err.Error()})
		return
	}
	defer rows.Close()

	histories := []models.StudyHistory{} // nullではなく空配列を返すように初期化

	// 2. 取得した結果を1行ずつ処理
	for rows.Next() {
		var h models.StudyHistory
		var categoryStr sql.NullString // カテゴリが0個の場合を考慮して NullString を使う

		err := rows.Scan(&h.ID, &h.Description, &h.Content, &h.Date, &h.Time, &categoryStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan history: " + err.Error()})
			return
		}

		// カンマ区切りの文字列をスライスに変換
		if categoryStr.Valid && categoryStr.String != "" {
			h.Categories = strings.Split(categoryStr.String, ",")
		} else {
			h.Categories = []string{} // 空配列をセット
		}

		histories = append(histories, h)
	}

	// 3. レスポンスの返却
	data := models.StudyHistoryResponse{
		Status:    200,
		Histories: histories,
	}
	c.JSON(http.StatusOK, gin.H{"message": "get all histories", "data": data})
}

// POSTメソッドの場合(データを追加する)
func CreateStudyHistory(c *gin.Context) {
	var input models.StudyHistory

	// 1. JSONのバリデーション
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// トランザクション開始
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// 2. study_logs テーブルに保存
	query := `INSERT INTO study_logs (description, content, date, time) VALUES (?, ?, ?, ?) RETURNING id`
	var logID int
	err = tx.QueryRow(query, input.Description, input.Content, input.Date, input.Time).Scan(&logID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert log: " + err.Error()})
		return
	}

	// 3. カテゴリの処理と紐付け
	for _, catName := range input.Categories {
		var catID int
		// カテゴリが存在するか確認
		err := tx.QueryRow("SELECT id FROM categories WHERE category_name = ?", catName).Scan(&catID)

		if err != nil {
			if err == sql.ErrNoRows {
				// カテゴリが存在しない場合は新規作成
				err = tx.QueryRow("INSERT INTO categories (category_name) VALUES (?) RETURNING id", catName).Scan(&catID)
				if err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category: " + err.Error()})
					return
				}
			} else {
				// その他のDBエラー
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
				return
			}
		}

		// 中間テーブルに保存
		_, err = tx.Exec("INSERT INTO study_log_categories (study_log_id, category_id) VALUES (?, ?)", logID, catID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link category: " + err.Error()})
			return
		}
	}

	// コミットして確定
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "history created", "id": logID})
}
