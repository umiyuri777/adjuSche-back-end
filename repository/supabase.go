package repository

import (
	"github.com/gin-gonic/gin"
)

func CreateEvent(c *gin.Context) {
	db := supabase.CreateClient()
	
}
