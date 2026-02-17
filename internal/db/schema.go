package db

import (
	"database/sql"
	"strings"
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
	// Migrations for existing databases
	migrations := []string{
		`ALTER TABLE events ADD COLUMN content TEXT;`,
		`ALTER TABLE events ADD COLUMN attendees_json TEXT;`,
		`ALTER TABLE events ADD COLUMN notion_updated_at TEXT;`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			if !strings.Contains(err.Error(), "duplicate column") {
				return err
			}
		}
	}
	// Migrate old sync_records schema to new simplified one
	if err := migrateSyncRecords(db); err != nil {
		return err
	}
	return nil
}

func migrateSyncRecords(db *sql.DB) error {
	// Rebuild sync_records when legacy columns remain or attempted is missing.
	rows, err := db.Query(`PRAGMA table_info(sync_records)`)
	if err != nil {
		return nil // table might not exist yet
	}
	defer rows.Close()

	hasLegacyColumns := false
	hasSyncStatus := false
	hasSynced := false
	hasAttempted := false
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk); err != nil {
			return err
		}
		switch name {
		case "calendar_updated_at", "last_synced_at", "sync_status", "notion_updated_at":
			hasLegacyColumns = true
		}
		if name == "sync_status" {
			hasSyncStatus = true
		}
		if name == "synced" {
			hasSynced = true
		}
		if name == "attempted" {
			hasAttempted = true
		}
	}
	if !hasLegacyColumns && hasAttempted {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS sync_records_new (
		notion_page_id TEXT PRIMARY KEY,
		calendar_event_id TEXT NOT NULL DEFAULT '',
		attempted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);`); err != nil {
		_ = tx.Rollback()
		return err
	}

	var insertSQL string
	switch {
	case hasAttempted && hasSynced:
		insertSQL = `INSERT OR IGNORE INTO sync_records_new (notion_page_id, calendar_event_id, attempted, synced)
		SELECT notion_page_id, COALESCE(calendar_event_id, ''), COALESCE(attempted, 0), COALESCE(synced, 0)
		FROM sync_records;`
	case hasSyncStatus:
		insertSQL = `INSERT OR IGNORE INTO sync_records_new (notion_page_id, calendar_event_id, attempted, synced)
		SELECT notion_page_id, COALESCE(calendar_event_id, ''),
			1,
			CASE WHEN sync_status = 'synced' THEN 1 ELSE 0 END
		FROM sync_records;`
	case hasSynced:
		insertSQL = `INSERT OR IGNORE INTO sync_records_new (notion_page_id, calendar_event_id, attempted, synced)
		SELECT notion_page_id, COALESCE(calendar_event_id, ''), 1, COALESCE(synced, 0)
		FROM sync_records;`
	default:
		insertSQL = `INSERT OR IGNORE INTO sync_records_new (notion_page_id, calendar_event_id, attempted, synced)
		SELECT notion_page_id, COALESCE(calendar_event_id, ''), 1, 0
		FROM sync_records;`
	}
	if _, err := tx.Exec(insertSQL); err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`DROP TABLE sync_records;`); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`ALTER TABLE sync_records_new RENAME TO sync_records;`); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
