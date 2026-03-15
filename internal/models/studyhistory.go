package models

type StudyHistory struct {
	ID          string   `json:"id"`
	Description string   `json:"description" binding:"required"`
	Content     string   `json:"content" binding:"required"`
	Categories  []string `json:"categories"`
	Date        string   `json:"date" binding:"required,datetime=2006-01-02"`
	Time        string   `json:"time" binding:"required,datetime=15:04"`
}

// StudyHistoryResponse GETで返す全体のレスポンス形式
type StudyHistoryResponse struct {
	Status    int            `json:"status"`
	Histories []StudyHistory `json:"histories"` // 配列（スライス）として持つ
}
