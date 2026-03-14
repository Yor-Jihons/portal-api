package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// テーブル作成
	schema := `
	CREATE TABLE categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category_name TEXT NOT NULL UNIQUE
	);
	CREATE TABLE study_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		description TEXT NOT NULL,
		content TEXT,
		date DATE NOT NULL DEFAULT (CURRENT_DATE),
		time TEXT NOT NULL
	);
	CREATE TABLE study_log_categories (
		study_log_id INTEGER REFERENCES study_logs(id) ON DELETE CASCADE,
		category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
		PRIMARY KEY (study_log_id, category_id)
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	return db
}

func TestGetStudyHistories(t *testing.T) {
	// 1. セットアップ
	db := setupTestDB(t)
	defer db.Close()

	// テストデータの挿入
	_, _ = db.Exec("INSERT INTO study_logs (id, description, content, date, time) VALUES (1, 'Test Title', 'Test Content', '2024-03-14', '1h')")
	_, _ = db.Exec("INSERT INTO categories (id, category_name) VALUES (1, 'Go')")
	_, _ = db.Exec("INSERT INTO study_log_categories (study_log_id, category_id) VALUES (1, 1)")

	handler := NewStudyHistoryHandler(db)

	// 2. HTTPリクエストのシミュレーション
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 3. ハンドラーの実行
	handler.GetStudyHistories(c)

	// 4. 検証 (testify/assertを使用)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "get all histories", response["message"])
	
	data := response["data"].(map[string]interface{})
	histories := data["histories"].([]interface{})
	assert.Len(t, histories, 1)

	first := histories[0].(map[string]interface{})
	assert.Equal(t, "Test Title", first["description"])
	assert.Equal(t, "Test Content", first["content"])
	
	categories := first["categories"].([]interface{})
	assert.Len(t, categories, 1)
	assert.Equal(t, "Go", categories[0])
}

func TestCreateStudyHistory(t *testing.T) {
	// 1. セットアップ
	db := setupTestDB(t)
	defer db.Close()

	handler := NewStudyHistoryHandler(db)

	// 2. HTTPリクエストのシミュレーション (POST)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// JSON ボディの作成
	body := `{"description": "New Study", "content": "Learning Tests", "date": "2024-03-14", "time": "2h", "categories": ["Go", "Testing"]}`
	c.Request = httptest.NewRequest("POST", "/study-histories", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// 3. ハンドラーの実行
	handler.CreateStudyHistory(c)

	// 4. 検証
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "history created", response["message"])
	assert.NotNil(t, response["id"])

	// DBに正しく保存されたか確認
	var count int
	err = db.QueryRow("SELECT count(*) FROM study_logs").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	var catCount int
	err = db.QueryRow("SELECT count(*) FROM categories").Scan(&catCount)
	assert.NoError(t, err)
	assert.Equal(t, 2, catCount) // "Go" と "Testing"
}
