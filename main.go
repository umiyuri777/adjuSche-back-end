package main

import (
	"adjuSche-back-end/presentation"
	"adjuSche-back-end/repository"
	"adjuSche-back-end/servise"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("RENDER") == "" {
		err := godotenv.Load("./env/.env")
		if err != nil {
			log.Fatalf("環境変数の読み込みに失敗しました: %v\n", err)
		}
	}

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
		repo.InsertUser()
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
