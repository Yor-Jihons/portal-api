package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetPing(t *testing.T) {
	// 準備
	gin.SetMode(gin.TestMode)
	
	// DBは使用しないが、StudyHistoryHandlerの構造体に必要なのでnilで渡すか
	// あるいはモック等が必要な場合は既存のsetupTestDBを利用する
	// pingはDBアクセスしないのでnilでも動作するはず
	handler := &StudyHistoryHandler{DB: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 実行
	handler.GetPing(c)

	// 検証
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "Pong!"}`, w.Body.String())
}
