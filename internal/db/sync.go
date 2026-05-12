package db

import (
	"context"
	"database/sql"
	"fmt"

	"notion-notifier/internal/models"
)

func (r *Repository) ListSyncRecords(ctx context.Context) ([]models.SyncRecord, error) {
	query := `SELECT notion_page_id, calendar_event_id, attempted, synced FROM sync_records;`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sync records: %w", err)
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
		return records, fmt.Errorf("failed to query sync record map: %w", err)
	}
	defer rows.Close()
	rowsRecords, err := scanSyncRecords(rows)
	if err != nil {
		return records, fmt.Errorf("failed to scan sync record map: %w", err)
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
	if err != nil {
		return fmt.Errorf("failed to upsert sync record for page %s: %w", record.NotionPageID, err)
	}
	return nil
}

func (r *Repository) ClearSyncRecords(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sync_records;`)
	if err != nil {
		return fmt.Errorf("failed to clear sync records: %w", err)
	}
	return nil
}

// ListOrphanedSyncRecords returns sync_records whose notion_page_id no longer exists in the events table.
func (r *Repository) ListOrphanedSyncRecords(ctx context.Context) ([]models.SyncRecord, error) {
	query := `SELECT s.notion_page_id, s.calendar_event_id, s.attempted, s.synced
	FROM sync_records s LEFT JOIN events e ON s.notion_page_id = e.notion_page_id
	WHERE e.notion_page_id IS NULL;`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query orphaned sync records: %w", err)
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
			return nil, fmt.Errorf("failed to scan sync record row: %w", err)
		}
		rec.Attempted = intToBool(attempted)
		rec.Synced = intToBool(synced)
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sync record rows: %w", err)
	}
	return out, nil
}

// DeleteSyncRecord removes a single sync record by Notion page ID.
func (r *Repository) DeleteSyncRecord(ctx context.Context, notionPageID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sync_records WHERE notion_page_id = ?;`, notionPageID)
	if err != nil {
		return fmt.Errorf("failed to delete sync record for page %s: %w", notionPageID, err)
	}
	return nil
}
