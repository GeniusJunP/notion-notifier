<script lang="ts">
    import type { Config } from "../lib/api";
    import { configStore } from "../lib/store";
    import { saveConfigWithStore } from "../lib/config-save";
    import WebhookSettingsCard from "../components/WebhookSettingsCard.svelte";
    import {
        Settings,
        Save,
        Database,
        FileText,
        Globe,
        Plus,
        Trash2,
        Hash,
    } from "lucide-svelte";

    let config: Config | null = null;
    configStore.subscribe((v) => (config = v));

    let isSaving = false;

    async function saveConfig() {
        isSaving = true;
        await saveConfigWithStore(config, {
            successMessage: "システム設定を保存しました",
            errorMessage: "保存失敗",
        });
        isSaving = false;
    }

    function addCustomMapping() {
        if (!config) return;
        config.property_mapping.custom = [
            ...config.property_mapping.custom,
            { variable: "", property: "" },
        ];
        configStore.set(config);
    }

    function removeCustomMapping(index: number) {
        if (!config) return;
        config.property_mapping.custom = config.property_mapping.custom.filter(
            (_, i) => i !== index,
        );
        configStore.set(config);
    }
</script>

<div class="space-y-8 max-w-5xl">
    <div class="flex items-center justify-between">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-3">
            <Settings size={28} class="text-brand-600 dark:text-brand-400" />
            システム設定
        </h2>
        <button
            on:click={saveConfig}
            disabled={isSaving}
            class="px-8 py-3 bg-brand-600 dark:bg-brand-500 text-white rounded-2xl font-bold flex items-center gap-2 hover:bg-brand-700 dark:hover:bg-brand-600 active:scale-95 disabled:opacity-50 shadow-xl shadow-brand-100 dark:shadow-brand-900 transition-all"
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
                <section
                    class="bg-white dark:bg-gray-800 p-6 rounded-3xl border border-gray-100 dark:border-gray-700 shadow-sm space-y-6"
                >
                    <h3
                        class="text-lg font-bold text-gray-800 dark:text-gray-200 flex items-center gap-2"
                    >
                        <Globe size={20} class="text-blue-500 dark:text-blue-400" />
                        基本設定
                    </h3>
                    <div class="grid grid-cols-1 gap-4">
                        <div>
                            <label
                                for="settings-timezone"
                                class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                                >タイムゾーン</label
                            >
                            <input
                                id="settings-timezone"
                                type="text"
                                bind:value={config.timezone}
                                class="w-full px-4 py-2.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
                            />
                            <p class="mt-1 text-[11px] text-gray-400 dark:text-gray-500">
                                IANA形式 (例: Asia/Tokyo)
                            </p>
                        </div>
                        <div>
                            <label
                                for="settings-sync-interval"
                                class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                                >Notion 同期間隔 (分)</label
                            >
                            <input
                                id="settings-sync-interval"
                                type="number"
                                bind:value={config.sync.check_interval}
                                class="w-full px-4 py-2.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
                            />
                        </div>
                    </div>
                </section>

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
                                <label
                                    for="settings-mapping-usage"
                                    class="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase mb-2 text-center"
                                    >用途</label
                                >
                                <div
                                    id="settings-mapping-usage"
                                    class="space-y-3"
                                >
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
                                                    !config.property_mapping
                                                        .attendees_enabled;
                                            }}
                                            class="h-8 px-3 rounded-xl border border-gray-200 dark:border-gray-600 text-xs font-bold transition-all {config
                                                .property_mapping
                                                .attendees_enabled
                                                ? 'bg-green-500 dark:bg-green-600 text-white border-green-500 dark:border-green-600'
                                                : 'bg-white dark:bg-gray-700 text-gray-400 dark:text-gray-500'}"
                                            aria-label="参加者プロパティの利用を切り替え"
                                            aria-pressed={config.property_mapping.attendees_enabled}
                                        >

                                            {config.property_mapping
                                                .attendees_enabled
                                                ? "ON"
                                                : "OFF"}

                                        </button>
                                    </div>
                                </div>
                            </div>
                            <div>
                                <label
                                    for="settings-mapping-properties"
                                    class="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase mb-2 text-center"
                                    >Notion プロパティ名</label
                                >
                                <div
                                    id="settings-mapping-properties"
                                    class="space-y-3"
                                >
                                    <input
                                        type="text"
                                        bind:value={
                                            config.property_mapping.title
                                        }
                                        class="w-full h-10 px-4 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400"
                                    />
                                    <input
                                        type="text"
                                        bind:value={
                                            config.property_mapping.date
                                        }
                                        class="w-full h-10 px-4 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400"
                                    />
                                    <input
                                        type="text"
                                        bind:value={
                                            config.property_mapping.location
                                        }
                                        class="w-full h-10 px-4 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400"
                                    />
                                    <input
                                        type="text"
                                        bind:value={
                                            config.property_mapping
                                                .attendees
                                        }
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
                                    <div
                                        class="flex items-center gap-2 group"
                                    >
                                        <div
                                            class="flex-1 grid grid-cols-2 gap-2"
                                        >
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
                                            on:click={() =>
                                                removeCustomMapping(idx)}
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
            </div>

            <!-- Content Rules & Advanced -->
            <div class="space-y-8">
                <section
                    class="bg-white dark:bg-gray-800 p-6 rounded-3xl border border-gray-100 dark:border-gray-700 shadow-sm space-y-6"
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
                                    bind:value={
                                        config.content_rules.start_heading
                                    }
                                    placeholder="例: メモ"
                                    class="flex-1 px-4 py-2.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all font-medium"
                                />
                                <button
                                    on:click={() => {
                                        if (!config) return;
                                        config.content_rules.include_start_heading =
                                            !config.content_rules
                                                .include_start_heading;
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
                            <label
                                class="flex items-center gap-3 cursor-pointer group"
                            >
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
                                        bind:checked={
                                            config.content_rules
                                                .stop_at_next_heading
                                        }
                                        class="sr-only"
                                    />
                                </div>
                                <span class="text-sm font-medium text-gray-700 dark:text-gray-300"
                                    >次の見出しで停止する</span
                                >
                            </label>

                            <label
                                class="flex items-center gap-3 cursor-pointer group"
                            >
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
                                        bind:checked={
                                            config.content_rules
                                                .stop_at_delimiter
                                        }
                                        class="sr-only"
                                    />
                                </div>
                                <span class="text-sm font-medium text-gray-700 dark:text-gray-300"
                                    >区切り文字で停止する</span
                                >
                            </label>
                        </div>

                        {#if config.content_rules.stop_at_delimiter}
                            <div
                                class="animate-in slide-in-from-top-2 duration-200"
                            >
                                <label
                                    for="settings-delimiter-text"
                                    class="block text-xs font-bold text-gray-400 dark:text-gray-500 uppercase mb-2"
                                    >区切り文字</label
                                >
                                <input
                                    id="settings-delimiter-text"
                                    type="text"
                                    bind:value={
                                        config.content_rules.delimiter_text
                                    }
                                    placeholder="---"
                                    class="w-full px-4 py-2.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
                                />
                            </div>
                        {/if}
                    </div>
                </section>

                <section
                    class="bg-white dark:bg-gray-800 p-6 rounded-3xl border border-gray-100 dark:border-gray-700 shadow-sm space-y-6"
                >
                    <h3
                        class="text-lg font-bold text-gray-800 dark:text-gray-200 flex items-center gap-2"
                    >
                        <Hash size={20} class="text-green-500 dark:text-green-400" />
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
