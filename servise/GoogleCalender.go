package servise

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// OAuth2設定とトークンファイルのパス
const (
	tokFile  = "env/token.json"
	credFile = "env/client_secret.json"
	// Google Cloud ConsoleのOAuth2設定で以下のURLを「承認済みのリダイレクトURI」に追加してください
	redirectURL = "http://localhost:8080/oauth/callback"
)

var (
	authCodeChan = make(chan string, 1)
	authMutex    sync.Mutex
)

// getClient はOAuth2トークンを取得し、認証されたHTTPクライアントを返します
func getClient(config *oauth2.Config) *http.Client {
	// トークンをファイルから取得またはWebから新規取得
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// getTokenFromWeb はWebブラウザを使用してOAuth2トークンを取得します
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authMutex.Lock()
	defer authMutex.Unlock()

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("ブラウザでこのURLにアクセスしてください: \n%v\n", authURL)
	fmt.Println("認証完了後、自動的に処理が続行されます...")

	// 認証コードの受信を待機
	var authCode string
	select {
	case authCode = <-authCodeChan:
		fmt.Println("認証コードを受信しました")
	case <-time.After(5 * time.Minute):
		log.Fatal("認証タイムアウト: 5分以内に認証を完了してください")
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("トークン取得に失敗しました: %v", err)
	}
	return tok
}

// tokenFromFile はファイルからトークンを取得します
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken はトークンをファイルに保存します
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("トークンをファイルに保存中: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("トークンファイルの保存に失敗しました: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// HandleOAuthCallback はOAuth2認証コールバックを処理します
func HandleOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "認証コードが見つかりません"})
		return
	}

	// 認証コードをチャンネルに送信
	select {
	case authCodeChan <- code:
		c.HTML(http.StatusOK, "", `
		<html>
		<head><title>認証完了</title></head>
		<body>
			<h1>✅ 認証が完了しました</h1>
			<p>このタブを閉じて、アプリケーションの画面に戻ってください。</p>
		</body>
		</html>
		`)
	default:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "認証処理が進行中です"})
	}
}

// GetGoogleCalendarEvents はGoogleカレンダーのイベントを取得してコンソールに表示します
func GetGoogleCalendarEvents(c *gin.Context) {
	// クライアントシークレットファイルを読み込み
	b, err := ioutil.ReadFile(credFile)
	if err != nil {
		log.Printf("クライアントシークレットファイルの読み込みに失敗しました: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "認証ファイルが見つかりません"})
		return
	}

	// OAuth2設定を生成
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Printf("OAuth2設定の作成に失敗しました: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OAuth2設定エラー"})
		return
	}
	config.RedirectURL = redirectURL

	// 認証されたHTTPクライアントを取得
	client := getClient(config)

	// Calendar APIサービスを作成
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Calendar APIサービスの作成に失敗しました: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Calendar APIサービスエラー"})
		return
	}

	// 今日から1週間後までのイベントを取得
	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Printf("イベントの取得に失敗しました: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "イベント取得エラー"})
		return
	}

	// イベントをコンソールに表示
	fmt.Println("今後のイベント:")
	if len(events.Items) == 0 {
		fmt.Println("今後のイベントはありません。")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("- %v (%v)\n", item.Summary, date)
		}
	}

	// HTTPレスポンスとして返却
	c.JSON(http.StatusOK, gin.H{
		"message": "カレンダーイベントをコンソールに表示しました",
		"events":  events.Items,
	})
}

// FetchAndDisplayCalendarEvents は単独でカレンダーイベントを取得・表示する関数です
func FetchAndDisplayCalendarEvents() error {
	// クライアントシークレットファイルを読み込み
	b, err := ioutil.ReadFile(credFile)
	if err != nil {
		return fmt.Errorf("クライアントシークレットファイルの読み込みに失敗しました: %v", err)
	}

	// OAuth2設定を生成
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return fmt.Errorf("OAuth2設定の作成に失敗しました: %v", err)
	}
	config.RedirectURL = redirectURL

	// 認証されたHTTPクライアントを取得
	client := getClient(config)

	// Calendar APIサービスを作成
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("calendar APIサービスの作成に失敗しました: %v", err)
	}

	// 今日から1週間後までのイベントを取得
	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		return fmt.Errorf("イベントの取得に失敗しました: %v", err)
	}

	// イベントをコンソールに表示
	fmt.Println("=== Googleカレンダーのイベント ===")
	if len(events.Items) == 0 {
		fmt.Println("今後のイベントはありません。")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("タイトル: %s\n", item.Summary)
			fmt.Printf("日時: %s\n", date)
			if item.Description != "" {
				fmt.Printf("説明: %s\n", item.Description)
			}
			fmt.Println("---")
		}
	}

	return nil
}
