# Function-Level Detailed Graph (All Functions)

Generated: 2026-02-17 JST

## Coverage
- total functions: 257
- go functions (including tests): 209
- frontend functions (ts+svelte): 47
- script functions: 1
- edge rule: static same-group call references (cross-group links are shown in the integrated graph).

## Group Index
| Group | Functions | Edges |
|---|---:|---:|
| `fe:web/src/App.svelte` | 10 | 4 |
| `fe:web/src/components/PreviewModal.svelte` | 3 | 1 |
| `fe:web/src/components/TemplateGuideSidebar.svelte` | 1 | 0 |
| `fe:web/src/lib/api.ts` | 3 | 0 |
| `fe:web/src/lib/store.ts` | 5 | 1 |
| `fe:web/src/routes/Calendar.svelte` | 3 | 0 |
| `fe:web/src/routes/Dashboard.svelte` | 7 | 3 |
| `fe:web/src/routes/History.svelte` | 3 | 0 |
| `fe:web/src/routes/Notifications.svelte` | 9 | 1 |
| `fe:web/src/routes/Settings.svelte` | 3 | 0 |
| `go:api` | 31 | 50 |
| `go:app` | 6 | 5 |
| `go:calendar` | 15 | 8 |
| `go:config` | 26 | 22 |
| `go:db` | 34 | 29 |
| `go:logging` | 2 | 0 |
| `go:main` | 1 | 0 |
| `go:middleware` | 3 | 1 |
| `go:notion` | 22 | 18 |
| `go:retry` | 5 | 0 |
| `go:scheduler` | 56 | 40 |
| `go:static` | 1 | 0 |
| `go:template` | 5 | 4 |
| `go:webhook` | 2 | 1 |
| `script:scripts/deploy-mac.sh` | 1 | 1 |

## fe:web/src/App.svelte

```mermaid
flowchart TD
  n1["web/src/App.svelte:checkHealth"]
  n2["web/src/App.svelte:clearSnooze"]
  n3["web/src/App.svelte:closeSidebar"]
  n4["web/src/App.svelte:handleGlobalKeydown"]
  n5["web/src/App.svelte:handleSync"]
  n6["web/src/App.svelte:openGuideModal"]
  n7["web/src/App.svelte:openSidebar"]
  n8["web/src/App.svelte:saveSnooze"]
  n9["web/src/App.svelte:toggleDarkMode"]
  n10["web/src/App.svelte:updateClock"]
  n2 --> n8
  n4 --> n3
  n5 --> n1
  n8 --> n1
```

## fe:web/src/components/PreviewModal.svelte

```mermaid
flowchart TD
  n1["web/src/components/PreviewModal.svelte:close"]
  n2["web/src/components/PreviewModal.svelte:handleKeydown"]
  n3["web/src/components/PreviewModal.svelte:renderMarkdown"]
  n2 --> n1
```

## fe:web/src/components/TemplateGuideSidebar.svelte

```mermaid
flowchart TD
  n1["web/src/components/TemplateGuideSidebar.svelte:openGuideDetail"]
```

## fe:web/src/lib/api.ts

```mermaid
flowchart TD
  n1["web/src/lib/api.ts:buildManualNotificationRequest"]
  n2["web/src/lib/api.ts:buildPreviewNotificationRequest"]
  n3["web/src/lib/api.ts:request"]
```

## fe:web/src/lib/store.ts

```mermaid
flowchart TD
  n1["web/src/lib/store.ts:addToast"]
  n2["web/src/lib/store.ts:navigate"]
  n3["web/src/lib/store.ts:saveConfig"]
  n4["web/src/lib/store.ts:setDarkMode"]
  n5["web/src/lib/store.ts:syncNotion"]
  n3 --> n1
```

## fe:web/src/routes/Calendar.svelte

```mermaid
flowchart TD
  n1["web/src/routes/Calendar.svelte:handleClear"]
  n2["web/src/routes/Calendar.svelte:handleConfigUpdate"]
  n3["web/src/routes/Calendar.svelte:handleSync"]
```

## fe:web/src/routes/Dashboard.svelte

```mermaid
flowchart TD
  n1["web/src/routes/Dashboard.svelte:formatEventDateTime"]
  n2["web/src/routes/Dashboard.svelte:handleManualPreview"]
  n3["web/src/routes/Dashboard.svelte:handleManualSend"]
  n4["web/src/routes/Dashboard.svelte:handleSync"]
  n5["web/src/routes/Dashboard.svelte:loadData"]
  n6["web/src/routes/Dashboard.svelte:loadDefaultTemplate"]
  n7["web/src/routes/Dashboard.svelte:openPreview"]
  n2 --> n7
  n3 --> n7
  n4 --> n5
```

