package pressable

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const DefaultBaseURL = "https://my.pressable.com"

type Client struct {
	baseURL      *url.URL
	accessToken  string
	clientID     string
	clientSecret string
	httpClient   *http.Client
	mu           sync.Mutex
	expiresAt    time.Time
}

type Option func(*Client)

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}

func New(baseURL, accessToken, clientID, clientSecret string, options ...Option) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultBaseURL
	}
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse base_url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("base_url must be absolute: %q", baseURL)
	}
	client := &Client{
		baseURL:      parsed,
		accessToken:  strings.TrimSpace(accessToken),
		clientID:     strings.TrimSpace(clientID),
		clientSecret: strings.TrimSpace(clientSecret),
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}
	for _, option := range options {
		option(client)
	}
	return client, nil
}

type Envelope struct {
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Errors  json.RawMessage `json:"errors"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type APIError struct {
	Status  int
	Method  string
	Path    string
	Message string
	Errors  string
	Body    string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("pressable %s %s failed (%d): %s", e.Method, e.Path, e.Status, e.Message)
	}
	return fmt.Sprintf("pressable %s %s failed (%d): %s", e.Method, e.Path, e.Status, e.Body)
}

func (c *Client) Request(ctx context.Context, method, path string, body any) (*Envelope, int, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, 0, err
	}

	payload, contentType, err := encodeBody(body)
	if err != nil {
		return nil, 0, err
	}

	requestURL := c.resolve(path)
	req, err := http.NewRequestWithContext(ctx, method, requestURL, payload)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, err
	}

	envelope := &Envelope{}
	if len(bytes.TrimSpace(resBody)) > 0 {
		if err := json.Unmarshal(resBody, envelope); err != nil {
			envelope.Data = json.RawMessage(resBody)
		}
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return envelope, res.StatusCode, &APIError{
			Status:  res.StatusCode,
			Method:  method,
			Path:    path,
			Message: envelope.Message,
			Errors:  string(envelope.Errors),
			Body:    string(resBody),
		}
	}

	return envelope, res.StatusCode, nil
}

func (c *Client) GetData(ctx context.Context, path string, target any) error {
	envelope, _, err := c.Request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return DecodeData(envelope, target)
}

func DecodeData(envelope *Envelope, target any) error {
	if envelope == nil {
		return errors.New("missing response envelope")
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return errors.New("response data is empty")
	}
	return json.Unmarshal(envelope.Data, target)
}

func RawString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "null"
	}
	return string(raw)
}

func (c *Client) ensureToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		return nil
	}
	if c.accessToken != "" && c.clientID == "" {
		return nil
	}
	if c.clientID == "" || c.clientSecret == "" {
		return nil
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.resolve("/auth/token"),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return &APIError{Status: res.StatusCode, Method: http.MethodPost, Path: "/auth/token", Body: string(body)}
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return err
	}
	if token.AccessToken == "" {
		return errors.New("Pressable token response did not include access_token")
	}
	expiresIn := token.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3599
	}
	c.accessToken = token.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(expiresIn-60) * time.Second)
	return nil
}

func (c *Client) resolve(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	resolved := *c.baseURL
	if strings.HasPrefix(path, "/") {
		resolved.Path = path
	} else {
		resolved.Path = "/" + path
	}
	return resolved.String()
}

func encodeBody(body any) (io.Reader, string, error) {
	if body == nil {
		return nil, "", nil
	}
	switch value := body.(type) {
	case string:
		if strings.TrimSpace(value) == "" {
			return nil, "", nil
		}
		return strings.NewReader(value), "application/json", nil
	case []byte:
		if len(value) == 0 {
			return nil, "", nil
		}
		return bytes.NewReader(value), "application/json", nil
	default:
		payload, err := json.Marshal(value)
		if err != nil {
			return nil, "", err
		}
		return bytes.NewReader(payload), "application/json", nil
	}
}
