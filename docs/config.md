# Config 仕様

## 概要

- `config.yaml` が唯一の永続設定ソース
- API経由で更新されるたびに正規化（`NormalizeConfig`）→ バリデーション（`ValidateConfig`）を通過

## 設定ファイル構造

```yaml
timezone: "Asia/Tokyo"   # 必須, IANA タイムゾーン

sync:
  check_interval: 15     # 必須, > 0, 分単位

notifications:
  advance:               # 事前通知（配列）
    - enabled: true
      minutes_before: 30  # 必須, > 0
      message: |
        ## 予定リマインド！⏰
        @everyone **{{.Name}}** が **{{.MinutesBefore}}分後** に始まります！

        ### 詳細
        - **日時:** {{.Date}} {{if .IsAllDay}}(終日){{else}}`{{.Time}}`{{end}}{{if .EndDate}} 〜 {{.EndDate}} {{if .EndTime}}`{{.EndTime}}`{{end}}{{end}}
        {{if .Location}}- **場所:** {{.Location}}{{end}}
        {{if .URL}}- **詳細:** {{.URL}}{{end}}
        {{with .Content}}- **メモ:** {{.}}{{end}}
      conditions:
        days_of_week: []       # 1-7 (月-日), 空配列=制限なし
        property_filters: []
  periodic:              # 定期通知（配列）
    - enabled: true
      days_of_week: [1, 4]    # 1-7 (月-日), 空配列=制限なし（毎日）
      time: "09:00"           # 必須, HH:mm
      days_ahead: 7           # 必須, > 0
      message: |
        {{if .Events}}
        ## 今週の予定！📣
        @everyone **今週は {{len .Events}} 件** あります！

        {{range .Events}}
        ### {{.Name}}
        - **日時:** {{.Date}} {{if .IsAllDay}}(終日){{else}}`{{.Time}}`{{end}}{{if .EndDate}} 〜 {{.EndDate}} {{if .EndTime}}`{{.EndTime}}`{{end}}{{end}}
        {{if .Location}}- **場所:** {{.Location}}{{end}}
        {{if .URL}}- **詳細:** {{.URL}}{{end}}
        {{with .Content}}- **メモ:** {{.}}{{end}}

        {{end}}
        {{else}}
        ## 今週の予定！📣
        @everyone 今週の予定はありません！
        {{end}}
  manual: |              # 手動通知テンプレート
    {{if .Events}}
    ## 今週の予定！📣
    @everyone **今週は {{len .Events}} 件** あります！
    {{range .Events}}
    ### {{.Name}}
    {{end}}
    {{else}}
    ## 今週の予定！📣
    @everyone 今週の予定はありません！
    {{end}}

webhook:
  is_test: false
  notification:
    content_type: "application/json"
    payload_template: '{"content":{{json .Message}}}'
  internal_notification:
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
  attendees: ""          # Notion people property name
  attendees_enabled: false
  custom: []

content_rules:
  start_heading: ""
  include_start_heading: false
  stop_at_next_heading: true
  stop_at_delimiter: true
  delimiter_text: ""

snooze:
  until: ""              # RFC3339 or ""
  mute_upcoming: true
  mute_periodic: true
```

## フィールド詳細

