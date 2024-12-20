package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
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
	logger      zerolog.Logger
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

func NewClient(serverToken string, logger zerolog.Logger, opts ...ClientOption) *Client {
	client := &Client{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		serverToken: serverToken,
		baseURL:     defaultBaseURL,
		logger:      logger,
	}

	client.logger.Debug().
		Str("base_url", client.baseURL).
		Int64("timeout_seconds", int64(client.httpClient.Timeout.Seconds())).
		Msg("Initializing Postmark client")

	for _, opt := range opts {
		opt(client)
	}
	return client
}

type ClientOption func(*Client)

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
		c.logger.Debug().
			Int64("timeout_seconds", int64(httpClient.Timeout.Seconds())).
			Msg("Setting custom HTTP client")
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
		c.logger.Debug().
			Str("base_url", baseURL).
			Msg("Setting custom base URL")
	}
}

func (c *Client) doRequest(params requestParams, dst interface{}) error {
	requestLog := c.logger.With().
		Str("method", params.method).
		Str("path", params.path).
		Str("token_type", params.tokenType).
		Logger()

	start := time.Now()

	var body io.Reader
	if params.payload != nil {
		payloadData, err := json.Marshal(params.payload)
		if err != nil {
			requestLog.Error().Err(err).Msg("Failed to marshal request payload")
			return fmt.Errorf("marshaling request payload: %w", err)
		}
		body = bytes.NewBuffer(payloadData)
		requestLog.Debug().
			RawJSON("payload", payloadData).
			Msg("Request payload prepared")
	}

	req, err := http.NewRequest(params.method, fmt.Sprintf("%s/%s", c.baseURL, params.path), body)
	if err != nil {
		requestLog.Error().Err(err).Msg("Failed to create request")
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", c.serverToken)

	requestLog.Debug().Msg("Sending request to Postmark API")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		requestLog.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("Request execution failed")
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		requestLog.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("Failed to read response body")
		return fmt.Errorf("reading response body: %w", err)
	}

	requestLog.Debug().
		Int("status_code", resp.StatusCode).
		Dur("duration", time.Since(start)).
		RawJSON("response", respBody).
		Msg("Received response")

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			requestLog.Error().
				Err(err).
				Int("status_code", resp.StatusCode).
				Str("response_body", string(respBody)).
				Msg("Failed to parse error response")
			return fmt.Errorf("unexpected error response: status=%d body=%s",
				resp.StatusCode, string(respBody))
		}
		requestLog.Error().
			Int("error_code", errResp.ErrorCode).
			Str("error_message", errResp.Message).
			Msg("Received error response from Postmark")
		return &errResp
	}

	if dst != nil {
		if err := json.Unmarshal(respBody, dst); err != nil {
			requestLog.Error().
				Err(err).
				Msg("Failed to unmarshal response")
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	requestLog.Info().
		Int("status_code", resp.StatusCode).
		Dur("duration", time.Since(start)).
		Msg("Request completed successfully")

	return nil
}

