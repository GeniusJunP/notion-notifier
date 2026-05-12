<script lang="ts">
    import { Save,Settings } from "lucide-svelte";

    import ContentRuleSettings from "../components/settings/ContentRuleSettings.svelte";
    import GeneralSettings from "../components/settings/GeneralSettings.svelte";
    import PropertyMappingSettings from "../components/settings/PropertyMappingSettings.svelte";
    import WebhookSettingsCard from "../components/WebhookSettingsCard.svelte";
    import { configStore, saveConfig as saveConfigState } from "../lib/store";
    import Button from "../lib/ui/Button.svelte";
    import Typography from "../lib/ui/Typography.svelte";

    let isSaving = $state(false);

    async function saveConfig() {
        if (!$configStore) return;
        isSaving = true;
        await saveConfigState($configStore, {
            successMessage: "システム設定を保存しました",
            errorMessage: "保存失敗",
        });
        isSaving = false;
    }
</script>

<div class="max-w-5xl space-y-8">
    <div class="flex items-center justify-between">
        <Typography variant="page-title" as="h2" class="flex items-center gap-3">
            <Settings size={28} class="text-brand-600 dark:text-brand-300" />
            システム設定
        </Typography>
        <Button onclick={saveConfig} disabled={isSaving} loading={isSaving} size="lg">
            {#if !isSaving}
                <Save size={20} />
            {/if}
            全ての変更を適用
        </Button>
    </div>

    {#if $configStore}
        <div class="grid grid-cols-1 gap-8 lg:grid-cols-2">
            <div class="space-y-8">
                <GeneralSettings bind:config={$configStore} />
                <PropertyMappingSettings bind:config={$configStore} />
            </div>

            <div class="space-y-8">
                <ContentRuleSettings bind:config={$configStore} />
                <WebhookSettingsCard config={$configStore} />
            </div>
        </div>
    {/if}
</div>
