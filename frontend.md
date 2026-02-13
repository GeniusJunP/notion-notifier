# Svelte SPA 完全移行 + Go バンドル設計プラン

## 概要
- 既存のGo内蔵HTML/Alpine/HTMX UIを完全削除し、Svelte SPAをGoにバンドルして配信
- 既存APIを活用しつつ、API/Configのドキュメントを明文化して乖離を防止
- 責務分離を徹底し、UI追加がConfig拡張で破綻しない構造へ整理

## 現在の機能一覧（再発明防止のための基準）

### Dashboard
- Notion同期の手動実行
- 次回/最終同期時刻の表示
- 予定プレビュー（同期ステータス付き）
- 手動通知（テンプレート、期間指定、プレビュー）

### 通知設定
- 事前通知（複数ルール、条件フィルタ、曜日）
- 定期通知（複数ルール、曜日、時刻、日数先）
- テンプレートリセット/保存
- リアルタイムプレビュー（API経由）
- スヌーズ設定

### カレンダー連携
- 有効/無効、頻度、同期範囲
- 手動同期
- 同期レコード削除
- 同期ステータス表示

### システム設定
- Notionプロパティマッピング（標準＋カスタム）
- コンテンツ抽出ルール
- Webhook設定（content_type / payload_template）

### 共通
- Basic認証
- 右上ステータス・トースト
- 保存中/未保存の表示
- ログ出力（HTTP/CONFIG/SYNC/CALENDAR/WEBHOOK）

---

## 目標アーキテクチャ

### 1. UI
- Svelte SPA + Vite + TypeScript
- 既存HTMLテンプレートとAlpine/HTMXは削除
- `web/` にフロントエンドを配置し、`web/dist` をGoにバンドル

### 2. バックエンド
- Goは API + Scheduler + DB に集中
- 画面描画はしない（SPAへ完全移譲）
- 静的配信: `web/dist` を embed.FS で配信
- `/` はSPAのindex.htmlを返す（SPAルーティング）

---

## API設計の整理（重要変更点）

- 既存のAPIは維持しつつ、SPAに必要なGET APIを追加

| メソッド | エンドポイント | 説明 |
|----------|----------------|------|
| GET      | /api/config    | 現在の設定を取得 |
| PUT      | /api/config    | 設定全体を保存（正規化後のConfigを返す） |
| GET      | /api/dashboard | ダッシュボード用集計データ |
| GET      | /api/history   | 通知履歴 |
| GET      | /api/events/upcoming | 予定プレビュー用データ |
| POST     | /api/sync      | Notion同期実行 |
| POST     | /api/calendar/sync | カレンダー手動同期 |
| POST     | /api/calendar/clear | 同期記録削除 |
| POST     | /api/history/clear | 履歴削除 |
| POST     | /api/notifications/preview | テンプレートプレビュー |
| POST     | /api/notifications/manual | 手動通知送信 |

### 重要な方針
- APIはJSONのみ（formやhtml返却をやめる）
- バリデーションエラーは 422 + details で統一
- UI側は APIから返った正規化済みConfig を再利用

---

## Config設計の強化

### 目的
- WebUIの機能追加にConfigが耐えられるようにする

### 方針
- Configが唯一の永続ソース
- `schema_version` を追加し、今後の拡張に備える
- `NormalizeConfig` で旧バージョンからの移行を吸収
- APIは必ず正規化済みConfigを返す

### Docs
- `config.md` に以下を明記:
    - フィールド一覧
    - 型/必須条件
    - デフォルト値
    - 正規化・移行ルール

---

## 責務分離（厳守）

- `internal/http/api` → APIハンドラのみ
- `internal/http/static` → SPA配信
- `internal/app` → 依存の束ねだけ（起動・DI）
- `internal/config` → 設定/正規化/バリデーション
- `internal/scheduler` → 同期・通知ロジック
- `internal/models` / `internal/db` → データ層

---

## SPAバンドル手順

1. `web/` に Svelte SPA を配置
2. `npm run build` → `web/dist`
3. Go側で `//go:embed web/dist/**`
4. `/` とそれ以外の非 `/api` は `index.html` を返す

---

## ドキュメント作成（乖離防止）

- `api.md`
    - エンドポイント一覧
    - リクエスト/レスポンス例
    - エラー形式
- `config.md`
    - Config仕様・正規化ルール・拡張方針
- `features.md`
    - 現在の機能一覧（再発明防止用）

---

## テスト/検証

- `go build ./...`
- SPA build: `npm run build`
- APIの代表ケース:
    - Config保存（正常/異常）
    - プレビュー
    - Notion同期
    - カレンダー同期

---

## 明示的な前提・決定事項

- Svelte SPAは Vite + TypeScript
- SPAは `/` に配置（既存UI削除）
- API/Configドキュメントは `docs/` に分離

---

## 残る確認事項（実装前に必要なら）

- Basic認証をSPA静的配信にも適用するか（現状維持が基本）- Basicは適用。
- SPAのデザインは現行UIを踏襲するか - 実装が軽ければそれでいいが、現状のデザインはhtmx由来なので、技術負債が残る可能性あり。
- APIやconfigの変更 - frontの機能追加などに柔軟に対応できるような土台、再利用可能な設計を心がける。
- フロントエンドのデザインはコンポーネント化を意識し、再利用性を高め、バラバラなスタイル、スタイルの再定義などを避ける。
