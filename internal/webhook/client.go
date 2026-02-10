package webhook

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"notion-notifier/internal/retry"
)

type Client struct {
	http  *http.Client
	retry retry.Config
}

func New(httpClient *http.Client, cfg retry.Config) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{http: httpClient, retry: cfg}
}

func (c *Client) Send(ctx context.Context, webhookURL, contentType string, payload []byte) error {
	if webhookURL == "" {
		return errors.New("webhook url is empty")
	}
	if contentType == "" {
		contentType = "application/json"
	}
	var lastErr error
	maxRetries := c.retry.WithDefaults().MaxRetries
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", contentType)
		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted {
				return nil
			}
			errMsg := string(bodyBytes)
			if !retry.IsRetryableStatus(resp.StatusCode) {
				return fmt.Errorf("webhook failed: status %d, body: %s", resp.StatusCode, errMsg)
			}
			lastErr = fmt.Errorf("webhook failed: status %d, body: %s", resp.StatusCode, errMsg)
			retryAfter, _ := retry.ParseRetryAfter(resp.Header.Get("Retry-After"), time.Now())
			delay := retry.BackoffDelay(c.retry, attempt, retryAfter)
			if err := retry.Sleep(ctx, delay); err != nil {
				return err
			}
			continue
		}
		delay := retry.BackoffDelay(c.retry, attempt, 0)
		if err := retry.Sleep(ctx, delay); err != nil {
			return err
		}
	}
	if lastErr != nil {
		return lastErr
	}
	return retry.ErrRetriesExhausted
}
