<script lang="ts">
    import { Database, Plus, Trash2 } from "lucide-svelte";
    import type { Config } from "../../lib/api";

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

<section
    class="bg-white dark:bg-gray-800 p-6 rounded-3xl border border-gray-100 dark:border-gray-700 shadow-sm space-y-6"
>
    <div class="flex items-center justify-between">
        <h3
            class="text-lg font-bold text-gray-800 dark:text-gray-200 flex items-center gap-2"
        >
            <Database size={20} class="text-indigo-500 dark:text-indigo-400" />
            Notion プロパティマッピング
        </h3>
        <button
            on:click={addCustomMapping}
            class="p-1 px-2 bg-indigo-50 dark:bg-indigo-900 text-indigo-600 dark:text-indigo-400 rounded-lg text-xs font-bold flex items-center gap-1 hover:bg-indigo-100 dark:hover:bg-indigo-800 transition-colors"
        >
            <Plus size={14} /> カスタム
        </button>
    </div>

    <div class="space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
                <span
                    class="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase mb-2 text-center"
                    >用途</span
                >
                <div class="space-y-3">
                    <div
                        class="h-10 flex items-center px-4 bg-gray-100 dark:bg-gray-700 rounded-xl text-sm font-bold text-gray-600 dark:text-gray-300"
                    >
                        タイトル
                    </div>
                    <div
                        class="h-10 flex items-center px-4 bg-gray-100 dark:bg-gray-700 rounded-xl text-sm font-bold text-gray-600 dark:text-gray-300"
                    >
                        日付 (Date)
                    </div>
                    <div
                        class="h-10 flex items-center px-4 bg-gray-100 dark:bg-gray-700 rounded-xl text-sm font-bold text-gray-600 dark:text-gray-300"
                    >
                        場所
                    </div>
                    <div
                        class="h-10 flex items-center justify-between px-4 bg-gray-100 dark:bg-gray-700 rounded-xl text-sm font-bold text-gray-600 dark:text-gray-300"
                    >
                        参加者
                        <button
                            on:click={() => {
                                if (!config) return;
                                config.property_mapping.attendees_enabled =
                                    !config.property_mapping.attendees_enabled;
                            }}
                            class="h-8 px-3 rounded-xl border border-gray-200 dark:border-gray-600 text-xs font-bold transition-all {config
                                .property_mapping.attendees_enabled
                                ? 'bg-green-500 dark:bg-green-600 text-white border-green-500 dark:border-green-600'
                                : 'bg-white dark:bg-gray-700 text-gray-400 dark:text-gray-500'}"
                            aria-label="参加者プロパティの利用を切り替え"
                            aria-pressed={config.property_mapping
                                .attendees_enabled}
                        >
                            {config.property_mapping.attendees_enabled
                                ? "ON"
                                : "OFF"}
                        </button>
                    </div>
                </div>
            </div>
            <div>
                <span
                    class="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase mb-2 text-center"
                    >Notion プロパティ名</span
                >
                <div class="space-y-3">
                    <input
                        type="text"
                        bind:value={config.property_mapping.title}
                        class="w-full h-10 px-4 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400"
                    />
                    <input
                        type="text"
                        bind:value={config.property_mapping.date}
                        class="w-full h-10 px-4 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400"
                    />
                    <input
                        type="text"
                        bind:value={config.property_mapping.location}
                        class="w-full h-10 px-4 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400"
                    />
                    <input
                        type="text"
                        bind:value={config.property_mapping.attendees}
                        class="w-full h-10 px-4 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400"
                    />
                </div>
            </div>
        </div>

        {#if config.property_mapping.custom.length > 0}
            <div
                class="pt-4 border-t border-gray-100 dark:border-gray-600 space-y-3"
            >
                <p
                    class="block text-xs font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest"
                >
                    カスタムマッピング
                </p>
                {#each config.property_mapping.custom as custom, idx}
                    <div class="flex items-center gap-2 group">
                        <div class="flex-1 grid grid-cols-2 gap-2">
                            <input
                                type="text"
                                bind:value={custom.variable}
                                placeholder="変数名"
                                class="h-9 px-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-sm font-mono"
                            />
                            <input
                                type="text"
                                bind:value={custom.property}
                                placeholder="Notion属性"
                                class="h-9 px-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-sm"
                            />
                        </div>
                        <button
                            on:click={() => removeCustomMapping(idx)}
                            class="text-gray-300 dark:text-gray-600 hover:text-red-500 dark:hover:text-red-400 transition-colors opacity-100 md:opacity-0 md:group-hover:opacity-100"
                            aria-label={`カスタムマッピング ${idx + 1} を削除`}
                        >
                            <Trash2 size={16} />
                        </button>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</section>
