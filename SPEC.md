# notion-notifier 仕様書

## 概要

Notionデータベース上の予定を任意Webhookに通知し、Googleカレンダーと同期する個人用ツール。

## 動作環境

- **OS**: Linux（Asahi Linux/Ubuntu）
- **アクセス**: LAN（localhost）
- **データベース**: SQLite（ファイルベース、Pure Go ドライバ）
- **配布形態**: 単一バイナリ（CGO不要）

## アーキテクチャ

```
+------------------+     +------------------+
|    Notion API    |     | Google Calendar  |
+--------+---------+     +--------+---------+
         |                        |
         v                        v
+------------------------------------------+
|          notion-notifier (Go)            |
|                                          |
|  +-----------+  +-----------+  +-------+ |
|  | Scheduler |  | Web UI    |  |  DB   | |
|  |           |  | (SSR)     |  |       | |
|  +-----------+  +-----------+  +-------+ |
+--------+---------+--------+--------------+
         |                  |
         v                  v
+--------+---------+  +-----+--------+
|    Webhook      |  |  Browser     |
| (Discord等)     |  | (SSR HTML)   |
+-----------------+  +--------------+
```

## 設定方式

**重要**: API認証情報（Notion API Key、Webhook URL、Google credentialsなど）はWebUIで編集禁止。`env.yaml`ファイルを直接編集して設定する。設定値は `config.yaml` で管理し、認証情報は環境変数でも上書き可能。

```
config.yaml.sample  # 設定ファイルのテンプレート（git管理対象）
env.yaml.sample     # 認証情報テンプレート（git管理対象）
config.yaml         # 設定ファイル（git管理対象外）
env.yaml            # 認証情報（git管理対象外）
data.db             # SQLiteデータ（git管理対象外）
```

### 環境変数による上書き

env.yaml の認証情報は以下の環境変数で上書きできる：

| 環境変数                       | 対応設定                       |
| ------------------------------ | ------------------------------ |
| `NOTION_API_KEY`               | `notion.api_key`               |
| `NOTION_DATABASE_ID`           | `notion.database_id`           |
| `SCHEDULE_WEBHOOK_URL`         | `webhook.schedule_url`         |
| `NOTIFICATION_WEBHOOK_URL`     | `webhook.notification_url`     |
| `GOOGLE_CALENDAR_ID`           | `google.calendar_id`           |
| `GOOGLE_SERVICE_ACCOUNT_KEY`   | `google.service_account_key`   |
| `BASIC_AUTH_USERNAME`          | `security.basic_auth.username` |
| `BASIC_AUTH_PASSWORD`          | `security.basic_auth.password` |

### 設定編集について

| 項目                                     | 直接編集 | WebUI編集 |
| ---------------------------------------- | -------- | --------- |
| **API認証情報**（Notion/Webhook/Google/Basic認証） | ○        | ×         |
| **通知テンプレート**                     | ○        | ○         |
| **スケジュール設定**（間隔など）         | ○        | ○         |
| **プロパティマッピング**                 | ○        | ○         |
| **コンテンツ抽出ルール**                 | ○        | ○         |
| **スヌーズ/ミュート**                    | ○        | ○         |

## 機能一覧

### 1. データ同期

| 機能             | 説明                                         |
| ---------------- | -------------------------------------------- |
| Notionデータ取得 | データベースから予定を取得、SQLiteに保存     |
| 自動同期         | 起動時・指定間隔（デフォルト15分）で自動取得 |
| 手動同期         | ダッシュボードから手動実行                   |
| 予定一覧         | 本日〜2週間先の予定を一覧表示                |

### 2. 通知機能

#### 2.1 事前通知

| 設定項目   | 説明                                         |
| ---------- | -------------------------------------------- |
| 有効/無効  | 個別にON/OFF                                 |
| 分前       | 5/10/15/30/45/60/120/180/360/720分前から選択 |
| メッセージ | テンプレート使用（Go text/template構文）      |
| 発火条件   | 曜日指定、プロパティフィルタ                 |

**精度**: 事前通知はデータ同期時に次回発火時刻を計算し、`time.AfterFunc` で正確にスケジューリングする。チェック間隔（デフォルト15分）に依存しない。

#### 2.2 定期通知

| 設定項目   | 説明                 |
| ---------- | -------------------- |
| 有効/無効  | ON/OFF               |
| 曜日       | 複数選択可（月〜日） |
| 時刻       | HH:mm形式            |
| 範囲       | 1/3/7/14/30日先まで  |
| メッセージ | テンプレート使用     |

#### 2.3 手動通知

