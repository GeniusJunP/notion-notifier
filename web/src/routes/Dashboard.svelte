<script lang="ts">
    import { onMount } from "svelte";
    import { api, type DashboardData, type UpcomingEvent } from "../lib/api";
    import { addToast } from "../lib/store";
    import PreviewModal from "../components/PreviewModal.svelte";
    import {
        RefreshCw,
        CalendarDays,
        ArrowRight,
        AlertTriangle,
        Clock,
        CheckCircle,
        Send,
        Play,
        RotateCcw,
    } from "lucide-svelte";

    let dashboard: DashboardData | null = null;
    let upcoming: UpcomingEvent[] = [];
    let isLoading = true;
    let isSyncing = false;

    // Manual notification state
    let manualTemplate = "";
    let manualFromDate = new Date().toISOString().split("T")[0];
    let manualToDate = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)
        .toISOString()
        .split("T")[0];
    let isPreviewLoading = false;
    let isSending = false;
    let previewOpen = false;
    let previewTitle = "";
    let previewContent = "";

    function openPreview(title: string, content: string) {
        previewTitle = title;
        previewContent = content;
        previewOpen = true;
    }

    const calendarStateMeta: Record<
        UpcomingEvent["calendar_state"],
        { className: string; label: string }
    > = {
        disabled: { className: "bg-gray-100 text-gray-600", label: "連携オフ" },
        needs_sync: { className: "bg-amber-100 text-amber-700", label: "要同期" },
        synced: { className: "bg-green-100 text-green-700", label: "反映済み" },
        error: { className: "bg-red-100 text-red-700", label: "連携エラー" },
    };

    function formatEventDateTime(event: UpcomingEvent): string {
        const startDate = event.start_date;
        const endDate = event.end_date || event.start_date;
        if (event.is_all_day) {
            if (endDate && endDate !== startDate) {
                return `${startDate} - ${endDate} (終日)`;
            }
            return `${startDate} (終日)`;
        }
        const start = event.start_time
            ? `${startDate} ${event.start_time}`
            : startDate;
        if (!event.end_time) {
            return start;
        }
        if (endDate && endDate !== startDate) {
            return `${start} - ${endDate} ${event.end_time}`;
        }
        return `${start} - ${event.end_time}`;
    }

    async function loadData() {
        isLoading = true;
        try {
            const [d, u] = await Promise.all([
                api.getDashboard(),
                api.getUpcomingEvents(),
            ]);
            dashboard = d;
            upcoming = u;
        } catch (e) {
            addToast("データの取得に失敗しました", "error");
        } finally {
            isLoading = false;
        }
    }

    onMount(async () => {
        await loadData();
        // Load default template
        try {
            const defaults = await api.getDefaultTemplates();
            if (defaults.periodic && !manualTemplate) {
                manualTemplate = defaults.periodic;
            }
        } catch {}
    });

    async function handleSync() {
        isSyncing = true;
        try {
            const res = await api.syncNotion();
            addToast(`${res.count}件の項目を同期しました`, "success");
            await loadData();
        } catch (e) {
            addToast("同期に失敗しました", "error");
        } finally {
            isSyncing = false;
        }
    }

    async function handleManualPreview() {
        if (!manualTemplate.trim()) {
            addToast("テンプレートを入力してください", "error");
            return;
        }
        isPreviewLoading = true;
        try {
            const res = await api.previewNotification({
                template: manualTemplate,
                from_date: manualFromDate,
                to_date: manualToDate,
            });
            openPreview("手動通知プレビュー", res.message);
        } catch (e: any) {
            addToast(`プレビュー失敗: ${e.error || "不明なエラー"}`, "error");
        } finally {
            isPreviewLoading = false;
        }
    }

    async function handleManualSend() {
        if (!manualTemplate.trim()) {
            addToast("テンプレートを入力してください", "error");
            return;
        }
        if (!confirm("手動通知を送信しますか？")) return;
        isSending = true;
        try {
            const res = await api.sendManualNotification({
                template: manualTemplate,
                from_date: manualFromDate,
                to_date: manualToDate,
            });
            addToast("通知を送信しました", "success");
            openPreview("送信メッセージ", res.message);
        } catch (e: any) {
            addToast(`送信失敗: ${e.error || "不明なエラー"}`, "error");
        } finally {
            isSending = false;
        }
    }

    async function loadDefaultTemplate() {
        try {
            const defaults = await api.getDefaultTemplates();
            manualTemplate = defaults.periodic || "";
            addToast("デフォルトテンプレートを読み込みました", "info");
        } catch {
            addToast("デフォルトテンプレートの取得に失敗しました", "error");
        }
    }

