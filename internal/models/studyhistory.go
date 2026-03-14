package models

type StudyHistory struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Categories  []string `json:"categories"`
	Date        string   `json:"date"`
	Time        string   `json:"time"`
}

// StudyHistoryResponse GETで返す全体のレスポンス形式
type StudyHistoryResponse struct {
	Status    int            `json:"status"`
	Histories []StudyHistory `json:"histories"` // 配列（スライス）として持つ
}
