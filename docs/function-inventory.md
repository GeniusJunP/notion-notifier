# Function Inventory (Global Audit)

Generated: 2026-02-17 09:50:21 JST

## Scope
- scanned with `ls -R` and `rg`
- target dirs: `cmd`, `internal`, `web/src`, `scripts`, `docs`

## Directory Snapshot
```text
cmd:
notion-notifier

cmd/notion-notifier:
main.go

docs:
api-contract-audit.md
api.md
config.md
data-flow.md
deployment-tutorial.md
features.md
function-inventory.md

internal:
app
calendar
config
db
http
logging
models
notion
retry
scheduler
template
webhook

internal/app:
app.go

internal/calendar:
google.go

internal/config:
config.go
config_test.go
manager.go

internal/db:
db.go
db_test.go
schema.go

internal/http:
api
middleware
static

internal/http/api:
handler.go
handler_test.go
helpers.go
timeutil.go

internal/http/middleware:
middleware.go

internal/http/static:
spa.go

internal/logging:
logging.go

internal/models:
models.go

internal/notion:
client.go
client_test.go

internal/retry:
retry.go

internal/scheduler:
runtime.go
worker.go
worker_test.go

internal/template:
renderer.go

internal/webhook:
client.go

scripts:
deploy-linux.sh
deploy-mac.sh
deploy-windows.ps1

web/src:
App.svelte
app.css
components
lib
main.ts
routes

web/src/components:
PreviewModal.svelte
SidebarButton.svelte
TemplateGuideSidebar.svelte
ToastContainer.svelte
WebhookSettingsCard.svelte

web/src/lib:
api.ts
store.ts

web/src/routes:
Calendar.svelte
Dashboard.svelte
History.svelte
Notifications.svelte
Settings.svelte
```

## Function Count
```text
go(all)=200
go(prod only)=172
ts(function decl)=4
svelte(function decl)=40
shell(function decl)=1
powershell(function decl)=0
```

