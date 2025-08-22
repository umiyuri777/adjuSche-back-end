package main

import (
	"adjuSche-back-end/servise"
	"log"
    "adjuSche-back-end/repository"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/hello", getHello)

	r.GET("/calendar", servise.GetGoogleCalendarEvents)

    r.GET("/test", func(c *gin.Context) {
        repo, err := repository.NewSupabaseRepository()
        if err != nil {
            c.String(500, "Error connecting to database: %v", err)
            return
        }
        repo.CreateUserTable()
        c.String(200, "User table created successfully")
    })
    log.Println("サーバーを起動しています... http://localhost:8080")
    r.Run(":8080")
}

func getHello(c *gin.Context) {
	c.String(200, "Hello, World!")
}
