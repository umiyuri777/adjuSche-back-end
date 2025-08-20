package main

// Ginを使用した場合の例
import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    r.GET("/hello", getHello)
	r.GET("/getSchedule", getSchedule)
    r.Run(":8080")
}

func getHello(c *gin.Context) {
	c.String(200, "Hello, World!")
}

func getSchedule(c *gin.Context) {
	
}