| 設定項目     | 説明                                          |
| ------------ | --------------------------------------------- |
| 日付         | カレンダー選択                                |
| テンプレート | 定期テンプレート/カスタムから選択             |
| プレビュー   | 送信前にWebhookペイロードをプレビュー         |
| 送信         | Webhookに即座に通知                           |

#### 2.4 通知抑制

| 機能     | 説明                                                           |
| -------- | -------------------------------------------------------------- |
| スヌーズ | 今後の予定通知のみ停止（1週間/2週間/1ヶ月/3ヶ月/カスタム日付） |
| ミュート | 全通知を停止（休暇用）                                         |

### 3. Google Calendar連携

| 機能         | 説明                   |
| ------------ | ---------------------- |
| 有効/無効    | ON/OFF                 |
| 自動同期間隔 | 1/3/6/12/24時間        |
| 同期対象日数 | 7/14/30日など           |
| 手動同期     | 日付範囲指定           |
| 状態表示     | 最終同期時刻、同期件数 |
| 履歴クリア   | 同期レコード削除       |

#### 同期トラッキング

Google CalendarのイベントにNotionページIDを埋め込み、逆同期を可能にする：

- **description**: `Notion: https://notion.so/{page_id}`
- **extendedProperties.private.notion_page_id**: `{page_id}`

これにより：

- CalendarイベントからNotionページへの逆引きが可能
- UIに「Notionで開く」リンクを表示可能

### 4. 通知履歴

| 機能       | 説明                        |
| ---------- | --------------------------- |
| 表示       | 直近50件                    |
| ステータス | 成功/失敗、時刻、メッセージ |
| クリア     | 手動で履歴削除              |

## 設定ファイル

### config.yaml

```yaml
# タイムゾーン
timezone: "Asia/Tokyo"

# 基本設定
sync:
  check_interval: 15 # データ同期間隔（分）

# 事前通知（複数設定可）
notifications:
  advance:
    - enabled: true
      minutes_before: 30
      message: "⏰ {{.MinutesBefore}}分後に「{{.Name}}」が始まります"
      location: "{{.Location}}"
      url: "{{.URL}}"
      conditions:
        enabled: false
        days_of_week: []
        property_filters: []

  # 定期通知（複数設定可）
  periodic:
    - enabled: true
      days_of_week: [1, 4] # 月曜、木曜
      time: "09:00"
      days_ahead: 7
      message: |
        @everyone 今週の予定をお届けします！

        {{range .Events}}
        📅 **{{.Name}}** ({{.Date}})
        🕐 {{.Time}}
        📍 {{.Location}}

        {{end}}

# Webhook設定（任意JSON）
webhook:
  schedule:
    content_type: "application/json"
    payload_template: "{\"content\":\"{{.Message}}\"}"
  notification:
    content_type: "application/json"
    payload_template: "{\"content\":\"{{.Message}}\"}"

# Google Calendar連携
calendar_sync:
  enabled: false
  interval_hours: 6
  lookahead_days: 30

# Notionプロパティマッピング
property_mapping:
  title: "名前"
  date: "日付"
  location: "場所"
  custom:
    - variable: "speaker"
      property: "登壇者"
    - variable: "category"
      property: "カテゴリ"

# コンテンツ抽出ルール
content_rules:
  start_heading: "活動内容"
  include_start_heading: false
  stop_at_next_heading: true
  stop_at_delimiter: true

# コンテンツ抽出ルールの挙動
# - start_heading と一致する見出し（h1〜h3）を起点に本文を抽出
# - stop_at_next_heading が true の場合、次の見出しで終了
# - stop_at_delimiter が true の場合、Notionの区切り線（dividerブロック）で終了
# - 抽出結果は {{.Content}} で参照可能

# 通知抑制
snooze_until: "" # ISO8601日時、空白=無効
mute_until: "" # ISO8601日時、空白=無効

# セキュリティ（オプション）
security:
  basic_auth:
    enabled: false
    # 認証情報は env.yaml または環境変数で指定
```

### env.yaml

```yaml
# API認証情報（WebUIでは編集不可、環境変数で上書き可）
notion:
  api_key: ""
  database_id: ""

webhook:
  schedule_url: ""
  notification_url: ""

google:
  calendar_id: ""
  service_account_key: ""
```

## テンプレート

### テンプレートエンジン

通知メッセージは **Go の `text/template`** を使用し、Web UIは **Go の `html/template`** を使用する。

- 変数参照: `{{.Name}}`, `{{.Date}}`
- ループ: `{{range .Events}}...{{end}}`
- 条件分岐: `{{if .Location}}📍 {{.Location}}{{end}}`

### 標準変数

