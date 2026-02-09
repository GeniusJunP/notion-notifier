# フロントエンド要件書（Modern Dashboard Architecture）

## 1. コンセプト & デザイン方針
Notion予定通知ツールの管理画面を、**「直感的な操作性」** と **「洗練された美しさ」** を兼ね備えたモダンなダッシュボードとして再構築する。

- **Tech Stack**: 
  - **Go HTML Templates**: サーバーサイドレンダリング (SEO不要だが高速表示のため)
  - **Tailwind CSS**: ユーティリティファーストなスタイリング
  - **Alpine.js**: 軽量なクライアントサイドのインタラクション（モーダル、ドロップダウン、ステート管理）
  - **htmx**: サーバーとの非同期通信・部分HTML更新（SPA風の挙動を実現）
  - **Lucide Icons**: シンプルで美しいSVGアイコン

- **デザインテーマ**:
  - **Clean & Airy**: 十分な余白（Whitespace）とクリアなタイポグラフィ
  - **Card-Based Layout**: 情報のまとまりをカードで表現
  - **Feedback-Oriented**: 操作に対する即座のフィードバック（Toast通知、ローディング状態）
- **Micro-Interactions**: ホバー効果やトランジションによる心地よい操作感
  - **Basic認証設定**: UIでは扱わず、`config.yaml` で設定

## 2. 画面構成 & ナビゲーション

### レイアウト構造
- **Sidebar (PC) / Drawer (Mobile)**:
  - ロゴエリア
  - メインナビゲーション（アイコン付き）
  - システムステータス（簡易表示）
- **Main Content Area**:
  - ページヘッダー（タイトル + 主要アクション）
  - パンくずリスト（必要に応じて）
  - コンテンツ本体
- **Toast Container**:
  - 画面右上または右下にフローティング表示

### サイトマップ
1. **ダッシュボード (`/`)**
   - **Metrics Card**: 本日の予定件数、次の同期までの時間、エラー発生状況
   - **Notification Status**: ミュート/スヌーズ状態の表示
   - **Manual Notification**: テンプレート選択と手動送信
   - **Recent Activity**: 直近の通知履歴タイムライン
   - **Upcoming**: 直近の予定リスト（カード形式）
   - **Quick Actions**: サイドバー内に配置（手動同期/ミュート/履歴クリア）

2. **通知設定 (`/notifications`)**
   - **Rules Engine UI**: 
     - 複雑な設定をアコーディオンまたはタブで整理
     - **Alpine.js** を利用した「ルールの動的追加/削除」UI（フォームの配列操作）
   - **Template Editor**: リアルタイムプレビュー機能付きのエディタ

3. **カレンダー連携 (`/calendar`)**
   - NotionデータベースとGoogleカレンダーの同期状態可視化
   - 同期設定フォーム（スイッチUI、スライダー等）

4. **システム設定 (`/settings`)**
   - プロパティマッピング設定
   - セキュリティ（Basic認証）
   - コンテンツ抽出ルール（正規表現テスター付き）

## 3. Alpine.js & htmx の役割分担

### Alpine.js (Client-side interactivity)
- **UI State**: モーダルの開閉、ドロップダウンメニュー、サイドバーのトグル
- **Form Handling**: 
  - 複数入力フィールドの動的追加（例：通知ルールの追加）
  - バリデーションエラーのリアルタイム表示
  - スイッチ（Toggle）コンポーネントの制御
- **Feedback**: Toast通知の表示アニメーション、自動消去タイマー

### htmx (Server interaction)
- **Data Fetching**: 検索、フィルタリング、ページネーション
- **Form Submission**: 
  - 設定の保存（`hx-post`）
  - 保存後のトースト表示（`hx-trigger`）
  - コンテンツの部分リロード（`hx-target`）
- **Polling**: バックグラウンドジョブの進捗表示（同期処理中など）

## 4. UIコンポーネント要件

### モダンコンポーネント
- **Cards**: 柔らかなシャドウと角丸（`rounded-xl`, `shadow-sm`）
- **Stats**: 大きな数字とトレンドインジケータ
- **Data Tables**: 
  - ストライプなし、下線のみのクリーンなデザイン
  - 行ホバーアクション
  - ステータスバッジ（Pill shape）
- **Forms**:
  - フローティングラベルまたは明確なラベル階層
  - インプットグループの明確な区分け
  - スイッチ（Checkboxのスタイル化）によるON/OFF
- **Modals**: 背景ぼかし（Backdrop blur）効果

## 5. 実装フェーズ（推奨手順）

1. **Base Layout**: Goテンプレートでのレイアウト枠（Sidebar + Main）作成とTailwindセットアップ
2. **Alpine Components**: モーダル、トースト、ドロップダウンのベース実装
3. **Core Pages**: ダッシュボードと設定画面のマークアップ
4. **Interactivity**: htmxによるフォーム送信とAlpineによる動的UIの統合

---
*備考: この構成はシングルバイナリでの配布（Goの`embed`機能利用）とも相性が良く、ビルドプロセスもシンプルに保てます。*

---
## 未実装（2026-02-09時点）
以下は現状まだ未実装であり、UIにはプレースホルダまたは静的表示が含まれます。

- なし（現時点で未実装項目なし）
