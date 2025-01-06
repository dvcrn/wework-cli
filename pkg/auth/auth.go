package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type Auth0Config struct {
	ClientID    string `json:"client_id"`
	Domain      string `json:"domain"`
	RedirectURI string `json:"redirect_uri"`
	Audience    string `json:"audience"`
}

type WeWorkAuth struct {
	username      string
	password      string
	client        *http.Client
	config        *Auth0Config
	codeVerifier  string
	codeChallenge string
}

type TokenResponse struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

type WeWorkLoginResponse struct {
	Token   string `json:"token"`
	IDToken string `json:"idToken"`
}

func NewWeWorkAuth(username, password string) (*WeWorkAuth, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %v", err)
	}

	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	auth := &WeWorkAuth{
		username: username,
		password: password,
		client:   client,
	}

	if err := auth.getAuth0Config(); err != nil {
		return nil, err
	}

	auth.codeVerifier = generateCodeVerifier()
	auth.codeChallenge = generateCodeChallenge(auth.codeVerifier)

	return auth, nil
}

func (w *WeWorkAuth) getAuth0Config() error {
	baseURL := "https://members.wework.com/workplaceone/api/auth0/config"
	params := url.Values{}
	params.Add("companyId", "00000000-0000-0000-0000-000000000000")
	params.Add("domain", "members.wework.com")

	resp, err := w.client.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return fmt.Errorf("failed to get auth0 config: %v", err)
	}
	defer resp.Body.Close()

	var config Auth0Config
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return fmt.Errorf("failed to decode auth0 config: %v", err)
	}

	w.config = &config
	return nil
}

