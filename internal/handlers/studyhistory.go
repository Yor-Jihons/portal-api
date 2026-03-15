package handlers

import (
	"database/sql"
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

// GETメソッドの場合
func (h *StudyHistoryHandler) GetStudyHistories(c *gin.Context) {
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

	rows, err := h.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch histories: " + err.Error()})
		return
	}
	defer rows.Close()

	histories := []models.StudyHistory{} // nullではなく空配列を返すように初期化

	// 2. 取得した結果を1行ずつ処理
	for rows.Next() {
		var h_data models.StudyHistory
		var categoryStr sql.NullString // カテゴリが0個の場合を考慮して NullString を使う

		err := rows.Scan(&h_data.ID, &h_data.Description, &h_data.Content, &h_data.Date, &h_data.Time, &categoryStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan history: " + err.Error()})
			return
		}

		// カンマ区切りの文字列をスライスに変換
		if categoryStr.Valid && categoryStr.String != "" {
			h_data.Categories = strings.Split(categoryStr.String, ",")
		} else {
			h_data.Categories = []string{} // 空配列をセット
		}

		histories = append(histories, h_data)
	}

	// 3. レスポンスの返却
	data := models.StudyHistoryResponse{
		Status:    200,
		Histories: histories,
	}
	c.JSON(http.StatusOK, gin.H{"message": "get all histories", "data": data})
}

// POSTメソッドの場合(データを追加する)
func (h *StudyHistoryHandler) CreateStudyHistory(c *gin.Context) {
	var input models.StudyHistory

	// 1. JSONのバリデーション
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// サニタイズ（XSS対策）
	p := bluemonday.StrictPolicy()
	input.Description = p.Sanitize(input.Description)
	input.Content = p.Sanitize(input.Content)

	// トランザクション開始
	tx, err := h.DB.Begin()
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
	processedCategories := make(map[string]bool)
	for _, rawCatName := range input.Categories {
		// カテゴリ名の正規化（空白削除、小文字化）
		catName := strings.ToLower(strings.TrimSpace(rawCatName))
		if catName == "" || processedCategories[catName] {
			continue // 空文字または重複したカテゴリ名はスキップ
		}
		processedCategories[catName] = true

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
