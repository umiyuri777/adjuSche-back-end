package presentation

import (
	"adjuSche-back-end/application"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type eventConditions struct {
	PeriodStart string `json:"periodStart" binding:"required"`
	PeriodEnd   string `json:"periodEnd" binding:"required"`
	TimeStart   string `json:"timeStart"`
	TimeEnd     string `json:"timeEnd"`
	DurationMin int    `json:"durationMin" binding:"required"`
}

type CreateEventRequest struct {
	HostUserID       string          `json:"hostUserID" binding:"required"`
	Title            string          `json:"title" binding:"required"`
	Memo             string          `json:"memo"`
	ParticipantCount int             `json:"participantCount" binding:"required"`
	Conditions       eventConditions `json:"conditions" binding:"required"`
}

type CreateEventResponse struct {
	Status  string `json:"status"`
	EventID string `json:"event_id"`
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

	id, err := application.CreateEventAndCondition(c.Request.Context(), application.CreateEventInput{
		HostUserID:       req.HostUserID,
		Title:            req.Title,
		Memo:             req.Memo,
		ParticipantCount: req.ParticipantCount,
		PeriodStart:      req.Conditions.PeriodStart,
		PeriodEnd:        req.Conditions.PeriodEnd,
		TimeStart:        req.Conditions.TimeStart,
		TimeEnd:          req.Conditions.TimeEnd,
		DurationMin:      req.Conditions.DurationMin,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CreateEventResponse{
		Status:  "success",
		EventID: strconv.FormatInt(id, 10),
	})
}
