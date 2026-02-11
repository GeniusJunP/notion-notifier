# Config 仕様

## 概要

- `config.yaml` が唯一の永続設定ソース
- API経由で更新されるたびに正規化（`NormalizeConfig`）→ バリデーション（`ValidateConfig`）を通過
- `schema_version` で将来の移行に対応

## 設定ファイル構造

```yaml
schema_version: 1        # 自動付与
timezone: "Asia/Tokyo"   # 必須, IANA タイムゾーン

sync:
  check_interval: 15     # 必須, > 0, 分単位

notifications:
  advance:               # 事前通知（配列）
    - enabled: true
      minutes_before: 30  # 必須, > 0
      message: "テンプレート"
      location: ""
      url: ""
      conditions:
        enabled: false
        days_of_week: []       # 1-7 (月-日)
        property_filters: []
  periodic:              # 定期通知（配列）
    - enabled: true
      days_of_week: [1, 4]    # 1-7 (月-日)
      time: "09:00"           # 必須, HH:mm
      days_ahead: 7           # 必須, > 0
      message: "テンプレート"

webhook:
  schedule:
    content_type: "application/json"
    payload_template: '{"content":{{json .Message}}}'
  notification:
    content_type: "application/json"
    payload_template: '{"content":{{json .Message}}}'

calendar_sync:
  enabled: false
  interval_hours: 6      # > 0
  lookahead_days: 30     # > 0

property_mapping:
  title: "名前"
  date: "日付"
  location: "場所"
  custom: []

content_rules:
  start_heading: ""
  include_start_heading: false
  stop_at_next_heading: true
  stop_at_delimiter: true
  delimiter_text: ""

snooze_until: ""         # RFC3339 or ""
mute_until: ""           # RFC3339 or ""

security:
  basic_auth:
    enabled: false
```

## フィールド詳細

| フィールド | 型 | 必須 | デフォルト | 説明 |
|---|---|---|---|---|
| `schema_version` | int | 自動 | 1 | 正規化時に自動設定 |
| `timezone` | string | ○ | "Asia/Tokyo" | IANA タイムゾーン |
| `sync.check_interval` | int | ○ | 15 | 同期間隔（分） |
| `notifications.advance[].enabled` | bool | - | false | 有効/無効 |
| `notifications.advance[].minutes_before` | int | ○ | - | 何分前に通知 |
| `notifications.advance[].message` | string | - | "" | Go template |
| `notifications.advance[].conditions.enabled` | bool | - | false | 条件フィルタ有効化 |
| `notifications.advance[].conditions.days_of_week` | []int | - | [] | 1-7 |
| `notifications.advance[].conditions.property_filters` | []obj | - | [] | プロパティフィルタ |
| `notifications.periodic[].enabled` | bool | - | false | 有効/無効 |
| `notifications.periodic[].days_of_week` | []int | - | [] | 1-7 |
| `notifications.periodic[].time` | string | ○ | - | HH:mm |
| `notifications.periodic[].days_ahead` | int | ○ | - | 何日先まで |
| `notifications.periodic[].message` | string | - | "" | Go template |
| `webhook.schedule.content_type` | string | - | "application/json" | Content-Type |
| `webhook.schedule.payload_template` | string | - | `{"content":"{{.Message}}"}` | ペイロードテンプレート |
| `calendar_sync.enabled` | bool | - | false | カレンダー同期有効化 |
| `calendar_sync.interval_hours` | int | - | 6 | 自動同期間隔（時間） |
| `calendar_sync.lookahead_days` | int | - | 30 | 同期対象日数 |
| `snooze_until` | string | - | "" | スヌーズ期限 (RFC3339) |
| `mute_until` | string | - | "" | ミュート期限 (RFC3339) |
| `security.basic_auth.enabled` | bool | - | false | Basic認証有効化 |

## 正規化ルール (`NormalizeConfig`)

1. `schema_version` を最新版 (1) に設定
2. `notifications.weekly` → `notifications.periodic` への移行
3. `timezone` が空なら `"Asia/Tokyo"`
4. `sync.check_interval` が 0以下なら `15`
5. `calendar_sync.interval_hours` が 0以下なら `6`
6. `calendar_sync.lookahead_days` が 0以下なら `30`
7. Webhook の `content_type` が空なら `"application/json"`
8. Webhook の `payload_template` が空ならデフォルトテンプレート
9. テンプレート文字列の `\r\n` → `\n` 変換

## バリデーションルール (`ValidateConfig`)

- `timezone`: 必須, `time.LoadLocation` で有効なこと
- `sync.check_interval`: > 0
- `calendar_sync.interval_hours`: > 0
- `calendar_sync.lookahead_days`: > 0
- `notifications.advance[].minutes_before`: > 0
- `notifications.periodic[].days_ahead`: > 0
- `notifications.periodic[].time`: HH:mm 形式
- `notifications.periodic[].days_of_week`: 各要素 1-7
- `snooze_until`, `mute_until`: 空 or RFC3339

## 認証情報 (`env.yaml`)

認証情報は `env.yaml` に分離。WebUI/API では編集不可。

```yaml
notion:
  api_key: ""
  database_id: ""
webhook:
  schedule_url: ""
  notification_url: ""
google:
  calendar_id: ""
  oauth_client_id: ""
  oauth_client_secret: ""
  oauth_refresh_token: ""
security:
  basic_auth:
    username: ""
    password: ""
```

環境変数で上書き可能（詳細は SPEC.md 参照）。
