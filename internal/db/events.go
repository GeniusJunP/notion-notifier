package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"notion-notifier/internal/models"
)

func (r *Repository) UpsertEvents(ctx context.Context, events []models.Event) error {
	if len(events) == 0 {
		return nil
	}
	query := `INSERT INTO events (
		notion_page_id, title, start_date, start_time, end_date, end_time, is_all_day, location, url, content, attendees_json, raw_properties, notion_updated_at, fetched_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(notion_page_id) DO UPDATE SET
		title=excluded.title,
		start_date=excluded.start_date,
		start_time=excluded.start_time,
		end_date=excluded.end_date,
		end_time=excluded.end_time,
		is_all_day=excluded.is_all_day,
		location=excluded.location,
		url=excluded.url,
		content=excluded.content,
		attendees_json=excluded.attendees_json,
		raw_properties=excluded.raw_properties,
		notion_updated_at=excluded.notion_updated_at,
		fetched_at=excluded.fetched_at;`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for upsert events: %w", err)
	}
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to prepare upsert events statement: %w", err)
	}
	defer stmt.Close()
	for _, ev := range events {
		_, err := stmt.ExecContext(ctx,
			ev.NotionPageID,
			ev.Title,
			ev.StartDate,
			ev.StartTime,
			ev.EndDate,
			ev.EndTime,
			boolToInt(ev.IsAllDay),
			ev.Location,
			ev.URL,
			ev.Content,
			encodeStringSlice(ev.Attendees),
			ev.RawPropsJSON,
			ev.NotionUpdatedAt,
			ev.FetchedAt.Format(time.RFC3339),
		)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute upsert events statement for page %s: %w", ev.NotionPageID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit upsert events transaction: %w", err)
	}
	return nil
}

func (r *Repository) ListEventsBetween(ctx context.Context, from, to time.Time) ([]models.Event, error) {
	query := `SELECT notion_page_id, title, start_date, start_time, end_date, end_time, is_all_day, location, url, content, attendees_json, raw_properties, notion_updated_at, fetched_at
	FROM events
	WHERE start_date <= ? AND (end_date IS NULL OR end_date = '' OR end_date >= ?)
	ORDER BY start_date ASC, start_time ASC;`
	rows, err := r.db.QueryContext(ctx, query, to.Format("2006-01-02"), from.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to query events between %s and %s: %w", from.Format("2006-01-02"), to.Format("2006-01-02"), err)
	}
	defer rows.Close()
	return scanEvents(rows)
}

func (r *Repository) ListUpcomingEvents(ctx context.Context, days int, now time.Time) ([]models.Event, error) {
	to := now.AddDate(0, 0, days)
	return r.ListEventsBetween(ctx, now, to)
}

func (r *Repository) GetEvent(ctx context.Context, notionPageID string) (models.Event, bool, error) {
	query := `SELECT notion_page_id, title, start_date, start_time, end_date, end_time, is_all_day, location, url, content, attendees_json, raw_properties, notion_updated_at, fetched_at
	FROM events WHERE notion_page_id = ?;`
	row := r.db.QueryRowContext(ctx, query, notionPageID)
	var ev models.Event
	var isAllDay int
	var attendeesJSON string
	var fetchedAt string
	if err := row.Scan(&ev.NotionPageID, &ev.Title, &ev.StartDate, &ev.StartTime, &ev.EndDate, &ev.EndTime, &isAllDay, &ev.Location, &ev.URL, &ev.Content, &attendeesJSON, &ev.RawPropsJSON, &ev.NotionUpdatedAt, &fetchedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ev, false, nil
		}
		return ev, false, fmt.Errorf("failed to scan event %s: %w", notionPageID, err)
	}
	ev.IsAllDay = intToBool(isAllDay)
	ev.Attendees = decodeStringSlice(attendeesJSON)
	ev.FetchedAt = parseRFC3339(fetchedAt)
	return ev, true, nil
}

func scanEvents(rows *sql.Rows) ([]models.Event, error) {
	var events []models.Event
	for rows.Next() {
		var ev models.Event
		var isAllDay int
		var attendeesJSON string
		var fetchedAt string
		if err := rows.Scan(&ev.NotionPageID, &ev.Title, &ev.StartDate, &ev.StartTime, &ev.EndDate, &ev.EndTime, &isAllDay, &ev.Location, &ev.URL, &ev.Content, &attendeesJSON, &ev.RawPropsJSON, &ev.NotionUpdatedAt, &fetchedAt); err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}
		ev.IsAllDay = intToBool(isAllDay)
		ev.Attendees = decodeStringSlice(attendeesJSON)
		ev.FetchedAt = parseRFC3339(fetchedAt)
		events = append(events, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %w", err)
	}
	return events, nil
}

func (r *Repository) DeleteEventsNotIn(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		_, err := r.db.ExecContext(ctx, `DELETE FROM events;`)
		if err != nil {
			return fmt.Errorf("failed to delete all events: %w", err)
		}
		return nil
	}
	placeholders := inPlaceholders(len(ids))
	query := fmt.Sprintf("DELETE FROM events WHERE notion_page_id NOT IN (%s);", placeholders)
	args := toAnySlice(ids)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete events not in list: %w", err)
	}
	return nil
}

func encodeStringSlice(values []string) string {
	if len(values) == 0 {
		return ""
	}
	data, err := json.Marshal(values)
	if err != nil {
		return ""
	}
	return string(data)
}

func decodeStringSlice(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil
	}
	if len(values) == 0 {
		return nil
	}
	return values
}
