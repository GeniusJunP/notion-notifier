<script lang="ts">
    import {
        api,
        buildPreviewNotificationRequest,
        type Config,
        type UpcomingNotification,
        type PeriodicNotification,
    } from "../lib/api";
    import {
        configStore,
        addToast,
        saveConfig as saveConfigState,
    } from "../lib/store";
    import PreviewModal from "../components/PreviewModal.svelte";
    import { Plus, Save } from "lucide-svelte";
    import UpcomingRuleCard from "../components/notifications/UpcomingRuleCard.svelte";
    import PeriodicRuleCard from "../components/notifications/PeriodicRuleCard.svelte";

    $: config = $configStore;

    let isSaving = false;
    let previewOpen = false;
    let previewTitle = "";
    let previewContent = "";

    function openPreview(title: string, content: string) {
        previewTitle = title;
        previewContent = content;
        previewOpen = true;
    }

    async function saveConfig() {
        isSaving = true;
        await saveConfigState(config, {
            successMessage: "設定を保存しました",
            errorMessage: "保存に失敗しました",
        });
        isSaving = false;
    }

    function addUpcomingRule() {
        if (!config) return;
        const newRule: UpcomingNotification = {
            enabled: true,
            minutes_before: 30,
            message: "",
            conditions: {
                days_of_week: [],
                property_filters: [],
            },
        };
        config.notifications.upcoming = [
            ...(config.notifications.upcoming || []),
            newRule,
        ];
        configStore.set(config);
    }

    function addPeriodicRule() {
        if (!config) return;
        const newRule: PeriodicNotification = {
            enabled: true,
            days_of_week: [],
            time: "09:00",
            days_ahead: 7,
            message: "",
        };
        config.notifications.periodic = [
            ...(config.notifications.periodic || []),
            newRule,
        ];
        configStore.set(config);
    }

    function removeUpcomingRule(index: number) {
        if (!config) return;
        config.notifications.upcoming = (
            config.notifications.upcoming || []
        ).filter((_, i) => i !== index);
        configStore.set(config);
    }

    function removePeriodicRule(index: number) {
        if (!config) return;
        config.notifications.periodic = (
            config.notifications.periodic || []
        ).filter((_, i) => i !== index);
        configStore.set(config);
    }

    async function previewTemplate(
        template: string,
        title: string,
        minutes_before?: number,
        days_ahead?: number,
    ) {
        try {
            const req = buildPreviewNotificationRequest(template, {
                minutesBefore: minutes_before,
                daysAhead: days_ahead,
            });
            const res = await api.previewNotification(req);
            openPreview(title, res.message);
        } catch {
            addToast("プレビューに失敗しました", "error");
        }
    }

    async function resetTemplate(type: "upcoming" | "periodic", index: number) {
        if (!config) return;
        try {
            const defaults = await api.getDefaultTemplates();
            const message = defaults[type] || "";
            if (type === "upcoming") {
                config.notifications.upcoming[index].message = message;
            } else {
                config.notifications.periodic[index].message = message;
            }
            configStore.set(config);
            addToast("デフォルトテンプレートを適用しました", "info");
        } catch {
            addToast("デフォルトテンプレートの取得に失敗しました", "error");
        }
    }
</script>

<div class="space-y-6">
    <div class="flex items-center justify-between gap-4 flex-wrap">
        <button
            on:click={saveConfig}
            disabled={isSaving}
            class="px-6 py-2.5 bg-brand-600 dark:bg-brand-500 text-white rounded-xl font-bold flex items-center gap-2 hover:bg-brand-700 dark:hover:bg-brand-600 active:scale-95 disabled:opacity-50 transition-all"
        >
            {#if isSaving}
                <div
                    class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
                ></div>
            {:else}
                <Save size={18} />
            {/if}
            変更を保存
        </button>
    </div>

    {#if config}
        <!-- 並列表示: 小画面は縦並び、lg以上で2カラム横並び -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <!-- 事前通知カラム -->
            <div class="space-y-4">
                <div class="flex items-center justify-between">
                    <h2
                        class="text-xl font-bold text-gray-800 dark:text-gray-200"
                    >
                        事前通知ルール
                    </h2>
                    <button
                        on:click={addUpcomingRule}
                        class="text-brand-600 dark:text-brand-400 flex items-center gap-1 text-sm font-bold hover:underline"
                    >
                        <Plus size={16} /> ルールを追加
                    </button>
                </div>

                {#each config.notifications.upcoming || [] as rule, i}
                    <UpcomingRuleCard
                        bind:rule
                        index={i}
                        on:remove={(e) => removeUpcomingRule(e.detail)}
                        on:preview={(e) =>
                            previewTemplate(
                                e.detail.template,
                                e.detail.title,
                                e.detail.minutes_before,
                            )}
                        on:reset={(e) => resetTemplate("upcoming", e.detail)}
                    />
                {/each}
            </div>

            <!-- 定期通知カラム -->
            <div class="space-y-4">
                <div class="flex items-center justify-between">
                    <h2
                        class="text-xl font-bold text-gray-800 dark:text-gray-200"
                    >
                        定期通知ルール
                    </h2>
                    <button
                        on:click={addPeriodicRule}
                        class="text-brand-600 dark:text-brand-400 flex items-center gap-1 text-sm font-bold hover:underline"
                    >
                        <Plus size={16} /> ルールを追加
                    </button>
                </div>

                {#each config.notifications.periodic || [] as rule, i}
                    <PeriodicRuleCard
                        bind:rule
                        index={i}
                        on:remove={(e) => removePeriodicRule(e.detail)}
                        on:preview={(e) =>
                            previewTemplate(
                                e.detail.template,
                                e.detail.title,
                                undefined,
                                e.detail.days_ahead,
                            )}
                        on:reset={(e) => resetTemplate("periodic", e.detail)}
                    />
                {/each}
            </div>
        </div>
    {/if}
</div>

<PreviewModal
    open={previewOpen}
    title={previewTitle}
    content={previewContent}
    mode="webhook"
    on:close={() => (previewOpen = false)}
/>