## Go Functions (All, includes tests)
```text
cmd/notion-notifier/main.go:16:func main() {
internal/app/app.go:110:func buildRouter(cfg *config.Manager, repo *db.Repository, sched *scheduler.Scheduler) http.Handler {
internal/app/app.go:133:func resolveServerRuntime(env config.Env) (string, tlsRuntime, error) {
internal/app/app.go:153:func generateSelfSignedCert() (tls.Certificate, error) {
internal/app/app.go:185:func (a *App) Start(ctx context.Context) error {
internal/app/app.go:206:func (a *App) Close() error {
internal/app/app.go:50:func New(cfgPath, envPath, dbPath string) (*App, error) {
internal/calendar/google.go:102:func (c *Client) DeleteEvent(ctx context.Context, eventID string) error {
internal/calendar/google.go:123:func (c *Client) ListEvents(ctx context.Context, from, to time.Time) ([]CalendarEvent, error) {
internal/calendar/google.go:176:func EventMatchesNotion(calEvent CalendarEvent, ev models.Event, tz *time.Location) bool {
internal/calendar/google.go:199:func sameDateOrDateTime(currentDate, currentDateTime, desiredDate, desiredDateTime string) bool {
internal/calendar/google.go:210:func normalizeDateTime(value string) string {
internal/calendar/google.go:221:func equalEmails(a, b []string) bool {
internal/calendar/google.go:233:func extractEmails(attendees []*calendarapi.EventAttendee) []string {
internal/calendar/google.go:260:func mapEvent(ev models.Event, tz *time.Location) *calendarapi.Event {
internal/calendar/google.go:286:func buildStartEnd(ev models.Event, tz *time.Location) (*calendarapi.EventDateTime, *calendarapi.EventDateTime) {
internal/calendar/google.go:33:func (o ClientOptions) normalize() ClientOptions {
internal/calendar/google.go:41:func (o ClientOptions) Validate() error {
internal/calendar/google.go:52:func (o ClientOptions) IsConfigured() bool {
internal/calendar/google.go:56:func (o ClientOptions) Fingerprint() string {
internal/calendar/google.go:67:func NewClient(ctx context.Context, opts ClientOptions) (*Client, error) {
internal/calendar/google.go:86:func (c *Client) UpsertEvent(ctx context.Context, ev models.Event, existingID string, tz *time.Location) (string, string, error) {
internal/config/config.go:150:func LoadConfig(path string) (Config, error) {
internal/config/config.go:166:func LoadEnv(path string) (Env, error) {
internal/config/config.go:181:func ApplyEnvOverrides(env Env) Env {
internal/config/config.go:199:func pickEnv(key, fallback string) string {
internal/config/config.go:206:func pickEnvBool(key string, fallback bool) bool {
internal/config/config.go:218:func pickEnvInt(key string, fallback int) int {
internal/config/config.go:230:func ValidateConfig(cfg Config) error {
internal/config/config.go:279:func validateHHMM(value string) error {
internal/config/config.go:287:func IsSnoozed(cfg Config, now time.Time) bool {
internal/config/config.go:298:func WriteConfig(path string, cfg Config) error {
internal/config/config.go:311:func NormalizeConfig(cfg Config) Config {
internal/config/config.go:363:func SanitizeTemplate(input string) string {
internal/config/config.go:396:func DefaultTemplates() map[string]string {
internal/config/config_test.go:51:func TestApplyEnvOverridesBasicAuthEnabled(t *testing.T) {
internal/config/config_test.go:69:func TestApplyEnvOverridesAppPort(t *testing.T) {
internal/config/config_test.go:77:func TestApplyEnvOverridesTLSFiles(t *testing.T) {
internal/config/config_test.go:89:func TestApplyEnvOverridesWebhookURLs(t *testing.T) {
internal/config/config_test.go:8:func TestDefaultTemplatesContainExpectedTokens(t *testing.T) {
internal/config/manager.go:11:func (e ValidationError) Error() string {
internal/config/manager.go:18:func (e ValidationError) Unwrap() error {
internal/config/manager.go:30:func NewManager(cfgPath, envPath string) (*Manager, error) {
internal/config/manager.go:43:func (m *Manager) Snapshot() (Config, Env) {
internal/config/manager.go:49:func (m *Manager) Config() Config {
internal/config/manager.go:55:func (m *Manager) Env() Env {
internal/config/manager.go:61:func (m *Manager) UpdateConfig(cfg Config) (Config, error) {
internal/config/manager.go:75:func (m *Manager) Reload() error {
internal/db/db.go:110:func (r *Repository) ListUpcomingEvents(ctx context.Context, days int, now time.Time) ([]models.Event, error) {
internal/db/db.go:115:func (r *Repository) GetEvent(ctx context.Context, notionPageID string) (models.Event, bool, error) {
internal/db/db.go:139:func scanEvents(rows *sql.Rows) ([]models.Event, error) {
internal/db/db.go:161:func encodeStringSlice(values []string) string {
internal/db/db.go:172:func decodeStringSlice(raw string) []string {
internal/db/db.go:186:func (r *Repository) InsertNotificationHistory(ctx context.Context, h models.NotificationHistory) error {
internal/db/db.go:192:func (r *Repository) ListNotificationHistory(ctx context.Context, limit int) ([]models.NotificationHistory, error) {
internal/db/db.go:20:func Open(path string) (*Repository, error) {
internal/db/db.go:219:func (r *Repository) ClearNotificationHistory(ctx context.Context) error {
internal/db/db.go:224:func (r *Repository) ReplaceAdvanceSchedules(ctx context.Context, schedules []models.AdvanceSchedule) error {
internal/db/db.go:273:func (r *Repository) ListPendingAdvanceSchedules(ctx context.Context) ([]models.AdvanceSchedule, error) {
internal/db/db.go:297:func (r *Repository) MarkAdvanceScheduleFired(ctx context.Context, id int64) error {
internal/db/db.go:302:func (r *Repository) ListSyncRecords(ctx context.Context) ([]models.SyncRecord, error) {
internal/db/db.go:325:func (r *Repository) GetSyncRecordMap(ctx context.Context, notionPageIDs []string) (map[string]models.SyncRecord, error) {
internal/db/db.go:32:func (r *Repository) Close() error {
internal/db/db.go:356:func (r *Repository) UpsertSyncRecord(ctx context.Context, record models.SyncRecord) error {
internal/db/db.go:36:func (r *Repository) UpsertEvents(ctx context.Context, events []models.Event) error {
internal/db/db.go:380:func (r *Repository) ClearSyncRecords(ctx context.Context) error {
internal/db/db.go:386:func (r *Repository) ListOrphanedSyncRecords(ctx context.Context) ([]models.SyncRecord, error) {
internal/db/db.go:411:func (r *Repository) DeleteSyncRecord(ctx context.Context, notionPageID string) error {
internal/db/db.go:416:func (r *Repository) DeleteEventsNotIn(ctx context.Context, ids []string) error {
internal/db/db.go:97:func (r *Repository) ListEventsBetween(ctx context.Context, from, to time.Time) ([]models.Event, error) {
internal/db/db_test.go:15:func TestUpsertEventsPersistsAttendees(t *testing.T) {
internal/db/db_test.go:164:func TestReplaceAdvanceSchedulesClearsAllWhenEmpty(t *testing.T) {
internal/db/db_test.go:192:func TestUpsertSyncRecordPersistsAttempted(t *testing.T) {
internal/db/db_test.go:224:func TestMigrateSyncRecordsAddsAttemptedFromLegacySchema(t *testing.T) {
internal/db/db_test.go:51:func TestReplaceAdvanceSchedulesPreservesFiredForSameFireAt(t *testing.T) {
internal/db/db_test.go:92:func TestReplaceAdvanceSchedulesResetsFiredWhenFireAtChangesAndDeletesStale(t *testing.T) {
internal/db/schema.go:75:func migrateSyncRecords(db *sql.DB) error {
internal/db/schema.go:8:func initSchema(db *sql.DB) error {
internal/http/api/handler.go:161:func (h *Handler) handleUpcomingEvents(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:236:func (h *Handler) handleHistory(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:24:func NewHandler(cfg *config.Manager, repo *db.Repository, sched *scheduler.Scheduler) *Handler {
internal/http/api/handler.go:264:func (h *Handler) handleSync(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:286:func (h *Handler) handleCalendarSync(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:29:func (h *Handler) Register(mux *http.ServeMux) {
internal/http/api/handler.go:317:func (h *Handler) handleCalendarClear(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:335:func (h *Handler) handleHistoryClear(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:364:func (h *Handler) handlePreviewNotification(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:404:func (h *Handler) handleManualNotification(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:442:func (h *Handler) handleDefaultTemplates(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:45:func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:56:func (h *Handler) getConfig(w http.ResponseWriter, _ *http.Request) {
internal/http/api/handler.go:60:func (h *Handler) putConfig(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:98:func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler_test.go:139:func TestHandleManualNotificationPersistsTemplateBeforeSend(t *testing.T) {
internal/http/api/handler_test.go:188:func TestHandleDefaultTemplates(t *testing.T) {
internal/http/api/handler_test.go:22:func TestHandleUpcomingEventsCalendarState(t *testing.T) {
internal/http/api/handler_test.go:231:func setupAPIHandler(t *testing.T, calendarEnabled bool) (*http.ServeMux, *db.Repository, *config.Manager) {
internal/http/api/handler_test.go:302:func fetchCalendarStates(t *testing.T, mux *http.ServeMux) map[string]string {
internal/http/api/handler_test.go:325:func postJSON(t *testing.T, mux *http.ServeMux, path string, body any) map[string]any {
internal/http/api/handler_test.go:92:func TestHandlePreviewNotificationReturnsMessageOnly(t *testing.T) {
internal/http/api/helpers.go:18:func respondError(w http.ResponseWriter, status int, message string) {
internal/http/api/helpers.go:23:func respondValidationError(w http.ResponseWriter, message string, details map[string]string) {
internal/http/api/helpers.go:9:func respondJSON(w http.ResponseWriter, status int, data any) {
internal/http/api/timeutil.go:11:func parseDateRange(fromStr, toStr string, cfg *config.Manager) (time.Time, time.Time, error) {
internal/http/api/timeutil.go:46:func parseDateInput(value string, loc *time.Location) (time.Time, error) {
internal/http/api/timeutil.go:68:func formatDurationShort(d time.Duration) string {
internal/http/api/timeutil.go:87:func loadLocationOrLocal(name string) *time.Location {
internal/http/middleware/middleware.go:12:func Logging(next http.Handler) http.Handler {
internal/http/middleware/middleware.go:23:func BasicAuth(cfg *config.Manager) func(http.Handler) http.Handler {
internal/http/middleware/middleware.go:47:func (r *responseRecorder) WriteHeader(status int) {
internal/http/static/spa.go:12:func NewSPAHandler(distFS fs.FS) http.Handler {
internal/logging/logging.go:12:func Error(category, format string, args ...interface{}) {
internal/logging/logging.go:8:func Info(category, format string, args ...interface{}) {
internal/notion/client.go:115:func (c *Client) doRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
internal/notion/client.go:157:func MapPagesToEvents(pages []page, mapping config.PropertyMapping, tz *time.Location) []models.Event {
internal/notion/client.go:199:func parseDateRange(prop any, tz *time.Location) (startDate, startTime, endDate, endTime string, isAllDay bool) {
internal/notion/client.go:221:func splitDateTime(value string, tz *time.Location) (date string, tm string, allDay bool) {
internal/notion/client.go:243:func ExtractString(prop any) string {
internal/notion/client.go:356:func ExtractEmails(prop any) []string {
internal/notion/client.go:385:func joinRichText(value any) string {
internal/notion/client.go:430:func (c *Client) FetchContent(ctx context.Context, pageID string, rules config.ContentRules) (string, error) {
internal/notion/client.go:441:func (c *Client) listBlocks(ctx context.Context, blockID string) ([]block, error) {
internal/notion/client.go:466:func extractContentFromBlocks(blocks []block, rules config.ContentRules) string {
internal/notion/client.go:501:func blockInfo(b block) (text string, kind string, level int, isHeading bool, isDivider bool, checked bool) {
internal/notion/client.go:524:func joinBlockText(bt *blockText) string {
internal/notion/client.go:537:func headingMatches(text, start string) bool {
internal/notion/client.go:541:func formatHeading(level int, text string) string {
internal/notion/client.go:552:func formatBlock(kind string, level int, text string, checked bool) string {
internal/notion/client.go:57:func New(httpClient *http.Client, apiKey string, cfg retry.Config) *Client {
internal/notion/client.go:69:func (c *Client) QueryDatabase(ctx context.Context, databaseID string) ([]page, error) {
internal/notion/client.go:73:func (c *Client) QueryDatabaseOnOrAfter(ctx context.Context, databaseID, dateProperty, onOrAfter string) ([]page, error) {
internal/notion/client_test.go:106:func TestQueryDatabaseOnOrAfter_SendsDateFilter(t *testing.T) {
internal/notion/client_test.go:151:func TestQueryDatabase_NoFilterWhenDateConfigMissing(t *testing.T) {
internal/notion/client_test.go:17:func TestMapPagesToEvents_MapsAttendees(t *testing.T) {
internal/notion/client_test.go:63:func TestMapPagesToEvents_DisabledAttendees(t *testing.T) {
internal/retry/retry.go:17:func (c Config) WithDefaults() Config {
internal/retry/retry.go:30:func IsRetryableStatus(status int) bool {
internal/retry/retry.go:34:func BackoffDelay(cfg Config, attempt int, retryAfter time.Duration) time.Duration {
internal/retry/retry.go:49:func Sleep(ctx context.Context, d time.Duration) error {
internal/retry/retry.go:63:func ParseRetryAfter(header string) (time.Duration, bool) {
internal/scheduler/runtime.go:17:func (s *Scheduler) cancelRuntime() {
internal/scheduler/runtime.go:27:func (s *Scheduler) runtimeContext() (context.Context, error) {
internal/scheduler/runtime.go:36:func (s *Scheduler) newRuntimeOpContext(timeout time.Duration) (context.Context, context.CancelFunc, error) {
internal/scheduler/runtime.go:49:func (s *Scheduler) withRuntimeOp(timeout time.Duration, fn func(context.Context) error) error {
internal/scheduler/runtime.go:60:func (s *Scheduler) periodicSent(idx int, key string) bool {
internal/scheduler/runtime.go:66:func (s *Scheduler) markPeriodicSent(idx int, key string) {
internal/scheduler/runtime.go:72:func (s *Scheduler) clearAdvanceTimers() {
internal/scheduler/runtime.go:81:func (s *Scheduler) currentTimezone() string {
internal/scheduler/runtime.go:89:func (s *Scheduler) NotionSyncStatus() SyncStatus {
internal/scheduler/runtime.go:8:func (s *Scheduler) setRuntimeContext(parent context.Context) {
internal/scheduler/runtime.go:95:func (s *Scheduler) setNotionStatus(count int, err error) {
internal/scheduler/worker.go:101:func (s *Scheduler) Reload() error {
internal/scheduler/worker.go:108:func (s *Scheduler) syncLoop() {
internal/scheduler/worker.go:130:func (s *Scheduler) periodicLoop() {
internal/scheduler/worker.go:176:func (s *Scheduler) calendarLoop() {
internal/scheduler/worker.go:207:func (s *Scheduler) SyncNotion() (int, error) {
internal/scheduler/worker.go:217:func (s *Scheduler) syncNotion(ctx context.Context) (int, error) {
internal/scheduler/worker.go:277:func (s *Scheduler) RebuildAdvanceSchedules() error {
internal/scheduler/worker.go:281:func (s *Scheduler) rebuildAdvanceSchedules(ctx context.Context) error {
internal/scheduler/worker.go:296:func (s *Scheduler) SchedulePendingFromDB() error {
internal/scheduler/worker.go:300:func (s *Scheduler) schedulePendingFromDB(ctx context.Context) error {
internal/scheduler/worker.go:336:func (s *Scheduler) fireAdvance(ctx context.Context, sched models.AdvanceSchedule) error {
internal/scheduler/worker.go:361:func (s *Scheduler) sendPeriodic(ctx context.Context, now time.Time, rule config.PeriodicNotification) error {
internal/scheduler/worker.go:378:func (s *Scheduler) SendManualNotification(ctx context.Context, template string, from, to time.Time) (string, error) {
internal/scheduler/worker.go:395:func (s *Scheduler) PreviewAdvanceTemplate(ctx context.Context, template string, minutesBefore int) (string, error) {
internal/scheduler/worker.go:411:func (s *Scheduler) PreviewManualTemplate(ctx context.Context, template string, from, to time.Time) (string, error) {
internal/scheduler/worker.go:425:func (s *Scheduler) SyncCalendar(from, to time.Time) (int, error) {
internal/scheduler/worker.go:435:func (s *Scheduler) syncCalendar(ctx context.Context, from, to time.Time) (int, error) {
internal/scheduler/worker.go:608:func groupCalendarEvents(events []calendar.CalendarEvent) map[string][]calendar.CalendarEvent {
internal/scheduler/worker.go:616:func pickPrimaryCalendarEvent(events []calendar.CalendarEvent, record models.SyncRecord, hasRecord bool) (calendar.CalendarEvent, []calendar.CalendarEvent) {
internal/scheduler/worker.go:649:func (s *Scheduler) deleteCalendarEvents(ctx context.Context, notionID string, events []calendar.CalendarEvent) int {
internal/scheduler/worker.go:65:func New(cfg *config.Manager, repo *db.Repository, notionClient *notion.Client, webhookClient *webhook.Client, calendarClient *calendar.Client, renderer *tpl.Renderer) *Scheduler {
internal/scheduler/worker.go:662:func (s *Scheduler) sendWebhook(ctx context.Context, typ, message string, events []models.TemplateEvent, minutesBefore int, notionPageID string) error {
internal/scheduler/worker.go:708:func buildAdvanceSchedules(events []models.Event, cfg config.Config, now time.Time, loc *time.Location) []models.AdvanceSchedule {
internal/scheduler/worker.go:736:func parseEventStart(ev models.Event, loc *time.Location) time.Time {
internal/scheduler/worker.go:757:func notionOnOrAfterDate(now time.Time, loc *time.Location) string {
internal/scheduler/worker.go:768:func matchAdvanceConditions(ev models.Event, rule config.AdvanceNotification, cfg config.Config) bool {
internal/scheduler/worker.go:786:func buildFilterValues(ev models.Event, cfg config.Config) map[string]string {
internal/scheduler/worker.go:78:func (s *Scheduler) Start(ctx context.Context) {
internal/scheduler/worker.go:798:func matchFilter(value, operator, expected string) bool {
internal/scheduler/worker.go:813:func buildTemplateEvents(events []models.Event, mapping config.PropertyMapping) []models.TemplateEvent {
internal/scheduler/worker.go:822:func extractCustomValues(raw string, mapping config.PropertyMapping) map[string]string {
internal/scheduler/worker.go:837:func toTemplateEvent(ev models.Event, custom map[string]string) models.TemplateEvent {
internal/scheduler/worker.go:852:func scheduleKey(notionPageID string, ruleIndex int) string {
internal/scheduler/worker.go:856:func weekdayToConfig(day time.Weekday) int {
internal/scheduler/worker.go:863:func matchesDays(days []int, weekday int) bool {
internal/scheduler/worker.go:95:func (s *Scheduler) Stop() {
internal/scheduler/worker_test.go:116:func TestNotionOnOrAfterDate_JSTEarlyMorningUsesPreviousUTCDate(t *testing.T) {
internal/scheduler/worker_test.go:127:func TestNotionOnOrAfterDate_PSTUsesSameUTCDate(t *testing.T) {
internal/scheduler/worker_test.go:138:func TestToTemplateEvent_MapsEndDateAndTime(t *testing.T) {
internal/scheduler/worker_test.go:155:func TestSendWebhookRecordsHistoryOnPayloadRenderError(t *testing.T) {
internal/scheduler/worker_test.go:21:func TestMatchesDays(t *testing.T) {
internal/scheduler/worker_test.go:38:func TestMatchAdvanceConditions(t *testing.T) {
internal/template/renderer.go:14:func New() *Renderer {
internal/template/renderer.go:18:func newTemplate(name string) *template.Template {
internal/template/renderer.go:27:func (r *Renderer) RenderSingle(tmpl string, event models.TemplateEvent, minutesBefore int) (string, error) {
internal/template/renderer.go:50:func (r *Renderer) RenderList(tmpl string, events []models.TemplateEvent) (string, error) {
internal/template/renderer.go:64:func (r *Renderer) RenderPayload(tmpl string, ctx any) (string, error) {
internal/webhook/client.go:20:func New(httpClient *http.Client, cfg retry.Config) *Client {
internal/webhook/client.go:27:func (c *Client) Send(ctx context.Context, webhookURL, contentType string, payload []byte) error {
```

