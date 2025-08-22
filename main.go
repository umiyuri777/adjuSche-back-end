package main

// Ginを使用した場合の例
import (
    "github.com/gin-gonic/gin"
    "adjuSche-back-end/repository"
)

func main() {
    r := gin.Default()
    r.GET("/hello", getHello)
	r.GET("/getSchedule", getSchedule)
    r.GET("/test", func(c *gin.Context) {
        repo, err := repository.NewSupabaseRepository()
        if err != nil {
            c.String(500, "Error connecting to database: %v", err)
            return
        }
        repo.CreateUserTable()
        c.String(200, "User table created successfully")
    })
    r.Run(":8080")
}

func getHello(c *gin.Context) {
	c.String(200, "Hello, World!")
}

func getSchedule(c *gin.Context) {
	
}