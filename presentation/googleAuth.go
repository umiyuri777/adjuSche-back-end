package presentation

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// GoogleAuthService Google認証サービス
type GoogleAuthService struct {
	config *oauth2.Config
}

// TokenResponse アクセストークン取得のレスポンス構造体
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// NewGoogleAuthService Google認証サービスの初期化
func NewGoogleAuthService() (*GoogleAuthService, error) {
	credData, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		return nil, fmt.Errorf("クライアントシークレットファイルの読み込みに失敗しました: %v", err)
	}

	config, err := google.ConfigFromJSON(credData, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("OAuth2設定の作成に失敗しました: %v", err)
	}

	// リダイレクトURIを設定（テスト用）
	config.RedirectURL = "http://localhost:8080/oauth/callback"

	return &GoogleAuthService{config: config}, nil
}

// generateState CSRF攻撃防止用のstate文字列を生成
func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetGoogleAuthURL Google認証URLを取得するエンドポイント
func GetGoogleAuthURL(c *gin.Context) {
	authService, err := NewGoogleAuthService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Google認証サービスの初期化に失敗しました",
		})
		return
	}

	state, err := generateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "セキュリティトークンの生成に失敗しました",
		})
		return
	}

	// セッションにstateを保存（簡易的にCookieを使用）
	c.SetCookie("oauth_state", state, 3600, "/", "", false, true)

	authURL := authService.config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"auth_url": authURL,
		"message":  "以下のURLにアクセスしてGoogle認証を行ってください",
	})
}

// HandleGoogleCallback Google認証のコールバック処理
func HandleGoogleCallback(c *gin.Context) {
	authService, err := NewGoogleAuthService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Google認証サービスの初期化に失敗しました",
		})
		return
	}

	// stateパラメータの検証
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "不正なリクエストです（state不一致）",
		})
		return
	}

	// 認証コードの取得
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "認証コードが見つかりません",
		})
		return
	}

	// アクセストークンの取得
	token, err := authService.config.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "アクセストークンの取得に失敗しました",
		})
		return
	}

	// トークンをJSONとしてシリアライズ
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "トークンのシリアライズに失敗しました",
		})
		return
	}

	// Cookieをクリア
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"message":       "Google認証が完了しました",
		"access_token":  token.AccessToken,
		"token_type":    token.TokenType,
		"refresh_token": token.RefreshToken,
		"expires_in":    token.Expiry.Unix(),
		"full_token":    string(tokenJSON),
		"usage_note":    "full_tokenの値をAuthorizationヘッダー（Bearer <full_token>）またはX-Tokenヘッダーとして使用してください",
	})
}

// GetCurrentToken 現在保存されているトークンを取得（テスト用）
func GetCurrentToken(c *gin.Context) {
	tokenData, err := ioutil.ReadFile("env/token.json")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "保存されたトークンが見つかりません",
		})
		return
	}

	var token oauth2.Token
	err = json.Unmarshal(tokenData, &token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "トークンの解析に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"access_token":  token.AccessToken,
		"token_type":    token.TokenType,
		"refresh_token": token.RefreshToken,
		"expires_in":    token.Expiry.Unix(),
		"full_token":    string(tokenData),
		"usage_note":    "full_tokenの値をAuthorizationヘッダー（Bearer <full_token>）またはX-Tokenヘッダーとして使用してください",
	})
}