| フィールド | 型 | 必須 | デフォルト | 説明 |
|---|---|---|---|---|
| `timezone` | string | ○ | "Asia/Tokyo" | IANA タイムゾーン |
| `sync.check_interval` | int | ○ | 15 | 同期間隔（分） |
| `notifications.advance[].enabled` | bool | - | false | 有効/無効 |
| `notifications.advance[].minutes_before` | int | ○ | - | 何分前に通知 |
| `notifications.advance[].message` | string | - | "" | Go template |
| `notifications.advance[].conditions.days_of_week` | []int | - | [] | 1-7, 空配列は制限なし |
| `notifications.advance[].conditions.property_filters` | []obj | - | [] | プロパティフィルタ |
| `notifications.periodic[].enabled` | bool | - | false | 有効/無効 |
| `notifications.periodic[].days_of_week` | []int | - | [] | 1-7, 空配列は制限なし（毎日） |
| `notifications.periodic[].time` | string | ○ | - | HH:mm |
| `notifications.periodic[].days_ahead` | int | ○ | - | 何日先まで |
| `notifications.periodic[].message` | string | - | "" | Go template |
| `notifications.manual` | string | - | `DefaultManualMessage` | 手動通知で使う Go template |
| `webhook.is_test` | bool | - | false | 内部通知テンプレートと内部通知URLを使うテストモード |
| `webhook.notification.content_type` | string | - | "application/json" | Content-Type |
| `webhook.notification.payload_template` | string | - | `{"content":{{json .Message}}}` | ペイロードテンプレート |
| `webhook.internal_notification.content_type` | string | - | "application/json" | Content-Type |
| `webhook.internal_notification.payload_template` | string | - | `{"content":{{json .Message}}}` | ペイロードテンプレート |
| `calendar_sync.enabled` | bool | - | false | カレンダー同期有効化 |
| `calendar_sync.interval_hours` | int | - | 6 | 自動同期間隔（時間） |
| `calendar_sync.lookahead_days` | int | - | 30 | 同期対象日数 |
| `property_mapping.attendees` | string | - | "" | 参加者メール抽出元のNotion peopleプロパティ |
| `property_mapping.attendees_enabled` | bool | - | false | 参加者同期の有効化 |
| `snooze.until` | string | - | "" | スヌーズ期限 (RFC3339) |
| `snooze.mute_upcoming` | bool | - | true | 事前通知をスヌーズ対象にする |
| `snooze.mute_periodic` | bool | - | true | 定期通知をスヌーズ対象にする |

## 正規化ルール (`NormalizeConfig`)

1. `timezone` が空なら `"Asia/Tokyo"`
2. `sync.check_interval` が 0以下なら `15`
3. `calendar_sync.interval_hours` が 0以下なら `6`
4. `calendar_sync.lookahead_days` が 0以下なら `30`
5. Webhook の `content_type` が空なら `"application/json"`
6. Webhook の `payload_template` が空ならデフォルトテンプレート
7. `notifications.manual` が空なら `DefaultManualMessage` を設定
8. テンプレート文字列の `\r\n` → `\n` 変換

## バリデーションルール (`ValidateConfig`)

- `timezone`: 必須, `time.LoadLocation` で有効なこと
- `sync.check_interval`: > 0
- `calendar_sync.interval_hours`: > 0
- `calendar_sync.lookahead_days`: > 0
- `notifications.advance[].minutes_before`: > 0
- `notifications.periodic[].days_ahead`: > 0
- `notifications.periodic[].time`: HH:mm 形式
- `notifications.periodic[].days_of_week`: 各要素 1-7
- `notifications.advance[].conditions.days_of_week`: 各要素 1-7
- `snooze.until`: 空 or RFC3339

## 認証情報 (`env.yaml`)

認証情報は `env.yaml` に分離。WebUI/API では編集不可。
Google Calendar連携は Service Account で接続する。対象カレンダーの共有設定で、Service Account の `client_email` に予定の変更権限を付与する。

```yaml
notion:
  api_key: ""
  database_id: ""
webhook:
  notification_url: ""
  internal_notification_url: ""
google:
  calendar_id: ""
  service_account_key_file: ""
  service_account_key_json: "" # コンテナ等では GOOGLE_SERVICE_ACCOUNT_KEY_JSON env var 推奨
server:
  port: 18080
  tls:
    cert_file: ""
    key_file: ""
security:
  basic_auth:
    enabled: false
    username: ""
    password: ""
```

`server.tls.cert_file` と `server.tls.key_file` の両方が空の場合、サーバーは自己署名証明書を自動生成してHTTPSで起動します。

環境変数で上書き可能（詳細は SPEC.md 参照）。
