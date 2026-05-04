<script lang="ts">
    import { api, getErrorMessage } from "../lib/api";
    import {
        configStore,
        addToast,
    } from "../lib/store";
    import {
        Calendar,
        RefreshCw,
        Trash2,
        CalendarDays,
        Settings,
    } from "lucide-svelte";
    import Button from "../lib/ui/Button.svelte";
    import Card from "../lib/ui/Card.svelte";
    import FormField from "../lib/ui/FormField.svelte";
    import IconChip from "../lib/ui/IconChip.svelte";
    import Input from "../lib/ui/Input.svelte";
    import Toggle from "../lib/ui/Toggle.svelte";
    import { toLocalDateInputValue } from "../lib/utils";
    import { onMount, onDestroy } from "svelte";

    let config = $configStore;
    let unsubscribe: () => void;

    onMount(() => {
        unsubscribe = configStore.subscribe((value) => {
            config = value;
        });
    });

    onDestroy(() => {
        if (unsubscribe) unsubscribe();
    });

    let isSyncing = false;
    let isClearing = false;
    let syncRange = {
        from: toLocalDateInputValue(new Date()),
        to: toLocalDateInputValue(new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)),
    };

    async function handleSync() {
        isSyncing = true;
        try {
            const res = await api.syncCalendar(syncRange.from, syncRange.to);
            addToast(
                `${res.count}件の予定をカレンダーに同期しました`,
                "success",
            );
        } catch (e: unknown) {
            addToast(`同期失敗: ${getErrorMessage(e)}`, "error");
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
        } catch {
            addToast("削除に失敗しました", "error");
        } finally {
            isClearing = false;
        }
    }
</script>

<div class="max-w-4xl space-y-8">
    <Card radius="3xl" padding="lg" class="flex flex-col items-center gap-8 md:flex-row">
        <IconChip tone="brand" size="lg">
            <Calendar size={32} />
        </IconChip>
        <div class="flex-1 text-center md:text-left">
            <h2 class="ui-page-title mb-2">
                Google カレンダー同期
            </h2>
            <p class="ui-support-text">
                Notionを正としてGoogle
                カレンダーを同期します。カレンダー側の意図しない編集は次回同期でNotion内容に戻されます。
            </p>
        </div>
        <div class="flex items-center gap-3">
            <span class="text-sm font-semibold text-gray-500 dark:text-gray-400">
                同期有効化
            </span>
            {#if config}
                <Toggle
                    checked={config?.calendar_sync.enabled ?? false}
                    on:change={(e) => {
                        if (!config) return;
                        const target = e.currentTarget as HTMLInputElement;
                        configStore.update((cfg) => cfg ? {
                            ...cfg,
                            calendar_sync: { ...cfg.calendar_sync, enabled: target.checked }
                        } : null);
                    }}
                    ariaLabel="カレンダー同期の有効化を切り替え"
                    tone="success"
                />
            {/if}
        </div>
    </Card>

    {#if config}
        <div class="grid grid-cols-1 gap-8 lg:grid-cols-2">
            <div class="space-y-6">
                <h3 class="ui-section-title">
                    <Settings size={20} class="text-gray-400 dark:text-gray-500" />
                    同期設定
                </h3>

                <Card tone="muted" class="space-y-4" radius="2xl">
                    <FormField label="実行間隔 (時間)" forId="cal-interval-hours">
                        <Input
                            id="cal-interval-hours"
                            type="number"
                            value={config?.calendar_sync.interval_hours ?? 0}
                            on:input={(e) => {
                                if (!config) return;
                                const target = e.currentTarget as HTMLInputElement;
                                configStore.update((cfg) => cfg ? {
                                    ...cfg,
                                    calendar_sync: { ...cfg.calendar_sync, interval_hours: parseInt(target.value) || 0 }
                                } : null);
                            }}
                        />
                    </FormField>

                    <FormField label="同期先読み日数" forId="cal-lookahead-days">
                        <Input
                            id="cal-lookahead-days"
                            type="number"
                            value={config?.calendar_sync.lookahead_days ?? 0}
                            on:input={(e) => {
                                if (!config) return;
                                const target = e.currentTarget as HTMLInputElement;
                                configStore.update((cfg) => cfg ? {
                                    ...cfg,
                                    calendar_sync: { ...cfg.calendar_sync, lookahead_days: parseInt(target.value) || 0 }
                                } : null);
                            }}
                        />
                    </FormField>

                    <div
                        class="flex items-center justify-between border-t border-gray-200/70 pt-4 text-xs text-gray-500 dark:border-gray-800 dark:text-gray-400"
                    >
                        <span>
                            同期を無効にすると、スケジューラによる自動実行が停止します。
                        </span>
                    </div>
                </Card>
            </div>

            <div class="space-y-6">
                <h3 class="ui-section-title">
                    <RefreshCw size={20} class="text-gray-400 dark:text-gray-500" />
                    手動同期・管理
                </h3>

                <Card tone="muted" class="space-y-6" radius="2xl">
                    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                        <FormField
                            label="開始日"
                            forId="cal-sync-from"
                            variant="eyebrow"
                        >
                            <Input
                                id="cal-sync-from"
                                type="date"
                                bind:value={syncRange.from}
                                uiSize="sm"
                            />
                        </FormField>

                        <FormField
                            label="終了日"
                            forId="cal-sync-to"
                            variant="eyebrow"
                        >
                            <Input
                                id="cal-sync-to"
                                type="date"
                                bind:value={syncRange.to}
                                uiSize="sm"
                            />
                        </FormField>
                    </div>

                    <Button
                        on:click={handleSync}
                        disabled={isSyncing}
                        loading={isSyncing}
                        block
                        size="lg"
                    >
                        {#if !isSyncing}
                            <RefreshCw size={20} />
                        {/if}
                        {isSyncing ? "同期中..." : "指定範囲で同期実行"}
                    </Button>

                    <div class="border-t border-gray-200/70 pt-6 dark:border-gray-800">
                        <Button
                            on:click={handleClear}
                            disabled={isClearing}
                            block
                            size="md"
                            variant="danger"
                        >
                            <Trash2 size={16} />
                            同期記録をリセット
                        </Button>
                        <p class="ui-hint text-center">
                            カレンダーで重複が発生する場合や、再同期したい場合に実行してください。
                        </p>
                    </div>
                </Card>
            </div>
        </div>
    {/if}

    <Card tone="muted" radius="2xl" class="flex items-start gap-4">
        <IconChip tone="neutral" size="md" class="shrink-0">
            <CalendarDays size={20} />
        </IconChip>
        <div>
            <h4 class="mb-1 font-bold text-gray-900 dark:text-gray-100">
                同期の仕組み
            </h4>
            <p class="text-sm leading-relaxed text-gray-600 dark:text-gray-300">
                Calendar
                APIの追跡イベント（notion_page_id付き）を起点にDBへ逆引きし、差分があれば統一ロジックで更新します。
                Notionにない追跡イベントは削除し、重複イベントは1件に整理します。設定の「Notion
                プロパティマッピング」で、タイトル・日付・場所・参加者（メールアドレス）を紐付けてください。
            </p>
        </div>
    </Card>
</div>
