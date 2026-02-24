package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"notion-notifier/internal/models"
)

type Repository struct {
	db *sql.DB
}

func Open(path string) (*Repository, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := initSchema(db); err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

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
		return err
	}
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		_ = tx.Rollback()
		return err
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
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) ListEventsBetween(ctx context.Context, from, to time.Time) ([]models.Event, error) {
	query := `SELECT notion_page_id, title, start_date, start_time, end_date, end_time, is_all_day, location, url, content, attendees_json, raw_properties, notion_updated_at, fetched_at
	FROM events
	WHERE start_date <= ? AND (end_date IS NULL OR end_date = '' OR end_date >= ?)
	ORDER BY start_date ASC, start_time ASC;`
	rows, err := r.db.QueryContext(ctx, query, to.Format("2006-01-02"), from.Format("2006-01-02"))
	if err != nil {
		return nil, err
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
		if err == sql.ErrNoRows {
			return ev, false, nil
		}
		return ev, false, err
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
			return nil, err
		}
		ev.IsAllDay = intToBool(isAllDay)
		ev.Attendees = decodeStringSlice(attendeesJSON)
		ev.FetchedAt = parseRFC3339(fetchedAt)
		events = append(events, ev)
	}
	return events, rows.Err()
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

func (r *Repository) InsertNotificationHistory(ctx context.Context, h models.NotificationHistory) error {
	query := `INSERT INTO notification_history (type, status, message, notion_page_id, error, sent_at) VALUES (?, ?, ?, ?, ?, ?);`
	_, err := r.db.ExecContext(ctx, query, h.Type, h.Status, h.Message, h.NotionPageID, h.Error, h.SentAt.Format(time.RFC3339))
	return err
}

func (r *Repository) ListNotificationHistory(ctx context.Context, limit int) ([]models.NotificationHistory, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, type, status, message, notion_page_id, error, sent_at FROM notification_history ORDER BY sent_at DESC LIMIT ?;`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.NotificationHistory
	for rows.Next() {
		var h models.NotificationHistory
		var sentAt string
		if err := rows.Scan(&h.ID, &h.Type, &h.Status, &h.Message, &h.NotionPageID, &h.Error, &sentAt); err != nil {
			return nil, err
		}
		h.SentAt = parseRFC3339(sentAt)
		out = append(out, h)
	}
	return out, rows.Err()
}

func (r *Repository) ClearNotificationHistory(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notification_history;`)
	return err
}

