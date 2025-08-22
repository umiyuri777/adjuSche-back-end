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

	// カレンダー関連ルート
	r.GET("/calendar", servise.GetGoogleCalendarEvents)

	log.Println("サーバーを起動しています... http://localhost:8080")
	r.Run(":8080")
}

func getHello(c *gin.Context) {
	c.String(200, "Hello, World!")
}
