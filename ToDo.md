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