</script>

<div class="space-y-8">
    <!-- Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div
            class="bg-white dark:bg-gray-800 p-6 rounded-2xl shadow-sm dark:shadow-md border border-gray-100 dark:border-gray-700 flex flex-col justify-between"
        >
            <div class="flex items-center justify-between mb-4">
                <span class="text-sm font-medium text-gray-500 dark:text-gray-400">本日の予定</span
                >
                <div
                    class="w-10 h-10 bg-brand-50 dark:bg-brand-900/20 rounded-xl flex items-center justify-center text-brand-600 dark:text-brand-300"
                >
                    <CalendarDays size={20} />
                </div>
            </div>
            <div>
                <h3 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
                    {dashboard?.today_count ?? 0}
                </h3>
                <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">Notion 内予定数</p>
            </div>
        </div>

        <div
            class="bg-white dark:bg-gray-800 p-6 rounded-2xl shadow-sm dark:shadow-md border border-gray-100 dark:border-gray-700 flex flex-col justify-between"
        >
            <div class="flex items-center justify-between mb-4">
                <span class="text-sm font-medium text-gray-500 dark:text-gray-400">次回同期</span>
                <div
                    class="w-10 h-10 bg-orange-50 dark:bg-orange-900/20 rounded-xl flex items-center justify-center text-orange-600 dark:text-orange-300"
                >
                    <RefreshCw size={20} />
                </div>
            </div>
            <div>
                <h3 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
                    {dashboard?.next_sync_in ?? "--"}
                </h3>
                <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">
                    {dashboard?.next_sync
                        ? new Date(dashboard.next_sync).toLocaleTimeString()
                        : "Scheduled"}
                </p>
            </div>
        </div>

        <div
            class="bg-white dark:bg-gray-800 p-6 rounded-2xl shadow-sm dark:shadow-md border border-gray-100 dark:border-gray-700 flex flex-col justify-between"
        >
            <div class="flex items-center justify-between mb-4">
                <span class="text-sm font-medium text-gray-500 dark:text-gray-400"
                    >最終同期件数</span
                >
                <div
                    class="w-10 h-10 bg-green-50 dark:bg-green-900/20 rounded-xl flex items-center justify-center text-green-600 dark:text-green-300"
                >
                    <CheckCircle size={20} />
                </div>
            </div>
            <div>
                <h3 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
                    {dashboard?.last_sync_count ?? 0}
                </h3>
                <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">同期済み件数</p>
            </div>
        </div>

        <div
            class="bg-white dark:bg-gray-800 p-6 rounded-2xl shadow-sm dark:shadow-md border border-gray-100 dark:border-gray-700 flex flex-col justify-between"
        >
            <div class="flex items-center justify-between mb-4">
                <span class="text-sm font-medium text-gray-500 dark:text-gray-400">ステータス</span
                >
                <div
                    class="w-10 h-10 {dashboard?.last_sync_error
                        ? 'bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-300'
                        : 'bg-green-50 dark:bg-green-900/20 text-green-600 dark:text-green-300'} rounded-xl flex items-center justify-center"
                >
                    {#if dashboard?.last_sync_error}
                        <AlertTriangle size={20} />
                    {:else}
                        <CheckCircle size={20} />
                    {/if}
                </div>
            </div>
            <div>
                <h3 class="text-lg font-bold text-gray-900 dark:text-gray-100">
                    {dashboard?.last_sync_error ? "エラー" : "正常"}
                </h3>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 truncate">
                    {dashboard?.last_sync_error || "問題は検出されていません"}
                </p>
            </div>
        </div>
    </div>

    <!-- Actions -->
    <div
        class="bg-brand-600 p-8 rounded-3xl shadow-xl shadow-brand-100 dark:shadow-brand-900 text-white flex flex-col md:flex-row items-center justify-between gap-6 overflow-hidden relative"
    >
        <div class="relative z-10">
            <h2 class="text-2xl font-bold mb-2">手動同期を実行</h2>
            <p class="text-brand-100 max-w-md">
                Notion
                データベースから最新の変更を即座に取得し、カレンダーや通知スケジュールを更新します。
            </p>
        </div>
        <button
            on:click={handleSync}
            disabled={isSyncing}
            class="relative z-10 px-8 py-4 bg-white dark:bg-gray-800 text-brand-600 dark:text-brand-300 rounded-2xl font-bold shadow-lg dark:shadow-xl hover:bg-brand-50 dark:hover:bg-brand-900/20 transition-all flex items-center gap-3 active:scale-95 disabled:opacity-70 disabled:pointer-events-none"
        >
            {#if isSyncing}
                <RefreshCw size={20} class="animate-spin" />
                同期中...
            {:else}
                <RefreshCw size={20} />
                今すぐ同期
            {/if}
        </button>

        <!-- Abstract pattern -->
        <div
            class="absolute right-0 top-0 w-64 h-64 bg-white dark:bg-gray-800 opacity-[0.05] rounded-full translate-x-1/2 -translate-y-1/2"
        ></div>
        <div
            class="absolute left-0 bottom-0 w-32 h-32 bg-white dark:bg-gray-800 opacity-[0.05] rounded-full -translate-x-1/2 translate-y-1/2"
        ></div>
    </div>

    <!-- Manual Notification -->
    <div
        class="bg-white dark:bg-gray-800 rounded-3xl border border-gray-100 dark:border-gray-700 shadow-sm dark:shadow-md overflow-hidden"
    >
        <div
            class="p-6 border-b border-gray-50 dark:border-gray-700 flex items-center justify-between bg-gray-50/50 dark:bg-gray-700/50"
        >
            <div class="flex items-center gap-3">
                <div
                    class="w-10 h-10 bg-orange-50 dark:bg-orange-900/20 rounded-xl flex items-center justify-center text-orange-600 dark:text-orange-300"
                >
                    <Send size={20} />
                </div>
                <div>
                    <h2 class="text-lg font-bold text-gray-900 dark:text-gray-100">手動通知</h2>
                    <p class="text-xs text-gray-400">
                        テンプレートを使用して即座にWebhook通知を送信
                    </p>
                </div>
            </div>
            <button
                on:click={loadDefaultTemplate}
                class="text-xs font-bold text-gray-500 dark:text-gray-400 flex items-center gap-1 hover:text-brand-600 dark:hover:text-brand-300 transition-colors"
            >
                <RotateCcw size={12} />
                デフォルトに戻す
            </button>
        </div>

        <div class="p-6 space-y-5">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                    <label
                        for="manual-from-date"
                        class="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase mb-2"
                        >開始日</label
                    >
                    <input
                        id="manual-from-date"
                        type="date"
                        bind:value={manualFromDate}
                        class="w-full px-4 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
                    />
                </div>
                <div>
                    <label
                        for="manual-to-date"
                        class="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase mb-2"
                        >終了日</label
                    >
                    <input
                        id="manual-to-date"
                        type="date"
                        bind:value={manualToDate}
                        class="w-full px-4 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
                    />
                </div>
            </div>

            <div>
                <label
                    for="manual-template"
                    class="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase mb-2"
                    >メッセージテンプレート</label
                >
                <textarea
                    id="manual-template"
                    bind:value={manualTemplate}
                    placeholder="Go テンプレート形式で入力..."
                    class="w-full p-4 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all font-mono text-sm min-h-[120px]"
                ></textarea>
            </div>

            <div class="flex items-center gap-3">
                <button
                    on:click={handleManualPreview}
                    disabled={isPreviewLoading}
                    class="flex-1 py-3 bg-gray-100 text-gray-700 rounded-xl font-bold flex items-center justify-center gap-2 hover:bg-gray-200 transition-all active:scale-95 disabled:opacity-50"
                >
                    {#if isPreviewLoading}
                        <div
                            class="w-4 h-4 border-2 border-gray-400/30 border-t-gray-600 rounded-full animate-spin"
                        ></div>
                    {:else}
                        <Play size={16} />
                    {/if}
                    プレビュー
                </button>
                <button
                    on:click={handleManualSend}
                    disabled={isSending}
                    class="flex-1 py-3 bg-brand-600 text-white rounded-xl font-bold flex items-center justify-center gap-2 hover:bg-brand-700 transition-all active:scale-95 disabled:opacity-50 shadow-lg dark:shadow-xl shadow-brand-100 dark:shadow-brand-900"
                >
                    {#if isSending}
                        <div
                            class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
                        ></div>
                    {:else}
                        <Send size={16} />
                    {/if}
                    通知を送信
                </button>
            </div>

        </div>
    </div>

    <!-- Upcoming Events -->
    <div class="space-y-4">
        <div class="flex items-center justify-between">
            <h2 class="text-xl font-bold text-gray-800 dark:text-gray-100">直近の予定 (14日間)</h2>
            <button
                on:click={loadData}
                class="text-sm text-brand-600 font-medium hover:underline flex items-center gap-1"
            >
                <RefreshCw size={14} />
                更新
            </button>
        </div>

        {#if isLoading && !isSyncing}
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                {#each Array(4) as _}
                    <div
                        class="bg-white dark:bg-gray-800 p-6 rounded-2xl border border-gray-100 dark:border-gray-700 animate-pulse h-32"
                    ></div>
                {/each}
            </div>
        {:else if upcoming.length === 0}
            <div
                class="bg-white dark:bg-gray-800 p-12 rounded-3xl border border-dashed border-gray-200 dark:border-gray-600 text-center"
            >
                <div
                    class="w-16 h-16 bg-gray-50 dark:bg-gray-700 rounded-2xl flex items-center justify-center text-gray-300 dark:text-gray-500 mx-auto mb-4"
                >
                    <CalendarDays size={32} />
                </div>
                <h3 class="text-lg font-bold text-gray-900 dark:text-gray-100 mb-1">
                    予定が見つかりません
                </h3>
                <p class="text-gray-500 dark:text-gray-400 max-w-sm mx-auto">
                    同期された直近の予定はありません。Notion
                    データベースを確認してください。
                </p>
            </div>
        {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
                {#each upcoming as event}
                    <div
                        class="bg-white dark:bg-gray-800 p-5 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm dark:shadow-md hover:shadow-md dark:hover:shadow-lg transition-shadow group flex flex-col justify-between min-h-[140px]"
                    >
                        <div>
                            <div
                                class="flex items-start justify-between gap-3 mb-2"
                            >
                                <h3
                                    class="font-bold text-gray-900 dark:text-gray-100 line-clamp-2 leading-tight group-hover:text-brand-600 dark:group-hover:text-brand-300 transition-colors"
                                >
                                    {event.title}
                                </h3>
                                <div class="flex flex-col items-end gap-1">
                                    <span
                                        class={`px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider ${calendarStateMeta[event.calendar_state].className}`}
                                    >
                                        {calendarStateMeta[event.calendar_state].label}
                                    </span>
                                </div>
                            </div>
                            <div class="space-y-1.5">
                                <div
                                    class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400"
                                >
                                    <Clock size={14} class="text-gray-400 dark:text-gray-500" />
                                    <span>{formatEventDateTime(event)}</span>
                                </div>
                                {#if event.location}
                                    <div
                                        class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400"
                                    >
                                        <div
                                            class="w-3.5 flex items-center justify-center"
                                        >
                                            <div
                                                class="w-1 h-3.5 bg-brand-400 dark:bg-brand-500 rounded-full"
                                            ></div>
                                        </div>
                                        <span class="truncate"
                                            >{event.location}</span
                                        >
                                    </div>
                                {/if}
                            </div>
                        </div>

                        <div
                            class="mt-4 pt-3 border-t border-gray-50 flex items-center justify-between"
                        >
                            <a
                                href={event.url}
                                target="_blank"
                                rel="noopener noreferrer"
                                class="text-xs text-brand-600 font-bold flex items-center gap-1 hover:underline"
                            >
                                Notion で開く
                                <ArrowRight size={12} />
                            </a>
                        </div>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>

<PreviewModal
    open={previewOpen}
    title={previewTitle}
    content={previewContent}
    on:close={() => (previewOpen = false)}
/>
