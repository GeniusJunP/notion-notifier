<script lang="ts">
    import { FileText } from "lucide-svelte";
    import type { Config } from "../../lib/api";

    export let config: Config;
</script>

<section
    class="bg-white dark:bg-gray-800 p-6 rounded-3xl border border-gray-100 dark:border-gray-700 space-y-6"
>
    <h3
        class="text-lg font-bold text-gray-800 dark:text-gray-200 flex items-center gap-2"
    >
        <FileText size={20} class="text-orange-500 dark:text-orange-400" />
        コンテンツ抽出ルール
    </h3>
    <div class="space-y-4">
        <div>
            <label
                for="settings-start-heading"
                class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                >開始見出し</label
            >
            <div class="flex gap-2">
                <input
                    id="settings-start-heading"
                    type="text"
                    bind:value={config.content_rules.start_heading}
                    placeholder="例: メモ"
                    class="flex-1 px-4 py-2.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all font-medium"
                />
                <button
                    on:click={() => {
                        if (!config) return;
                        config.content_rules.include_start_heading =
                            !config.content_rules.include_start_heading;
                    }}
                    class="px-3 rounded-xl border border-gray-200 dark:border-gray-600 text-xs font-bold transition-all {config
                        .content_rules.include_start_heading
                        ? 'bg-orange-500 dark:bg-orange-600 text-white border-orange-500 dark:border-orange-600'
                        : 'bg-white dark:bg-gray-700 text-gray-400 dark:text-gray-500'}"
                    aria-label="開始見出しを通知本文に含める設定を切り替え"
                    aria-pressed={config.content_rules.include_start_heading}
                >
                    見出し含む
                </button>
            </div>
        </div>

        <div class="space-y-3 pt-2">
            <label class="flex items-center gap-3 cursor-pointer group">
                <div
                    class="w-10 h-6 bg-gray-200 dark:bg-gray-600 rounded-full relative transition-colors {config
                        .content_rules.stop_at_next_heading
                        ? 'bg-orange-500 dark:bg-orange-600'
                        : ''}"
                >
                    <div
                        class="absolute top-1 left-1 w-4 h-4 bg-white rounded-full transition-transform {config
                            .content_rules.stop_at_next_heading
                            ? 'translate-x-4'
                            : ''}"
                    ></div>
                    <input
                        type="checkbox"
                        bind:checked={config.content_rules.stop_at_next_heading}
                        class="sr-only"
                    />
                </div>
                <span
                    class="text-sm font-medium text-gray-700 dark:text-gray-300"
                    >次の見出しで停止する</span
                >
            </label>

            <label class="flex items-center gap-3 cursor-pointer group">
                <div
                    class="w-10 h-6 bg-gray-200 dark:bg-gray-600 rounded-full relative transition-colors {config
                        .content_rules.stop_at_delimiter
                        ? 'bg-orange-500 dark:bg-orange-600'
                        : ''}"
                >
                    <div
                        class="absolute top-1 left-1 w-4 h-4 bg-white rounded-full transition-transform {config
                            .content_rules.stop_at_delimiter
                            ? 'translate-x-4'
                            : ''}"
                    ></div>
                    <input
                        type="checkbox"
                        bind:checked={config.content_rules.stop_at_delimiter}
                        class="sr-only"
                    />
                </div>
                <span
                    class="text-sm font-medium text-gray-700 dark:text-gray-300"
                    >区切り文字で停止する</span
                >
            </label>
        </div>

        {#if config.content_rules.stop_at_delimiter}
            <div class="animate-in slide-in-from-top-2 duration-200">
                <label
                    for="settings-delimiter-text"
                    class="block text-xs font-bold text-gray-400 dark:text-gray-500 uppercase mb-2"
                    >区切り文字</label
                >
                <input
                    id="settings-delimiter-text"
                    type="text"
                    bind:value={config.content_rules.delimiter_text}
                    placeholder="---"
                    class="w-full px-4 py-2.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
                />
            </div>
        {/if}
    </div>
</section>
