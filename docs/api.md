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
  "timezone": "Asia/Tokyo",
  "sync": { "check_interval": 15 },
  "notifications": {
    "upcoming": [...],
    "periodic": [...],
    "manual": "{{if .Events}}..."
  },
  "webhook": { ... },
  "calendar_sync": { ... },
  "property_mapping": { ... },
  "content_rules": { ... },
  "snooze_until": ""
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
  "snooze_until": ""
}
```

### GET /api/events/upcoming

直近14日間の予定を返す。状態は `calendar_state` のみを返す。

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
    "calendar_state": "synced" // "disabled" | "needs_sync" | "synced" | "error"
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

カレンダー同期を手動実行。Notionを正として、Calendar APIで取得した追跡イベント（`notion_page_id`付き）を逆引き照合する。

- Notionにない追跡イベントは削除
- Notionとの差分（Calendar側の編集含む）はUpsertで上書き
- DBに存在してCalendarにない予定はUpsertで作成/復元
- 同一 `notion_page_id` の重複イベントは1件に整理

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

テンプレートをプレビュー。`minutes_before > 0` の場合は事前通知（upcoming）のテンプレートとしてプレビュー。

**Request Body:**
```json
{
  "template": "## 予定リマインド！⏰\n@everyone **{{.Name}}** が **{{.MinutesBefore}}分後** に始まります！\n\n### 詳細\n- **日時:** {{.Date}} {{if .IsAllDay}}(終日){{else}}`{{.Time}}`{{end}}",
  "from_date": "2026-02-10",
  "to_date": "2026-02-17",
  "minutes_before": 30
}
```

**Response 200:**
```json
{
  "message": "## 予定リマインド！⏰\n@everyone **ミーティング** が **30分後** に始まります！\n\n### 詳細\n- **日時:** 2026-02-11 `10:00`"
}
```

### POST /api/notifications/manual

手動通知をWebhookに送信。

**Request Body:**
```json
{
  "template": "{{if .Events}}\n## 今週の予定！📣\n@everyone **今週は {{len .Events}} 件** あります！\n{{range .Events}}\n### {{.Name}}\n- **日時:** {{.Date}} {{if .IsAllDay}}(終日){{else}}`{{.Time}}`{{end}}\n{{end}}\n{{else}}\n## 今週の予定！📣\n@everyone 今週の予定はありません！\n{{end}}",
  "from_date": "2026-02-10",
  "to_date": "2026-02-17"
}
```

**Response 200:**
```json
{
  "message": "## 今週の予定！📣\n@everyone **今週は 2 件** あります！\n\n### 企画会議\n- **日時:** 2026-02-11 `10:00`"
}
```

### GET /api/templates/defaults

デフォルトのメッセージテンプレートを取得。テンプレートリセット用。

**Response 200:**
```json
{
  "upcoming": "## 予定リマインド！⏰\n@everyone **{{.Name}}** が **{{.MinutesBefore}}分後** に始まります！...",
  "periodic": "{{if .Events}}\n## 今週の予定！📣\n@everyone **今週は {{len .Events}} 件** あります！...",
  "manual": "{{if .Events}}\n## 今週の予定！📣\n@everyone **今週は {{len .Events}} 件** あります！..."
}
```