## fe:web/src/routes/History.svelte

```mermaid
flowchart TD
  n1["web/src/routes/History.svelte:formatDate"]
  n2["web/src/routes/History.svelte:handleClear"]
  n3["web/src/routes/History.svelte:loadHistory"]
```

## fe:web/src/routes/Notifications.svelte

```mermaid
flowchart TD
  n1["web/src/routes/Notifications.svelte:addAdvanceRule"]
  n2["web/src/routes/Notifications.svelte:addPeriodicRule"]
  n3["web/src/routes/Notifications.svelte:openPreview"]
  n4["web/src/routes/Notifications.svelte:previewTemplate"]
  n5["web/src/routes/Notifications.svelte:removeAdvanceRule"]
  n6["web/src/routes/Notifications.svelte:removePeriodicRule"]
  n7["web/src/routes/Notifications.svelte:resetTemplate"]
  n8["web/src/routes/Notifications.svelte:saveConfig"]
  n9["web/src/routes/Notifications.svelte:toggleDay"]
  n4 --> n3
```

## fe:web/src/routes/Settings.svelte

```mermaid
flowchart TD
  n1["web/src/routes/Settings.svelte:addCustomMapping"]
  n2["web/src/routes/Settings.svelte:removeCustomMapping"]
  n3["web/src/routes/Settings.svelte:saveConfig"]
```

## go:api

```mermaid
flowchart TD
  n1["internal/http/api/handler.go:(Handler).Register"]
  n2["internal/http/api/handler.go:(Handler).getConfig"]
  n3["internal/http/api/handler.go:(Handler).handleCalendarClear"]
  n4["internal/http/api/handler.go:(Handler).handleCalendarSync"]
  n5["internal/http/api/handler.go:(Handler).handleConfig"]
  n6["internal/http/api/handler.go:(Handler).handleDashboard"]
  n7["internal/http/api/handler.go:(Handler).handleDefaultTemplates"]
  n8["internal/http/api/handler.go:(Handler).handleHistory"]
  n9["internal/http/api/handler.go:(Handler).handleHistoryClear"]
  n10["internal/http/api/handler.go:(Handler).handleManualNotification"]
  n11["internal/http/api/handler.go:(Handler).handlePreviewNotification"]
  n12["internal/http/api/handler.go:(Handler).handleSync"]
  n13["internal/http/api/handler.go:(Handler).handleUpcomingEvents"]
  n14["internal/http/api/handler.go:(Handler).putConfig"]
  n15["internal/http/api/handler.go:(Handler).saveConfig"]
  n16["internal/http/api/handler.go:NewHandler"]
  n17["internal/http/api/handler_test.go:TestHandleDefaultTemplates"]
  n18["internal/http/api/handler_test.go:TestHandleManualNotificationPersistsTemplateBeforeSend"]
  n19["internal/http/api/handler_test.go:TestHandlePreviewNotificationReturnsMessageOnly"]
  n20["internal/http/api/handler_test.go:TestHandleUpcomingEventsCalendarState"]
  n21["internal/http/api/handler_test.go:fetchCalendarStates"]
  n22["internal/http/api/handler_test.go:postJSON"]
  n23["internal/http/api/handler_test.go:setupAPIHandler"]
  n24["internal/http/api/helpers.go:requireMethod"]
  n25["internal/http/api/helpers.go:respondError"]
  n26["internal/http/api/helpers.go:respondJSON"]
  n27["internal/http/api/helpers.go:respondValidationError"]
  n28["internal/http/api/timeutil.go:formatDurationShort"]
  n29["internal/http/api/timeutil.go:loadLocationOrLocal"]
  n30["internal/http/api/timeutil.go:parseDateInput"]
  n31["internal/http/api/timeutil.go:parseDateRange"]
  n2 --> n26
  n3 --> n24
  n3 --> n25
  n4 --> n24
  n4 --> n25
  n4 --> n26
  n4 --> n31
  n5 --> n25
  n6 --> n24
  n6 --> n25
  n6 --> n26
  n6 --> n28
  n6 --> n29
  n7 --> n24
  n7 --> n26
  n8 --> n24
  n8 --> n25
  n8 --> n26
  n9 --> n24
  n9 --> n25
  n10 --> n24
  n10 --> n25
  n10 --> n26
  n10 --> n31
  n11 --> n24
  n11 --> n25
  n11 --> n26
  n11 --> n31
  n12 --> n24
  n12 --> n25
  n12 --> n26
  n13 --> n24
  n13 --> n25
  n13 --> n26
  n13 --> n29
  n14 --> n25
  n14 --> n26
  n14 --> n27
  n17 --> n23
  n18 --> n23
  n19 --> n22
  n19 --> n23
  n20 --> n21
  n20 --> n23
  n23 --> n16
  n24 --> n25
  n25 --> n26
  n27 --> n26
  n31 --> n29
  n31 --> n30
```

