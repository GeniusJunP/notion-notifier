package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/models"
	"notion-notifier/internal/retry"
)

const notionVersion = "2022-06-28"

type Client struct {
	http       *http.Client
	apiKey     string
	retryCfg   retry.Config
	baseURL    string
}

type queryRequest struct {
	StartCursor string `json:"start_cursor,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
}

type queryResponse struct {
	Results    []page `json:"results"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}

type page struct {
	ID             string                 `json:"id"`
	URL            string                 `json:"url"`
	LastEditedTime string                 `json:"last_edited_time"`
	Properties     map[string]any         `json:"properties"`
}

func New(httpClient *http.Client, apiKey string, cfg retry.Config) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{
		http:     httpClient,
		apiKey:   apiKey,
		retryCfg: cfg,
		baseURL:  "https://api.notion.com/v1",
	}
}

func (c *Client) QueryDatabase(ctx context.Context, databaseID string) ([]page, error) {
	if c.apiKey == "" {
		return nil, errors.New("notion api key is empty")
	}
	if databaseID == "" {
		return nil, errors.New("notion database id is empty")
	}
	var out []page
	cursor := ""
	for {
		reqBody := queryRequest{StartCursor: cursor, PageSize: 100}
		data, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}
		url := fmt.Sprintf("%s/databases/%s/query", c.baseURL, databaseID)
		resp, err := c.doRequest(ctx, http.MethodPost, url, data)
		if err != nil {
			return nil, err
		}
		var qr queryResponse
		if err := json.Unmarshal(resp, &qr); err != nil {
			return nil, err
		}
		out = append(out, qr.Results...)
		if !qr.HasMore || qr.NextCursor == "" {
			break
		}
		cursor = qr.NextCursor
	}
	return out, nil
}

func (c *Client) doRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	maxRetries := c.retryCfg.WithDefaults().MaxRetries
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Notion-Version", notionVersion)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			defer resp.Body.Close()
			data, _ := io.ReadAll(resp.Body)
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return data, nil
			}
			if !retry.IsRetryableStatus(resp.StatusCode) {
				return nil, fmt.Errorf("notion api error: status %d body=%s", resp.StatusCode, string(data))
			}
			lastErr = fmt.Errorf("notion api error: status %d", resp.StatusCode)
			retryAfter, _ := retry.ParseRetryAfter(resp.Header.Get("Retry-After"), time.Now())
			delay := retry.BackoffDelay(c.retryCfg, attempt, retryAfter)
			if err := retry.Sleep(ctx, delay); err != nil {
				return nil, err
			}
			continue
		}
		delay := retry.BackoffDelay(c.retryCfg, attempt, 0)
		if err := retry.Sleep(ctx, delay); err != nil {
			return nil, err
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, retry.ErrRetriesExhausted
}

func MapPagesToEvents(pages []page, mapping config.PropertyMapping, tz *time.Location) []models.Event {
	var events []models.Event
	for _, p := range pages {
		title := ExtractString(p.Properties[mapping.Title])
		dateProp := p.Properties[mapping.Date]
		startDate, startTime, endDate, endTime, isAllDay := parseDateRange(dateProp, tz)
		location := ExtractString(p.Properties[mapping.Location])
		custom := map[string]string{}
		for _, cm := range mapping.Custom {
			custom[cm.Variable] = ExtractString(p.Properties[cm.Property])
		}
		rawProps, _ := json.Marshal(p.Properties)
		events = append(events, models.Event{
			NotionPageID: p.ID,
			Title:        title,
			StartDate:    startDate,
			StartTime:    startTime,
			EndDate:      endDate,
			EndTime:      endTime,
			IsAllDay:     isAllDay,
			Location:     location,
			URL:          p.URL,
			Custom:       custom,
			RawPropsJSON: string(rawProps),
			FetchedAt:    time.Now().In(tz),
			NotionUpdatedAt: p.LastEditedTime,
		})
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].StartDate == events[j].StartDate {
			return events[i].StartTime < events[j].StartTime
		}
		return events[i].StartDate < events[j].StartDate
	})
	return events
}

