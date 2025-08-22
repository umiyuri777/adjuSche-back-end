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
)

// OAuth2設定とトークンファイルのパス
const (
	TokenFile   = "env/token.json"
	CredFile    = "env/client_secret.json"
	RedirectURL = "http://localhost:8080/oauth/callback"
)

var (
	authCodeChan = make(chan string, 1)
	authMutex    sync.Mutex
)

// AuthService はOAuth2認証を管理する構造体です
type AuthService struct {
	config *oauth2.Config
}

// NewAuthService は新しいAuthServiceインスタンスを作成します
func NewAuthService() (*AuthService, error) {
	// クライアントシークレットファイルを読み込み
	b, err := ioutil.ReadFile(CredFile)
	if err != nil {
		return nil, fmt.Errorf("クライアントシークレットファイルの読み込みに失敗しました: %v", err)
	}

	// OAuth2設定を生成
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("OAuth2設定の作成に失敗しました: %v", err)
	}
	config.RedirectURL = RedirectURL

	return &AuthService{config: config}, nil
}

// GetAuthenticatedClient は認証されたHTTPクライアントを返します
func (a *AuthService) GetAuthenticatedClient() (*http.Client, error) {
	// トークンをファイルから取得またはWebから新規取得
	tok, err := a.tokenFromFile(TokenFile)
	if err != nil {
		tok, err = a.getTokenFromWeb()
		if err != nil {
			return nil, fmt.Errorf("トークン取得に失敗しました: %v", err)
		}
		a.saveToken(TokenFile, tok)
	}
	return a.config.Client(context.Background(), tok), nil
}

// IsTokenValid は現在のトークンが有効かどうかを確認します
func (a *AuthService) IsTokenValid() bool {
	tok, err := a.tokenFromFile(TokenFile)
	if err != nil {
		return false
	}

	// トークンの有効期限をチェック
	if tok.Expiry.Before(time.Now()) {
		return false
	}

	return true
}

// RefreshToken はトークンを更新します
func (a *AuthService) RefreshToken() error {
	tok, err := a.tokenFromFile(TokenFile)
	if err != nil {
		return fmt.Errorf("既存のトークン読み込みに失敗しました: %v", err)
	}

	// トークンを更新
	newTok, err := a.config.TokenSource(context.Background(), tok).Token()
	if err != nil {
		return fmt.Errorf("トークンの更新に失敗しました: %v", err)
	}

	// 新しいトークンを保存
	a.saveToken(TokenFile, newTok)
	return nil
}

// getTokenFromWeb はWebブラウザを使用してOAuth2トークンを取得します
func (a *AuthService) getTokenFromWeb() (*oauth2.Token, error) {
	authMutex.Lock()
	defer authMutex.Unlock()

	authURL := a.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("ブラウザでこのURLにアクセスしてください: \n%v\n", authURL)
	fmt.Println("認証完了後、自動的に処理が続行されます...")

	// 認証コードの受信を待機
	var authCode string
	select {
	case authCode = <-authCodeChan:
		fmt.Println("認証コードを受信しました")
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("認証タイムアウト: 5分以内に認証を完了してください")
	}

	tok, err := a.config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("トークン取得に失敗しました: %v", err)
	}
	return tok, nil
}

// tokenFromFile はファイルからトークンを取得します
func (a *AuthService) tokenFromFile(file string) (*oauth2.Token, error) {
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
func (a *AuthService) saveToken(path string, token *oauth2.Token) {
	fmt.Printf("トークンをファイルに保存中: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("トークンファイルの保存に失敗しました: %v", err)
		return
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

// StartAuth は認証プロセスを開始するエンドポイントです
func StartAuth(c *gin.Context) {
	authService, err := NewAuthService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "認証サービスの初期化に失敗しました"})
		return
	}

	// 既存のトークンが有効かチェック
	if authService.IsTokenValid() {
		c.JSON(http.StatusOK, gin.H{"message": "既に認証済みです"})
		return
	}

	// 認証URLを生成してクライアントに返す
	authURL := authService.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.JSON(http.StatusOK, gin.H{
		"message":  "ブラウザで以下のURLにアクセスして認証を完了してください",
		"auth_url": authURL,
	})
}
