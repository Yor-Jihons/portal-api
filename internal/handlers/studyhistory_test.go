package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Yor-Jihons/portal-api/internal/middlewares"
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
		time INTEGER NOT NULL,
		ref TEXT
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
	db := setupTestDB(t)
	defer db.Close()

	_, _ = db.Exec("INSERT INTO study_logs (id, description, content, date, time, ref) VALUES (1, 'Test Title', 'Test Content', '2024-03-14', 60, 'https://example.com')")
	_, _ = db.Exec("INSERT INTO categories (id, category_name) VALUES (1, 'go')")
	_, _ = db.Exec("INSERT INTO study_log_categories (study_log_id, category_id) VALUES (1, 1)")

	handler := NewStudyHistoryHandler(db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.GetStudyHistories(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateStudyHistory_Validation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	handler := NewStudyHistoryHandler(db)

	tests := []struct {
		nameBody string
		body     string
		wantCode int
	}{
		{"Valid", `{"description": "New Study", "content": "Learning Tests", "ref": "https://example.com", "date": "2024-03-14", "time": 60, "categories": ["go"]}`, http.StatusCreated},
		{"Invalid Date Format", `{"description": "New Study", "content": "Learning Tests", "date": "2024/03/14", "time": 60}`, http.StatusBadRequest},
		{"Invalid Time Type", `{"description": "New Study", "content": "Learning Tests", "date": "2024-03-14", "time": "10:00"}`, http.StatusBadRequest},
		{"Time Too Small", `{"description": "New Study", "content": "Learning Tests", "date": "2024-03-14", "time": 0}`, http.StatusBadRequest},
		{"Missing Description", `{"content": "Learning Tests", "date": "2024-03-14", "time": 60}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.nameBody, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/study-histories", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateStudyHistory(c)
			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestCreateStudyHistory_Sanitize(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	handler := NewStudyHistoryHandler(db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// HTMLタグを含むボディ
	body := `{"description": "<b>Bold Title</b>", "content": "<script>alert(1)</script>Content", "ref": "<i>Ref</i>", "date": "2024-03-14", "time": 60}`
	c.Request = httptest.NewRequest("POST", "/study-histories", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateStudyHistory(c)

	if w.Code != http.StatusCreated {
		t.Logf("Response Body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	// DBの値を確認（タグが消えていること）
	var desc, content, ref string
	err := db.QueryRow("SELECT description, content, ref FROM study_logs LIMIT 1").Scan(&desc, &content, &ref)
	assert.NoError(t, err)
	assert.Equal(t, "Bold Title", desc)
	assert.Equal(t, "Content", content)
	assert.Equal(t, "Ref", ref)
}

func TestCreateStudyHistory_CategoryNormalization(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	handler := NewStudyHistoryHandler(db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 表記ゆれのあるカテゴリ
	body := `{"description": "Study", "content": "Content", "ref": "ref", "date": "2024-03-14", "time": 60, "categories": ["  Go  ", "GO", "go"]}`
	c.Request = httptest.NewRequest("POST", "/study-histories", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateStudyHistory(c)

	if w.Code != http.StatusCreated {
		t.Logf("Response Body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	// カテゴリテーブルに "go" だけが 1件登録されていることを確認
	var count int
	err := db.QueryRow("SELECT count(*) FROM categories WHERE category_name = 'go'").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	var totalCount int
	err = db.QueryRow("SELECT count(*) FROM categories").Scan(&totalCount)
	assert.NoError(t, err)
	assert.Equal(t, 1, totalCount)
}

func TestUpdateStudyHistory(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	handler := NewStudyHistoryHandler(db)

	// 初期データの挿入
	_, _ = db.Exec("INSERT INTO study_logs (id, description, content, date, time, ref) VALUES (1, 'Old Title', 'Old Content', '2024-03-14', 60, 'old-ref')")
	_, _ = db.Exec("INSERT INTO categories (id, category_name) VALUES (1, 'old-cat')")
	_, _ = db.Exec("INSERT INTO study_log_categories (study_log_id, category_id) VALUES (1, 1)")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	// 更新用ボディ
	body := `{"description": "New Title", "content": "New Content", "ref": "new-ref", "date": "2024-03-15", "time": 90, "categories": ["new-cat"]}`
	c.Request = httptest.NewRequest("PUT", "/study-histories/1", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateStudyHistory(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// DBの値が更新されているか確認
	var desc, content, ref, date string
	var timeVal int
	err := db.QueryRow("SELECT description, content, date, time, ref FROM study_logs WHERE id = 1").Scan(&desc, &content, &date, &timeVal, &ref)
	assert.NoError(t, err)
	assert.Equal(t, "New Title", desc)
	assert.Equal(t, "new-ref", ref)
	// SQLite の DATE 型は環境やドライバによって形式が異なる場合があるため、接頭辞でチェック
	assert.True(t, strings.HasPrefix(date, "2024-03-15"), "Expected date to start with 2024-03-15, got %s", date)
	assert.Equal(t, 90, timeVal)

	// カテゴリが更新されているか確認
	var catName string
	err = db.QueryRow("SELECT c.category_name FROM categories c JOIN study_log_categories r ON c.id = r.category_id WHERE r.study_log_id = 1").Scan(&catName)
	assert.NoError(t, err)
	assert.Equal(t, "new-cat", catName)
}

func TestDeleteStudyHistory(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	handler := NewStudyHistoryHandler(db)

	// 初期データの挿入
	_, _ = db.Exec("INSERT INTO study_logs (id, description, content, date, time, ref) VALUES (1, 'To be deleted', 'Content', '2024-03-14', 60, 'ref')")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	c.Request = httptest.NewRequest("DELETE", "/study-histories/1", nil)

	handler.DeleteStudyHistory(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// DBから消えているか確認
	var count int
	err := db.QueryRow("SELECT count(*) FROM study_logs WHERE id = 1").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 1秒間に 1リクエスト、バースト1の設定でテスト
	limiter := middlewares.NewIPRateLimiter(1, 1)
	r.Use(middlewares.RateLimitMiddleware(limiter))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// 1回目は成功
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "1.2.3.4:1234"
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 2回目は即座にエラーになるはず（1秒待っていないため）
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "1.2.3.4:1234"
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	// 別のIPからは成功する
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "5.6.7.8:1234"
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)

	// 1.1秒待てば成功する
	time.Sleep(1100 * time.Millisecond)
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/test", nil)
	req4.RemoteAddr = "1.2.3.4:1234"
	r.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)
}
