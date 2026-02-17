package db

import (
	"database/sql"
)

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
			content TEXT,
			attendees_json TEXT,
			raw_properties TEXT,
			notion_updated_at TEXT,
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
			calendar_event_id TEXT NOT NULL DEFAULT '',
			attempted INTEGER NOT NULL DEFAULT 0,
			synced INTEGER NOT NULL DEFAULT 0
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
