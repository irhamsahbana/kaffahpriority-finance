package adapter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// DropboxTokenManager manages OAuth2 token refresh for Dropbox
type DropboxTokenManager struct {
	mu           sync.RWMutex
	accessToken  string
	refreshToken string
	appKey       string
	appSecret    string
	expiresAt    time.Time
	client       *http.Client
}

// TokenResponse represents Dropbox OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// NewDropboxTokenManager creates a new token manager
func NewDropboxTokenManager(accessToken, refreshToken, appKey, appSecret string) *DropboxTokenManager {
	return &DropboxTokenManager{
		accessToken:  accessToken,
		refreshToken: refreshToken,
		appKey:       appKey,
		appSecret:    appSecret,
		expiresAt:    time.Now().Add(4 * time.Hour), // Default 4 hours
		client:       &http.Client{Timeout: 30 * time.Second},
	}
}

// GetValidToken returns a valid access token, refreshing if necessary
func (tm *DropboxTokenManager) GetValidToken() (string, error) {
	fnName := "adapter::GetValidToken"

	tm.mu.RLock()
	// Check if current token is still valid (with 5 minute buffer)
	if time.Now().Before(tm.expiresAt.Add(-5 * time.Minute)) {
		token := tm.accessToken
		tm.mu.RUnlock()
		return token, nil
	}
	tm.mu.RUnlock()

	// Token is expired or about to expire, refresh it
	log.Info().Msgf("%s - access token expired or about to expire, refreshing", fnName)
	return tm.refreshAccessToken()
}

// refreshAccessToken refreshes the access token using the refresh token
func (tm *DropboxTokenManager) refreshAccessToken() (string, error) {
	fnName := "adapter::refreshAccessToken"

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check if another goroutine already refreshed the token
	if time.Now().Before(tm.expiresAt.Add(-5 * time.Minute)) {
		return tm.accessToken, nil
	}

	if tm.refreshToken == "" {
		log.Error().Msgf("%s - no refresh token available", fnName)
		return "", fmt.Errorf("no refresh token available")
	}

	log.Info().Msgf("%s - refreshing Dropbox access token", fnName)

	// Prepare the refresh request
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {tm.refreshToken},
	}

	req, err := http.NewRequest("POST", "https://api.dropboxapi.com/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		log.Error().Err(err).Msgf("%s - failed to create refresh request", fnName)
		return "", fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(tm.appKey, tm.appSecret)

	// Make the refresh request
	resp, err := tm.client.Do(req)
	if err != nil {
		log.Error().Err(err).Msgf("%s - failed to send refresh request", fnName)
		return "", fmt.Errorf("failed to send refresh request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msgf("%s - failed to read refresh response", fnName)
		return "", fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Error().Int("statusCode", resp.StatusCode).Str("response", string(body)).Msgf("%s - token refresh failed", fnName)
		return "", fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		log.Error().Err(err).Msgf("%s - failed to parse token response", fnName)
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	// Update the tokens
	tm.accessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		tm.refreshToken = tokenResp.RefreshToken
	}

	// Set expiration time (default to 4 hours if not provided)
	expiresIn := tokenResp.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 14400 // 4 hours in seconds
	}
	tm.expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)

	log.Info().Time("expiresAt", tm.expiresAt).Msgf("%s - access token refreshed successfully", fnName)
	return tm.accessToken, nil
}

// GetAuthURL generates the OAuth2 authorization URL for initial setup
func GetDropboxAuthURL(appKey, redirectURL string) string {
	baseURL := "https://www.dropbox.com/oauth2/authorize"
	params := url.Values{
		"client_id":         {appKey},
		"response_type":     {"code"},
		"redirect_uri":      {redirectURL},
		"token_access_type": {"offline"}, // To get refresh token
	}
	return baseURL + "?" + params.Encode()
}

// ExchangeCodeForToken exchanges authorization code for access and refresh tokens
func ExchangeCodeForToken(code, appKey, appSecret, redirectURL string) (*TokenResponse, error) {
	fnName := "adapter::ExchangeCodeForToken"

	data := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURL},
	}

	req, err := http.NewRequest("POST", "https://api.dropboxapi.com/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		log.Error().Err(err).Msgf("%s - failed to create token exchange request", fnName)
		return nil, fmt.Errorf("failed to create token exchange request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(appKey, appSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msgf("%s - failed to send token exchange request", fnName)
		return nil, fmt.Errorf("failed to send token exchange request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msgf("%s - failed to read token exchange response", fnName)
		return nil, fmt.Errorf("failed to read token exchange response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Error().Int("statusCode", resp.StatusCode).Str("response", string(body)).Msgf("%s - token exchange failed", fnName)
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		log.Error().Err(err).Msgf("%s - failed to parse token exchange response", fnName)
		return nil, fmt.Errorf("failed to parse token exchange response: %w", err)
	}

	log.Info().Msgf("%s - token exchange completed successfully", fnName)
	return &tokenResp, nil
}
