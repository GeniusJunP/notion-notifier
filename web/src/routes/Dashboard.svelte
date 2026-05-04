<script lang="ts">
    import { onMount } from "svelte";
    import {
        api,
        buildManualNotificationRequest,
        getErrorMessage,
        type DashboardData,
        type UpcomingEvent,
    } from "../lib/api";
    import {
        addToast,
        configStore,
        syncNotion as syncNotionState,
    } from "../lib/store";
    import { toLocalDateInputValue } from "../lib/utils";
    import PreviewModal from "../components/PreviewModal.svelte";
    import StatCard from "../components/dashboard/StatCard.svelte";
    import ManualSyncCard from "../components/dashboard/ManualSyncCard.svelte";
    import ManualNotificationCard from "../components/dashboard/ManualNotificationCard.svelte";
    import UpcomingEventsList from "../components/dashboard/UpcomingEventsList.svelte";
    import {
        RefreshCw,
        CalendarDays,
        CheckCircle,
        AlertTriangle,
    } from "lucide-svelte";

    let dashboard: DashboardData | null = null;
    let upcoming: UpcomingEvent[] = [];
    $: config = $configStore;
    let isLoading = true;
    let isSyncing = false;

    // Manual notification state
    let manualTemplate = "";
    let manualFromDate = toLocalDateInputValue(new Date());
    let manualToDate = toLocalDateInputValue(
        new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
    );
    let isPreviewLoading = false;
    let isSending = false;
    let previewOpen = false;
    let previewTitle = "";
    let previewContent = "";
    $: if (config && manualTemplate === "") {
        manualTemplate = config.notifications.manual || "";
    }

    function openPreview(title: string, content: string) {
        previewTitle = title;
        previewContent = content;
        previewOpen = true;
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
        } catch {
            addToast("データの取得に失敗しました", "error");
        } finally {
            isLoading = false;
        }
    }

    onMount(async () => {
        await loadData();
    });

    async function handleSync() {
        isSyncing = true;
        try {
            await syncNotionState({
                successMessage: (count) => `${count}件の項目を同期しました`,
                errorMessage: "同期に失敗しました",
                onSynced: () => loadData(),
            });
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
            const req = buildManualNotificationRequest(
                manualTemplate,
                manualFromDate,
                manualToDate,
            );
            const res = await api.previewNotification(req);
            openPreview("手動通知プレビュー", res.message);
        } catch (e: unknown) {
            addToast(`プレビュー失敗: ${getErrorMessage(e)}`, "error");
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
            const req = buildManualNotificationRequest(
                manualTemplate,
                manualFromDate,
                manualToDate,
            );
            const res = await api.sendManualNotification(req);
            addToast("通知を送信しました", "success");
            openPreview("送信メッセージ", res.message);
            try {
                configStore.set(await api.getConfig());
            } catch {
                addToast("設定の再読み込みに失敗しました", "error");
            }
        } catch (e: unknown) {
            addToast(`送信失敗: ${getErrorMessage(e)}`, "error");
        } finally {
            isSending = false;
        }
    }

    async function loadDefaultTemplate() {
        try {
            const defaults = await api.getDefaultTemplates();
            manualTemplate = defaults.manual || defaults.periodic || "";
            addToast("デフォルトテンプレートを読み込みました", "info");
        } catch {
            addToast("デフォルトテンプレートの取得に失敗しました", "error");
        }
    }
</script>

<div class="space-y-8">
    <!-- Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
            label="本日の予定"
            value={dashboard?.today_count ?? 0}
            subLabel="Notion 内予定数"
            tone="brand"
        >
            <CalendarDays size={20} slot="icon" />
        </StatCard>

        <StatCard
            label="次回同期"
            value={dashboard?.next_sync_in ?? "--"}
            subLabel={dashboard?.next_sync
                ? new Date(dashboard.next_sync).toLocaleTimeString()
                : "Scheduled"}
            tone="warning"
        >
            <RefreshCw size={20} slot="icon" />
        </StatCard>

        <StatCard
            label="最終同期件数"
            value={dashboard?.last_sync_count ?? 0}
            subLabel="同期済み件数"
            tone="success"
        >
            <CheckCircle size={20} slot="icon" />
        </StatCard>

        <StatCard
            label="ステータス"
            value={dashboard?.last_sync_error ? "エラー" : "正常"}
            subLabel={dashboard?.last_sync_error || "問題は検出されていません"}
            tone={dashboard?.last_sync_error ? "danger" : "success"}
        >
            <svelte:fragment slot="icon">
                {#if dashboard?.last_sync_error}
                    <AlertTriangle size={20} />
                {:else}
                    <CheckCircle size={20} />
                {/if}
            </svelte:fragment>
        </StatCard>
    </div>

    <!-- Actions -->
    <ManualSyncCard {isSyncing} on:sync={handleSync} />

    <!-- Manual Notification -->
    <ManualNotificationCard
        bind:manualFromDate
        bind:manualToDate
        bind:manualTemplate
        {isPreviewLoading}
        {isSending}
        on:loadDefault={loadDefaultTemplate}
        on:preview={handleManualPreview}
        on:send={handleManualSend}
    />

    <!-- Upcoming Events -->
    <UpcomingEventsList
        {upcoming}
        {isLoading}
        {isSyncing}
        on:refresh={loadData}
    />
</div>

<PreviewModal
    open={previewOpen}
    title={previewTitle}
    content={previewContent}
    mode="webhook"
    on:close={() => (previewOpen = false)}
/>
