package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"notion-notifier/internal/models"
)

func (r *Repository) ReplaceUpcomingSchedules(ctx context.Context, schedules []models.UpcomingSchedule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for replace upcoming schedules: %w", err)
	}
	if len(schedules) == 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM upcoming_schedules;`); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to delete all upcoming schedules: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit delete all upcoming schedules transaction: %w", err)
		}
		return nil
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
		return fmt.Errorf("failed to prepare replace upcoming schedules statement: %w", err)
	}
	defer stmt.Close()
	for _, sched := range schedules {
		_, err := stmt.ExecContext(ctx, sched.NotionPageID, sched.RuleIndex, sched.FireAt.Format(time.RFC3339), boolToInt(sched.Fired))
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute replace upcoming schedules statement for page %s rule %d: %w", sched.NotionPageID, sched.RuleIndex, err)
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
		return fmt.Errorf("failed to delete stale upcoming schedules: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit replace upcoming schedules transaction: %w", err)
	}
	return nil
}

func (r *Repository) ListPendingUpcomingSchedules(ctx context.Context) ([]models.UpcomingSchedule, error) {
	query := `SELECT id, notion_page_id, rule_index, fire_at, fired FROM upcoming_schedules WHERE fired = 0 ORDER BY fire_at ASC;`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending upcoming schedules: %w", err)
	}
	defer rows.Close()
	var schedules []models.UpcomingSchedule
	for rows.Next() {
		var sched models.UpcomingSchedule
		var fireAt string
		var fired int
		if err := rows.Scan(&sched.ID, &sched.NotionPageID, &sched.RuleIndex, &fireAt, &fired); err != nil {
			return nil, fmt.Errorf("failed to scan pending upcoming schedule row: %w", err)
		}
		sched.Fired = intToBool(fired)
		sched.FireAt = parseRFC3339(fireAt)
		schedules = append(schedules, sched)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pending upcoming schedule rows: %w", err)
	}
	return schedules, nil
}

func (r *Repository) MarkUpcomingScheduleFired(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE upcoming_schedules SET fired = 1 WHERE id = ?;`, id)
	if err != nil {
		return fmt.Errorf("failed to mark upcoming schedule %d as fired: %w", id, err)
	}
	return nil
}
