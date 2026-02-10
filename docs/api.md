# API Reference

## 概要

全エンドポイントは JSON のみ。HTML/form 返却なし。  
認証: Basic認証が有効な場合、全エンドポイントに適用。

## 共通仕様

### レスポンスフォーマット

- 成功: `200 OK` + JSON body（または `204 No Content`）
- クライアントエラー: `400`/`422` + `{"error": "message"}` or `{"error": "message", "details": {...}}`
- サーバーエラー: `500` + `{"error": "message"}`

### 日時フォーマット

- 日付入力: `2006-01-02`, `2006-01-02T15:04`, RFC3339
- 日時出力: RFC3339 (`2006-01-02T15:04:05+09:00`)

---

## エンドポイント一覧

| Method | Path | Description |
|--------|------|-------------|
| GET    | `/api/config` | 現在の設定を取得 |
| PUT    | `/api/config` | 設定全体を保存 |
| GET    | `/api/dashboard` | ダッシュボード集計データ |
| GET    | `/api/history` | 通知履歴 |
| GET    | `/api/events/upcoming` | 今後の予定一覧 |
| POST   | `/api/sync` | Notion同期実行 |
| POST   | `/api/calendar/sync` | カレンダー手動同期 |
| POST   | `/api/calendar/clear` | 同期レコード削除 |
| POST   | `/api/history/clear` | 通知履歴削除 |
| POST   | `/api/notifications/preview` | テンプレートプレビュー |
| POST   | `/api/notifications/manual` | 手動通知送信 |
| GET    | `/api/templates/defaults` | デフォルトテンプレート取得 |

---

## 詳細

### GET /api/config

現在の正規化済み設定を取得。

**Response 200:**
```json
{
  "schema_version": 1,
  "timezone": "Asia/Tokyo",
  "sync": { "check_interval": 15 },
  "notifications": {
    "advance": [...],
    "periodic": [...]
  },
  "webhook": { ... },
  "calendar_sync": { ... },
  "property_mapping": { ... },
  "content_rules": { ... },
  "snooze_until": "",
  "mute_until": "",
  "security": { "basic_auth": { "enabled": false } }
}
```

### PUT /api/config

設定全体を保存。正規化 → バリデーション → 保存。  
成功時は正規化済みConfigを返す。

**Request Body:** Config JSON（全フィールド）

**Response 200:** 正規化済み Config JSON

**Response 422:**
```json
{
  "error": "validation failed",
  "details": { "config": "timezone is required" }
}
```

### GET /api/dashboard

**Response 200:**
```json
{
  "today_count": 3,
  "next_sync": "2026-02-10T10:15:00+09:00",
  "next_sync_in": "12m",
  "last_sync": "2026-02-10T10:00:00+09:00",
  "last_sync_count": 25,
  "last_sync_error": "",
  "snooze_active": false,
  "snooze_until": "",
  "mute_active": false,
  "mute_until": ""
}
```

### GET /api/events/upcoming

直近14日間の予定を同期ステータス付きで返す。

**Response 200:**
```json
[
  {
    "notion_page_id": "abc123",
    "title": "ミーティング",
    "start_date": "2026-02-11",
    "start_time": "14:00",
    "end_date": "2026-02-11",
    "end_time": "15:00",
    "is_all_day": false,
    "location": "会議室A",
    "url": "https://notion.so/abc123",
    "sync_status": "synced",  // cache status (backward compatibility)
    "cache_status": "synced", // "synced" | "pending" | "unsynced"
    "calendar_status": "present" // "present" | "missing" | "disabled" | "unavailable"
  }
]
```

### GET /api/history

直近50件の通知履歴。

**Response 200:**
```json
[
  {
    "id": 1,
    "type": "periodic",
    "status": "success",
    "message": "今週の予定...",
    "error": "",
    "sent_at": "2026-02-10T09:00:00+09:00"
  }
]
```

### POST /api/sync

Notion同期を手動実行。

**Response 200:**
```json
{ "count": 25 }
```

### POST /api/calendar/sync

カレンダー同期を手動実行。

**Request Body:**
```json
{
  "from_date": "2026-02-10",
  "to_date": "2026-03-10"
}
```

**Response 200:**
```json
{ "count": 5 }
```

### POST /api/calendar/clear

同期レコードを全削除。

**Response:** `204 No Content`

### POST /api/history/clear

通知履歴を全削除。

**Response:** `204 No Content`

### POST /api/notifications/preview

テンプレートをプレビュー。`minutes_before > 0` の場合は事前通知テンプレートとしてプレビュー。

**Request Body:**
```json
{
  "template": "⏰ {{.MinutesBefore}}分後に「{{.Name}}」",
  "from_date": "2026-02-10",
  "to_date": "2026-02-17",
  "minutes_before": 30
}
```

**Response 200:**
```json
{
  "message": "⏰ 30分後に「ミーティング」",
  "payload": "{\"content\":\"...\"}"
}
```

### POST /api/notifications/manual

手動通知をWebhookに送信。

**Request Body:**
```json
{
  "template": "今週の予定: {{range .Events}}...",
  "from_date": "2026-02-10",
  "to_date": "2026-02-17"
}
```

**Response 200:**
```json
{
  "message": "今週の予定: ..."
}
```

### GET /api/templates/defaults

デフォルトのメッセージテンプレートを取得。テンプレートリセット用。

**Response 200:**
```json
{
  "advance": "📢 まもなく「{{.Name}}」が始まります...",
  "periodic": "📋 今後の予定（{{len .Events}}件）..."
}
```
