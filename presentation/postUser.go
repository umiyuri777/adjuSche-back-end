package presentation

import (
	"adjuSche-back-end/application"
	"adjuSche-back-end/servise"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type InviteUserRequest struct {
	UserID  string `json:"userId"`
	EventID string `json:"eventId"`
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
	const CredFile = "client_secret.json"

	var req InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "無効なリクエストボディです",
		})
		return
	}

	tokenString, err := servise.ExtractTokenFromHeader(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "認証トークンが必要です",
		})
		return
	}

	// eventId を int64 へ
	eventID, err := strconv.ParseInt(req.EventID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "eventId は数値で指定してください"})
		return
	}

	summary, slots, err := application.BuildInviteResponse(c.Request.Context(), eventID, tokenString, CredFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	// 整形
	res := InviteUserResponse{
		EventName:   summary.EventName,
		VotedCount:  summary.VotedCount,
		Memo:        summary.Memo,
		PeriodStart: summary.PeriodStart.Format(time.RFC3339),
		PeriodEnd:   summary.PeriodEnd.Format(time.RFC3339),
		DurationMin: summary.DurationMin,
	}
	for _, s := range slots {
		res.PossibleDate = append(res.PossibleDate, possibleDate{
			ID:                   s.ID,
			Date:                 s.Date,
			PeriodStart:          s.PeriodStart.Format(time.RFC3339),
			PeriodEnd:            s.PeriodEnd.Format(time.RFC3339),
			ParticipateMemberNum: s.ParticipateMemberNum,
		})
	}

	c.JSON(http.StatusOK, res)
}