## go:app

```mermaid
flowchart TD
  n1["internal/app/app.go:(App).Close"]
  n2["internal/app/app.go:(App).Start"]
  n3["internal/app/app.go:New"]
  n4["internal/app/app.go:buildRouter"]
  n5["internal/app/app.go:generateSelfSignedCert"]
  n6["internal/app/app.go:resolveServerRuntime"]
  n3 --> n3
  n3 --> n4
  n3 --> n5
  n3 --> n6
  n6 --> n3
```

## go:calendar

```mermaid
flowchart TD
  n1["internal/calendar/google.go:(Client).DeleteEvent"]
  n2["internal/calendar/google.go:(Client).ListEvents"]
  n3["internal/calendar/google.go:(Client).UpsertEvent"]
  n4["internal/calendar/google.go:(ClientOptions).Fingerprint"]
  n5["internal/calendar/google.go:(ClientOptions).IsConfigured"]
  n6["internal/calendar/google.go:(ClientOptions).Validate"]
  n7["internal/calendar/google.go:(ClientOptions).normalize"]
  n8["internal/calendar/google.go:EventMatchesNotion"]
  n9["internal/calendar/google.go:NewClient"]
  n10["internal/calendar/google.go:buildStartEnd"]
  n11["internal/calendar/google.go:equalEmails"]
  n12["internal/calendar/google.go:extractEmails"]
  n13["internal/calendar/google.go:mapEvent"]
  n14["internal/calendar/google.go:normalizeDateTime"]
  n15["internal/calendar/google.go:sameDateOrDateTime"]
  n2 --> n12
  n3 --> n13
  n8 --> n11
  n8 --> n12
  n8 --> n13
  n8 --> n15
  n13 --> n10
  n15 --> n14
```

## go:config

```mermaid
flowchart TD
  n1["internal/config/config.go:ApplyEnvOverrides"]
  n2["internal/config/config.go:DefaultTemplates"]
  n3["internal/config/config.go:IsSnoozed"]
  n4["internal/config/config.go:LoadConfig"]
  n5["internal/config/config.go:LoadEnv"]
  n6["internal/config/config.go:NormalizeConfig"]
  n7["internal/config/config.go:SanitizeTemplate"]
  n8["internal/config/config.go:ValidateConfig"]
  n9["internal/config/config.go:WriteConfig"]
  n10["internal/config/config.go:pickEnv"]
  n11["internal/config/config.go:pickEnvBool"]
  n12["internal/config/config.go:pickEnvInt"]
  n13["internal/config/config.go:validateHHMM"]
  n14["internal/config/config_test.go:TestApplyEnvOverridesAppPort"]
  n15["internal/config/config_test.go:TestApplyEnvOverridesBasicAuthEnabled"]
  n16["internal/config/config_test.go:TestApplyEnvOverridesTLSFiles"]
  n17["internal/config/config_test.go:TestApplyEnvOverridesWebhookURLs"]
  n18["internal/config/config_test.go:TestDefaultTemplatesContainExpectedTokens"]
  n19["internal/config/manager.go:(Manager).Config"]
  n20["internal/config/manager.go:(Manager).Env"]
  n21["internal/config/manager.go:(Manager).Reload"]
  n22["internal/config/manager.go:(Manager).Snapshot"]
  n23["internal/config/manager.go:(Manager).UpdateConfig"]
  n24["internal/config/manager.go:(ValidationError).Error"]
  n25["internal/config/manager.go:(ValidationError).Unwrap"]
  n26["internal/config/manager.go:NewManager"]
  n1 --> n10
  n1 --> n11
  n1 --> n12
  n4 --> n6
  n4 --> n8
  n6 --> n7
  n8 --> n13
  n9 --> n6
  n14 --> n1
  n15 --> n1
  n16 --> n1
  n17 --> n1
  n18 --> n2
  n21 --> n1
  n21 --> n4
  n21 --> n5
  n23 --> n6
  n23 --> n8
  n23 --> n9
  n26 --> n1
  n26 --> n4
  n26 --> n5
```

