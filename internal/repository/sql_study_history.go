// internal/repository/sql_study_history.go
package repository

import (
	"database/sql"
	"os"

	"github.com/Yor-Jihons/portal-api/internal/models"
)

type SQLStudyHistoryRepository struct {
	DB *sql.DB
}

func InitDB() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// ローカル環境のSQLite
		dbURL = "file:dev.db"
	}

	// driverNameは "libsql" を使用
	db, err := sql.Open("libsql", dbURL)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (r *SQLStudyHistoryRepository) FetchAll() ([]models.StudyHistory, error) {
	rows, err := r.DB.Query("SELECT id, content, date FROM study_histories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []models.StudyHistory
	for rows.Next() {
		var h models.StudyHistory
		if err := rows.Scan(&h.ID, &h.Content, &h.Date); err != nil {
			return nil, err
		}
		histories = append(histories, h)
	}
	return histories, nil
}
