package models

type StudyHistory struct {
	ID          string   `json:"id"`
	Description string   `json:"description" binding:"required"`
	Content     string   `json:"content" binding:"required"`
	Ref         string   `json:"ref"`
	Categories  []string `json:"categories"`
	Date        string   `json:"date" binding:"required,datetime=2006-01-02"`
	Time        int      `json:"time" binding:"required,min=1"`
}

// StudyHistoryResponse GETで返す全体のレスポンス形式
type StudyHistoryResponse struct {
	Status    int            `json:"status"`
	Histories []StudyHistory `json:"histories"` // 配列（スライス）として持つ
}
