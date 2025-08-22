package servise

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type EventConditions struct {
	PeriodStart string  `json:"periodStart" binding:"required"`
	PeriodEnd   string  `json:"periodEnd" binding:"required"`
	TimeType    *string `json:"timeType,omitempty"`
	TimeStart   string  `json:"timeStart"`
	TimeEnd     string  `json:"timeEnd"`
	DurationMin int     `json:"durationMin" binding:"required"`
}

type CreateEventRequest struct {
	Title            string          `json:"title" binding:"required"`
	Note             string          `json:"note"`
	ParticipantCount int             `json:"participantCount" binding:"required"`
	Conditions       EventConditions `json:"conditions" binding:"required"`
}

type CreateEventResponse struct {
	Status    string `json:"status"`
	InviteURL string `json:"invite_url"`
}

func CreateEvent(c *gin.Context) {
	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "無効なリクエストボディです",
		})
		return
	}

	//TODO: イベントの作成処理 DBに保存

	// ここでは仕様確定前のため、保存等は行わず招待URLのみを生成
	inviteURL := fmt.Sprintf("https://example.com/invite/%d", time.Now().UnixNano())

	resp := CreateEventResponse{
		Status:    "success",
		InviteURL: inviteURL,
	}

	c.JSON(http.StatusOK, resp)
}
