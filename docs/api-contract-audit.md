# 通知条件簡素化 + 一覧状態再定義 + プレビュー統一（最終実装計画）

## Summary
今回のリファクタは「実装を減らして意味を明確にする」ことが目的です。  
合意済み仕様を固定し、後方互換は持たずに整理します。

- P1: `conditions.enabled` を削除し、条件評価を簡素化
- P2: 一覧APIの状態を `calendar_state` 1本に統一（行動可能な4状態）
- P3: `docs/api.md` / `docs/config.md` を実装に合わせて更新
- P4: プレビューAPI/画面を「メッセージ本文のみ」に統一し、共通モーダル化
- 追加合意: `periodic.days_of_week` も空配列なら制限なし（毎日扱い）

---

## 1) 仕様固定（Decision Complete）

### P1: 条件評価（事前通知）
- `advance.conditions.enabled` は **廃止**。
- `days_of_week`:
  - 空配列: 制限なし
  - 値あり: 一致した曜日のみ通過
- `property_filters`:
  - 空配列: 制限なし
  - 値あり: 全条件一致（AND）で通過

### 追加: periodic の曜日仕様
- `periodic.days_of_week` も同じく:
  - 空配列: 制限なし（毎日）
  - 値あり: 一致した曜日のみ発火

### P2: 一覧状態（`/api/events/upcoming`）
- 状態フィールドは `calendar_state` のみ。
- 値は `disabled | needs_sync | synced | error`。
- 判定は `calendar_sync.enabled` と `sync_records.attempted/synced` で決定。
- 判定順:
  1. `calendar_sync.enabled == false` -> `disabled`
  2. sync_record が無い、または `attempted == false` -> `needs_sync`
  3. `attempted == true && synced == true` -> `synced`
  4. `attempted == true && synced == false` -> `error`

### P4: プレビュー統一
- `POST /api/notifications/preview` は常に `message` のみ返す。
- 表示は共通モーダル1つを作成して、Dashboard/Notificationsで共用。
- プレビュー内容は「メッセージ本文」のみ。

---

## 2) Public API / Interface / Type 変更

### API 変更
- `GET /api/events/upcoming`
  - 削除: `cache_status`（および既存の未使用旧ステータス系）
  - 追加/維持: `calendar_state`（4値）
- `POST /api/notifications/preview`
  - 削除: `payload`
  - 返却: `{ "message": string }` のみ

### Config / JSON 型変更
- `AdvanceConditions` から `enabled` を削除
- `periodic.days_of_week` の空配列意味を「制限なし」に定義

### DB 変更（`sync_records`）
- `attempted` カラムを導入
- `sync_records` 正式定義:
  - `notion_page_id TEXT PRIMARY KEY`
  - `calendar_event_id TEXT NOT NULL DEFAULT ''`
  - `attempted INTEGER NOT NULL DEFAULT 0`
  - `synced INTEGER NOT NULL DEFAULT 0`
- 旧テーブルからの移行時:
  - 既存行は `attempted=1` を付与
  - `synced` は既存値を引き継ぐ

---

## 3) 実装方針（ファイル別）

### Backend

1. `/Users/junpei/working/_testCode/example-web/notion-notifier/internal/config/config.go`
- `AdvanceConditions.Enabled` を削除
- `ValidateConfig` から関連前提を整理
- コメント/構造体タグを更新

2. `/Users/junpei/working/_testCode/example-web/notion-notifier/internal/scheduler/worker.go`
- 共通ヘルパ導入:
  - `matchesDays(days []int, weekday int) bool`（空配列=true）
- `periodicLoop` の曜日判定を共通化
- `matchAdvanceConditions` から `enabled` 分岐を削除
- `property_filters` 判定は既存ANDを維持
- 同期結果書き込み時に `attempted/synced` を適切更新
  - 成功時: attempted=true, synced=true
  - 失敗時: attempted=true, synced=false

3. `/Users/junpei/working/_testCode/example-web/notion-notifier/internal/models/models.go`
- `SyncRecord` に `Attempted bool` 追加

4. `/Users/junpei/working/_testCode/example-web/notion-notifier/internal/db/db.go`
- `sync_records` の作成/移行ロジックを新スキーマへ
- `ListSyncRecords`, `UpsertSyncRecord`, `GetSyncStatusMap` を `attempted` 対応
- 一覧用に `calendar_state` 判定に必要な取得形式へ調整

