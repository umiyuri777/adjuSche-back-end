package servise

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type LineWebhookRequest struct {
	Message string `json:"message" binding:"required"`
}

type LineWebhookResponse struct {
	Status  string `json:"status"`
	FormURL string `json:"form_url"`
}

func HandleLineWebhook(c *gin.Context) {
	var req LineWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "無効なリクエストボディです",
		})
		return
	}

	// TODO: LINEのメッセージ解析やフォーム生成ロジックの実装
	formURL := fmt.Sprintf("https://example.com/form/%d", time.Now().UnixNano())

	resp := LineWebhookResponse{
		Status:  "success",
		FormURL: formURL,
	}

	c.JSON(http.StatusOK, resp)
}