| 変数                 | 説明                             |
| -------------------- | -------------------------------- |
| `{{.Name}}`          | 予定名（タイトル）               |
| `{{.Date}}`          | 日付（YYYY-MM-DD）               |
| `{{.Time}}`          | 開始時刻（HH:mm）               |
| `{{.EndTime}}`       | 終了時刻（HH:mm、空の場合あり） |
| `{{.IsAllDay}}`      | 終日イベントかどうか             |
| `{{.Location}}`      | 場所                             |
| `{{.URL}}`           | NotionページURL                  |
| `{{.Content}}`       | 抽出済み本文（コンテンツ抽出ルール適用） |
| `{{.MinutesBefore}}` | 事前通知の分数                   |

### Webhookペイロード用変数

Webhookの `payload_template` では以下の変数を使用できる：

| 変数                   | 説明                                   |
| ---------------------- | -------------------------------------- |
| `{{.Message}}`         | 生成済み通知メッセージ本文             |
| `{{.Type}}`            | 通知種別（advance/periodic/manual） |
| `{{.Events}}`          | 通知対象のイベント配列                 |
| `{{.Event}}`           | 先頭イベント（単一通知の簡易参照）     |
| `{{.MinutesBefore}}`   | 事前通知の分数（advanceのみ）          |

### カスタム変数

プロパティマッピングの `custom` で定義した変数は `{{.Custom.変数名}}` でアクセス：

| 変数                   | 説明                                   |
| ---------------------- | -------------------------------------- |
| `{{.Custom.speaker}}`  | 登壇者（プロパティマッピングで設定）   |
| `{{.Custom.category}}` | カテゴリ（プロパティマッピングで設定） |

### テンプレート例

```
# 定期通知
@everyone 次の予定一覧！

{{range .Events}}
- **{{.Name}}** ({{.Date}} {{.Time}}){{if .Location}} 📍{{.Location}}{{end}}
{{end}}
```

## Web UI

### フロントエンド技術スタック

| 技術          | 用途                                       |
| ------------- | ------------------------------------------ |
| **SSR**       | `html/template` でサーバー側HTML生成        |
| **Pico CSS**  | 軽量CSSフレームワーク（クラスレス基本スタイル） |
| `style.css`   | カスタムスタイルの上書き                   |
| `app.js`      | 追加の小さなUI補助（任意、フレームワーク不要） |

### 静的アセット配信

Go の `embed` パッケージで全静的アセット（HTML, CSS, JS）をバイナリに埋め込む。外部ファイル配置不要。

```go
//go:embed web/templates/* web/static/*
var webAssets embed.FS
```

### ページ構成

| タブ           | 機能                                              |
| -------------- | ------------------------------------------------- |
| ダッシュボード | 予定一覧、同期ボタン、スヌーズ/ミュート、手動通知 |
| 通知設定       | 事前通知/定期通知の設定                          |
| カレンダー連携 | Google Calendar連携設定                           |
| プロパティ     | Notionカラムとテンプレートのマッピング            |
| 履歴           | 通知履歴表示・クリア                              |

### 画面遷移と操作

Web UIはページ遷移とフォーム送信で操作する。HTMLは通常の `html/template` 記述でよく、必要に応じて `app.js` で最小限のUI補助（入力の表示切替など）を行う。

### 認証

**Basic認証（オプション）**: `config.yaml`で有効化。Web UIに適用。

## 技術仕様

### データベース（SQLite）

```sql
-- 予定テーブル
CREATE TABLE events (
    notion_page_id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    start_date TEXT NOT NULL,       -- YYYY-MM-DD
    start_time TEXT,                -- HH:mm（終日イベントはNULL）
    end_date TEXT,                  -- YYYY-MM-DD（NULLの場合 = start_date）
    end_time TEXT,                  -- HH:mm
    is_all_day INTEGER DEFAULT 0,  -- 1=終日
    location TEXT,
    url TEXT,
    content TEXT,
    raw_properties TEXT,           -- JSONで保存
    fetched_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- 通知履歴テーブル
CREATE TABLE notification_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,        -- advance/periodic/manual
    status TEXT NOT NULL,      -- success/failed
    message TEXT,
    notion_page_id TEXT,
    error TEXT,
    sent_at TEXT NOT NULL
);

-- 同期レコードテーブル
CREATE TABLE sync_records (
    notion_page_id TEXT PRIMARY KEY,
    calendar_event_id TEXT,
    notion_updated_at TEXT,     -- Notion側の最終更新日時
    calendar_updated_at TEXT,   -- Calendar側の最終更新日時
    last_synced_at TEXT,
    sync_status TEXT DEFAULT 'synced'  -- synced/pending/error
);

-- 事前通知スケジュール（精密スケジューリング用）
CREATE TABLE advance_schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    notion_page_id TEXT NOT NULL,
    rule_index INTEGER NOT NULL,   -- config内の事前通知ルールのインデックス
    fire_at TEXT NOT NULL,         -- 発火予定時刻（ISO8601）
    fired INTEGER DEFAULT 0,      -- 1=発火済み
    UNIQUE(notion_page_id, rule_index)
);
```