## go:db

```mermaid
flowchart TD
  n1["internal/db/db.go:(Repository).ClearNotificationHistory"]
  n2["internal/db/db.go:(Repository).ClearSyncRecords"]
  n3["internal/db/db.go:(Repository).Close"]
  n4["internal/db/db.go:(Repository).DeleteEventsNotIn"]
  n5["internal/db/db.go:(Repository).DeleteSyncRecord"]
  n6["internal/db/db.go:(Repository).GetEvent"]
  n7["internal/db/db.go:(Repository).GetSyncRecordMap"]
  n8["internal/db/db.go:(Repository).InsertNotificationHistory"]
  n9["internal/db/db.go:(Repository).ListEventsBetween"]
  n10["internal/db/db.go:(Repository).ListNotificationHistory"]
  n11["internal/db/db.go:(Repository).ListOrphanedSyncRecords"]
  n12["internal/db/db.go:(Repository).ListPendingAdvanceSchedules"]
  n13["internal/db/db.go:(Repository).ListSyncRecords"]
  n14["internal/db/db.go:(Repository).ListUpcomingEvents"]
  n15["internal/db/db.go:(Repository).MarkAdvanceScheduleFired"]
  n16["internal/db/db.go:(Repository).ReplaceAdvanceSchedules"]
  n17["internal/db/db.go:(Repository).UpsertEvents"]
  n18["internal/db/db.go:(Repository).UpsertSyncRecord"]
  n19["internal/db/db.go:Open"]
  n20["internal/db/db.go:boolToInt"]
  n21["internal/db/db.go:decodeStringSlice"]
  n22["internal/db/db.go:encodeStringSlice"]
  n23["internal/db/db.go:inPlaceholders"]
  n24["internal/db/db.go:intToBool"]
  n25["internal/db/db.go:parseRFC3339"]
  n26["internal/db/db.go:scanEvents"]
  n27["internal/db/db.go:scanSyncRecords"]
  n28["internal/db/db.go:toAnySlice"]
  n29["internal/db/db_test.go:TestReplaceAdvanceSchedulesClearsAllWhenEmpty"]
  n30["internal/db/db_test.go:TestReplaceAdvanceSchedulesPreservesFiredForSameFireAt"]
  n31["internal/db/db_test.go:TestReplaceAdvanceSchedulesResetsFiredWhenFireAtChangesAndDeletesStale"]
  n32["internal/db/db_test.go:TestUpsertEventsPersistsAttendees"]
  n33["internal/db/db_test.go:TestUpsertSyncRecordPersistsAttempted"]
  n34["internal/db/schema.go:initSchema"]
  n4 --> n23
  n4 --> n28
  n6 --> n21
  n6 --> n24
  n6 --> n25
  n7 --> n23
  n7 --> n27
  n7 --> n28
  n9 --> n26
  n10 --> n25
  n11 --> n27
  n12 --> n24
  n12 --> n25
  n13 --> n27
  n16 --> n20
  n17 --> n20
  n17 --> n22
  n18 --> n20
  n19 --> n19
  n19 --> n34
  n26 --> n21
  n26 --> n24
  n26 --> n25
  n27 --> n24
  n29 --> n19
  n30 --> n19
  n31 --> n19
  n32 --> n19
  n33 --> n19
```

## go:logging

```mermaid
flowchart TD
  n1["internal/logging/logging.go:Error"]
  n2["internal/logging/logging.go:Info"]
```

## go:main

```mermaid
flowchart TD
  n1["cmd/notion-notifier/main.go:main"]
```

## go:middleware

```mermaid
flowchart TD
  n1["internal/http/middleware/middleware.go:(responseRecorder).WriteHeader"]
  n2["internal/http/middleware/middleware.go:BasicAuth"]
  n3["internal/http/middleware/middleware.go:Logging"]
  n2 --> n2
```

## go:notion

