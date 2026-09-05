package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// GoogleOAuthClient はGoogle OAuth 2.0のAuthorization Code Flowを扱う。
// 依存を増やさないよう golang.org/x/oauth2 は使わず、標準ライブラリのみで実装している。
type GoogleOAuthClient struct {
	clientID     string
	clientSecret string
	redirectURL  string
	httpClient   *http.Client
}

func NewGoogleOAuthClient(clientID, clientSecret, redirectURL string) *GoogleOAuthClient {
	return &GoogleOAuthClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		httpClient:   http.DefaultClient,
	}
}

const (
	googleAuthURL     = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL    = "https://oauth2.googleapis.com/token"
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

// AuthURL はユーザーをGoogleの同意画面へ送るためのURLを組み立てる。
// state はCSRF対策のランダム文字列で、ハンドラー側でCookieに保存した値と突き合わせる。
func (c *GoogleOAuthClient) AuthURL(state string) string {
	q := url.Values{
		"client_id":     {c.clientID},
		"redirect_uri":  {c.redirectURL},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"access_type":   {"online"},
		"prompt":        {"select_account"},
	}
	return googleAuthURL + "?" + q.Encode()
}

type GoogleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ExchangeCode は認可コードをアクセストークンに交換し、続けてユーザー情報を取得する。
func (c *GoogleOAuthClient) ExchangeCode(ctx context.Context, code string) (*GoogleUserInfo, error) {
	token, err := c.exchangeToken(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange token: %w", err)
	}
	return c.fetchUserInfo(ctx, token)
}

func (c *GoogleOAuthClient) exchangeToken(ctx context.Context, code string) (string, error) {
	form := url.Values{
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {c.redirectURL},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("google token endpoint returned %d: %s", resp.StatusCode, body)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}
	return tokenResp.AccessToken, nil
}

func (c *GoogleOAuthClient) fetchUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google userinfo endpoint returned %d: %s", resp.StatusCode, body)
	}

	var info GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}
