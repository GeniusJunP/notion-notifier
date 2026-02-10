package app

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"notion-notifier/internal/config"
)

func (s *Server) render(w http.ResponseWriter, name string, data map[string]interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, env := s.cfg.Get()
		if !cfg.Security.BasicAuth.Enabled {
			next(w, r)
			return
		}

		user, pass, ok := r.BasicAuth()
		if !ok || user != env.Security.BasicAuth.Username || pass != env.Security.BasicAuth.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func isTruthy(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1"
	case float64:
		return v != 0
	default:
		return false
	}
}

func normalizeBoolField(m map[string]interface{}, key string) {
	raw, ok := m[key]
	if !ok {
		return
	}
	switch v := raw.(type) {
	case bool:
		return
	case string:
		m[key] = v == "true" || v == "1"
	case float64:
		m[key] = v != 0
	}
}

func normalizeBoolSlice(raw interface{}, key string) {
	items, ok := raw.([]interface{})
	if !ok {
		return
	}
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		normalizeBoolField(m, key)
	}
}

type pathToken struct {
	key   string
	index *int
}

func expandFlatUpdates(updates map[string]interface{}) map[string]interface{} {
	expanded := map[string]interface{}{}
	pathEntries := map[string]interface{}{}
	for k, v := range updates {
		if isPathKey(k) {
			pathEntries[k] = v
			continue
		}
		expanded[k] = v
	}
	for k, v := range pathEntries {
		tokens := parsePath(k)
		if len(tokens) == 0 {
			expanded[k] = v
			continue
		}
		if root, ok := setPathValue(expanded, tokens, v).(map[string]interface{}); ok {
			expanded = root
		}
	}
	return expanded
}

func isPathKey(key string) bool {
	return strings.Contains(key, ".") || strings.Contains(key, "[")
}

func parsePath(path string) []pathToken {
	parts := strings.Split(path, ".")
	tokens := make([]pathToken, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		for part != "" {
			open := strings.Index(part, "[")
			if open == -1 {
				tokens = append(tokens, pathToken{key: part})
				part = ""
				break
			}
			if open > 0 {
				tokens = append(tokens, pathToken{key: part[:open]})
			}
			rest := part[open+1:]
			close := strings.Index(rest, "]")
			if close == -1 {
				tokens = append(tokens, pathToken{key: part})
				part = ""
				break
			}
			if idx, ok := parseIndex(rest[:close]); ok {
				tokens = append(tokens, pathToken{index: &idx})
			}
			part = rest[close+1:]
		}
	}
	return tokens
}

func parseIndex(value string) (int, bool) {
	if value == "" {
		return 0, false
	}
	num, err := strconv.Atoi(value)
	if err != nil || num < 0 {
		return 0, false
	}
	return num, true
}

func setPathValue(container interface{}, tokens []pathToken, value interface{}) interface{} {
	if len(tokens) == 0 {
		return container
	}
	token := tokens[0]
	if token.key != "" {
		current, _ := container.(map[string]interface{})
		if current == nil {
			current = map[string]interface{}{}
		}
		if len(tokens) == 1 {
			current[token.key] = value
			return current
		}
		current[token.key] = setPathValue(current[token.key], tokens[1:], value)
		return current
	}
	if token.index != nil {
		current, _ := container.([]interface{})
		if current == nil {
			current = []interface{}{}
		}
		index := *token.index
		for len(current) <= index {
			current = append(current, nil)
		}
		if len(tokens) == 1 {
			current[index] = value
			return current
		}
		current[index] = setPathValue(current[index], tokens[1:], value)
		return current
	}
	return container
}