func (r *Repository) ReplaceUpcomingSchedules(ctx context.Context, schedules []models.UpcomingSchedule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if len(schedules) == 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM upcoming_schedules;`); err != nil {
			_ = tx.Rollback()
			return err
		}
		return tx.Commit()
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO upcoming_schedules (notion_page_id, rule_index, fire_at, fired) VALUES (?, ?, ?, ?)
	ON CONFLICT(notion_page_id, rule_index) DO UPDATE SET
		fire_at=excluded.fire_at,
		fired=CASE
			WHEN upcoming_schedules.fire_at = excluded.fire_at THEN upcoming_schedules.fired
			ELSE excluded.fired
		END;`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, sched := range schedules {
		_, err := stmt.ExecContext(ctx, sched.NotionPageID, sched.RuleIndex, sched.FireAt.Format(time.RFC3339), boolToInt(sched.Fired))
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	conditions := make([]string, 0, len(schedules))
	args := make([]any, 0, len(schedules)*2)
	for _, sched := range schedules {
		conditions = append(conditions, "(notion_page_id = ? AND rule_index = ?)")
		args = append(args, sched.NotionPageID, sched.RuleIndex)
	}
	query := `DELETE FROM upcoming_schedules WHERE NOT (` + strings.Join(conditions, " OR ") + `);`
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *Repository) ListPendingUpcomingSchedules(ctx context.Context) ([]models.UpcomingSchedule, error) {
	query := `SELECT id, notion_page_id, rule_index, fire_at, fired FROM upcoming_schedules WHERE fired = 0 ORDER BY fire_at ASC;`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var schedules []models.UpcomingSchedule
	for rows.Next() {
		var sched models.UpcomingSchedule
		var fireAt string
		var fired int
		if err := rows.Scan(&sched.ID, &sched.NotionPageID, &sched.RuleIndex, &fireAt, &fired); err != nil {
			return nil, err
		}
		sched.Fired = intToBool(fired)
		sched.FireAt = parseRFC3339(fireAt)
		schedules = append(schedules, sched)
	}
	return schedules, rows.Err()
}

func (r *Repository) MarkUpcomingScheduleFired(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE upcoming_schedules SET fired = 1 WHERE id = ?;`, id)
	return err
}

func (r *Repository) ListSyncRecords(ctx context.Context) ([]models.SyncRecord, error) {
	query := `SELECT notion_page_id, calendar_event_id, attempted, synced FROM sync_records;`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSyncRecords(rows)
}

func (r *Repository) GetSyncRecordMap(ctx context.Context, notionPageIDs []string) (map[string]models.SyncRecord, error) {
	records := make(map[string]models.SyncRecord)
	if len(notionPageIDs) == 0 {
		return records, nil
	}
	placeholders := inPlaceholders(len(notionPageIDs))
	query := fmt.Sprintf("SELECT notion_page_id, calendar_event_id, attempted, synced FROM sync_records WHERE notion_page_id IN (%s);", placeholders)
	args := toAnySlice(notionPageIDs)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return records, err
	}
	defer rows.Close()
	rowsRecords, err := scanSyncRecords(rows)
	if err != nil {
		return records, err
	}
	for _, rec := range rowsRecords {
		records[rec.NotionPageID] = rec
	}
	return records, nil
}

func (r *Repository) UpsertSyncRecord(ctx context.Context, record models.SyncRecord) error {
	query := `INSERT INTO sync_records (notion_page_id, calendar_event_id, attempted, synced)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(notion_page_id) DO UPDATE SET
		calendar_event_id=excluded.calendar_event_id,
		attempted=excluded.attempted,
		synced=excluded.synced;`
	_, err := r.db.ExecContext(ctx, query,
		record.NotionPageID,
		record.CalendarEventID,
		boolToInt(record.Attempted),
		boolToInt(record.Synced),
	)
	return err
}

func (r *Repository) ClearSyncRecords(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sync_records;`)
	return err
}

// ListOrphanedSyncRecords returns sync_records whose notion_page_id no longer exists in the events table.
func (r *Repository) ListOrphanedSyncRecords(ctx context.Context) ([]models.SyncRecord, error) {
	query := `SELECT s.notion_page_id, s.calendar_event_id, s.attempted, s.synced
	FROM sync_records s LEFT JOIN events e ON s.notion_page_id = e.notion_page_id
	WHERE e.notion_page_id IS NULL;`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSyncRecords(rows)
}

func scanSyncRecords(rows *sql.Rows) ([]models.SyncRecord, error) {
	var out []models.SyncRecord
	for rows.Next() {
		var rec models.SyncRecord
		var attempted int
		var synced int
		if err := rows.Scan(&rec.NotionPageID, &rec.CalendarEventID, &attempted, &synced); err != nil {
			return nil, err
		}
		rec.Attempted = intToBool(attempted)
		rec.Synced = intToBool(synced)
		out = append(out, rec)
	}
	return out, rows.Err()
}

// DeleteSyncRecord removes a single sync record by Notion page ID.
func (r *Repository) DeleteSyncRecord(ctx context.Context, notionPageID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sync_records WHERE notion_page_id = ?;`, notionPageID)
	return err
}

func (r *Repository) DeleteEventsNotIn(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		_, err := r.db.ExecContext(ctx, `DELETE FROM events;`)
		return err
	}
	placeholders := inPlaceholders(len(ids))
	query := fmt.Sprintf("DELETE FROM events WHERE notion_page_id NOT IN (%s);", placeholders)
	args := toAnySlice(ids)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func inPlaceholders(length int) string {
	return strings.TrimRight(strings.Repeat("?,", length), ",")
}

func toAnySlice(values []string) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func intToBool(value int) bool {
	return value == 1
}

func parseRFC3339(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