```mermaid
flowchart TD
  n1["internal/notion/client.go:(Client).FetchContent"]
  n2["internal/notion/client.go:(Client).QueryDatabase"]
  n3["internal/notion/client.go:(Client).QueryDatabaseOnOrAfter"]
  n4["internal/notion/client.go:(Client).doRequest"]
  n5["internal/notion/client.go:(Client).listBlocks"]
  n6["internal/notion/client.go:ExtractEmails"]
  n7["internal/notion/client.go:ExtractString"]
  n8["internal/notion/client.go:MapPagesToEvents"]
  n9["internal/notion/client.go:New"]
  n10["internal/notion/client.go:blockInfo"]
  n11["internal/notion/client.go:extractContentFromBlocks"]
  n12["internal/notion/client.go:formatBlock"]
  n13["internal/notion/client.go:formatHeading"]
  n14["internal/notion/client.go:headingMatches"]
  n15["internal/notion/client.go:joinBlockText"]
  n16["internal/notion/client.go:joinRichText"]
  n17["internal/notion/client.go:parseDateRange"]
  n18["internal/notion/client.go:splitDateTime"]
  n19["internal/notion/client_test.go:TestMapPagesToEvents_DisabledAttendees"]
  n20["internal/notion/client_test.go:TestMapPagesToEvents_MapsAttendees"]
  n21["internal/notion/client_test.go:TestQueryDatabaseOnOrAfter_SendsDateFilter"]
  n22["internal/notion/client_test.go:TestQueryDatabase_NoFilterWhenDateConfigMissing"]
  n1 --> n11
  n3 --> n9
  n7 --> n7
  n7 --> n16
  n8 --> n6
  n8 --> n7
  n8 --> n17
  n10 --> n15
  n11 --> n10
  n11 --> n12
  n11 --> n13
  n11 --> n14
  n12 --> n13
  n17 --> n18
  n19 --> n8
  n20 --> n8
  n21 --> n9
  n22 --> n9
```

## go:retry

```mermaid
flowchart TD
  n1["internal/retry/retry.go:(Config).WithDefaults"]
  n2["internal/retry/retry.go:BackoffDelay"]
  n3["internal/retry/retry.go:IsRetryableStatus"]
  n4["internal/retry/retry.go:ParseRetryAfter"]
  n5["internal/retry/retry.go:Sleep"]
```

## go:scheduler

