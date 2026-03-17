package routes

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestFullRouteIntegration(t *testing.T) {
	// 準備
	db := setupTestDB(t)
	defer db.Close()

	// 環境変数の設定
	originalKey := os.Getenv("API_KEY")
	testKey := "test-secret-api-key"
	os.Setenv("API_KEY", testKey)
	defer os.Setenv("API_KEY", originalKey)

	gin.SetMode(gin.TestMode)
	router := SetupRouter(db)

	tests := []struct {
		name           string
		method         string
		path           string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "Ping with valid key",
			method:         "GET",
			path:           "/ping",
			apiKey:         testKey,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ping without key",
			method:         "GET",
			path:           "/ping",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Get histories with valid key",
			method:         "GET",
			path:           "/study-histories",
			apiKey:         testKey,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get histories with invalid key",
			method:         "GET",
			path:           "/study-histories",
			apiKey:         "wrong-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "CORS Preflight",
			method:         "OPTIONS",
			path:           "/ping",
			apiKey:         "",
			expectedStatus: http.StatusNoContent, // CORS middleware handles OPTIONS
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			
			// レート制限は IP ごとなので、テストケースごとに IP を変えることで回避
			req.RemoteAddr = "1.2.3." + string(rune('1'+i)) + ":1234"
			
			if tt.apiKey != "" {
				req.Header.Set("X-API-KEY", tt.apiKey)
			}
			
			// CORS 用のヘッダー（必要に応じて）
			if tt.method == "OPTIONS" {
				req.Header.Set("Origin", "http://localhost:3000")
				req.Header.Set("Access-Control-Request-Method", "GET")
			}

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}

	// データの書き込みテスト（一貫性の確認）
	t.Run("Create History Integration", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"description": "Integration Test", "content": "Running full integration test", "date": "2024-03-14", "time": "12:00", "categories": ["test"]}`
		req, _ := http.NewRequest("POST", "/study-histories", strings.NewReader(body))
		req.RemoteAddr = "1.2.3.9:1234"
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", testKey)

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// 実際に書き込まれたか確認
		var count int
		err := db.QueryRow("SELECT count(*) FROM study_logs WHERE description = 'Integration Test'").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}
