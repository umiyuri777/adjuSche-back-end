package servise

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type CalendarService struct {
	service *calendar.Service
}

func NewCalendarServiceFromTokenString(tokenString, credFile string) (*CalendarService, error) {
	// クライアントシークレットファイルを読み込み
	credData, err := ioutil.ReadFile(credFile)
	if err != nil {
		return nil, fmt.Errorf("クライアントシークレットファイルの読み込みに失敗しました: %v", err)
	}

	// OAuth2設定を生成
	config, err := google.ConfigFromJSON(credData, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("OAuth2設定の作成に失敗しました: %v", err)
	}

	// token文字列をOAuth2トークンに変換
	token, err := parseTokenFromString(tokenString)
	if err != nil {
		return nil, fmt.Errorf("トークンの解析に失敗しました: %v", err)
	}

	// HTTPクライアントを作成
	client := config.Client(context.Background(), token)

	// カレンダーサービスを作成
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("calendar APIサービスの作成に失敗しました: %v", err)
	}

	return &CalendarService{service: srv}, nil
}

func parseTokenFromString(tokenString string) (*oauth2.Token, error) {
	token := &oauth2.Token{}
	err := json.Unmarshal([]byte(tokenString), token)
	if err != nil {
		return nil, fmt.Errorf("トークンのパースに失敗しました: %v", err)
	}
	return token, nil
}

func extractTokenFromHeader(c *gin.Context) (string, error) {
	// Authorizationヘッダーから取得
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// "Bearer token_data" 形式から token_data 部分を抽出
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer "), nil
		}
		// "Bearer" プレフィックスがない場合はそのまま使用
		return authHeader, nil
	}

	// X-Token ヘッダーから取得
	tokenHeader := c.GetHeader("X-Token")
	if tokenHeader != "" {
		return tokenHeader, nil
	}

	return "", fmt.Errorf("リクエストヘッダにトークンが見つかりません")
}

type CalendarEvent struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Location    string `json:"location"`
}

type DateRangeRequest struct {
	StartDate  string `json:"start_date" binding:"required"` // RFC3339形式の開始日時
	EndDate    string `json:"end_date" binding:"required"`   // RFC3339形式の終了日時
}

func (cs *CalendarService) GetEventsInDateRange(startDate, endDate time.Time) ([]*CalendarEvent, error) {
	timeMin := startDate.Format(time.RFC3339)
	timeMax := endDate.Format(time.RFC3339)

	// イベントの最大取得件数を1000件に設定
	maxResults := int64(1000)

	events, err := cs.service.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(timeMin).TimeMax(timeMax).
		MaxResults(maxResults).OrderBy("startTime").Do()
	if err != nil {
		return nil, fmt.Errorf("指定期間のイベント取得に失敗しました: %v", err)
	}

	var calendarEvents []*CalendarEvent
	for _, item := range events.Items {
		event := &CalendarEvent{
			Summary:     item.Summary,
			Description: item.Description,
			Location:    item.Location,
		}

		if item.Start.DateTime != "" {
			event.StartTime = item.Start.DateTime
		} else {
			event.StartTime = item.Start.Date
		}

		if item.End.DateTime != "" {
			event.EndTime = item.End.DateTime
		} else {
			event.EndTime = item.End.Date
		}

		calendarEvents = append(calendarEvents, event)
	}

	return calendarEvents, nil
}

func GetGoogleCalendarEvents(c *gin.Context) {
	const CredFile = "env/client_secret.json"

	tokenString, err := extractTokenFromHeader(c)
	if err != nil {
		log.Printf("トークンの取得に失敗しました: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "認証トークンが必要です",
		})
		return
	}

	calendarService, err := NewCalendarServiceFromTokenString(tokenString, CredFile)
	if err != nil {
		log.Printf("カレンダーサービスの初期化に失敗しました: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "提供されたトークンが無効です",
		})
		return
	}

	var req DateRangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("リクエストボディのバインドに失敗しました: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "JSON形式のstart_date, end_dateを指定してください (RFC3339)",
		})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		log.Printf("開始日時の解析に失敗しました: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "start_dateはRFC3339形式で指定してください",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		log.Printf("終了日時の解析に失敗しました: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "end_dateはRFC3339形式で指定してください",
		})
		return
	}

	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "end_dateはstart_date以降である必要があります",
		})
		return
	}

	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 50
	}

	events, err := calendarService.GetEventsInDateRange(startTime, endTime)
	if err != nil {
		log.Printf("イベントの取得に失敗しました: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Googleカレンダーからのイベント取得に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
	})
}