func applyAdvanceConditionClears(cfg *config.Config, notifications map[string]interface{}) {
	raw, ok := notifications["advance"].([]interface{})
	if !ok {
		return
	}
	for i, item := range raw {
		if i >= len(cfg.Notifications.Advance) {
			break
		}
		rule, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		cond, ok := rule["conditions"].(map[string]interface{})
		if !ok {
			continue
		}
		if isTruthy(cond["days_of_week_clear"]) {
			cfg.Notifications.Advance[i].Conditions.DaysOfWeek = nil
		}
		if isTruthy(cond["property_filters_clear"]) {
			cfg.Notifications.Advance[i].Conditions.PropertyFilters = nil
		}
	}
}

func applyPeriodicDayClears(cfg *config.Config, notifications map[string]interface{}) {
	raw, ok := notifications["periodic"].([]interface{})
	if !ok {
		return
	}
	for i, item := range raw {
		if i >= len(cfg.Notifications.Periodic) {
			break
		}
		rule, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if isTruthy(rule["days_of_week_clear"]) {
			cfg.Notifications.Periodic[i].DaysOfWeek = nil
		}
	}
}

func decodeNotificationRequest(r *http.Request, req *notificationRequest) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return fmt.Errorf("empty body")
	}
	if err := json.Unmarshal(body, req); err == nil {
		return nil
	}
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return err
	}
	req.Template = values.Get("template")
	req.FromDate = values.Get("from_date")
	req.ToDate = values.Get("to_date")
	if v := values.Get("minutes_before"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			req.MinutesBefore = parsed
		}
	}
	req.PreviewPayload = values.Get("preview_payload") == "true"
	if req.Template == "" && req.FromDate == "" && req.ToDate == "" && req.MinutesBefore == 0 {
		return fmt.Errorf("invalid payload")
	}
	return nil
}

func parseDateRange(fromStr, toStr string, cfg *config.Manager) (time.Time, time.Time, error) {
	current, _ := cfg.Get()
	loc, _ := time.LoadLocation(current.Timezone)
	now := time.Now().In(loc)
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	to := from

	fromStr = strings.TrimSpace(fromStr)
	toStr = strings.TrimSpace(toStr)

	if fromStr != "" {
		parsed, err := parseDateInput(fromStr, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		from = parsed
	}

	if toStr != "" {
		parsed, err := parseDateInput(toStr, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		to = parsed
	} else if fromStr != "" {
		to = from
	}

	if to.Before(from) {
		return time.Time{}, time.Time{}, fmt.Errorf("to_date must be after from_date")
	}

	return from, to, nil
}

func parseDateInput(value string, loc *time.Location) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.In(loc), nil
	}
	layouts := []string{
		"2006-01-02",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, loc); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format")
}

func normalizeDateInput(value string, loc *time.Location) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if len(value) == len("2006-01-02") && strings.Count(value, "-") == 2 {
		parsed, err := time.ParseInLocation("2006-01-02", value, loc)
		if err != nil {
			return "", err
		}
		return parsed.Add(24 * time.Hour).Format(time.RFC3339), nil
	}
	parsed, err := parseDateInput(value, loc)
	if err != nil {
		return "", err
	}
	return parsed.Format(time.RFC3339), nil
}

func writePreviewHTML(w http.ResponseWriter, message string, payload string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	escaped := template.HTMLEscapeString(message)
	parts := []string{
		fmt.Sprintf(`<div class="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest mb-2">Message</div><pre class="whitespace-pre-wrap font-mono text-xs leading-relaxed">%s</pre>`, escaped),
	}
	if payload != "" {
		parts = append(parts, fmt.Sprintf(`<div class="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest mt-6 mb-2">Payload</div><pre class="whitespace-pre-wrap font-mono text-xs leading-relaxed">%s</pre>`, template.HTMLEscapeString(payload)))
	}
	fmt.Fprintf(w, `<div class="mt-4 rounded-2xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/50 p-4 text-sm text-slate-700 dark:text-slate-200">%s</div>`, strings.Join(parts, ""))
}

func formatDateOnly(value string, loc *time.Location) string {
	if value == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return ""
	}
	if loc == nil {
		loc = time.Local
	}
	return parsed.In(loc).Format("2006-01-02")
}
