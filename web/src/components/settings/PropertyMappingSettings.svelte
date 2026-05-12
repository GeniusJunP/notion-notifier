<script lang="ts">
    import { Database, Plus, Trash2 } from "lucide-svelte";
    import type { Snippet } from "svelte";

    import type { Config, CustomMapping } from "../../lib/api";
    import Button from "../../lib/ui/Button.svelte";
    import Divider from "../../lib/ui/Divider.svelte";
    import FormGrid from "../../lib/ui/FormGrid.svelte";
    import Input from "../../lib/ui/Input.svelte";
    import SectionCard from "../../lib/ui/SectionCard.svelte";
    import Toggle from "../../lib/ui/Toggle.svelte";
    import Typography from "../../lib/ui/Typography.svelte";

    let { config = $bindable() } = $props<{ config: Config }>();

    function addCustomMapping() {
        if (!config) return;
        config.property_mapping.custom = [
            ...config.property_mapping.custom,
            { variable: "", property: "" },
        ];
    }

    function removeCustomMapping(index: number) {
        if (!config) return;
        config.property_mapping.custom = config.property_mapping.custom.filter(
            (_: CustomMapping, i: number) => i !== index,
        );
    }
</script>

{#snippet listHeader(text: string)}
    <Typography variant="label-caps" class="mb-2 text-center">
        {text}
    </Typography>
{/snippet}

{#snippet propertyLabel(text: string, rightContent?: Snippet)}
    <div class="flex h-10 items-center justify-between rounded-xl bg-gray-100 px-4 text-sm font-bold text-gray-600 dark:bg-gray-800 dark:text-gray-300">
        {text}
        {#if rightContent}
            {@render rightContent()}
        {/if}
    </div>
{/snippet}

<SectionCard>
    <div class="flex items-center justify-between">
        <Typography variant="section-title" as="h3">
            <Database size={20} class="text-brand-500 dark:text-brand-300" />
            Notion プロパティマッピング
        </Typography>
        <Button onclick={addCustomMapping} variant="secondary" size="sm">
            <Plus size={14} /> カスタム
        </Button>
    </div>

    <div class="space-y-4">
        <FormGrid>
            <div>
                {@render listHeader("用途")}
                <div class="space-y-3">
                    {@render propertyLabel("タイトル")}
                    {@render propertyLabel("日付 (Date)")}
                    {@render propertyLabel("場所")}
                    {#snippet toggleSnippet()}
                        <Toggle
                            bind:checked={config.property_mapping.attendees_enabled}
                            ariaLabel="参加者プロパティの利用を切り替え"
                            tone="success"
                            size="sm"
                        />
                    {/snippet}
                    {@render propertyLabel("参加者", toggleSnippet)}
                </div>
            </div>

            <div>
                {@render listHeader("Notion プロパティ名")}
                <div class="space-y-3">
                    <Input type="text" bind:value={config.property_mapping.title} uiSize="sm" />
                    <Input type="text" bind:value={config.property_mapping.date} uiSize="sm" />
                    <Input type="text" bind:value={config.property_mapping.location} uiSize="sm" />
                    <Input type="text" bind:value={config.property_mapping.attendees} uiSize="sm" />
                </div>
            </div>
        </FormGrid>

        {#if config.property_mapping.custom.length > 0}
            <Divider spacing="md" class="space-y-3" />
            <div class="space-y-3">
                <Typography variant="label-caps-wide" class="tracking-widest">
                    カスタムマッピング
                </Typography>

                {#each config.property_mapping.custom as custom, idx (idx)}
                    <div class="group flex items-center gap-2">
                        <div class="grid flex-1 grid-cols-2 gap-2">
                            <Input
                                type="text"
                                bind:value={custom.variable}
                                placeholder="変数名"
                                uiSize="sm"
                                mono
                            />
                            <Input
                                type="text"
                                bind:value={custom.property}
                                placeholder="Notion属性"
                                uiSize="sm"
                            />
                        </div>
                        <Button
                            onclick={() => removeCustomMapping(idx)}
                            class="opacity-100 md:opacity-0 md:group-hover:opacity-100"
                            variant="ghost"
                            size="icon"
                            aria-label={`カスタムマッピング ${idx + 1} を削除`}
                        >
                            <Trash2 size={16} />
                        </Button>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</SectionCard>
