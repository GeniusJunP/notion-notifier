<script lang="ts">
    import type { Config } from "../lib/api";
    import { configStore, saveConfig as saveConfigState } from "../lib/store";
    import WebhookSettingsCard from "../components/WebhookSettingsCard.svelte";
    import GeneralSettings from "../components/settings/GeneralSettings.svelte";
    import PropertyMappingSettings from "../components/settings/PropertyMappingSettings.svelte";
    import ContentRuleSettings from "../components/settings/ContentRuleSettings.svelte";
    import { Settings, Save, Hash } from "lucide-svelte";

    let config: Config | null = null;
    configStore.subscribe((v) => (config = v));

    let isSaving = false;

    async function saveConfig() {
        isSaving = true;
        await saveConfigState(config, {
            successMessage: "システム設定を保存しました",
            errorMessage: "保存失敗",
        });
        isSaving = false;
    }
</script>

<div class="space-y-8 max-w-5xl">
    <div class="flex items-center justify-between">
        <h2
            class="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-3"
        >
            <Settings size={28} class="text-brand-600 dark:text-brand-400" />
            システム設定
        </h2>
        <button
            on:click={saveConfig}
            disabled={isSaving}
            class="px-8 py-3 bg-brand-600 dark:bg-brand-500 text-white rounded-2xl font-bold flex items-center gap-2 hover:bg-brand-700 dark:hover:bg-brand-600 active:scale-95 disabled:opacity-50 transition-all"
        >
            {#if isSaving}
                <div
                    class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
                ></div>
            {:else}
                <Save size={20} />
            {/if}
            全ての変更を適用
        </button>
    </div>

    {#if config}
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <!-- General & Notion Mapping -->
            <div class="space-y-8">
                <GeneralSettings bind:config />
                <PropertyMappingSettings bind:config />
            </div>

            <!-- Content Rules & Advanced -->
            <div class="space-y-8">
                <ContentRuleSettings bind:config />

                <section
                    class="bg-white dark:bg-gray-800 p-6 rounded-3xl border border-gray-100 dark:border-gray-700 space-y-6"
                >
                    <h3
                        class="text-lg font-bold text-gray-800 dark:text-gray-200 flex items-center gap-2"
                    >
                        <Hash
                            size={20}
                            class="text-green-500 dark:text-green-400"
                        />
                        その他
                    </h3>
                    <div class="space-y-4">
                        <div class="grid grid-cols-1 gap-4">
                            <div>
                                <label
                                    for="settings-snooze"
                                    class="block text-xs font-bold text-gray-500 dark:text-gray-400 mb-2"
                                    >スヌーズ (Snooze)</label
                                >
                                <input
                                    id="settings-snooze"
                                    type="datetime-local"
                                    bind:value={config.snooze_until}
                                    class="w-full px-3 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-xs"
                                />
                            </div>
                        </div>
                    </div>
                </section>

                <WebhookSettingsCard {config} />
            </div>
        </div>
    {/if}
</div>