## Go Functions (Production only)
```text
cmd/notion-notifier/main.go:16:func main() {
internal/app/app.go:110:func buildRouter(cfg *config.Manager, repo *db.Repository, sched *scheduler.Scheduler) http.Handler {
internal/app/app.go:133:func resolveServerRuntime(env config.Env) (string, tlsRuntime, error) {
internal/app/app.go:153:func generateSelfSignedCert() (tls.Certificate, error) {
internal/app/app.go:185:func (a *App) Start(ctx context.Context) error {
internal/app/app.go:206:func (a *App) Close() error {
internal/app/app.go:50:func New(cfgPath, envPath, dbPath string) (*App, error) {
internal/calendar/google.go:102:func (c *Client) DeleteEvent(ctx context.Context, eventID string) error {
internal/calendar/google.go:123:func (c *Client) ListEvents(ctx context.Context, from, to time.Time) ([]CalendarEvent, error) {
internal/calendar/google.go:176:func EventMatchesNotion(calEvent CalendarEvent, ev models.Event, tz *time.Location) bool {
internal/calendar/google.go:199:func sameDateOrDateTime(currentDate, currentDateTime, desiredDate, desiredDateTime string) bool {
internal/calendar/google.go:210:func normalizeDateTime(value string) string {
internal/calendar/google.go:221:func equalEmails(a, b []string) bool {
internal/calendar/google.go:233:func extractEmails(attendees []*calendarapi.EventAttendee) []string {
internal/calendar/google.go:260:func mapEvent(ev models.Event, tz *time.Location) *calendarapi.Event {
internal/calendar/google.go:286:func buildStartEnd(ev models.Event, tz *time.Location) (*calendarapi.EventDateTime, *calendarapi.EventDateTime) {
internal/calendar/google.go:33:func (o ClientOptions) normalize() ClientOptions {
internal/calendar/google.go:41:func (o ClientOptions) Validate() error {
internal/calendar/google.go:52:func (o ClientOptions) IsConfigured() bool {
internal/calendar/google.go:56:func (o ClientOptions) Fingerprint() string {
internal/calendar/google.go:67:func NewClient(ctx context.Context, opts ClientOptions) (*Client, error) {
internal/calendar/google.go:86:func (c *Client) UpsertEvent(ctx context.Context, ev models.Event, existingID string, tz *time.Location) (string, string, error) {
internal/config/config.go:150:func LoadConfig(path string) (Config, error) {
internal/config/config.go:166:func LoadEnv(path string) (Env, error) {
internal/config/config.go:181:func ApplyEnvOverrides(env Env) Env {
internal/config/config.go:199:func pickEnv(key, fallback string) string {
internal/config/config.go:206:func pickEnvBool(key string, fallback bool) bool {
internal/config/config.go:218:func pickEnvInt(key string, fallback int) int {
internal/config/config.go:230:func ValidateConfig(cfg Config) error {
internal/config/config.go:279:func validateHHMM(value string) error {
internal/config/config.go:287:func IsSnoozed(cfg Config, now time.Time) bool {
internal/config/config.go:298:func WriteConfig(path string, cfg Config) error {
internal/config/config.go:311:func NormalizeConfig(cfg Config) Config {
internal/config/config.go:363:func SanitizeTemplate(input string) string {
internal/config/config.go:396:func DefaultTemplates() map[string]string {
internal/config/manager.go:11:func (e ValidationError) Error() string {
internal/config/manager.go:18:func (e ValidationError) Unwrap() error {
internal/config/manager.go:30:func NewManager(cfgPath, envPath string) (*Manager, error) {
internal/config/manager.go:43:func (m *Manager) Snapshot() (Config, Env) {
internal/config/manager.go:49:func (m *Manager) Config() Config {
internal/config/manager.go:55:func (m *Manager) Env() Env {
internal/config/manager.go:61:func (m *Manager) UpdateConfig(cfg Config) (Config, error) {
internal/config/manager.go:75:func (m *Manager) Reload() error {
internal/db/db.go:110:func (r *Repository) ListUpcomingEvents(ctx context.Context, days int, now time.Time) ([]models.Event, error) {
internal/db/db.go:115:func (r *Repository) GetEvent(ctx context.Context, notionPageID string) (models.Event, bool, error) {
internal/db/db.go:139:func scanEvents(rows *sql.Rows) ([]models.Event, error) {
internal/db/db.go:161:func encodeStringSlice(values []string) string {
internal/db/db.go:172:func decodeStringSlice(raw string) []string {
internal/db/db.go:186:func (r *Repository) InsertNotificationHistory(ctx context.Context, h models.NotificationHistory) error {
internal/db/db.go:192:func (r *Repository) ListNotificationHistory(ctx context.Context, limit int) ([]models.NotificationHistory, error) {
internal/db/db.go:20:func Open(path string) (*Repository, error) {
internal/db/db.go:219:func (r *Repository) ClearNotificationHistory(ctx context.Context) error {
internal/db/db.go:224:func (r *Repository) ReplaceAdvanceSchedules(ctx context.Context, schedules []models.AdvanceSchedule) error {
internal/db/db.go:273:func (r *Repository) ListPendingAdvanceSchedules(ctx context.Context) ([]models.AdvanceSchedule, error) {
internal/db/db.go:297:func (r *Repository) MarkAdvanceScheduleFired(ctx context.Context, id int64) error {
internal/db/db.go:302:func (r *Repository) ListSyncRecords(ctx context.Context) ([]models.SyncRecord, error) {
internal/db/db.go:325:func (r *Repository) GetSyncRecordMap(ctx context.Context, notionPageIDs []string) (map[string]models.SyncRecord, error) {
internal/db/db.go:32:func (r *Repository) Close() error {
internal/db/db.go:356:func (r *Repository) UpsertSyncRecord(ctx context.Context, record models.SyncRecord) error {
internal/db/db.go:36:func (r *Repository) UpsertEvents(ctx context.Context, events []models.Event) error {
internal/db/db.go:380:func (r *Repository) ClearSyncRecords(ctx context.Context) error {
internal/db/db.go:386:func (r *Repository) ListOrphanedSyncRecords(ctx context.Context) ([]models.SyncRecord, error) {
internal/db/db.go:411:func (r *Repository) DeleteSyncRecord(ctx context.Context, notionPageID string) error {
internal/db/db.go:416:func (r *Repository) DeleteEventsNotIn(ctx context.Context, ids []string) error {
internal/db/db.go:97:func (r *Repository) ListEventsBetween(ctx context.Context, from, to time.Time) ([]models.Event, error) {
internal/db/schema.go:75:func migrateSyncRecords(db *sql.DB) error {
internal/db/schema.go:8:func initSchema(db *sql.DB) error {
internal/http/api/handler.go:161:func (h *Handler) handleUpcomingEvents(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:236:func (h *Handler) handleHistory(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:24:func NewHandler(cfg *config.Manager, repo *db.Repository, sched *scheduler.Scheduler) *Handler {
internal/http/api/handler.go:264:func (h *Handler) handleSync(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:286:func (h *Handler) handleCalendarSync(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:29:func (h *Handler) Register(mux *http.ServeMux) {
internal/http/api/handler.go:317:func (h *Handler) handleCalendarClear(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:335:func (h *Handler) handleHistoryClear(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:364:func (h *Handler) handlePreviewNotification(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:404:func (h *Handler) handleManualNotification(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:442:func (h *Handler) handleDefaultTemplates(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:45:func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:56:func (h *Handler) getConfig(w http.ResponseWriter, _ *http.Request) {
internal/http/api/handler.go:60:func (h *Handler) putConfig(w http.ResponseWriter, r *http.Request) {
internal/http/api/handler.go:98:func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
internal/http/api/helpers.go:18:func respondError(w http.ResponseWriter, status int, message string) {
internal/http/api/helpers.go:23:func respondValidationError(w http.ResponseWriter, message string, details map[string]string) {
internal/http/api/helpers.go:9:func respondJSON(w http.ResponseWriter, status int, data any) {
internal/http/api/timeutil.go:11:func parseDateRange(fromStr, toStr string, cfg *config.Manager) (time.Time, time.Time, error) {
internal/http/api/timeutil.go:46:func parseDateInput(value string, loc *time.Location) (time.Time, error) {
internal/http/api/timeutil.go:68:func formatDurationShort(d time.Duration) string {
internal/http/api/timeutil.go:87:func loadLocationOrLocal(name string) *time.Location {
internal/http/middleware/middleware.go:12:func Logging(next http.Handler) http.Handler {
internal/http/middleware/middleware.go:23:func BasicAuth(cfg *config.Manager) func(http.Handler) http.Handler {
internal/http/middleware/middleware.go:47:func (r *responseRecorder) WriteHeader(status int) {
internal/http/static/spa.go:12:func NewSPAHandler(distFS fs.FS) http.Handler {
internal/logging/logging.go:12:func Error(category, format string, args ...interface{}) {
internal/logging/logging.go:8:func Info(category, format string, args ...interface{}) {
internal/notion/client.go:115:func (c *Client) doRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
internal/notion/client.go:157:func MapPagesToEvents(pages []page, mapping config.PropertyMapping, tz *time.Location) []models.Event {
internal/notion/client.go:199:func parseDateRange(prop any, tz *time.Location) (startDate, startTime, endDate, endTime string, isAllDay bool) {
internal/notion/client.go:221:func splitDateTime(value string, tz *time.Location) (date string, tm string, allDay bool) {
internal/notion/client.go:243:func ExtractString(prop any) string {
internal/notion/client.go:356:func ExtractEmails(prop any) []string {
internal/notion/client.go:385:func joinRichText(value any) string {
internal/notion/client.go:430:func (c *Client) FetchContent(ctx context.Context, pageID string, rules config.ContentRules) (string, error) {
internal/notion/client.go:441:func (c *Client) listBlocks(ctx context.Context, blockID string) ([]block, error) {
internal/notion/client.go:466:func extractContentFromBlocks(blocks []block, rules config.ContentRules) string {
internal/notion/client.go:501:func blockInfo(b block) (text string, kind string, level int, isHeading bool, isDivider bool, checked bool) {
internal/notion/client.go:524:func joinBlockText(bt *blockText) string {
internal/notion/client.go:537:func headingMatches(text, start string) bool {
internal/notion/client.go:541:func formatHeading(level int, text string) string {
internal/notion/client.go:552:func formatBlock(kind string, level int, text string, checked bool) string {
internal/notion/client.go:57:func New(httpClient *http.Client, apiKey string, cfg retry.Config) *Client {
internal/notion/client.go:69:func (c *Client) QueryDatabase(ctx context.Context, databaseID string) ([]page, error) {
internal/notion/client.go:73:func (c *Client) QueryDatabaseOnOrAfter(ctx context.Context, databaseID, dateProperty, onOrAfter string) ([]page, error) {
internal/retry/retry.go:17:func (c Config) WithDefaults() Config {
internal/retry/retry.go:30:func IsRetryableStatus(status int) bool {
internal/retry/retry.go:34:func BackoffDelay(cfg Config, attempt int, retryAfter time.Duration) time.Duration {
internal/retry/retry.go:49:func Sleep(ctx context.Context, d time.Duration) error {
internal/retry/retry.go:63:func ParseRetryAfter(header string) (time.Duration, bool) {
internal/scheduler/runtime.go:17:func (s *Scheduler) cancelRuntime() {
internal/scheduler/runtime.go:27:func (s *Scheduler) runtimeContext() (context.Context, error) {
internal/scheduler/runtime.go:36:func (s *Scheduler) newRuntimeOpContext(timeout time.Duration) (context.Context, context.CancelFunc, error) {
internal/scheduler/runtime.go:49:func (s *Scheduler) withRuntimeOp(timeout time.Duration, fn func(context.Context) error) error {
internal/scheduler/runtime.go:60:func (s *Scheduler) periodicSent(idx int, key string) bool {
internal/scheduler/runtime.go:66:func (s *Scheduler) markPeriodicSent(idx int, key string) {
internal/scheduler/runtime.go:72:func (s *Scheduler) clearAdvanceTimers() {
internal/scheduler/runtime.go:81:func (s *Scheduler) currentTimezone() string {
internal/scheduler/runtime.go:89:func (s *Scheduler) NotionSyncStatus() SyncStatus {
internal/scheduler/runtime.go:8:func (s *Scheduler) setRuntimeContext(parent context.Context) {
internal/scheduler/runtime.go:95:func (s *Scheduler) setNotionStatus(count int, err error) {
internal/scheduler/worker.go:101:func (s *Scheduler) Reload() error {
internal/scheduler/worker.go:108:func (s *Scheduler) syncLoop() {
internal/scheduler/worker.go:130:func (s *Scheduler) periodicLoop() {
internal/scheduler/worker.go:176:func (s *Scheduler) calendarLoop() {
internal/scheduler/worker.go:207:func (s *Scheduler) SyncNotion() (int, error) {
internal/scheduler/worker.go:217:func (s *Scheduler) syncNotion(ctx context.Context) (int, error) {
internal/scheduler/worker.go:277:func (s *Scheduler) RebuildAdvanceSchedules() error {
internal/scheduler/worker.go:281:func (s *Scheduler) rebuildAdvanceSchedules(ctx context.Context) error {
internal/scheduler/worker.go:296:func (s *Scheduler) SchedulePendingFromDB() error {
internal/scheduler/worker.go:300:func (s *Scheduler) schedulePendingFromDB(ctx context.Context) error {
internal/scheduler/worker.go:336:func (s *Scheduler) fireAdvance(ctx context.Context, sched models.AdvanceSchedule) error {
internal/scheduler/worker.go:361:func (s *Scheduler) sendPeriodic(ctx context.Context, now time.Time, rule config.PeriodicNotification) error {
internal/scheduler/worker.go:378:func (s *Scheduler) SendManualNotification(ctx context.Context, template string, from, to time.Time) (string, error) {
internal/scheduler/worker.go:395:func (s *Scheduler) PreviewAdvanceTemplate(ctx context.Context, template string, minutesBefore int) (string, error) {
internal/scheduler/worker.go:411:func (s *Scheduler) PreviewManualTemplate(ctx context.Context, template string, from, to time.Time) (string, error) {
internal/scheduler/worker.go:425:func (s *Scheduler) SyncCalendar(from, to time.Time) (int, error) {
internal/scheduler/worker.go:435:func (s *Scheduler) syncCalendar(ctx context.Context, from, to time.Time) (int, error) {
internal/scheduler/worker.go:608:func groupCalendarEvents(events []calendar.CalendarEvent) map[string][]calendar.CalendarEvent {
internal/scheduler/worker.go:616:func pickPrimaryCalendarEvent(events []calendar.CalendarEvent, record models.SyncRecord, hasRecord bool) (calendar.CalendarEvent, []calendar.CalendarEvent) {
internal/scheduler/worker.go:649:func (s *Scheduler) deleteCalendarEvents(ctx context.Context, notionID string, events []calendar.CalendarEvent) int {
internal/scheduler/worker.go:65:func New(cfg *config.Manager, repo *db.Repository, notionClient *notion.Client, webhookClient *webhook.Client, calendarClient *calendar.Client, renderer *tpl.Renderer) *Scheduler {
internal/scheduler/worker.go:662:func (s *Scheduler) sendWebhook(ctx context.Context, typ, message string, events []models.TemplateEvent, minutesBefore int, notionPageID string) error {
internal/scheduler/worker.go:708:func buildAdvanceSchedules(events []models.Event, cfg config.Config, now time.Time, loc *time.Location) []models.AdvanceSchedule {
internal/scheduler/worker.go:736:func parseEventStart(ev models.Event, loc *time.Location) time.Time {
internal/scheduler/worker.go:757:func notionOnOrAfterDate(now time.Time, loc *time.Location) string {
internal/scheduler/worker.go:768:func matchAdvanceConditions(ev models.Event, rule config.AdvanceNotification, cfg config.Config) bool {
internal/scheduler/worker.go:786:func buildFilterValues(ev models.Event, cfg config.Config) map[string]string {
internal/scheduler/worker.go:78:func (s *Scheduler) Start(ctx context.Context) {
internal/scheduler/worker.go:798:func matchFilter(value, operator, expected string) bool {
internal/scheduler/worker.go:813:func buildTemplateEvents(events []models.Event, mapping config.PropertyMapping) []models.TemplateEvent {
internal/scheduler/worker.go:822:func extractCustomValues(raw string, mapping config.PropertyMapping) map[string]string {
internal/scheduler/worker.go:837:func toTemplateEvent(ev models.Event, custom map[string]string) models.TemplateEvent {
internal/scheduler/worker.go:852:func scheduleKey(notionPageID string, ruleIndex int) string {
internal/scheduler/worker.go:856:func weekdayToConfig(day time.Weekday) int {
internal/scheduler/worker.go:863:func matchesDays(days []int, weekday int) bool {
internal/scheduler/worker.go:95:func (s *Scheduler) Stop() {
internal/template/renderer.go:14:func New() *Renderer {
internal/template/renderer.go:18:func newTemplate(name string) *template.Template {
internal/template/renderer.go:27:func (r *Renderer) RenderSingle(tmpl string, event models.TemplateEvent, minutesBefore int) (string, error) {
internal/template/renderer.go:50:func (r *Renderer) RenderList(tmpl string, events []models.TemplateEvent) (string, error) {
internal/template/renderer.go:64:func (r *Renderer) RenderPayload(tmpl string, ctx any) (string, error) {
internal/webhook/client.go:20:func New(httpClient *http.Client, cfg retry.Config) *Client {
internal/webhook/client.go:27:func (c *Client) Send(ctx context.Context, webhookURL, contentType string, payload []byte) error {
```

