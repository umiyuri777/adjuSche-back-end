package presentation

import (
	"adjuSche-back-end/application"
	"adjuSche-back-end/servise"
	"fmt"
	"log"
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

	// userId(UUID) が提供されていれば、空き時間を Availabilities に保存
	if req.UserID != "" {
		log.Println("req.UserID", req.UserID)
		log.Printf("eventID: %d", eventID)
		intervals := make([]servise.TimeInterval, 0, len(slots))
		for _, s := range slots {
			intervals = append(intervals, servise.TimeInterval{Start: s.PeriodStart, End: s.PeriodEnd})
		}
		log.Printf("保存対象の空き時間スロット数: %d", len(intervals))

		err := application.SaveUserAvailabilitiesFromCalendar(c.Request.Context(), eventID, req.UserID, intervals)
		if err != nil {
			log.Printf("空き時間の保存に失敗しました: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  fmt.Sprintf("空き時間の保存に失敗しました: %v", err),
			})
			return
		}
		log.Println("空き時間の保存が完了しました")
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
