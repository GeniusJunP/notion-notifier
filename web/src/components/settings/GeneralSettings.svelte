<script lang="ts">
    import { Globe } from "lucide-svelte";
    import type { Config } from "../../lib/api";
    import FormField from "../../lib/ui/FormField.svelte";
    import Input from "../../lib/ui/Input.svelte";
    import SectionCard from "../../lib/ui/SectionCard.svelte";

    export let config: Config;
</script>

<SectionCard>
    <h3 class="ui-section-title">
        <Globe size={20} class="text-brand-500 dark:text-brand-300" />
        基本設定
    </h3>

    <div class="grid grid-cols-1 gap-4">
        <FormField
            label="タイムゾーン"
            forId="settings-timezone"
            hint="IANA形式 (例: Asia/Tokyo)"
        >
            <Input
                id="settings-timezone"
                type="text"
                bind:value={config.timezone}
                placeholder="Asia/Tokyo"
            />
        </FormField>

        <FormField label="Notion 同期間隔 (分)" forId="settings-sync-interval">
            <Input
                id="settings-sync-interval"
                type="number"
                bind:value={config.sync.check_interval}
                min="1"
            />
        </FormField>
    </div>
</SectionCard>
