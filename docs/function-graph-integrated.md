# Integrated Function Graph (Unified AST View)

Generated: 2026-02-17 JST

## Coverage
- total functions: 247
- go functions (including tests): 201
- frontend functions (ts+svelte): 45
- script functions: 1
- total inferred edges: 247

## Edge Matrix (Group to Group)
| From -> To | Edges |
|---|---:|
| `fe:web/src/App.svelte -> fe:web/src/App.svelte` | 4 |
| `fe:web/src/App.svelte -> fe:web/src/lib/store.ts` | 3 |
| `fe:web/src/components/PreviewModal.svelte -> fe:web/src/components/PreviewModal.svelte` | 1 |
| `fe:web/src/lib/store.ts -> fe:web/src/lib/store.ts` | 1 |
| `fe:web/src/routes/Calendar.svelte -> fe:web/src/lib/store.ts` | 3 |
| `fe:web/src/routes/Dashboard.svelte -> fe:web/src/lib/store.ts` | 5 |
| `fe:web/src/routes/Dashboard.svelte -> fe:web/src/routes/Dashboard.svelte` | 3 |
| `fe:web/src/routes/History.svelte -> fe:web/src/lib/store.ts` | 2 |
| `fe:web/src/routes/Notifications.svelte -> fe:web/src/lib/store.ts` | 4 |
| `fe:web/src/routes/Notifications.svelte -> fe:web/src/routes/Notifications.svelte` | 1 |
| `fe:web/src/routes/Settings.svelte -> fe:web/src/lib/store.ts` | 1 |
| `go:api -> go:api` | 50 |
| `go:api -> go:config` | 6 |
| `go:api -> go:db` | 1 |
| `go:api -> go:logging` | 13 |
| `go:api -> go:scheduler` | 1 |
| `go:api -> go:template` | 1 |
| `go:app -> go:api` | 1 |
| `go:app -> go:app` | 5 |
| `go:app -> go:config` | 1 |
| `go:app -> go:db` | 1 |
| `go:app -> go:middleware` | 2 |
| `go:app -> go:notion` | 1 |
| `go:app -> go:scheduler` | 1 |
| `go:app -> go:static` | 1 |
| `go:app -> go:template` | 1 |
| `go:app -> go:webhook` | 1 |
| `go:calendar -> go:calendar` | 8 |
| `go:config -> go:config` | 22 |
| `go:db -> go:db` | 11 |
| `go:main -> go:app` | 1 |
| `go:middleware -> go:logging` | 1 |
| `go:middleware -> go:middleware` | 1 |
| `go:notion -> go:notion` | 18 |
| `go:notion -> go:retry` | 4 |
| `go:scheduler -> go:calendar` | 2 |
| `go:scheduler -> go:config` | 4 |
| `go:scheduler -> go:db` | 1 |
| `go:scheduler -> go:logging` | 9 |
| `go:scheduler -> go:notion` | 3 |
| `go:scheduler -> go:scheduler` | 32 |
| `go:scheduler -> go:template` | 1 |
| `go:scheduler -> go:webhook` | 1 |
| `go:template -> go:config` | 3 |
| `go:template -> go:template` | 4 |
| `go:webhook -> go:retry` | 4 |
| `go:webhook -> go:webhook` | 1 |
| `script:scripts/deploy-mac.sh -> script:scripts/deploy-mac.sh` | 1 |

## Unified Graph

