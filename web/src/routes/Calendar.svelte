<script lang="ts">
    import { onMount } from "svelte";
    import { api, type Config } from "../lib/api";
    import { configStore, addToast } from "../lib/store";
    import {
        Calendar,
        RefreshCw,
        Trash2,
        CalendarDays,
        Settings,
        ArrowRight,
        Search,
    } from "lucide-svelte";

    let config: Config | null = null;
    configStore.subscribe((v) => (config = v));

    let isSyncing = false;
    let isClearing = false;
    let syncRange = {
        from: new Date().toISOString().split("T")[0],
        to: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)
            .toISOString()
            .split("T")[0],
    };

    async function handleConfigUpdate() {
        if (!config) return;
        try {
            const saved = await api.updateConfig(config);
            configStore.set(saved);
            addToast("設定を保存しました", "success");
        } catch (e) {
            addToast("保存に失敗しました", "error");
        }
    }

    async function handleSync() {
        isSyncing = true;
        try {
            const res = await api.syncCalendar(syncRange.from, syncRange.to);
            addToast(
                `${res.count}件の予定をカレンダーに同期しました`,
                "success",
            );
        } catch (e: any) {
            addToast(`同期失敗: ${e.error || "不明なエラー"}`, "error");
        } finally {
            isSyncing = false;
        }
    }

    async function handleClear() {
        if (
            !confirm(
                "全ての同期記録を削除しますか？Google カレンダー自体の予定は削除されませんが、次回の同期で再作成される可能性があります。",
            )
        )
            return;
        isClearing = true;
        try {
            await api.clearCalendarSync();
            addToast("同期記録を削除しました", "success");
        } catch (e) {
            addToast("削除に失敗しました", "error");
        } finally {
            isClearing = false;
        }
    }
</script>