func parseDateRange(prop any, tz *time.Location) (startDate, startTime, endDate, endTime string, isAllDay bool) {
	props, ok := prop.(map[string]any)
	if !ok {
		return "", "", "", "", false
	}
	if props["type"] != "date" {
		return "", "", "", "", false
	}
	dateVal, ok := props["date"].(map[string]any)
	if !ok {
		return "", "", "", "", false
	}
	startRaw, _ := dateVal["start"].(string)
	endRaw, _ := dateVal["end"].(string)
	startDate, startTime, isAllDay = splitDateTime(startRaw, tz)
	endDate, endTime, _ = splitDateTime(endRaw, tz)
	if endDate == "" {
		endDate = startDate
	}
	return startDate, startTime, endDate, endTime, isAllDay
}

func splitDateTime(value string, tz *time.Location) (date string, tm string, allDay bool) {
	if value == "" {
		return "", "", false
	}
	if strings.Contains(value, "T") {
		t, err := time.Parse(time.RFC3339, value)
		if err == nil {
			if tz != nil {
				t = t.In(tz)
			}
			return t.Format("2006-01-02"), t.Format("15:04"), false
		}
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		if tz != nil {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, tz)
		}
		return t.Format("2006-01-02"), "", true
	}
	return "", "", false
}

func ExtractString(prop any) string {
	props, ok := prop.(map[string]any)
	if !ok {
		return ""
	}
	typ, _ := props["type"].(string)
	switch typ {
	case "title":
		return joinRichText(props["title"])
	case "rich_text":
		return joinRichText(props["rich_text"])
	case "select":
		if v, ok := props["select"].(map[string]any); ok {
			if name, ok := v["name"].(string); ok {
				return name
			}
		}
	case "multi_select":
		if arr, ok := props["multi_select"].([]any); ok {
			var names []string
			for _, item := range arr {
				if m, ok := item.(map[string]any); ok {
					if name, ok := m["name"].(string); ok {
						names = append(names, name)
					}
				}
			}
			return strings.Join(names, ", ")
		}
	case "people":
		if arr, ok := props["people"].([]any); ok {
			var names []string
			for _, item := range arr {
				if m, ok := item.(map[string]any); ok {
					if name, ok := m["name"].(string); ok {
						names = append(names, name)
					}
				}
			}
			return strings.Join(names, ", ")
		}
	case "number":
		if v, ok := props["number"].(float64); ok {
			return fmt.Sprintf("%v", v)
		}
	case "checkbox":
		if v, ok := props["checkbox"].(bool); ok {
			if v {
				return "true"
			}
			return "false"
		}
	case "url":
		if v, ok := props["url"].(string); ok {
			return v
		}
	case "email":
		if v, ok := props["email"].(string); ok {
			return v
		}
	case "phone_number":
		if v, ok := props["phone_number"].(string); ok {
			return v
		}
	case "status":
		if v, ok := props["status"].(map[string]any); ok {
			if name, ok := v["name"].(string); ok {
				return name
			}
		}
	case "formula":
		if v, ok := props["formula"].(map[string]any); ok {
			formType, _ := v["type"].(string)
			if formType != "" {
				return ExtractString(map[string]any{"type": formType, formType: v[formType]})
			}
		}
	case "date":
		if v, ok := props["date"].(map[string]any); ok {
			if start, ok := v["start"].(string); ok {
				return start
			}
		}
	case "relation":
		if arr, ok := props["relation"].([]any); ok {
			var ids []string
			for _, item := range arr {
				if m, ok := item.(map[string]any); ok {
					if id, ok := m["id"].(string); ok {
						ids = append(ids, id)
					}
				}
			}
			return strings.Join(ids, ", ")
		}
	case "files":
		if arr, ok := props["files"].([]any); ok {
			var names []string
			for _, item := range arr {
				if m, ok := item.(map[string]any); ok {
					if name, ok := m["name"].(string); ok {
						names = append(names, name)
					}
				}
			}
			return strings.Join(names, ", ")
		}
	}
	return ""
}

func joinRichText(value any) string {
	arr, ok := value.([]any)
	if !ok {
		return ""
	}
	var parts []string
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			if text, ok := m["plain_text"].(string); ok {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, "")
}

type blockListResponse struct {
	Results    []block `json:"results"`
	HasMore    bool    `json:"has_more"`
	NextCursor string  `json:"next_cursor"`
}

type block struct {
	ID               string        `json:"id"`
	Type             string        `json:"type"`
	HasChildren      bool          `json:"has_children"`
	Paragraph        *blockText    `json:"paragraph,omitempty"`
	Heading1         *blockText    `json:"heading_1,omitempty"`
	Heading2         *blockText    `json:"heading_2,omitempty"`
	Heading3         *blockText    `json:"heading_3,omitempty"`
	BulletedListItem *blockText    `json:"bulleted_list_item,omitempty"`
	NumberedListItem *blockText    `json:"numbered_list_item,omitempty"`
	ToDo             *blockText    `json:"to_do,omitempty"`
	Divider          *struct{}     `json:"divider,omitempty"`
}

