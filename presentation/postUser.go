package presentation

import (
	"adjuSche-back-end/application"
	"net/http"

	"github.com/gin-gonic/gin"
)

type InviteUserRequest struct {
	UserID  string `json:"userId" binding:"required"`
	EventID string `json:"eventId" binding:"required"`
}

type possibleDate struct {
	ID                   int    `json:"id"`
	Date                 string `json:"date,omitempty"`
	PeriodStart          string `json:"periodStart"`
	PeriodEnd            string `json:"periodEnd"`
	ParticipateMemberNum int    `json:"participate_member_num"`
}

type InviteUserResponse struct {
	EventName    string         `json:"eventName"`
	VotedCount   int            `json:"votedCount"`
	Memo         string         `json:"memo"`
	PeriodStart  string         `json:"periodStart"`
	PeriodEnd    string         `json:"periodEnd"`
	DurationMin  int            `json:"durationMin"`
	PossibleDate []possibleDate `json:"possibleDate"`
}

func InviteUser(c *gin.Context) {
	var req InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "無効なリクエストボディです",
		})
		return
	}

	application.InviteUser()

	resp := InviteUserResponse{
		EventName:   "サンプルイベント",
		VotedCount:  5,
		Memo:        "メモの例",
		PeriodStart: "2025-08-01T00:00:00Z",
		PeriodEnd:   "2025-08-31T23:59:59Z",
		DurationMin: 60,
		PossibleDate: []possibleDate{
			{ID: 1, Date: "2025-08-10", PeriodStart: "2025-08-10T09:00:00Z", PeriodEnd: "2025-08-10T18:00:00Z", ParticipateMemberNum: 3},
			{ID: 2, PeriodStart: "2025-08-15T13:00:00Z", PeriodEnd: "2025-08-15T17:00:00Z", ParticipateMemberNum: 2},
		},
	}

	c.JSON(http.StatusOK, resp)
}
