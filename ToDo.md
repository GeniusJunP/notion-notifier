# ToDo (KISS Refactor)

## Goal
- Remove complexity and keep behavior explicit.
- Avoid "accepted input gets ignored/overwritten" patterns.
- Keep current `is_test` routing behavior intact:
  - `false`: `webhook.notification` + `webhook.notification_url`
  - `true`: `webhook.internal_notification` + `webhook.internal_notification_url`

## Tasks
- [x] Refactor scheduler public methods to stop discarding caller `context.Context`.
- [x] Remove duplicated runtime-op wrappers (`withRuntimeErrOp` / `withRuntimeIntOp`) and unify execution path.
- [x] Eliminate unnecessary closures/temporary variables in runtime-wrapper call sites.
- [x] Keep line count flat or reduced in `internal/scheduler/worker.go`.
- [x] Ensure failed history is recorded when selected webhook URL is empty.
- [x] Keep snooze behavior unchanged (`is_test=true` bypasses snooze).
- [x] Re-run backend tests for touched packages.
- [x] Re-run frontend type check.
- [x] Update docs/UI copy only if needed to stay consistent with behavior.

## Verification Commands
- [x] `go test ./internal/scheduler ./internal/http/api ./internal/config`
- [x] `npm run check`

## Global Audit Backlog (KISS)
- [x] Remove double normalize/validate path in config update flow (`api.putConfig` + `config.Manager.UpdateConfig`).
- [x] Simplify config manager read API to avoid ambiguous tuple usage (`cfg, _ := m.Get()` pattern).
- [x] Consolidate Calendar client initialization responsibility (currently split between app bootstrap and scheduler runtime sync path).
- [x] Reduce silent error drops in API read endpoints (`dashboard`, `upcoming`, `history`) with explicit fallback/error policy.
- [x] Split large backend files by responsibility (`internal/scheduler/worker.go`, `internal/http/api/handler.go`, `internal/db/db.go`).
- [x] Split large frontend route components (`web/src/routes/Settings.svelte`, `web/src/App.svelte`) into focused sections/components.

## Global Function Inventory Audit (2026-02-17)
- [x] Capture global directory tree snapshot (`ls -R cmd internal web/src scripts docs`).
- [x] Generate all-function inventory file (`docs/function-inventory.md`).
- [x] Count functions by language and prod/test scope.
- [x] Classify every production function by responsibility boundary (App/API/Scheduler/DB/Config/UI).
- [x] Generate all-function-level Mermaid graph (`docs/function-graph-detailed.md`).
- [x] Detect duplicated write paths (config updates, history writes, sync triggers).
- [x] Detect config-change paths that require scheduler reload/rebuild and unify the hook point.
- [x] Detect semantic drift between preview/send behavior across routes.
- [x] Produce remediation map (`keep/merge/delete`) with expected line reduction per module.
- [x] Publish responsibility map with Mermaid graph (`docs/function-responsibility-map.md`).

## Refactor Execution Backlog (from Global Audit)
- [ ] Unify backend config-change hook (`UpdateConfig` success path + scheduler rebuild/reload policy in one place).
- [ ] Align periodic preview with periodic send semantics (`days_ahead` must drive preview query range).
- [ ] Separate manual notification concerns (template persistence and webhook send) to avoid mixed responsibilities.
- [ ] Deduplicate frontend `saveConfig` flow across `App/Calendar/Notifications/Settings`.
- [ ] Split `internal/scheduler/worker.go` into domain-focused files while preserving behavior.
