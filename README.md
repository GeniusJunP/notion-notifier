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

サービスが起動したら、ブラウザで Web ダッシュボードにアクセスしてください：
**http://localhost:18080** (デフォルトポート)

* **デフォルトユーザー名:** `admin`
* **デフォルトパスワード:** `password`

ダッシュボードの **システム設定** タブから、Notion の API キー、Webhook URL などを入力し、安全に保存・適用できます。

### 設定ファイル・データベースの物理パス
デフォルトでは、設定ファイル(`config.yaml`, `env.yaml`) と SQLite データベースファイルは、OS標準のユーザーディレクトリに配置されます。

* **Linux:** `~/.config/notion-notifier/` / データファイルは `~/.local/share/notion-notifier/`
* **macOS:** `~/Library/Application Support/notion-notifier/`
* **Windows:** `%APPDATA%\notion-notifier\` / データファイルは `%LOCALAPPDATA%\notion-notifier\`

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
