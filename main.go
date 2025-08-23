package main

import (
	"adjuSche-back-end/middleware"
	"adjuSche-back-end/presentation"
	"adjuSche-back-end/repository"
	"adjuSche-back-end/servise"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/v7/linebot"
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

		ctx := context.Background()

		mockUser := &repository.Users{
			GoogleID:  "mock-google-id" + fmt.Sprintf("%d", time.Now().UnixNano()),
			Name:      "Mock User",
			Email:     "msisisis@gmail.com" + fmt.Sprintf("%d", time.Now().UnixNano()),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = repo.CreateUser(ctx, mockUser)
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}

		//Event作成
		mockEvent := &repository.Events{
			HostUserID: mockUser.ID,
			Title:      "Mock Event",
			Note:   sql.NullString{String: "This is a mock event note.", Valid: true},
			ParticipantCount: 1,
			Status:     0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err = repo.CreateEvent(ctx, mockEvent)
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
	})

	allowOrigins := []string{"*"} // TODO: 本番環境では"*"を使用しない

	r.Use(middleware.CorsMiddleware(allowOrigins))

	log.Println("サーバーを起動しています... http://localhost:8080")
	r.Run(":8080")
}

func handleLineWebhook(c *gin.Context) {
	bot, err := linebot.New(
		os.Getenv("LINE_BOT_CHANNEL_SECRET"),
		os.Getenv("LINE_BOT_CHANNEL_TOKEN"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "LINEボットの初期化に失敗しました",
		})
		return
	}

	// リクエスト処理
	events, berr := bot.ParseRequest(c.Request)
	if berr != nil {
		fmt.Println(berr.Error())
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				_, rerr := bot.ReplyMessage(
					event.ReplyToken,
					linebot.NewTextMessage(getResMessage(message.Text)),
				).Do()
				if rerr != nil {
					fmt.Println(rerr.Error())
				}
			}
		}
	}
}

func getResMessage(message string) string {
	if message == "日程調整" {
		formURL := getFormURL()
		return formURL
	}
	return "日程調整をしたい場合は、「日程調整」と入力してください。"
}

func getFormURL() string {
	// TODO: LINEのメッセージ解析やフォーム生成ロジックの実装
	return "https://amazon.com"
}
