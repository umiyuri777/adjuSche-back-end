package main

import (
	"adjuSche-back-end/repository"
	"adjuSche-back-end/servise"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/hello", getHello)

	r.POST("/calendar", servise.GetGoogleCalendarEvents)

	r.POST("/line/webhook", servise.HandleLineWebhook)

	r.POST("/event", servise.CreateEvent)

	r.POST("/invite", servise.InviteUser)

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

func getHello(c *gin.Context) {
	c.String(200, "Hello, World!")
}
