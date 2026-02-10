# 機能一覧

## Dashboard
- Notion同期の手動実行 (`POST /api/sync`)
- 次回/最終同期時刻の表示 (`GET /api/dashboard`)
- 予定プレビュー（カレンダー・SQLiteそれぞれの同期ステータス付き）(`GET /api/events/upcoming`)
- 手動通知（テンプレート、周知期間指定、プレビュー）(`POST /api/notifications/manual`, `/preview`)

## 通知設定
- 事前通知（複数ルール、条件フィルタ、曜日）
- 定期通知（複数ルール対応、条件フィルタ、曜日、時刻、通知の日数範囲）
- 各通知のメッセージテンプレートのデフォルト値取得 (`GET /api/templates/defaults`)
- メッセージのプレビュー（API経由）
- スヌーズ／ミュート設定

## カレンダー連携
- 有効/無効、頻度、同期範囲
- 手動同期 (`POST /api/calendar/sync`)
- 同期レコード削除 (`POST /api/calendar/clear`)
- 同期ステータス表示
- Calendar API読み込みによる双方向比較同期
  - 予定の作成/更新/削除反映
  - `extendedProperties.private.notion_page_id` による重複防止・同期済み予定の識別
  - Notion側で削除された予定のCalendar側自動削除
  - 孤立sync_recordsの自動クリーンアップ
- 招待ユーザーのメールアドレスをNotionプロパティのユーザー（people）欄から抽出
  - `property_mapping.attendees` にNotionプロパティ名を設定
  - Calendar予定のattendeesに自動追加

## システム設定
- Notionプロパティマッピング（標準＋カスタム＋attendees）
- コンテンツ抽出ルール
- Webhook設定（content_type / payload_template）

## 共通
- Basic認証（SPA静的配信にも適用）
- ステータス・トースト（フロントで実装）
- 保存中/未保存の表示（フロントで実装）
- ログ出力（HTTP/CONFIG/SYNC/CALENDAR/WEBHOOK）