type blockText struct {
	RichText []richText `json:"rich_text"`
	Checked  bool       `json:"checked"`
}

type richText struct {
	PlainText string `json:"plain_text"`
}

func (c *Client) FetchContent(ctx context.Context, pageID string, rules config.ContentRules) (string, error) {
	if c == nil || pageID == "" || rules.StartHeading == "" {
		return "", nil
	}
	blocks, err := c.listBlocks(ctx, pageID)
	if err != nil {
		return "", err
	}
	return extractContentFromBlocks(blocks, rules), nil
}

func (c *Client) listBlocks(ctx context.Context, blockID string) ([]block, error) {
	var out []block
	cursor := ""
	for {
		url := fmt.Sprintf("%s/blocks/%s/children?page_size=100", c.baseURL, blockID)
		if cursor != "" {
			url += "&start_cursor=" + cursor
		}
		resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		var br blockListResponse
		if err := json.Unmarshal(resp, &br); err != nil {
			return nil, err
		}
		out = append(out, br.Results...)
		if !br.HasMore || br.NextCursor == "" {
			break
		}
		cursor = br.NextCursor
	}
	return out, nil
}

func extractContentFromBlocks(blocks []block, rules config.ContentRules) string {
	if rules.StartHeading == "" {
		return ""
	}
	delimiterText := strings.TrimSpace(rules.DelimiterText)
	var lines []string
	started := false
	for _, b := range blocks {
		text, kind, level, isHeading, isDivider, checked := blockInfo(b)
		if !started {
			if isHeading && headingMatches(text, rules.StartHeading) {
				started = true
				if rules.IncludeStart && text != "" {
					lines = append(lines, formatHeading(level, text))
				}
			}
			continue
		}
		if rules.StopAtDelimiter && isDivider {
			break
		}
		if rules.StopAtDelimiter && delimiterText != "" && strings.TrimSpace(text) == delimiterText {
			break
		}
		if rules.StopAtNextHeading && isHeading {
			break
		}
		if text == "" {
			continue
		}
		lines = append(lines, formatBlock(kind, level, text, checked))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func blockInfo(b block) (text string, kind string, level int, isHeading bool, isDivider bool, checked bool) {
	switch b.Type {
	case "paragraph":
		return joinBlockText(b.Paragraph), "paragraph", 0, false, false, false
	case "heading_1":
		return joinBlockText(b.Heading1), "heading", 1, true, false, false
	case "heading_2":
		return joinBlockText(b.Heading2), "heading", 2, true, false, false
	case "heading_3":
		return joinBlockText(b.Heading3), "heading", 3, true, false, false
	case "bulleted_list_item":
		return joinBlockText(b.BulletedListItem), "bulleted", 0, false, false, false
	case "numbered_list_item":
		return joinBlockText(b.NumberedListItem), "numbered", 0, false, false, false
	case "to_do":
		return joinBlockText(b.ToDo), "todo", 0, false, false, b.ToDo != nil && b.ToDo.Checked
	case "divider":
		return "", "divider", 0, false, true, false
	default:
		return "", "", 0, false, false, false
	}
}

func joinBlockText(bt *blockText) string {
	if bt == nil || len(bt.RichText) == 0 {
		return ""
	}
	var parts []string
	for _, rt := range bt.RichText {
		if rt.PlainText != "" {
			parts = append(parts, rt.PlainText)
		}
	}
	return strings.Join(parts, "")
}

func headingMatches(text, start string) bool {
	return strings.EqualFold(strings.TrimSpace(text), strings.TrimSpace(start))
}

func formatHeading(level int, text string) string {
	prefix := "#"
	switch level {
	case 2:
		prefix = "##"
	case 3:
		prefix = "###"
	}
	return prefix + " " + text
}

func formatBlock(kind string, level int, text string, checked bool) string {
	switch kind {
	case "bulleted":
		return "- " + text
	case "numbered":
		return "1. " + text
	case "todo":
		if checked {
			return "- [x] " + text
		}
		return "- [ ] " + text
	default:
		if kind == "heading" {
			return formatHeading(level, text)
		}
		return text
	}
}
