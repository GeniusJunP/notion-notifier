<script lang="ts">
    import { configStore, saveConfig as saveConfigState } from "../lib/store";
    import WebhookSettingsCard from "../components/WebhookSettingsCard.svelte";
    import GeneralSettings from "../components/settings/GeneralSettings.svelte";
    import PropertyMappingSettings from "../components/settings/PropertyMappingSettings.svelte";
    import ContentRuleSettings from "../components/settings/ContentRuleSettings.svelte";
    import { Settings, Save } from "lucide-svelte";
    import Button from "../lib/ui/Button.svelte";
    import { onMount, onDestroy } from "svelte";

    let config = $configStore;
    let isSaving = false;
    let unsubscribe: () => void;

    onMount(() => {
        unsubscribe = configStore.subscribe((value) => {
            config = value;
        });
    });

    onDestroy(() => {
        if (unsubscribe) unsubscribe();
    });

    async function saveConfig() {
        isSaving = true;
        await saveConfigState(config, {
            successMessage: "システム設定を保存しました",
            errorMessage: "保存失敗",
        });
        isSaving = false;
    }
</script>

<div class="max-w-5xl space-y-8">
    <div class="flex items-center justify-between">
        <h2 class="ui-page-title flex items-center gap-3">
            <Settings size={28} class="text-brand-600 dark:text-brand-300" />
            システム設定
        </h2>
        <Button on:click={saveConfig} disabled={isSaving} loading={isSaving} size="lg">
            {#if !isSaving}
                <Save size={20} />
            {/if}
            全ての変更を適用
        </Button>
    </div>

    {#if config}
        <div class="grid grid-cols-1 gap-8 lg:grid-cols-2">
            <div class="space-y-8">
                <GeneralSettings bind:config />
                <PropertyMappingSettings bind:config />
            </div>

            <div class="space-y-8">
                <ContentRuleSettings bind:config />
                <WebhookSettingsCard config={config ?? undefined} />
            </div>
        </div>
    {/if}
</div>