## TypeScript Functions
```text
web/src/lib/api.ts:126:async function request<T>(path: string, options?: RequestInit): Promise<T> {
web/src/lib/store.ts:14:export function addToast(message: string, type: Toast['type'] = 'info') {
web/src/lib/store.ts:29:export function navigate(path: string) {
web/src/lib/store.ts:37:export function setDarkMode(value: boolean) {
```

## Svelte Functions
```text
web/src/App.svelte:111:  async function saveSnooze() {
web/src/App.svelte:122:  async function clearSnooze() {
web/src/App.svelte:129:  function toggleDarkMode() {
web/src/App.svelte:137:  function openGuideModal(event: CustomEvent<{ title: string; content: string }>) {
web/src/App.svelte:69:  function updateClock() {
web/src/App.svelte:73:  function closeSidebar() {
web/src/App.svelte:77:  function openSidebar() {
web/src/App.svelte:81:  function handleGlobalKeydown(event: KeyboardEvent) {
web/src/App.svelte:87:  async function checkHealth() {
web/src/App.svelte:97:  async function handleSync() {
web/src/components/PreviewModal.svelte:27:    function renderMarkdown(source: string, currentMode: "webhook" | "guide"): string {
web/src/components/PreviewModal.svelte:43:    function close() {
web/src/components/PreviewModal.svelte:47:    function handleKeydown(event: KeyboardEvent) {
web/src/components/TemplateGuideSidebar.svelte:97:    function openGuideDetail() {
web/src/routes/Calendar.svelte:27:    async function handleConfigUpdate() {
web/src/routes/Calendar.svelte:38:    async function handleSync() {
web/src/routes/Calendar.svelte:53:    async function handleClear() {
web/src/routes/Dashboard.svelte:102:    async function handleSync() {
web/src/routes/Dashboard.svelte:115:    async function handleManualPreview() {
web/src/routes/Dashboard.svelte:135:    async function handleManualSend() {
web/src/routes/Dashboard.svelte:160:    async function loadDefaultTemplate() {
web/src/routes/Dashboard.svelte:45:    function openPreview(title: string, content: string) {
web/src/routes/Dashboard.svelte:61:    function formatEventDateTime(event: UpcomingEvent): string {
web/src/routes/Dashboard.svelte:82:    async function loadData() {
web/src/routes/History.svelte:20:    async function loadHistory() {
web/src/routes/History.svelte:33:    async function handleClear() {
web/src/routes/History.svelte:49:    function formatDate(isoString: string) {
web/src/routes/Notifications.svelte:100:    async function previewTemplate(
web/src/routes/Notifications.svelte:116:    async function resetAdvanceTemplate(index: number) {
web/src/routes/Notifications.svelte:129:    async function resetPeriodicTemplate(index: number) {
web/src/routes/Notifications.svelte:145:    function toggleDay(list: number[], day: number) {
web/src/routes/Notifications.svelte:27:    function openPreview(title: string, content: string) {
web/src/routes/Notifications.svelte:33:    async function saveConfig() {
web/src/routes/Notifications.svelte:50:    function addAdvanceRule() {
web/src/routes/Notifications.svelte:68:    function addPeriodicRule() {
web/src/routes/Notifications.svelte:84:    function removeAdvanceRule(index: number) {
web/src/routes/Notifications.svelte:92:    function removePeriodicRule(index: number) {
web/src/routes/Settings.svelte:21:    async function saveConfig() {
web/src/routes/Settings.svelte:35:    function addCustomMapping() {
web/src/routes/Settings.svelte:44:    function removeCustomMapping(index: number) {
```

## Script Functions
```text
scripts/deploy-mac.sh:8:usage() {
```
