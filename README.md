# Notion Notifier

Notion Notifier は、**Notion データベース**、**Google Calendar**、および **Webhook サービス** (Discord, Slack, Microsoft Teams など) を連携させる、プロアクティブなセルフホスト自動化ツールです。

定期的な同期状況、通知ルールの設定、カレンダー連携の管理などをブラウザから直接行える、美しく高速な **Svelte SPA ダッシュボード**を内蔵しています。

![Dashboard Preview](https://via.placeholder.com/800x450.png?text=Notion+Notifier+Dashboard)

## 🌟 主な機能

* **Google カレンダーへの一方向/双方向同期:**
  Notion データベースの予定を監視し、自動的に Google カレンダーへ同期します。Notion の `people` プロパティから参加者を抽出し、連携することも可能です。
* **柔軟な Webhook 通知管理:**
  今後の Notion の予定に基づいて、任意の Webhook URL にメッセージを送信します。
  * **Upcoming ルール (直前通知):** 予定開始の「○日前」「○分前」に個別に通知します。
  * **Periodic ルール (サマリー通知):** 特定の曜日・時間に、今後の予定をまとめて通知します。
* **モダンな Web インターフェース:**
  Go のバイナリに直接バンドルされた Svelte SPA です。設定ファイルの直接編集なしに、ブラウザからテンプレートの管理、Webhook のテスト、同期状況の確認が可能です。
* **クロスプラットフォーム OS サービス:**
  Linux (`systemd`), macOS (`launchd`), Windows のネイティブなバックグラウンドサービスとして、コマンド一発でインストール・永続稼働が可能です。

## 🚀 クイックスタート (インストール)

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

## ⚙️ 設定と使い方

このシステムは「UIで完結する運用設定」と「テキストで管理する機密設定」に分かれています。

### 1. 機密設定 (`env.yaml` / 環境変数)
セキュリティ上、APIキーなどの**認証情報やWebhook URLはブラウザのUIからは編集できません**。初回起動時に自動生成される設定ファイルを直接手動で編集してください。

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

## 🛠️ ソースからのビルドと開発

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

## 📚 ドキュメント

技術仕様、機能設計、APIの詳細は以下のドキュメント（主に日本語）を参照してください。
- [Technical Specification & Architecture](docs/specification.md)
- [Feature Details](docs/features.md)
- [API Reference](docs/api.md)

## 📄 ライセンス

MIT License
