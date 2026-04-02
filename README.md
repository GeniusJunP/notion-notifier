# Notion Notifier

Notion Notifier は、**Notion データベース**、**Google Calendar**、および **Webhook サービス** (Discord, Slack, Microsoft Teams など) を連携させる、プロアクティブなセルフホスト自動化ツールです。

定期的な同期状況、通知ルールの設定、カレンダー連携の管理などをブラウザから直接行える、美しく高速な **Svelte SPA ダッシュボード**を内蔵しています。

## 主な機能

1. **Notion データベース連携 (Property Mapping & Filtering)**
   単なるタイトルや日付の取得にとどまりません。設定の `property_mapping` や `content_rules` に基づき、Notion の任意のプロパティを変数として抽出し、通知テンプレートやカレンダー連携で利用できます。また、特定のプロパティ条件（「ステータスが進行中」など）に一致する予定だけを対象にするフィルタリングも可能です。

2. **複数 Webhook への柔軟なルーティング (Multi-Webhook Routing)**
   標準の通知先（例: 全体チャンネル）とは別に、エラー報告や管理者向けの通知（Internal Notification）を別の Webhook URL にルーティングできます。通知の内容に応じて送り先を制御可能です。

3. **Go テンプレートエンジンによる完全なメッセージカスタマイズ**
   Webhook のペイロードや通知メッセージは、Go の標準 `text/template` 構文を使用して完全にカスタマイズできます。`{{.Title}}` や `{{.CustomProperty}}` などを組み合わせて、送信先プラットフォーム（Discord, Slack, Teams）に最適なリッチメッセージを構築できます。

4. **Google カレンダーへの一方向の一括同期 (Uni-directional Sync)**
   Notion の予定リストを監視し、Google カレンダー側に自動で反映（作成・更新・削除）します。機能は完全に一方向ですが、連携時に Notion の `people` プロパティから参加者のメールアドレスを抽出し、Google カレンダーのイベントの「ゲスト (Attendees)」として自動追加する構成オプション (`attendees_enabled`) を備えています。

5. **セルフアップデート対応の OS バックグラウンドサービス**
   Linux (`systemd`), macOS (`launchd`), Windows のネイティブサービスとして登録可能（ユーザーレベルでの安全な実行に対応）。さらに `./notion-notifier update` コマンドで、GitHub から最新のバイナリを取得して自動でアップデートし、再起動する仕組みを備えています。

6. **ビルドイン Web ダッシュボード (Svelte SPA)**
   設定ファイル（`config.yaml`）を直接編集するだけでなく、ポート 18080 で起動するモダンな管理画面から、同期状況の確認、通知ルールの無効化、テンプレートのテストプレビュー、スヌーズ（通知の一時停止）を直感的に行えます。

## インストール

事前ビルドされたシングルバイナリをダウンロードして配置するだけで動作します。

1. [Releases ページ](https://github.com/GeniusJunP/notion-notifier/releases) から最新の OS/アーキテクチャに合った圧縮ファイルをダウンロードし、解凍します。
2. ターミナル（またはコマンドプロンプト/PowerShell）で、解凍した `notion-notifier` 実行ファイルがある場所に移動します。

### OSサービスとしてインストール・起動

以下のコマンドを実行すると、適切な場所に設定ファイルが生成され、バックグラウンドサービスとして自動起動・自動再起動（PC起動時含む）するようになります。

```bash
# インストールとOSサービスへの登録
./notion-notifier install
# サービスの開始
./notion-notifier start

# ※ サービスを停止する場合は stop、登録解除は uninstall
```

*(Linux環境での留意点: システム全体ではなく、現在ログインしているユーザーのサービス (`systemd --user`) として登録したい場合は、`./notion-notifier --user install` を使用してください。)*

## 設定と使い方

### 1. 機密設定 (`env.yaml` / 環境変数)
セキュリティ上、APIキーなどの認証情報やWebhook URLはブラウザのUIからは編集できません。初回起動時に自動生成される設定ファイルを直接手動で編集してください。

**編集が必要なファイル:**
* **Linux:** `~/.config/notion-notifier/env.yaml`
* **macOS:** `~/Library/Application Support/notion-notifier/env.yaml`
* **Windows:** `%APPDATA%\notion-notifier\env.yaml`

主な設定項目:
- Notion API キー / データベース ID
- Discord / Slack などの Webhook URL
- Google Calendar API 認証情報
- 管理画面の Basic 認証 (ID/PW)

設定後はサービスを再起動（`./notion-notifier restart`）して適用します。

### 2. 運用設定 (Web ダッシュボード)
機密設定が完了すると、ブラウザで Web ダッシュボードにアクセスできます：
**http://localhost:18080** (デフォルトポート)

ダッシュボードでは、以下の運用設定(`config.yaml`)をブラウザから直接変更・適用できます：
* **システム設定:** タイムゾーン、同期頻度、Notionプロパティのマッピング、テストモードの切替
* **通知ルール:** Upcoming(直前通知)、Periodic(サマリー通知) の自由な追加・編集・一時停止
* **テンプレート:** Go テンプレート記法による各Webhookメッセージの自由なカスタマイズ
* **スヌーズ設定:** 指定日時までの通知を一時停止

データの実体（イベント履歴や次回同期状態）は、以下のパスにある SQLite データベースに保存されます。
* Linux: `~/.local/share/notion-notifier/data.db`
* Windows: `%LOCALAPPDATA%\notion-notifier\data.db`

## ソースからのビルドと開発

このプロジェクトは Go バックエンドと Svelte フロントエンドで構成されています。

1. リポジトリのクローン
   ```bash
   git clone https://github.com/GeniusJunP/notion-notifier.git
   cd notion-notifier
   ```
2. Svelte フロントエンドのビルド
   ```bash
   cd web
   npm install
   npm run build
   cd ..
   ```
3. Go サーバーの開発実行
   ```bash
   go run cmd/notion-notifier/main.go
   ```

## ドキュメント

技術仕様、機能設計、APIの詳細は以下のドキュメント（主に日本語）を参照してください。
- [Technical Specification & Architecture](docs/specification.md)
- [Feature Details](docs/features.md)
- [API Reference](docs/api.md)

## ライセンス

MIT License
