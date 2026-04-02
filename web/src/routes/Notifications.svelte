<script lang="ts">
    import {
        api,
        buildPreviewNotificationRequest,
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
    import Button from "../lib/ui/Button.svelte";

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
            allday_base_time: "09:00",
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
    <div class="flex flex-wrap items-center justify-between gap-4">
        <Button on:click={saveConfig} disabled={isSaving} loading={isSaving} size="md">
            {#if !isSaving}
                <Save size={18} />
            {/if}
            変更を保存
        </Button>
    </div>

    {#if config}
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <div class="space-y-4">
                <div class="flex items-center justify-between">
                    <h2 class="ui-block-title">
                        事前通知ルール
                    </h2>
                    <Button on:click={addUpcomingRule} variant="text" size="sm">
                        <Plus size={16} /> ルールを追加
                    </Button>
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

            <div class="space-y-4">
                <div class="flex items-center justify-between">
                    <h2 class="ui-block-title">
                        定期通知ルール
                    </h2>
                    <Button on:click={addPeriodicRule} variant="text" size="sm">
                        <Plus size={16} /> ルールを追加
                    </Button>
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