### ディレクトリ構成

```
notion-notifier/
├── main.go                        # エントリーポイント、embed宣言
├── internal/
│   ├── config/
│   │   └── config.go              # YAML読み込み、環境変数上書き
│   ├── models/
│   │   └── models.go              # 共通データ構造
│   ├── db/
│   │   └── db.go                  # SQLite接続、自動マイグレーション
│   ├── notion/
│   │   └── client.go              # Notion API クライアント
│   ├── webhook/
│   │   └── client.go              # Webhook 送信
│   ├── calendar/
│   │   └── google.go              # Google Calendar同期
│   ├── scheduler/
│   │   └── worker.go              # 定期実行、事前通知スケジューリング
│   ├── template/
│   │   └── renderer.go            # 通知テンプレートのレンダリング
│   └── server/
│       ├── server.go              # HTTPサーバー、ルーティング
│       ├── middleware.go          # Basic認証、Content-Type判定
│       ├── dashboard.go
│       ├── settings.go
│       ├── calendar_handler.go
│       └── history.go
├── web/
│   ├── templates/
│   │   ├── layout.html
│   │   ├── dashboard.html
│   │   ├── notifications.html
│   │   ├── calendar.html
│   │   ├── property.html
│   │   └── history.html
│   └── static/
│       ├── pico.min.css           # Pico CSS
│       ├── style.css              # カスタムスタイル
│       └── app.js                 # 最小限のUI補助（任意）
├── config.yaml.sample
├── .gitignore
└── README.md
```

### 依存ライブラリ

```go
import (
    // HTTPルーティング
    "github.com/go-chi/chi/v5"

    // テンプレート
    "html/template"  // Web UI用
    "text/template"  // 通知メッセージ用

    // SQLite（Pure Go、CGO不要）
    "modernc.org/sqlite"
    "database/sql"

    // YAML設定
    "gopkg.in/yaml.v3"

    // 静的アセット埋め込み
    "embed"

    // 標準ライブラリ
    "time"
    "encoding/json"
    "net/http"
)
```

### エラーハンドリングとリトライ

外部APIとの通信には統一的なリトライ戦略を適用する。

| API             | レートリミット            | リトライ戦略                             |
| --------------- | ------------------------- | ---------------------------------------- |
| Notion API      | 3 req/sec                 | 指数バックオフ（最大3回、429で待機）     |
| Webhook（Discord互換の場合） | 30 req/60sec per webhook  | 指数バックオフ（最大3回）                |
| Google Calendar | 制限あり                  | 指数バックオフ（最大3回）                |

```go
// リトライの基本実装イメージ
type RetryConfig struct {
    MaxRetries int
    BaseDelay  time.Duration // 1秒
    MaxDelay   time.Duration // 30秒
}
```

- **429 Too Many Requests**: `Retry-After` ヘッダーを尊重して待機
- **5xx エラー**: リトライ対象
- **4xx エラー（429以外）**: リトライしない（設定ミスの可能性）
- **通知失敗**: `notification_history` に `failed` として記録し、次回スケジュールには影響しない

## 運用

### インストール

```bash
# バイナリをコピー
cp notion-notifier /usr/local/bin/

# 設定ファイル作成
mkdir -p /etc/notion-notifier
cp config.yaml.sample /etc/notion-notifier/config.yaml
cp env.yaml.sample /etc/notion-notifier/env.yaml
# env.yamlを編集してAPI認証情報を設定

# systemdサービス
cp etc/notion-notifier.service /etc/systemd/system/
systemctl enable --now notion-notifier
```

### 設定

```bash
# 設定ファイルを編集
nano /etc/notion-notifier/config.yaml
nano /etc/notion-notifier/env.yaml

# または環境変数で認証情報を設定（systemd override推奨）
systemctl edit notion-notifier
# [Service]
# Environment="NOTION_API_KEY=secret_xxx"

# サービス再起動
systemctl restart notion-notifier
```

### 更新

```bash
systemctl stop notion-notifier
cp new-binary /usr/local/bin/notion-notifier
systemctl start notion-notifier
```

## セキュリティ

- API認証情報は `env.yaml`（git管理対象外）または環境変数で管理
- `config.yaml.sample` と `env.yaml.sample` のみgit管理（認証情報は空欄）
- `data.db` は `.gitignore` で除外
- WebUIからの認証情報編集不可
- 必要に応じてBasic認証を有効化
- LANアクセス（localhost）のみ
