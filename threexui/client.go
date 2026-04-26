package threexui

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type threeXUIClient struct {
	config     Config
	httpClient *http.Client
	cookieMu   sync.Mutex
	cookie     *http.Cookie
	cookieExp  time.Time
	logger     *log.Logger
}

func newClient(cfg Config, logger *log.Logger) *threeXUIClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 15 * time.Second
	}
	return &threeXUIClient{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: cfg.InsecureSkipVerify,
				},
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
			Timeout: cfg.Timeout,
		},
	}
}

func (c *threeXUIClient) getAuthToken() (*http.Cookie, error) {
	c.cookieMu.Lock()
	defer c.cookieMu.Unlock()

	if c.cookie != nil && time.Now().Before(c.cookieExp) {
		return c.cookie, nil
	}

	data := url.Values{
		"username": {c.config.Username},
		"password": {c.config.Password},
	}
	req, err := http.NewRequest(http.MethodPost,
		c.config.BaseURL+"/login",
		strings.NewReader(data.Encode()))
	if err != nil {
		c.logger.Printf("3x-ui: create login request failed: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Printf("3x-ui: login request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Printf("3x-ui: read login response failed: %v", err)
		return nil, err
	}

	var loginResp struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
	}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		c.logger.Printf("3x-ui: unmarshal login response failed: %v, body: %s", err, string(body))
		return nil, err
	}
	if !loginResp.Success {
		c.logger.Printf("3x-ui: authentication failed: %s", loginResp.Msg)
		return nil, fmt.Errorf("authentication failed: %s", loginResp.Msg)
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "3x-ui" {
			c.cookie = cookie
			c.cookieExp = time.Now().Add(59 * time.Minute)
			return cookie, nil
		}
	}
	c.logger.Println("3x-ui: no 3x-ui session cookie found")
	return nil, fmt.Errorf("no 3x-ui session cookie found")
}

func (c *threeXUIClient) do(method, path string, cookie *http.Cookie) ([]byte, error) {
	req, err := http.NewRequest(method, c.config.BaseURL+path, nil)
	if err != nil {
		c.logger.Printf("3x-ui: create request %s %s failed: %v", method, path, err)
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.AddCookie(cookie)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Printf("3x-ui: %s %s request failed: %v", method, path, err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Printf("3x-ui: read response body %s %s failed: %v", method, path, err)
		return nil, err
	}
	return body, nil
}
