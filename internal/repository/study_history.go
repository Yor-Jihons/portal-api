// internal/repository/study_history.go
package repository

import "github.com/Yor-Jihons/portal-api/internal/models"

type StudyHistoryRepository interface {
	FetchAll() ([]models.StudyHistory, error)
	// Create(history model.StudyHistory) error // 今後追加予定
}