func (w *WeWorkAuth) Authenticate() (*WeWorkLoginResponse, error) {
	// Step 1: Get initial state
	nonce := generateNonce()
	authParams := url.Values{}
	authParams.Add("redirect_uri", w.config.RedirectURI)
	authParams.Add("client_id", w.config.ClientID)
	authParams.Add("audience", w.config.Audience)
	authParams.Add("scope", "openid profile email offline_access")
	authParams.Add("response_type", "code")
	authParams.Add("response_mode", "query")
	authParams.Add("nonce", nonce)
	authParams.Add("code_challenge", w.codeChallenge)
	authParams.Add("code_challenge_method", "S256")
	authParams.Add("auth0Client", "eyJuYW1lIjoiQGF1dGgwL2F1dGgwLWFuZ3VsYXIiLCJ2ZXJzaW9uIjoiMS4xMS4xLmN1c3RvbSIsImVudiI6eyJhbmd1bGFyL2NvcmUiOiIxMy4xLjEifX0=")

	authURL := fmt.Sprintf("https://%s/authorize?%s", w.config.Domain, authParams.Encode())
	resp, err := w.client.Get(authURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth URL: %v", err)
	}
	defer resp.Body.Close()

	loginURL := resp.Header.Get("Location")
	if loginURL == "" {
		return nil, fmt.Errorf("no login URL in response")
	}

	parsedURL, err := url.Parse(loginURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse login URL: %v", err)
	}

	state := parsedURL.Query().Get("state")
	if state == "" {
		return nil, fmt.Errorf("no state in login URL")
	}

	// Handle IDP redirect
	idpResp, err := w.client.Get(fmt.Sprintf("https://%s%s", w.config.Domain, loginURL))
	if err != nil {
		return nil, fmt.Errorf("failed to handle IDP redirect: %v", err)
	}
	defer idpResp.Body.Close()

	if idpResp.StatusCode >= 300 && idpResp.StatusCode < 400 {
		loginURL = idpResp.Header.Get("Location")
	}

	var csrfToken string
	for _, cookie := range w.client.Jar.Cookies(parsedURL) {
		if cookie.Name == "_csrf" {
			csrfToken = cookie.Value
			break
		}
	}

	// Step 2: Perform login
	loginData := map[string]interface{}{
		"client_id":     w.config.ClientID,
		"redirect_uri":  w.config.RedirectURI,
		"tenant":        "wework-prod",
		"response_type": "code",
		"scope":         "openid profile email offline_access",
		"audience":      w.config.Audience,
		"state":         state,
		"nonce":         nonce,
		"connection":    "id-wework",
		"username":      w.username,
		"password":      w.password,
		"_csrf":         csrfToken,
		"_intstate":     "deprecated",
		"protocol":      "oauth2",
		"popup_options": map[string]interface{}{},
	}

	loginBody, err := json.Marshal(loginData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login data: %v", err)
	}

	loginReq, err := http.NewRequest("POST", fmt.Sprintf("https://%s/usernamepassword/login", w.config.Domain), bytes.NewBuffer(loginBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %v", err)
	}

	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := w.client.Do(loginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to perform login: %v", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(loginResp.Body)
		return nil, fmt.Errorf("login failed with status %d: %s", loginResp.StatusCode, string(body))
	}

	// Step 3: Extract form data
	doc, err := goquery.NewDocumentFromReader(loginResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse login response: %v", err)
	}

	form := doc.Find("form")
	if form.Length() == 0 {
		return nil, fmt.Errorf("no form found in response")
	}

	formAction, exists := form.Attr("action")
	if !exists {
		return nil, fmt.Errorf("no form action found")
	}

	formData := url.Values{}
	form.Find("input").Each(func(_ int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		value, _ := s.Attr("value")
		if name != "" && value != "" {
			formData.Add(name, value)
		}
	})

	// Step 4: Submit form and handle redirects
	formResp, err := w.client.PostForm(formAction, formData)
	if err != nil {
		return nil, fmt.Errorf("failed to submit form: %v", err)
	}
	defer formResp.Body.Close()

	var code string
	currentResp := formResp
	for currentResp.StatusCode >= 300 && currentResp.StatusCode < 400 {
		redirectURL := currentResp.Header.Get("Location")
		if redirectURL == "" {
			return nil, fmt.Errorf("no redirect URL found")
		}

		if !strings.HasPrefix(redirectURL, "http") {
			redirectURL = fmt.Sprintf("https://%s%s", w.config.Domain, redirectURL)
		}

		if strings.Contains(redirectURL, "code=") {
			parsedRedirect, err := url.Parse(redirectURL)
			if err != nil {
				return nil, fmt.Errorf("failed to parse redirect URL: %v", err)
			}
			code = parsedRedirect.Query().Get("code")
			break
		}

		currentResp, err = w.client.Get(redirectURL)
		if err != nil {
			return nil, fmt.Errorf("failed to follow redirect: %v", err)
		}
		defer currentResp.Body.Close()
	}

	if code == "" {
		return nil, fmt.Errorf("no code found in redirects")
	}

	// Step 5: Exchange code for tokens
	tokenData := map[string]string{
		"client_id":     w.config.ClientID,
		"code_verifier": w.codeVerifier,
		"grant_type":    "authorization_code",
		"code":          code,
		"redirect_uri":  w.config.RedirectURI,
	}

	tokenBody, err := json.Marshal(tokenData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token data: %v", err)
	}

	tokenReq, err := http.NewRequest("POST", fmt.Sprintf("https://%s/oauth/token", w.config.Domain), bytes.NewBuffer(tokenBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %v", err)
	}

	tokenReq.Header.Set("Content-Type", "application/json")
	tokenResp, err := w.client.Do(tokenReq)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for tokens: %v", err)
	}
	defer tokenResp.Body.Close()

	var tokens TokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %v", err)
	}

	// Step 6: Login to WeWork backend
	return w.loginToWeWork(&tokens)
}

func (w *WeWorkAuth) loginToWeWork(tokens *TokenResponse) (*WeWorkLoginResponse, error) {
	loginURL := "https://members.wework.com/workplaceone/api/auth0/login-by-auth0-token"
	loginData := map[string]interface{}{
		"id_token":      tokens.IDToken,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
		"scope":         tokens.Scope,
		"token_type":    tokens.TokenType,
		"client_id":     w.config.ClientID,
		"audience":      w.config.Audience,
	}

	loginBody, err := json.Marshal(loginData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal WeWork login data: %v", err)
	}

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(loginBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create WeWork login request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Request-Source", "com.wework.ondemand/WorkplaceOne/Prod/iOS/2.68.0(18.2)")
	req.Header.Set("User-Agent", "Mobile Safari 16.1")

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to login to WeWork: %v", err)
	}
	defer resp.Body.Close()

	var loginResp WeWorkLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode WeWork login response: %v", err)
	}

	return &loginResp, nil
}

func generateCodeVerifier() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func generateNonce() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
