package db

import (
	"context"
	"fmt"
	"time"

	"notion-notifier/internal/models"
)

func (r *Repository) InsertNotificationHistory(ctx context.Context, h models.NotificationHistory) error {
	query := `INSERT INTO notification_history (type, status, message, notion_page_id, error, sent_at) VALUES (?, ?, ?, ?, ?, ?);`
	_, err := r.db.ExecContext(ctx, query, h.Type, h.Status, h.Message, h.NotionPageID, h.Error, h.SentAt.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to insert notification history: %w", err)
	}
	return nil
}

func (r *Repository) ListNotificationHistory(ctx context.Context, limit int) ([]models.NotificationHistory, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, type, status, message, notion_page_id, error, sent_at FROM notification_history ORDER BY sent_at DESC LIMIT ?;`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query notification history: %w", err)
	}
	defer rows.Close()
	var out []models.NotificationHistory
	for rows.Next() {
		var h models.NotificationHistory
		var sentAt string
		if err := rows.Scan(&h.ID, &h.Type, &h.Status, &h.Message, &h.NotionPageID, &h.Error, &sentAt); err != nil {
			return nil, fmt.Errorf("failed to scan notification history row: %w", err)
		}
		h.SentAt = parseRFC3339(sentAt)
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notification history rows: %w", err)
	}
	return out, nil
}

func (r *Repository) ClearNotificationHistory(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notification_history;`)
	if err != nil {
		return fmt.Errorf("failed to clear notification history: %w", err)
	}
	return nil
}
