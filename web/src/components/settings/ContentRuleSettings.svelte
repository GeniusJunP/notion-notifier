<script lang="ts">
    import { FileText } from "lucide-svelte";
    import type { Config } from "../../lib/api";
    import Button from "../../lib/ui/Button.svelte";
    import FormField from "../../lib/ui/FormField.svelte";
    import Input from "../../lib/ui/Input.svelte";
    import SectionCard from "../../lib/ui/SectionCard.svelte";
    import Toggle from "../../lib/ui/Toggle.svelte";

    export let config: Config;
</script>

<SectionCard>
    <h3 class="ui-section-title">
        <FileText size={20} class="text-brand-500 dark:text-brand-300" />
        コンテンツ抽出ルール
    </h3>

    <div class="space-y-4">
        <FormField label="開始見出し" forId="settings-start-heading">
            <div class="flex gap-2">
                <Input
                    id="settings-start-heading"
                    type="text"
                    bind:value={config.content_rules.start_heading}
                    placeholder="例: メモ"
                    class="flex-1"
                />
                <Button
                    on:click={() => {
                        if (!config) return;
                        config.content_rules.include_start_heading =
                            !config.content_rules.include_start_heading;
                    }}
                    variant={config.content_rules.include_start_heading
                        ? "primary"
                        : "secondary"}
                    size="sm"
                    aria-label="開始見出しを通知本文に含める設定を切り替え"
                    aria-pressed={config.content_rules.include_start_heading}
                >
                    見出し含む
                </Button>
            </div>
        </FormField>

        <div class="space-y-3 pt-2">
            <div class="flex items-center gap-3">
                <Toggle
                    bind:checked={config.content_rules.stop_at_next_heading}
                    ariaLabel="次の見出しで停止する"
                    tone="warning"
                    size="sm"
                />
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                    次の見出しで停止する
                </span>
            </div>

            <div class="flex items-center gap-3">
                <Toggle
                    bind:checked={config.content_rules.stop_at_delimiter}
                    ariaLabel="区切り文字で停止する"
                    tone="warning"
                    size="sm"
                />
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                    区切り文字で停止する
                </span>
            </div>
        </div>

        {#if config.content_rules.stop_at_delimiter}
            <FormField
                class="animate-panel-in"
                label="区切り文字"
                forId="settings-delimiter-text"
                variant="eyebrow"
            >
                <Input
                    id="settings-delimiter-text"
                    type="text"
                    bind:value={config.content_rules.delimiter_text}
                    placeholder="---"
                />
            </FormField>
        {/if}
    </div>
</SectionCard>
