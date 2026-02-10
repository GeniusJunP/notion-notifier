# 機能一覧（再発明防止用）

## Dashboard
- Notion同期の手動実行 (`POST /api/sync`)
- 次回/最終同期時刻の表示 (`GET /api/dashboard`)
- 予定プレビュー（同期ステータス付き）(`GET /api/events/upcoming`)
- 手動通知（テンプレート、期間指定、プレビュー）(`POST /api/notifications/manual`, `/preview`)

## 通知設定
- 事前通知（複数ルール、条件フィルタ、曜日）
- 定期通知（複数ルール、曜日、時刻、日数先）
  - TODO:複数の定期通知に対応
- テンプレートリセット/保存
- リアルタイムプレビュー（API経由）
- スヌーズ／ミュート設定

## カレンダー連携
- 有効/無効、頻度、同期範囲
- 手動同期 (`POST /api/calendar/sync`)
- 同期レコード削除 (`POST /api/calendar/clear`)
- 同期ステータス表示

## システム設定
- Notionプロパティマッピング（標準＋カスタム）
- コンテンツ抽出ルール
- Webhook設定（content_type / payload_template）

## 共通
- Basic認証（SPA静的配信にも適用）
- ステータス・トースト（フロントで実装）
- 保存中/未保存の表示（フロントで実装）
- ログ出力（HTTP/CONFIG/SYNC/CALENDAR/WEBHOOK）