5. `/Users/junpei/working/_testCode/example-web/notion-notifier/internal/http/api/handler.go`
- `eventResponse` を `calendar_state` ベースに変更
- `/api/events/upcoming` の状態判定を新仕様へ
- `previewResponse` を `message` のみに統一
- `/api/notifications/preview` の manual/advance 両分岐で同一レスポンス形

### Frontend

6. `/Users/junpei/working/_testCode/example-web/notion-notifier/web/src/lib/api.ts`
- `AdvanceConditions` から `enabled` 削除
- `UpcomingEvent` を `calendar_state` のみに変更
- `previewNotification` レスポンス型を `message` のみに

7. `/Users/junpei/working/_testCode/example-web/notion-notifier/web/src/routes/Notifications.svelte`
- `conditions.enabled` 前提を削除（データ構築含む）
- 既存インラインプレビュー表示を共通モーダル呼び出しに変更

8. `/Users/junpei/working/_testCode/example-web/notion-notifier/web/src/routes/Dashboard.svelte`
- 一覧バッジ表示を `calendar_state` に置換
- 手動プレビュー表示を共通モーダル呼び出しに変更

9. `/Users/junpei/working/_testCode/example-web/notion-notifier/web/src/components/PreviewModal.svelte`（新規）
- `open`, `title`, `content` を受けるシンプルモーダル
- Dashboard/Notifications から共通利用

### Docs

10. `/Users/junpei/working/_testCode/example-web/notion-notifier/docs/api.md`
- `/api/events/upcoming` のレスポンスを `calendar_state` に更新
- `/api/notifications/preview` のレスポンスを `message` のみに更新

11. `/Users/junpei/working/_testCode/example-web/notion-notifier/docs/config.md`
- `notifications.advance.conditions.enabled` 記述を削除
- `periodic.days_of_week` 空配列の意味を明記（制限なし）
- `payload_template` 既定値を `{"content":{{json .Message}}}` に修正
- `property_mapping.attendees`, `attendees_enabled` を明記

---

## 4) 共通化（periodic / advance）で実施する範囲
「同じロジックを使っている部分」のみ安全に統一し、機能拡張はしない。

- 統一対象:
  - 曜日判定（空配列=制限なし）
  - 曜日変換（`weekdayToConfig`）の利用経路
- 非対象:
  - periodic に property_filters を追加する拡張（今回はやらない）
  - 通知内容生成ロジックの混在化（責務は現状維持）

---

## 5) テスト計画

### Unit / Integration（Go）
1. `matchAdvanceConditions`
- days空 + filters空 -> true
- days不一致 -> false
- filters一部不一致 -> false
- filters全一致 -> true

2. periodic曜日判定
- `days_of_week=[]` で当日発火条件を満たせば通る
- 指定曜日外なら通らない

3. `sync_records` マイグレーション
- 旧schema -> 新schemaへ変換
- `attempted` が正しく付与される

4. `/api/events/upcoming`
- `disabled/needs_sync/synced/error` が想定通り返る

5. `/api/notifications/preview`
- advance/manual どちらも `message` のみ返却
- `payload` が返らないことを確認

### Frontend checks
- `npm --prefix /Users/junpei/working/_testCode/example-web/notion-notifier/web run check`
- `npm --prefix /Users/junpei/working/_testCode/example-web/notion-notifier/web run build`

### Backend checks
- `go test ./...`

---

## 6) 受け入れ基準（ユーザー目線）
1. 曜日未指定のルールは「毎日扱い」で直感通り動く。  
2. 一覧の状態表示だけで次の行動が決められる。  
3. 事前通知条件から不要概念（enabled）が消え、設定が理解しやすい。  
4. プレビュー体験は画面間で統一され、見える内容はメッセージ本文だけ。  
5. docsと実装の契約不一致が解消される。

---

## Assumptions / Defaults
- 後方互換は不要（API/DB/UI すべて新仕様へ切替）。
- `calendar_state` は唯一の公開状態フィールドとして扱う。
- `attempted=true && synced=false` は「連携エラー」として公開する。
- periodicへの新機能追加（property_filters導入）は今回のスコープ外。