```mermaid
flowchart LR
  subgraph sg1["fe:web/src/App.svelte"]
    f1["web/src/App.svelte:checkHealth"]
    f2["web/src/App.svelte:clearSnooze"]
    f3["web/src/App.svelte:closeSidebar"]
    f4["web/src/App.svelte:handleGlobalKeydown"]
    f5["web/src/App.svelte:handleSync"]
    f6["web/src/App.svelte:openGuideModal"]
    f7["web/src/App.svelte:openSidebar"]
    f8["web/src/App.svelte:saveSnooze"]
    f9["web/src/App.svelte:toggleDarkMode"]
    f10["web/src/App.svelte:updateClock"]
  end
  subgraph sg2["fe:web/src/components/PreviewModal.svelte"]
    f11["web/src/components/PreviewModal.svelte:close"]
    f12["web/src/components/PreviewModal.svelte:handleKeydown"]
    f13["web/src/components/PreviewModal.svelte:renderMarkdown"]
  end
  subgraph sg3["fe:web/src/components/TemplateGuideSidebar.svelte"]
    f14["web/src/components/TemplateGuideSidebar.svelte:openGuideDetail"]
  end
  subgraph sg4["fe:web/src/lib/api.ts"]
    f15["web/src/lib/api.ts:request"]
  end
  subgraph sg5["fe:web/src/lib/store.ts"]
    f16["web/src/lib/store.ts:addToast"]
    f17["web/src/lib/store.ts:navigate"]
    f18["web/src/lib/store.ts:saveConfig"]
    f19["web/src/lib/store.ts:setDarkMode"]
  end
  subgraph sg6["fe:web/src/routes/Calendar.svelte"]
    f20["web/src/routes/Calendar.svelte:handleClear"]
    f21["web/src/routes/Calendar.svelte:handleConfigUpdate"]
    f22["web/src/routes/Calendar.svelte:handleSync"]
  end
  subgraph sg7["fe:web/src/routes/Dashboard.svelte"]
    f23["web/src/routes/Dashboard.svelte:formatEventDateTime"]
    f24["web/src/routes/Dashboard.svelte:handleManualPreview"]
    f25["web/src/routes/Dashboard.svelte:handleManualSend"]
    f26["web/src/routes/Dashboard.svelte:handleSync"]
    f27["web/src/routes/Dashboard.svelte:loadData"]
    f28["web/src/routes/Dashboard.svelte:loadDefaultTemplate"]
    f29["web/src/routes/Dashboard.svelte:openPreview"]
  end
  subgraph sg8["fe:web/src/routes/History.svelte"]
    f30["web/src/routes/History.svelte:formatDate"]
    f31["web/src/routes/History.svelte:handleClear"]
    f32["web/src/routes/History.svelte:loadHistory"]
  end
  subgraph sg9["fe:web/src/routes/Notifications.svelte"]
    f33["web/src/routes/Notifications.svelte:addAdvanceRule"]
    f34["web/src/routes/Notifications.svelte:addPeriodicRule"]
    f35["web/src/routes/Notifications.svelte:openPreview"]
    f36["web/src/routes/Notifications.svelte:previewTemplate"]
    f37["web/src/routes/Notifications.svelte:removeAdvanceRule"]
    f38["web/src/routes/Notifications.svelte:removePeriodicRule"]
    f39["web/src/routes/Notifications.svelte:resetAdvanceTemplate"]
    f40["web/src/routes/Notifications.svelte:resetPeriodicTemplate"]
    f41["web/src/routes/Notifications.svelte:saveConfig"]
    f42["web/src/routes/Notifications.svelte:toggleDay"]
  end
  subgraph sg10["fe:web/src/routes/Settings.svelte"]
    f43["web/src/routes/Settings.svelte:addCustomMapping"]
    f44["web/src/routes/Settings.svelte:removeCustomMapping"]
    f45["web/src/routes/Settings.svelte:saveConfig"]
  end
  subgraph sg11["go:api"]
    f46["internal/http/api/handler.go:(Handler).Register"]
    f47["internal/http/api/handler.go:(Handler).getConfig"]
    f48["internal/http/api/handler.go:(Handler).handleCalendarClear"]
    f49["internal/http/api/handler.go:(Handler).handleCalendarSync"]
    f50["internal/http/api/handler.go:(Handler).handleConfig"]
    f51["internal/http/api/handler.go:(Handler).handleDashboard"]
    f52["internal/http/api/handler.go:(Handler).handleDefaultTemplates"]
    f53["internal/http/api/handler.go:(Handler).handleHistory"]
    f54["internal/http/api/handler.go:(Handler).handleHistoryClear"]
    f55["internal/http/api/handler.go:(Handler).handleManualNotification"]
    f56["internal/http/api/handler.go:(Handler).handlePreviewNotification"]
    f57["internal/http/api/handler.go:(Handler).handleSync"]
    f58["internal/http/api/handler.go:(Handler).handleUpcomingEvents"]
    f59["internal/http/api/handler.go:(Handler).putConfig"]
    f60["internal/http/api/handler.go:(Handler).saveConfig"]
    f61["internal/http/api/handler.go:NewHandler"]
    f62["internal/http/api/handler_test.go:TestHandleDefaultTemplates"]
    f63["internal/http/api/handler_test.go:TestHandleManualNotificationPersistsTemplateBeforeSend"]
    f64["internal/http/api/handler_test.go:TestHandlePreviewNotificationReturnsMessageOnly"]
    f65["internal/http/api/handler_test.go:TestHandleUpcomingEventsCalendarState"]
    f66["internal/http/api/handler_test.go:fetchCalendarStates"]
    f67["internal/http/api/handler_test.go:postJSON"]
    f68["internal/http/api/handler_test.go:setupAPIHandler"]
    f69["internal/http/api/helpers.go:requireMethod"]
    f70["internal/http/api/helpers.go:respondError"]
    f71["internal/http/api/helpers.go:respondJSON"]
    f72["internal/http/api/helpers.go:respondValidationError"]
    f73["internal/http/api/timeutil.go:formatDurationShort"]
    f74["internal/http/api/timeutil.go:loadLocationOrLocal"]
    f75["internal/http/api/timeutil.go:parseDateInput"]
    f76["internal/http/api/timeutil.go:parseDateRange"]
  end
  subgraph sg12["go:app"]
    f77["internal/app/app.go:(App).Close"]
    f78["internal/app/app.go:(App).Start"]
    f79["internal/app/app.go:New"]
    f80["internal/app/app.go:buildRouter"]
    f81["internal/app/app.go:generateSelfSignedCert"]
    f82["internal/app/app.go:resolveServerRuntime"]
  end
  subgraph sg13["go:calendar"]
    f83["internal/calendar/google.go:(Client).DeleteEvent"]
    f84["internal/calendar/google.go:(Client).ListEvents"]
    f85["internal/calendar/google.go:(Client).UpsertEvent"]
    f86["internal/calendar/google.go:(ClientOptions).Fingerprint"]
    f87["internal/calendar/google.go:(ClientOptions).IsConfigured"]
    f88["internal/calendar/google.go:(ClientOptions).Validate"]
    f89["internal/calendar/google.go:(ClientOptions).normalize"]
    f90["internal/calendar/google.go:EventMatchesNotion"]
    f91["internal/calendar/google.go:NewClient"]
    f92["internal/calendar/google.go:buildStartEnd"]
    f93["internal/calendar/google.go:equalEmails"]
    f94["internal/calendar/google.go:extractEmails"]
    f95["internal/calendar/google.go:mapEvent"]
    f96["internal/calendar/google.go:normalizeDateTime"]
    f97["internal/calendar/google.go:sameDateOrDateTime"]
  end
  subgraph sg14["go:config"]
    f98["internal/config/config.go:ApplyEnvOverrides"]
    f99["internal/config/config.go:DefaultTemplates"]
    f100["internal/config/config.go:IsSnoozed"]
    f101["internal/config/config.go:LoadConfig"]
    f102["internal/config/config.go:LoadEnv"]
    f103["internal/config/config.go:NormalizeConfig"]
    f104["internal/config/config.go:SanitizeTemplate"]
    f105["internal/config/config.go:ValidateConfig"]
    f106["internal/config/config.go:WriteConfig"]
    f107["internal/config/config.go:pickEnv"]
    f108["internal/config/config.go:pickEnvBool"]
    f109["internal/config/config.go:pickEnvInt"]
    f110["internal/config/config.go:validateHHMM"]
    f111["internal/config/config_test.go:TestApplyEnvOverridesAppPort"]
    f112["internal/config/config_test.go:TestApplyEnvOverridesBasicAuthEnabled"]
    f113["internal/config/config_test.go:TestApplyEnvOverridesTLSFiles"]
    f114["internal/config/config_test.go:TestApplyEnvOverridesWebhookURLs"]
    f115["internal/config/config_test.go:TestDefaultTemplatesContainExpectedTokens"]
    f116["internal/config/manager.go:(Manager).Config"]
    f117["internal/config/manager.go:(Manager).Env"]
    f118["internal/config/manager.go:(Manager).Reload"]
    f119["internal/config/manager.go:(Manager).Snapshot"]
    f120["internal/config/manager.go:(Manager).UpdateConfig"]
    f121["internal/config/manager.go:(ValidationError).Error"]
    f122["internal/config/manager.go:(ValidationError).Unwrap"]
    f123["internal/config/manager.go:NewManager"]
  end
  subgraph sg15["go:db"]
    f124["internal/db/db.go:(Repository).ClearNotificationHistory"]
    f125["internal/db/db.go:(Repository).ClearSyncRecords"]
    f126["internal/db/db.go:(Repository).Close"]
    f127["internal/db/db.go:(Repository).DeleteEventsNotIn"]
    f128["internal/db/db.go:(Repository).DeleteSyncRecord"]
    f129["internal/db/db.go:(Repository).GetEvent"]
    f130["internal/db/db.go:(Repository).GetSyncRecordMap"]
    f131["internal/db/db.go:(Repository).InsertNotificationHistory"]
    f132["internal/db/db.go:(Repository).ListEventsBetween"]
    f133["internal/db/db.go:(Repository).ListNotificationHistory"]
    f134["internal/db/db.go:(Repository).ListOrphanedSyncRecords"]
    f135["internal/db/db.go:(Repository).ListPendingAdvanceSchedules"]
    f136["internal/db/db.go:(Repository).ListSyncRecords"]
    f137["internal/db/db.go:(Repository).ListUpcomingEvents"]
    f138["internal/db/db.go:(Repository).MarkAdvanceScheduleFired"]
    f139["internal/db/db.go:(Repository).ReplaceAdvanceSchedules"]
    f140["internal/db/db.go:(Repository).UpsertEvents"]
    f141["internal/db/db.go:(Repository).UpsertSyncRecord"]
    f142["internal/db/db.go:Open"]
    f143["internal/db/db.go:decodeStringSlice"]
    f144["internal/db/db.go:encodeStringSlice"]
    f145["internal/db/db.go:scanEvents"]
    f146["internal/db/db_test.go:TestReplaceAdvanceSchedulesClearsAllWhenEmpty"]
    f147["internal/db/db_test.go:TestReplaceAdvanceSchedulesPreservesFiredForSameFireAt"]
    f148["internal/db/db_test.go:TestReplaceAdvanceSchedulesResetsFiredWhenFireAtChangesAndDeletesStale"]
    f149["internal/db/db_test.go:TestUpsertEventsPersistsAttendees"]
    f150["internal/db/db_test.go:TestUpsertSyncRecordPersistsAttempted"]
    f151["internal/db/schema.go:initSchema"]
  end
  subgraph sg16["go:logging"]
    f152["internal/logging/logging.go:Error"]
    f153["internal/logging/logging.go:Info"]
  end
  subgraph sg17["go:main"]
    f154["cmd/notion-notifier/main.go:main"]
  end
  subgraph sg18["go:middleware"]
    f155["internal/http/middleware/middleware.go:(responseRecorder).WriteHeader"]
    f156["internal/http/middleware/middleware.go:BasicAuth"]
    f157["internal/http/middleware/middleware.go:Logging"]
  end
  subgraph sg19["go:notion"]
    f158["internal/notion/client.go:(Client).FetchContent"]
    f159["internal/notion/client.go:(Client).QueryDatabase"]
    f160["internal/notion/client.go:(Client).QueryDatabaseOnOrAfter"]
    f161["internal/notion/client.go:(Client).doRequest"]
    f162["internal/notion/client.go:(Client).listBlocks"]
    f163["internal/notion/client.go:ExtractEmails"]
    f164["internal/notion/client.go:ExtractString"]
    f165["internal/notion/client.go:MapPagesToEvents"]
    f166["internal/notion/client.go:New"]
    f167["internal/notion/client.go:blockInfo"]
    f168["internal/notion/client.go:extractContentFromBlocks"]
    f169["internal/notion/client.go:formatBlock"]
    f170["internal/notion/client.go:formatHeading"]
    f171["internal/notion/client.go:headingMatches"]
    f172["internal/notion/client.go:joinBlockText"]
    f173["internal/notion/client.go:joinRichText"]
    f174["internal/notion/client.go:parseDateRange"]
    f175["internal/notion/client.go:splitDateTime"]
    f176["internal/notion/client_test.go:TestMapPagesToEvents_DisabledAttendees"]
    f177["internal/notion/client_test.go:TestMapPagesToEvents_MapsAttendees"]
    f178["internal/notion/client_test.go:TestQueryDatabaseOnOrAfter_SendsDateFilter"]
    f179["internal/notion/client_test.go:TestQueryDatabase_NoFilterWhenDateConfigMissing"]
  end
  subgraph sg20["go:retry"]
    f180["internal/retry/retry.go:(Config).WithDefaults"]
    f181["internal/retry/retry.go:BackoffDelay"]
    f182["internal/retry/retry.go:IsRetryableStatus"]
    f183["internal/retry/retry.go:ParseRetryAfter"]
    f184["internal/retry/retry.go:Sleep"]
  end
  subgraph sg21["go:scheduler"]
    f185["internal/scheduler/runtime.go:(Scheduler).NotionSyncStatus"]
    f186["internal/scheduler/runtime.go:(Scheduler).cancelRuntime"]
    f187["internal/scheduler/runtime.go:(Scheduler).clearAdvanceTimers"]
    f188["internal/scheduler/runtime.go:(Scheduler).currentTimezone"]
    f189["internal/scheduler/runtime.go:(Scheduler).markPeriodicSent"]
    f190["internal/scheduler/runtime.go:(Scheduler).newRuntimeOpContext"]
    f191["internal/scheduler/runtime.go:(Scheduler).periodicSent"]
    f192["internal/scheduler/runtime.go:(Scheduler).runtimeContext"]
    f193["internal/scheduler/runtime.go:(Scheduler).setNotionStatus"]
    f194["internal/scheduler/runtime.go:(Scheduler).setRuntimeContext"]
    f195["internal/scheduler/runtime.go:(Scheduler).withRuntimeOp"]
    f196["internal/scheduler/worker.go:(Scheduler).PreviewAdvanceTemplate"]
    f197["internal/scheduler/worker.go:(Scheduler).PreviewManualTemplate"]
    f198["internal/scheduler/worker.go:(Scheduler).RebuildAdvanceSchedules"]
    f199["internal/scheduler/worker.go:(Scheduler).Reload"]
    f200["internal/scheduler/worker.go:(Scheduler).SchedulePendingFromDB"]
    f201["internal/scheduler/worker.go:(Scheduler).SendManualNotification"]
    f202["internal/scheduler/worker.go:(Scheduler).Start"]
    f203["internal/scheduler/worker.go:(Scheduler).Stop"]
    f204["internal/scheduler/worker.go:(Scheduler).SyncCalendar"]
    f205["internal/scheduler/worker.go:(Scheduler).SyncNotion"]
    f206["internal/scheduler/worker.go:(Scheduler).calendarLoop"]
    f207["internal/scheduler/worker.go:(Scheduler).deleteCalendarEvents"]
    f208["internal/scheduler/worker.go:(Scheduler).fireAdvance"]
    f209["internal/scheduler/worker.go:(Scheduler).periodicLoop"]
    f210["internal/scheduler/worker.go:(Scheduler).rebuildAdvanceSchedules"]
    f211["internal/scheduler/worker.go:(Scheduler).renderListFromRange"]
    f212["internal/scheduler/worker.go:(Scheduler).schedulePendingFromDB"]
    f213["internal/scheduler/worker.go:(Scheduler).sendPeriodic"]
    f214["internal/scheduler/worker.go:(Scheduler).sendWebhook"]
    f215["internal/scheduler/worker.go:(Scheduler).syncCalendar"]
    f216["internal/scheduler/worker.go:(Scheduler).syncLoop"]
    f217["internal/scheduler/worker.go:(Scheduler).syncNotion"]
    f218["internal/scheduler/worker.go:New"]
    f219["internal/scheduler/worker.go:buildAdvanceSchedules"]
    f220["internal/scheduler/worker.go:buildFilterValues"]
    f221["internal/scheduler/worker.go:buildTemplateEvents"]
    f222["internal/scheduler/worker.go:extractCustomValues"]
    f223["internal/scheduler/worker.go:groupCalendarEvents"]
    f224["internal/scheduler/worker.go:matchAdvanceConditions"]
    f225["internal/scheduler/worker.go:matchFilter"]
    f226["internal/scheduler/worker.go:matchesDays"]
    f227["internal/scheduler/worker.go:notionOnOrAfterDate"]
    f228["internal/scheduler/worker.go:parseEventStart"]
    f229["internal/scheduler/worker.go:pickPrimaryCalendarEvent"]
    f230["internal/scheduler/worker.go:scheduleKey"]
    f231["internal/scheduler/worker.go:toTemplateEvent"]
    f232["internal/scheduler/worker.go:weekdayToConfig"]
    f233["internal/scheduler/worker_test.go:TestMatchAdvanceConditions"]
    f234["internal/scheduler/worker_test.go:TestMatchesDays"]
    f235["internal/scheduler/worker_test.go:TestNotionOnOrAfterDate_JSTEarlyMorningUsesPreviousUTCDate"]
    f236["internal/scheduler/worker_test.go:TestNotionOnOrAfterDate_PSTUsesSameUTCDate"]
    f237["internal/scheduler/worker_test.go:TestSendWebhookRecordsHistoryOnPayloadRenderError"]
    f238["internal/scheduler/worker_test.go:TestToTemplateEvent_MapsEndDateAndTime"]
  end
  subgraph sg22["go:static"]
    f239["internal/http/static/spa.go:NewSPAHandler"]
  end
  subgraph sg23["go:template"]
    f240["internal/template/renderer.go:(Renderer).RenderList"]
    f241["internal/template/renderer.go:(Renderer).RenderPayload"]
    f242["internal/template/renderer.go:(Renderer).RenderSingle"]
    f243["internal/template/renderer.go:New"]
    f244["internal/template/renderer.go:newTemplate"]
  end
  subgraph sg24["go:webhook"]
    f245["internal/webhook/client.go:(Client).Send"]
    f246["internal/webhook/client.go:New"]
  end
  subgraph sg25["script:scripts/deploy-mac.sh"]
    f247["scripts/deploy-mac.sh:usage"]
  end
  f2 --> f8
  f4 --> f3
  f5 --> f1
  f5 --> f16
  f8 --> f1
  f8 --> f18
  f9 --> f19
  f12 --> f11
  f18 --> f16
  f20 --> f16
  f21 --> f18
  f22 --> f16
  f24 --> f16
  f24 --> f29
  f25 --> f16
  f25 --> f29
  f26 --> f16
  f26 --> f27
  f27 --> f16
  f28 --> f16
  f31 --> f16
  f32 --> f16
  f36 --> f16
  f36 --> f35
  f39 --> f16
  f40 --> f16
  f41 --> f18
  f45 --> f18
  f154 --> f79
  f79 --> f79
  f79 --> f80
  f79 --> f81
  f79 --> f82
  f79 --> f123
  f79 --> f142
  f79 --> f166
  f79 --> f218
  f79 --> f243
  f79 --> f246
  f80 --> f61
  f80 --> f156
  f80 --> f157
  f80 --> f239
  f82 --> f79
  f84 --> f94
  f85 --> f95
  f90 --> f93
  f90 --> f94
  f90 --> f95
  f90 --> f97
  f95 --> f92
  f97 --> f96
  f98 --> f107
  f98 --> f108
  f98 --> f109
  f101 --> f103
  f101 --> f105
  f103 --> f104
  f105 --> f110
  f106 --> f103
  f111 --> f98
  f112 --> f98
  f113 --> f98
  f114 --> f98
  f115 --> f99
  f118 --> f98
  f118 --> f101
  f118 --> f102
  f120 --> f103
  f120 --> f105
  f120 --> f106
  f123 --> f98
  f123 --> f101
  f123 --> f102
  f129 --> f143
  f132 --> f145
  f140 --> f144
  f142 --> f142
  f142 --> f151
  f145 --> f143
  f146 --> f142
  f147 --> f142
  f148 --> f142
  f149 --> f142
  f150 --> f142
  f47 --> f71
  f48 --> f69
  f48 --> f70
  f48 --> f152
  f48 --> f153
  f49 --> f69
  f49 --> f70
  f49 --> f71
  f49 --> f76
  f49 --> f152
  f49 --> f153
  f50 --> f70
  f51 --> f100
  f51 --> f69
  f51 --> f70
  f51 --> f71
  f51 --> f73
  f51 --> f74
  f51 --> f152
  f52 --> f99
  f52 --> f69
  f52 --> f71
  f53 --> f69
  f53 --> f70
  f53 --> f71
  f53 --> f152
  f54 --> f69
  f54 --> f70
  f54 --> f152
  f54 --> f153
  f55 --> f104
  f55 --> f69
  f55 --> f70
  f55 --> f71
  f55 --> f76
  f55 --> f152
  f56 --> f69
  f56 --> f70
  f56 --> f71
  f56 --> f76
  f57 --> f69
  f57 --> f70
  f57 --> f71
  f57 --> f152
  f58 --> f69
  f58 --> f70
  f58 --> f71
  f58 --> f74
  f58 --> f152
  f59 --> f70
  f59 --> f71
  f59 --> f72
  f59 --> f152
  f59 --> f153
  f62 --> f68
  f63 --> f68
  f64 --> f67
  f64 --> f68
  f65 --> f66
  f65 --> f68
  f68 --> f103
  f68 --> f106
  f68 --> f123
  f68 --> f142
  f68 --> f61
  f68 --> f218
  f68 --> f243
  f69 --> f70
  f70 --> f71
  f72 --> f71
  f76 --> f74
  f76 --> f75
  f156 --> f156
  f157 --> f153
  f158 --> f168
  f160 --> f166
  f161 --> f181
  f161 --> f182
  f161 --> f183
  f161 --> f184
  f164 --> f164
  f164 --> f173
  f165 --> f163
  f165 --> f164
  f165 --> f174
  f167 --> f172
  f168 --> f167
  f168 --> f169
  f168 --> f170
  f168 --> f171
  f169 --> f170
  f174 --> f175
  f176 --> f165
  f177 --> f165
  f178 --> f166
  f179 --> f166
  f196 --> f222
  f196 --> f231
  f207 --> f152
  f207 --> f153
  f208 --> f222
  f208 --> f231
  f209 --> f226
  f209 --> f232
  f210 --> f219
  f211 --> f221
  f212 --> f153
  f212 --> f230
  f214 --> f100
  f214 --> f152
  f214 --> f153
  f214 --> f218
  f215 --> f90
  f215 --> f91
  f215 --> f152
  f215 --> f153
  f215 --> f218
  f215 --> f223
  f215 --> f229
  f217 --> f152
  f217 --> f153
  f217 --> f165
  f217 --> f166
  f217 --> f218
  f217 --> f227
  f219 --> f224
  f219 --> f228
  f220 --> f222
  f221 --> f222
  f221 --> f231
  f222 --> f164
  f224 --> f220
  f224 --> f225
  f224 --> f226
  f224 --> f232
  f233 --> f224
  f233 --> f232
  f234 --> f226
  f234 --> f232
  f235 --> f227
  f236 --> f227
  f237 --> f103
  f237 --> f106
  f237 --> f123
  f237 --> f142
  f237 --> f218
  f237 --> f243
  f237 --> f246
  f238 --> f231
  f240 --> f104
  f240 --> f244
  f241 --> f104
  f241 --> f244
  f242 --> f104
  f242 --> f244
  f244 --> f243
  f245 --> f181
  f245 --> f182
  f245 --> f183
  f245 --> f184
  f245 --> f246
  f247 --> f247
```
