package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL   = "https://api.postmarkapp.com"
	TokenTypeServer  = "server"
	TokenTypeAccount = "account"
)

type Client struct {
	httpClient  *http.Client
	serverToken string
	baseURL     string
}

type requestParams struct {
	method    string
	path      string
	payload   interface{}
	tokenType string
}

type ErrorResponse struct {
	ErrorCode int    `json:"ErrorCode"`
	Message   string `json:"Message"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("postmark: %s (code: %d)", e.Message, e.ErrorCode)
}

func NewClient(serverToken string, opts ...ClientOption) *Client {
	client := &Client{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		serverToken: serverToken,
		baseURL:     defaultBaseURL,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

type ClientOption func(*Client)

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func (c *Client) doRequest(params requestParams, dst interface{}) error {
	var body io.Reader
	if params.payload != nil {
		payloadData, err := json.Marshal(params.payload)
		if err != nil {
			return fmt.Errorf("marshaling request payload: %w", err)
		}
		body = bytes.NewBuffer(payloadData)
	}

	req, err := http.NewRequest(params.method, fmt.Sprintf("%s/%s", c.baseURL, params.path), body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", c.serverToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return fmt.Errorf("unexpected error response: status=%d body=%s",
				resp.StatusCode, string(respBody))
		}
		return &errResp
	}

	if dst != nil {
		if err := json.Unmarshal(respBody, dst); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	return nil
}