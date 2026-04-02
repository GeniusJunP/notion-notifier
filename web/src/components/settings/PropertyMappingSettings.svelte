<script lang="ts">
    import { Database, Plus, Trash2 } from "lucide-svelte";
    import type { Config } from "../../lib/api";
    import Button from "../../lib/ui/Button.svelte";
    import Input from "../../lib/ui/Input.svelte";
    import SectionCard from "../../lib/ui/SectionCard.svelte";
    import Toggle from "../../lib/ui/Toggle.svelte";

    export let config: Config;

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
            (_, i) => i !== index,
        );
    }
</script>

<SectionCard>
    <div class="flex items-center justify-between">
        <h3 class="ui-section-title">
            <Database size={20} class="text-brand-500 dark:text-brand-300" />
            Notion プロパティマッピング
        </h3>
        <Button on:click={addCustomMapping} variant="secondary" size="sm">
            <Plus size={14} /> カスタム
        </Button>
    </div>

    <div class="space-y-4">
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
                <span
                    class="mb-2 block text-center text-xs font-bold uppercase text-gray-500 dark:text-gray-400"
                >
                    用途
                </span>
                <div class="space-y-3">
                    <div
                        class="flex h-10 items-center rounded-xl bg-gray-100 px-4 text-sm font-bold text-gray-600 dark:bg-gray-800 dark:text-gray-300"
                    >
                        タイトル
                    </div>
                    <div
                        class="flex h-10 items-center rounded-xl bg-gray-100 px-4 text-sm font-bold text-gray-600 dark:bg-gray-800 dark:text-gray-300"
                    >
                        日付 (Date)
                    </div>
                    <div
                        class="flex h-10 items-center rounded-xl bg-gray-100 px-4 text-sm font-bold text-gray-600 dark:bg-gray-800 dark:text-gray-300"
                    >
                        場所
                    </div>
                    <div
                        class="flex h-10 items-center justify-between rounded-xl bg-gray-100 px-4 text-sm font-bold text-gray-600 dark:bg-gray-800 dark:text-gray-300"
                    >
                        参加者
                        <Toggle
                            bind:checked={config.property_mapping
                                .attendees_enabled}
                            ariaLabel="参加者プロパティの利用を切り替え"
                            tone="success"
                            size="sm"
                        />
                    </div>
                </div>
            </div>

            <div>
                <span
                    class="mb-2 block text-center text-xs font-bold uppercase text-gray-500 dark:text-gray-400"
                >
                    Notion プロパティ名
                </span>
                <div class="space-y-3">
                    <Input type="text" bind:value={config.property_mapping.title} uiSize="sm" />
                    <Input type="text" bind:value={config.property_mapping.date} uiSize="sm" />
                    <Input type="text" bind:value={config.property_mapping.location} uiSize="sm" />
                    <Input type="text" bind:value={config.property_mapping.attendees} uiSize="sm" />
                </div>
            </div>
        </div>

        {#if config.property_mapping.custom.length > 0}
            <div class="space-y-3 border-t border-gray-200/70 pt-4 dark:border-gray-800">
                <p
                    class="block text-xs font-bold uppercase tracking-widest text-gray-500 dark:text-gray-400"
                >
                    カスタムマッピング
                </p>

                {#each config.property_mapping.custom as custom, idx}
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
                            on:click={() => removeCustomMapping(idx)}
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
