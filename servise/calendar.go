package servise

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Googleカレンダーへのアクセスを管理する構造体
type CalendarService struct {
	service *calendar.Service
}

// token.jsonファイルからCalendarServiceを作成
func NewCalendarServiceFromToken(tokenFile, credFile string) (*CalendarService, error) {
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

	// トークンファイルを読み込み
	token, err := loadTokenFromFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("トークンファイルの読み込みに失敗しました: %v", err)
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

// トークンファイルからOAuth2トークンを読み込み
func loadTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("トークンファイルを開けませんでした: %v", err)
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	if err != nil {
		return nil, fmt.Errorf("トークンのデコードに失敗しました: %v", err)
	}

	return token, nil
}

// token文字列からCalendarServiceを作成します
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

// parseTokenFromString は、token文字列をOAuth2トークンに変換します
func parseTokenFromString(tokenString string) (*oauth2.Token, error) {
	token := &oauth2.Token{}
	err := json.Unmarshal([]byte(tokenString), token)
	if err != nil {
		return nil, fmt.Errorf("トークンのパースに失敗しました: %v", err)
	}
	return token, nil
}

// extractTokenFromHeader は、リクエストヘッダからtokenを抽出します
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

// HTTPエンドポイント用のカレンダーイベント取得
func GetGoogleCalendarEvents(c *gin.Context) {
	const CredFile = "env/client_secret.json"

	// リクエストヘッダからtokenを取得
	tokenString, err := extractTokenFromHeader(c)
	if err != nil {
		log.Printf("トークンの取得に失敗しました: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "認証トークンが必要です",
			"detail": "AuthorizationヘッダーまたはX-Tokenヘッダーにトークンを設定してください",
		})
		return
	}

	// カレンダーサービスをヘッダーのtokenから初期化
	calendarService, err := NewCalendarServiceFromTokenString(tokenString, CredFile)
	if err != nil {
		log.Printf("カレンダーサービスの初期化に失敗しました: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":  "認証に失敗しました",
			"detail": "提供されたトークンが無効です",
		})
		return
	}

	// 今後のイベントを取得
	events, err := calendarService.GetUpcomingEvents()
	if err != nil {
		log.Printf("イベントの取得に失敗しました: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "イベント取得エラー",
			"detail": "Googleカレンダーからのイベント取得に失敗しました",
		})
		return
	}

	// HTTPレスポンスとして返却
	c.JSON(http.StatusOK, gin.H{
		"events": events,
	})
}
