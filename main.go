package main

import (
	"adjuSche-back-end/presentation"
	"adjuSche-back-end/repository"
	"adjuSche-back-end/servise"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})

	r.POST("/calendar", servise.GetGoogleCalendarEvents)

	r.POST("/line/webhook", handleLineWebhook)

	r.POST("/event", presentation.CreateEvent)

	r.POST("/invite", presentation.InviteUser)

	r.GET("/test", func(c *gin.Context) {
		repo, err := repository.NewSupabaseRepository()
		if err != nil {
			c.String(500, "Error connecting to database: %v", err)
			return
		}

		ctx := context.Background()

		mockUser := &repository.User{
			GoogleID:  "mock-google-id",
			Name:      "Mock User",
			Email:     "mockuser@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = repo.CreateUser(ctx, mockUser)
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
	})
	log.Println("サーバーを起動しています... http://localhost:8080")
	r.Run(":8080")
}

type LineWebhookRequest struct {
	Message string `json:"message" binding:"required"`
}

type LineWebhookResponse struct {
	Status  string `json:"status"`
	FormURL string `json:"form_url"`
}

func handleLineWebhook(c *gin.Context) {
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
