package db

import (
	"context"
	"database/sql"
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

func initSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS events (
			notion_page_id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			start_date TEXT NOT NULL,
			start_time TEXT,
			end_date TEXT,
			end_time TEXT,
			is_all_day INTEGER DEFAULT 0,
			location TEXT,
			url TEXT,
			raw_properties TEXT,
			fetched_at TEXT DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS notification_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			status TEXT NOT NULL,
			message TEXT,
			notion_page_id TEXT,
			error TEXT,
			sent_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS sync_records (
			notion_page_id TEXT PRIMARY KEY,
			calendar_event_id TEXT,
			notion_updated_at TEXT,
			calendar_updated_at TEXT,
			last_synced_at TEXT,
			sync_status TEXT DEFAULT 'synced'
		);`,
		`CREATE TABLE IF NOT EXISTS advance_schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			notion_page_id TEXT NOT NULL,
			rule_index INTEGER NOT NULL,
			fire_at TEXT NOT NULL,
			fired INTEGER DEFAULT 0,
			UNIQUE(notion_page_id, rule_index)
		);`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) UpsertEvents(ctx context.Context, events []models.Event) error {
	if len(events) == 0 {
		return nil
	}
	query := `INSERT INTO events (
		notion_page_id, title, start_date, start_time, end_date, end_time, is_all_day, location, url, raw_properties, fetched_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(notion_page_id) DO UPDATE SET
		title=excluded.title,
		start_date=excluded.start_date,
		start_time=excluded.start_time,
		end_date=excluded.end_date,
		end_time=excluded.end_time,
		is_all_day=excluded.is_all_day,
		location=excluded.location,
		url=excluded.url,
		raw_properties=excluded.raw_properties,
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
		isAllDay := 0
		if ev.IsAllDay {
			isAllDay = 1
		}
		_, err := stmt.ExecContext(ctx,
			ev.NotionPageID,
			ev.Title,
			ev.StartDate,
			ev.StartTime,
			ev.EndDate,
			ev.EndTime,
			isAllDay,
			ev.Location,
			ev.URL,
			ev.RawPropsJSON,
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
	query := `SELECT notion_page_id, title, start_date, start_time, end_date, end_time, is_all_day, location, url, raw_properties, fetched_at
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
	query := `SELECT notion_page_id, title, start_date, start_time, end_date, end_time, is_all_day, location, url, raw_properties, fetched_at
	FROM events WHERE notion_page_id = ?;`
	row := r.db.QueryRowContext(ctx, query, notionPageID)
	var ev models.Event
	var isAllDay int
	var fetchedAt string
	if err := row.Scan(&ev.NotionPageID, &ev.Title, &ev.StartDate, &ev.StartTime, &ev.EndDate, &ev.EndTime, &isAllDay, &ev.Location, &ev.URL, &ev.RawPropsJSON, &fetchedAt); err != nil {
		if err == sql.ErrNoRows {
			return ev, false, nil
		}
		return ev, false, err
	}
	ev.IsAllDay = isAllDay == 1
	if fetchedAt != "" {
		if t, err := time.Parse(time.RFC3339, fetchedAt); err == nil {
			ev.FetchedAt = t
		}
	}
	return ev, true, nil
}

func scanEvents(rows *sql.Rows) ([]models.Event, error) {
	var events []models.Event
	for rows.Next() {
		var ev models.Event
		var isAllDay int
		var fetchedAt string
		if err := rows.Scan(&ev.NotionPageID, &ev.Title, &ev.StartDate, &ev.StartTime, &ev.EndDate, &ev.EndTime, &isAllDay, &ev.Location, &ev.URL, &ev.RawPropsJSON, &fetchedAt); err != nil {
			return nil, err
		}
		ev.IsAllDay = isAllDay == 1
		if fetchedAt != "" {
			if t, err := time.Parse(time.RFC3339, fetchedAt); err == nil {
				ev.FetchedAt = t
			}
		}
		events = append(events, ev)
	}
	return events, rows.Err()
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
		if sentAt != "" {
			if t, err := time.Parse(time.RFC3339, sentAt); err == nil {
				h.SentAt = t
			}
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (r *Repository) ClearNotificationHistory(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notification_history;`)
	return err
}

func (r *Repository) ReplaceAdvanceSchedules(ctx context.Context, schedules []models.AdvanceSchedule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM advance_schedules;`); err != nil {
		_ = tx.Rollback()
		return err
	}
	if len(schedules) == 0 {
		return tx.Commit()
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO advance_schedules (notion_page_id, rule_index, fire_at, fired) VALUES (?, ?, ?, ?)
	ON CONFLICT(notion_page_id, rule_index) DO UPDATE SET fire_at=excluded.fire_at, fired=excluded.fired;`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, sched := range schedules {
		fired := 0
		if sched.Fired {
			fired = 1
		}
		_, err := stmt.ExecContext(ctx, sched.NotionPageID, sched.RuleIndex, sched.FireAt.Format(time.RFC3339), fired)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) ListPendingAdvanceSchedules(ctx context.Context) ([]models.AdvanceSchedule, error) {
	query := `SELECT id, notion_page_id, rule_index, fire_at, fired FROM advance_schedules WHERE fired = 0 ORDER BY fire_at ASC;`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var schedules []models.AdvanceSchedule
	for rows.Next() {
		var sched models.AdvanceSchedule
		var fireAt string
		var fired int
		if err := rows.Scan(&sched.ID, &sched.NotionPageID, &sched.RuleIndex, &fireAt, &fired); err != nil {
			return nil, err
		}
		sched.Fired = fired == 1
		if t, err := time.Parse(time.RFC3339, fireAt); err == nil {
			sched.FireAt = t
		}
		schedules = append(schedules, sched)
	}
	return schedules, rows.Err()
}

func (r *Repository) MarkAdvanceScheduleFired(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE advance_schedules SET fired = 1 WHERE id = ?;`, id)
	return err
}

func (r *Repository) GetSyncRecord(ctx context.Context, notionPageID string) (models.SyncRecord, bool, error) {
	var record models.SyncRecord
	query := `SELECT notion_page_id, calendar_event_id, notion_updated_at, calendar_updated_at, last_synced_at, sync_status FROM sync_records WHERE notion_page_id = ?;`
	row := r.db.QueryRowContext(ctx, query, notionPageID)
	if err := row.Scan(&record.NotionPageID, &record.CalendarEventID, &record.NotionUpdatedAt, &record.CalendarUpdatedAt, &record.LastSyncedAt, &record.SyncStatus); err != nil {
		if err == sql.ErrNoRows {
			return record, false, nil
		}
		return record, false, err
	}
	return record, true, nil
}

func (r *Repository) GetSyncStatusMap(ctx context.Context, notionPageIDs []string) (map[string]string, error) {
	statuses := make(map[string]string)
	if len(notionPageIDs) == 0 {
		return statuses, nil
	}
	placeholders := strings.Repeat("?,", len(notionPageIDs))
	placeholders = strings.TrimRight(placeholders, ",")
	query := fmt.Sprintf("SELECT notion_page_id, sync_status FROM sync_records WHERE notion_page_id IN (%s);", placeholders)
	args := make([]interface{}, len(notionPageIDs))
	for i, id := range notionPageIDs {
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return statuses, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var status string
		if err := rows.Scan(&id, &status); err != nil {
			return statuses, err
		}
		statuses[id] = status
	}
	return statuses, rows.Err()
}

func (r *Repository) UpsertSyncRecord(ctx context.Context, record models.SyncRecord) error {
	query := `INSERT INTO sync_records (notion_page_id, calendar_event_id, notion_updated_at, calendar_updated_at, last_synced_at, sync_status)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(notion_page_id) DO UPDATE SET
		calendar_event_id=excluded.calendar_event_id,
		notion_updated_at=excluded.notion_updated_at,
		calendar_updated_at=excluded.calendar_updated_at,
		last_synced_at=excluded.last_synced_at,
		sync_status=excluded.sync_status;`
	_, err := r.db.ExecContext(ctx, query,
		record.NotionPageID,
		record.CalendarEventID,
		record.NotionUpdatedAt,
		record.CalendarUpdatedAt,
		record.LastSyncedAt,
		record.SyncStatus,
	)
	return err
}

func (r *Repository) ClearSyncRecords(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sync_records;`)
	return err
}

func (r *Repository) GetSyncSummary(ctx context.Context) (string, int, error) {
	query := `SELECT MAX(last_synced_at), COUNT(*) FROM sync_records;`
	row := r.db.QueryRowContext(ctx, query)
	var lastSynced sql.NullString
	var count int
	if err := row.Scan(&lastSynced, &count); err != nil {
		return "", 0, err
	}
	if lastSynced.Valid {
		return lastSynced.String, count, nil
	}
	return "", count, nil
}

func (r *Repository) DeleteEventsNotIn(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		_, err := r.db.ExecContext(ctx, `DELETE FROM events;`)
		return err
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = strings.TrimRight(placeholders, ",")
	query := fmt.Sprintf("DELETE FROM events WHERE notion_page_id NOT IN (%s);", placeholders)
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}
