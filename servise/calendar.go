package servise

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// CalendarService はGoogleカレンダーへのアクセスを管理する構造体です
type CalendarService struct {
	service *calendar.Service
}

// NewCalendarService は新しいCalendarServiceインスタンスを作成します
func NewCalendarService(client *http.Client) (*CalendarService, error) {
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("calendar APIサービスの作成に失敗しました: %v", err)
	}

	return &CalendarService{service: srv}, nil
}

// CalendarEvent はカレンダーイベントの情報を格納する構造体です
type CalendarEvent struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Location    string `json:"location"`
}

// GetEvents は指定された期間のカレンダーイベントを取得します
func (cs *CalendarService) GetEvents(maxResults int64, timeMin time.Time) ([]*CalendarEvent, error) {
	t := timeMin.Format(time.RFC3339)
	events, err := cs.service.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(maxResults).OrderBy("startTime").Do()
	if err != nil {
		return nil, fmt.Errorf("イベントの取得に失敗しました: %v", err)
	}

	var calendarEvents []*CalendarEvent
	for _, item := range events.Items {
		event := &CalendarEvent{
			Summary:     item.Summary,
			Description: item.Description,
			Location:    item.Location,
		}

		// 開始時間の設定
		if item.Start.DateTime != "" {
			event.StartTime = item.Start.DateTime
		} else {
			event.StartTime = item.Start.Date
		}

		// 終了時間の設定
		if item.End.DateTime != "" {
			event.EndTime = item.End.DateTime
		} else {
			event.EndTime = item.End.Date
		}

		calendarEvents = append(calendarEvents, event)
	}

	return calendarEvents, nil
}

// GetUpcomingEvents は今後のイベントを取得します（デフォルトで10件）
func (cs *CalendarService) GetUpcomingEvents() ([]*CalendarEvent, error) {
	return cs.GetEvents(10, time.Now())
}

// GetEventsInDateRange は指定された日付範囲のイベントを取得します
func (cs *CalendarService) GetEventsInDateRange(startDate, endDate time.Time, maxResults int64) ([]*CalendarEvent, error) {
	timeMin := startDate.Format(time.RFC3339)
	timeMax := endDate.Format(time.RFC3339)

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

// DisplayEvents はイベントをコンソールに表示します
func (cs *CalendarService) DisplayEvents(events []*CalendarEvent) {
	fmt.Println("=== Googleカレンダーのイベント ===")
	if len(events) == 0 {
		fmt.Println("該当するイベントはありません。")
		return
	}

	for _, event := range events {
		fmt.Printf("タイトル: %s\n", event.Summary)
		fmt.Printf("日時: %s\n", event.StartTime)
		if event.Description != "" {
			fmt.Printf("説明: %s\n", event.Description)
		}
		if event.Location != "" {
			fmt.Printf("場所: %s\n", event.Location)
		}
		fmt.Println("---")
	}
}

// GetGoogleCalendarEvents はHTTPエンドポイント用のカレンダーイベント取得関数です
func GetGoogleCalendarEvents(c *gin.Context) {
	// 認証サービスを初期化
	authService, err := NewAuthService()
	if err != nil {
		log.Printf("認証サービスの初期化に失敗しました: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "認証サービスの初期化エラー"})
		return
	}

	// 認証されたHTTPクライアントを取得
	client, err := authService.GetAuthenticatedClient()
	if err != nil {
		log.Printf("認証クライアントの取得に失敗しました: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です。/auth/start にアクセスして認証を完了してください。"})
		return
	}

	// カレンダーサービスを初期化
	calendarService, err := NewCalendarService(client)
	if err != nil {
		log.Printf("カレンダーサービスの初期化に失敗しました: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "カレンダーサービスの初期化エラー"})
		return
	}

	// 今後のイベントを取得
	events, err := calendarService.GetUpcomingEvents()
	if err != nil {
		log.Printf("イベントの取得に失敗しました: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "イベント取得エラー"})
		return
	}

	// イベントをコンソールに表示
	calendarService.DisplayEvents(events)

	// HTTPレスポンスとして返却
	c.JSON(http.StatusOK, gin.H{
		"message": "カレンダーイベントを正常に取得しました",
		"events":  events,
		"count":   len(events),
	})
}
