package main

import (
	"adjuSche-back-end/servise"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Ginサーバーを起動
	r := gin.Default()

	// 基本ルート
	r.GET("/hello", getHello)

	// 認証関連ルート
	r.GET("/auth/start", servise.StartAuth)
	r.GET("/oauth/callback", servise.HandleOAuthCallback)

	// カレンダー関連ルート
	r.GET("/calendar", servise.GetGoogleCalendarEvents)

	log.Println("サーバーを起動しています... http://localhost:8080")
	log.Println("利用可能なエンドポイント:")
	log.Println("  GET /hello - ヘルスチェック")
	log.Println("  GET /auth/start - 認証プロセスを開始")
	log.Println("  GET /oauth/callback - OAuth認証コールバック")
	log.Println("  GET /calendar - Googleカレンダーのイベント取得")
	r.Run(":8080")
}

func getHello(c *gin.Context) {
	c.String(200, "Hello, World!")
}
