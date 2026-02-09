package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"notion-notifier/internal/retry"
)

type Client struct {
	http    *http.Client
	retry   retry.Config
}

type payload struct {
	Content string `json:"content"`
}

func New(httpClient *http.Client, cfg retry.Config) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{http: httpClient, retry: cfg}
}

func (c *Client) Send(ctx context.Context, webhookURL, message string) error {
	if webhookURL == "" {
		return errors.New("discord webhook url is empty")
	}
	body, err := json.Marshal(payload{Content: message})
	if err != nil {
		return err
	}
	var lastErr error
	maxRetries := c.retry.WithDefaults().MaxRetries
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
				return nil
			}
			if !retry.IsRetryableStatus(resp.StatusCode) {
				return fmt.Errorf("discord webhook failed: status %d", resp.StatusCode)
			}
			lastErr = fmt.Errorf("discord webhook failed: status %d", resp.StatusCode)
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