```mermaid
flowchart TD
  n1["internal/scheduler/helpers.go:buildAdvanceSchedules"]
  n2["internal/scheduler/helpers.go:buildFilterValues"]
  n3["internal/scheduler/helpers.go:buildTemplateEvents"]
  n4["internal/scheduler/helpers.go:extractCustomValues"]
  n5["internal/scheduler/helpers.go:groupCalendarEvents"]
  n6["internal/scheduler/helpers.go:loadLocationOrLocal"]
  n7["internal/scheduler/helpers.go:matchAdvanceConditions"]
  n8["internal/scheduler/helpers.go:matchFilter"]
  n9["internal/scheduler/helpers.go:matchesDays"]
  n10["internal/scheduler/helpers.go:notionOnOrAfterDate"]
  n11["internal/scheduler/helpers.go:parseEventStart"]
  n12["internal/scheduler/helpers.go:pickPrimaryCalendarEvent"]
  n13["internal/scheduler/helpers.go:scheduleKey"]
  n14["internal/scheduler/helpers.go:toTemplateEvent"]
  n15["internal/scheduler/helpers.go:weekdayToConfig"]
  n16["internal/scheduler/runtime.go:(Scheduler).NotionSyncStatus"]
  n17["internal/scheduler/runtime.go:(Scheduler).cancelRuntime"]
  n18["internal/scheduler/runtime.go:(Scheduler).clearAdvanceTimers"]
  n19["internal/scheduler/runtime.go:(Scheduler).currentTimezone"]
  n20["internal/scheduler/runtime.go:(Scheduler).markPeriodicSent"]
  n21["internal/scheduler/runtime.go:(Scheduler).newRuntimeOpContext"]
  n22["internal/scheduler/runtime.go:(Scheduler).periodicSent"]
  n23["internal/scheduler/runtime.go:(Scheduler).runtimeContext"]
  n24["internal/scheduler/runtime.go:(Scheduler).setNotionStatus"]
  n25["internal/scheduler/runtime.go:(Scheduler).setRuntimeContext"]
  n26["internal/scheduler/runtime.go:(Scheduler).withRuntimeOp"]
  n27["internal/scheduler/worker.go:(Scheduler).PreviewAdvanceTemplate"]
  n28["internal/scheduler/worker.go:(Scheduler).PreviewManualTemplate"]
  n29["internal/scheduler/worker.go:(Scheduler).RebuildAdvanceSchedules"]
  n30["internal/scheduler/worker.go:(Scheduler).Reload"]
  n31["internal/scheduler/worker.go:(Scheduler).SchedulePendingFromDB"]
  n32["internal/scheduler/worker.go:(Scheduler).SendManualNotification"]
  n33["internal/scheduler/worker.go:(Scheduler).Start"]
  n34["internal/scheduler/worker.go:(Scheduler).Stop"]
  n35["internal/scheduler/worker.go:(Scheduler).SyncCalendar"]
  n36["internal/scheduler/worker.go:(Scheduler).SyncNotion"]
  n37["internal/scheduler/worker.go:(Scheduler).calendarLoop"]
  n38["internal/scheduler/worker.go:(Scheduler).deleteCalendarEvents"]
  n39["internal/scheduler/worker.go:(Scheduler).fireAdvance"]
  n40["internal/scheduler/worker.go:(Scheduler).periodicLoop"]
  n41["internal/scheduler/worker.go:(Scheduler).rebuildAdvanceSchedules"]
  n42["internal/scheduler/worker.go:(Scheduler).renderListFromRange"]
  n43["internal/scheduler/worker.go:(Scheduler).schedulePendingFromDB"]
  n44["internal/scheduler/worker.go:(Scheduler).sendPeriodic"]
  n45["internal/scheduler/worker.go:(Scheduler).sendWebhook"]
  n46["internal/scheduler/worker.go:(Scheduler).syncCalendar"]
  n47["internal/scheduler/worker.go:(Scheduler).syncLoop"]
  n48["internal/scheduler/worker.go:(Scheduler).syncNotion"]
  n49["internal/scheduler/worker.go:New"]
  n50["internal/scheduler/worker_test.go:TestMatchAdvanceConditions"]
  n51["internal/scheduler/worker_test.go:TestMatchesDays"]
  n52["internal/scheduler/worker_test.go:TestNotionOnOrAfterDate_JSTEarlyMorningUsesPreviousUTCDate"]
  n53["internal/scheduler/worker_test.go:TestNotionOnOrAfterDate_PSTUsesSameUTCDate"]
  n54["internal/scheduler/worker_test.go:TestSendManualNotificationRoutesByIsTest"]
  n55["internal/scheduler/worker_test.go:TestSendWebhookRecordsHistoryOnPayloadRenderError"]
  n56["internal/scheduler/worker_test.go:TestToTemplateEvent_MapsEndDateAndTime"]
  n1 --> n7
  n1 --> n11
  n2 --> n4
  n3 --> n4
  n3 --> n14
  n7 --> n2
  n7 --> n8
  n7 --> n9
  n7 --> n15
  n27 --> n4
  n27 --> n6
  n27 --> n14
  n39 --> n4
  n39 --> n14
  n40 --> n6
  n40 --> n9
  n40 --> n15
  n41 --> n1
  n41 --> n6
  n42 --> n3
  n43 --> n6
  n43 --> n13
  n44 --> n6
  n45 --> n49
  n46 --> n5
  n46 --> n6
  n46 --> n12
  n46 --> n49
  n48 --> n6
  n48 --> n10
  n48 --> n49
  n50 --> n7
  n50 --> n15
  n51 --> n9
  n51 --> n15
  n52 --> n10
  n53 --> n10
  n54 --> n49
  n55 --> n49
  n56 --> n14
```

## go:static

```mermaid
flowchart TD
  n1["internal/http/static/spa.go:NewSPAHandler"]
```

## go:template

```mermaid
flowchart TD
  n1["internal/template/renderer.go:(Renderer).RenderList"]
  n2["internal/template/renderer.go:(Renderer).RenderPayload"]
  n3["internal/template/renderer.go:(Renderer).RenderSingle"]
  n4["internal/template/renderer.go:New"]
  n5["internal/template/renderer.go:newTemplate"]
  n1 --> n5
  n2 --> n5
  n3 --> n5
  n5 --> n4
```

## go:webhook

```mermaid
flowchart TD
  n1["internal/webhook/client.go:(Client).Send"]
  n2["internal/webhook/client.go:New"]
  n1 --> n2
```

## script:scripts/deploy-mac.sh

```mermaid
flowchart TD
  n1["scripts/deploy-mac.sh:usage"]
  n1 --> n1
```