<div class="space-y-8 max-w-4xl">
    <!-- Status Header -->
    <div
        class="bg-white p-8 rounded-3xl border border-gray-100 shadow-sm flex flex-col md:flex-row items-center gap-8"
    >
        <div
            class="w-16 h-16 bg-blue-50 rounded-2xl flex items-center justify-center text-blue-600"
        >
            <Calendar size={32} />
        </div>
        <div class="flex-1 text-center md:text-left">
            <h2 class="text-2xl font-bold text-gray-900 mb-2">
                Google カレンダー同期
            </h2>
            <p class="text-gray-500">
                Notionを正としてGoogle
                カレンダーを同期します。カレンダー側の意図しない編集は次回同期でNotion内容に戻されます。
            </p>
        </div>
        <div class="flex items-center gap-3">
            <span class="text-sm font-bold text-gray-400">同期有効化</span>
            {#if config}
                <button
                    on:click={() => {
                        if (!config) return;
                        config.calendar_sync.enabled =
                            !config.calendar_sync.enabled;
                        handleConfigUpdate();
                    }}
                    class="w-14 h-8 rounded-full transition-all duration-300 flex items-center p-1 {config
                        .calendar_sync.enabled
                        ? 'bg-green-500'
                        : 'bg-gray-200'}"
                >
                    <div
                        class="w-6 h-6 bg-white rounded-full shadow-sm transform transition-transform duration-300 {config
                            .calendar_sync.enabled
                            ? 'translate-x-6'
                            : 'translate-x-0'}"
                    ></div>
                </button>
            {/if}
        </div>
    </div>

    {#if config}
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <!-- Sync Configuration -->
            <div class="space-y-6">
                <h3
                    class="text-lg font-bold text-gray-800 flex items-center gap-2"
                >
                    <Settings size={20} class="text-gray-400" />
                    同期設定
                </h3>

                <div
                    class="bg-white p-6 rounded-2xl border border-gray-100 space-y-4"
                >
                    <div>
                        <label
                            for="cal-interval-hours"
                            class="block text-sm font-bold text-gray-700 mb-2"
                            >実行間隔 (時間)</label
                        >
                        <input
                            id="cal-interval-hours"
                            type="number"
                            bind:value={config.calendar_sync.interval_hours}
                            on:change={handleConfigUpdate}
                            class="w-full px-4 py-2 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand-500 transition-all"
                        />
                    </div>
                    <div>
                        <label
                            for="cal-lookahead-days"
                            class="block text-sm font-bold text-gray-700 mb-2"
                            >同期先読み日数</label
                        >
                        <input
                            id="cal-lookahead-days"
                            type="number"
                            bind:value={config.calendar_sync.lookahead_days}
                            on:change={handleConfigUpdate}
                            class="w-full px-4 py-2 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand-500 transition-all"
                        />
                    </div>
                    <div
                        class="pt-4 border-t border-gray-50 flex items-center justify-between text-xs text-gray-400"
                    >
                        <span
                            >同期を無効にすると、スケジューラによる自動実行が停止します。</span
                        >
                    </div>
                </div>
            </div>

            <!-- Manual Sync -->
            <div class="space-y-6">
                <h3
                    class="text-lg font-bold text-gray-800 flex items-center gap-2"
                >
                    <RefreshCw size={20} class="text-gray-400" />
                    手動同期・管理
                </h3>

                <div
                    class="bg-white p-6 rounded-2xl border border-gray-100 space-y-6"
                >
                    <div class="grid grid-cols-2 gap-4">
                        <div>
                            <label
                                for="cal-sync-from"
                                class="block text-xs font-bold text-gray-500 uppercase mb-2"
                                >開始日</label
                            >
                            <input
                                id="cal-sync-from"
                                type="date"
                                bind:value={syncRange.from}
                                class="w-full px-4 py-2 bg-gray-50 border border-gray-200 rounded-xl text-sm"
                            />
                        </div>
                        <div>
                            <label
                                for="cal-sync-to"
                                class="block text-xs font-bold text-gray-500 uppercase mb-2"
                                >終了日</label
                            >
                            <input
                                id="cal-sync-to"
                                type="date"
                                bind:value={syncRange.to}
                                class="w-full px-4 py-2 bg-gray-50 border border-gray-200 rounded-xl text-sm"
                            />
                        </div>
                    </div>

                    <button
                        on:click={handleSync}
                        disabled={isSyncing}
                        class="w-full py-4 bg-brand-600 text-white rounded-2xl font-bold flex items-center justify-center gap-2 hover:bg-brand-700 transition-all active:scale-95 disabled:opacity-50"
                    >
                        {#if isSyncing}
                            <RefreshCw size={20} class="animate-spin" />
                            同期中...
                        {:else}
                            <RefreshCw size={20} />
                            指定範囲で同期実行
                        {/if}
                    </button>

                    <div class="pt-6 border-t border-gray-100">
                        <button
                            on:click={handleClear}
                            disabled={isClearing}
                            class="w-full py-3 border border-red-100 text-red-600 rounded-2xl text-sm font-bold flex items-center justify-center gap-2 hover:bg-red-50 transition-all"
                        >
                            <Trash2 size={16} />
                            同期記録をリセット
                        </button>
                        <p class="mt-2 text-[11px] text-gray-400 text-center">
                            カレンダーで重複が発生する場合や、再同期したい場合に実行してください。
                        </p>
                    </div>
                </div>
            </div>
        </div>
    {/if}

    <!-- Help Section -->
    <div class="bg-gray-100 p-6 rounded-2xl flex items-start gap-4">
        <div
            class="w-10 h-10 bg-white rounded-xl flex items-center justify-center text-gray-400"
        >
            <CalendarDays size={20} />
        </div>
        <div>
            <h4 class="font-bold text-gray-900 mb-1">同期の仕組み</h4>
            <p class="text-sm text-gray-600 leading-relaxed">
                Calendar APIの追跡イベント（notion_page_id付き）を起点にDBへ逆引きし、差分があれば統一ロジックで更新します。
                Notionにない追跡イベントは削除し、重複イベントは1件に整理します。設定の「Notion
                プロパティマッピング」で、タイトル・日付・場所・参加者（メールアドレス）を紐付けてください。
            </p>
        </div>
    </div>
</div>
