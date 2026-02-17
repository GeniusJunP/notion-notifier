<script lang="ts">
    import {
        api,
        type Config,
        type AdvanceNotification,
        type PeriodicNotification,
    } from "../lib/api";
    import { configStore, addToast, saveConfig as saveConfigState } from "../lib/store";
    import PreviewModal from "../components/PreviewModal.svelte";
    import {
        Plus,
        Trash2,
        Save,
        Play,
        Clock,
        RotateCcw,
    } from "lucide-svelte";

    let config: Config | null = null;
    configStore.subscribe((v) => (config = v));

    let isSaving = false;
    let previewOpen = false;
    let previewTitle = "";
    let previewContent = "";

    function openPreview(title: string, content: string) {
        previewTitle = title;
        previewContent = content;
        previewOpen = true;
    }

    async function saveConfig() {
        isSaving = true;
        await saveConfigState(config, {
            successMessage: "設定を保存しました",
            errorMessage: "保存に失敗しました",
        });
        isSaving = false;
    }

    function addAdvanceRule() {
        if (!config) return;
        const newRule: AdvanceNotification = {
            enabled: true,
            minutes_before: 30,
            message: "",
            conditions: {
                days_of_week: [],
                property_filters: [],
            },
        };
        config.notifications.advance = [
            ...config.notifications.advance,
            newRule,
        ];
        configStore.set(config);
    }

    function addPeriodicRule() {
        if (!config) return;
        const newRule: PeriodicNotification = {
            enabled: true,
            days_of_week: [],
            time: "09:00",
            days_ahead: 7,
            message: "",
        };
        config.notifications.periodic = [
            ...config.notifications.periodic,
            newRule,
        ];
        configStore.set(config);
    }

    function removeAdvanceRule(index: number) {
        if (!config) return;
        config.notifications.advance = config.notifications.advance.filter(
            (_, i) => i !== index,
        );
        configStore.set(config);
    }

    function removePeriodicRule(index: number) {
        if (!config) return;
        config.notifications.periodic = config.notifications.periodic.filter(
            (_, i) => i !== index,
        );
        configStore.set(config);
    }

    async function previewTemplate(
        template: string,
        title: string,
        minutes_before?: number,
        days_ahead?: number,
    ) {
        try {
            const req: {
                template: string;
                minutes_before?: number;
                from_date?: string;
                to_date?: string;
            } = {
                template,
            };
            if (minutes_before && minutes_before > 0) {
                req.minutes_before = minutes_before;
            } else if (days_ahead && days_ahead > 0) {
                const from = new Date();
                const to = new Date(
                    Date.now() + days_ahead * 24 * 60 * 60 * 1000,
                );
                req.from_date = from.toISOString().split("T")[0];
                req.to_date = to.toISOString().split("T")[0];
            }
            const res = await api.previewNotification(req);
            openPreview(title, res.message);
        } catch (e) {
            addToast("プレビューに失敗しました", "error");
        }
    }

    async function resetAdvanceTemplate(index: number) {
        if (!config) return;
        try {
            const defaults = await api.getDefaultTemplates();
            config.notifications.advance[index].message =
                defaults.advance || "";
            configStore.set(config);
            addToast("デフォルトテンプレートを適用しました", "info");
        } catch {
            addToast("デフォルトテンプレートの取得に失敗しました", "error");
        }
    }

    async function resetPeriodicTemplate(index: number) {
        if (!config) return;
        try {
            const defaults = await api.getDefaultTemplates();
            config.notifications.periodic[index].message =
                defaults.periodic || "";
            configStore.set(config);
            addToast("デフォルトテンプレートを適用しました", "info");
        } catch {
            addToast("デフォルトテンプレートの取得に失敗しました", "error");
        }
    }

    const daysLabels = ["月", "火", "水", "木", "金", "土", "日"];
    const dayValues = [1, 2, 3, 4, 5, 6, 7];

    function toggleDay(list: number[], day: number) {
        if (list.includes(day)) {
            return list.filter((d) => d !== day);
        } else {
            return [...list, day].sort();
        }
    }
</script>

<div class="space-y-6">
    <div class="flex items-center justify-between gap-4 flex-wrap">
        <button
            on:click={saveConfig}
            disabled={isSaving}
            class="px-6 py-2.5 bg-brand-600 dark:bg-brand-500 text-white rounded-xl font-bold flex items-center gap-2 hover:bg-brand-700 dark:hover:bg-brand-600 active:scale-95 disabled:opacity-50 transition-all shadow-lg shadow-brand-100 dark:shadow-brand-900"
        >
            {#if isSaving}
                <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            {:else}
                <Save size={18} />
            {/if}
            変更を保存
        </button>
    </div>

    {#if config}
        <!-- 並列表示: 小画面は縦並び、lg以上で2カラム横並び -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <!-- 事前通知カラム -->
            <div class="space-y-4">
                <div class="flex items-center justify-between">
                    <h2 class="text-xl font-bold text-gray-800 dark:text-gray-200">
                        事前通知ルール
                    </h2>
                    <button
                        on:click={addAdvanceRule}
                        class="text-brand-600 dark:text-brand-400 flex items-center gap-1 text-sm font-bold hover:underline"
                    >
                        <Plus size={16} /> ルールを追加
                    </button>
                </div>

                {#each config.notifications.advance as rule, i}
                    <div
                        class="bg-white dark:bg-gray-800 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm overflow-hidden transition-all hover:border-brand-200 dark:hover:border-brand-300"
                    >
                        <div
                            class="p-5 border-b border-gray-50 dark:border-gray-600 flex items-center justify-between bg-gray-50/50 dark:bg-gray-700/50"
                        >
                            <div class="flex items-center gap-3">
                                <div
                                    class="w-8 h-8 rounded-lg bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 flex items-center justify-center font-bold text-gray-400 dark:text-gray-500"
                                >
                                    {i + 1}
                                </div>
                                <input
                                    type="checkbox"
                                    bind:checked={rule.enabled}
                                    class="w-5 h-5 accent-brand-600 rounded"
                                />
                                <span class="font-bold text-gray-900 dark:text-gray-100"
                                    >事前通知 {i + 1}</span
                                >
                            </div>
                            <button
                                on:click={() => removeAdvanceRule(i)}
                                class="text-gray-400 dark:text-gray-500 hover:text-red-500 dark:hover:text-red-400 transition-colors"
                                aria-label={`事前通知 ${i + 1} を削除`}
                            >
                                <Trash2 size={18} />
                            </button>
                        </div>

                        <div
                            class="p-6 grid grid-cols-1 gap-8"
                        >
                            <div class="space-y-4">
                                <div>
                                    <label
                                        for="adv-minutes-{i}"
                                        class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                                        >通知タイミング (分前)</label
                                    >
                                    <div class="relative">
                                        <Clock
                                            class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
                                            size={16}
                                        />
                                        <input
                                            id="adv-minutes-{i}"
                                            type="number"
                                            bind:value={rule.minutes_before}
                                            class="w-full pl-10 pr-4 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 focus:border-brand-500 dark:focus:border-brand-400 transition-all"
                                        />
                                    </div>
                                </div>

                                <div>
                                    <label
                                        for="adv-message-{i}"
                                        class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                                        >メッセージテンプレート</label
                                    >
                                    <textarea
                                        id="adv-message-{i}"
                                        bind:value={rule.message}
                                        placeholder="空欄の場合はデフォルトが使用されます"
                                        class="w-full p-4 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 focus:border-brand-500 dark:focus:border-brand-400 transition-all font-mono text-sm min-h-[120px]"
                                    ></textarea>
                                    <div
                                        class="mt-2 flex items-center gap-3"
                                    >
                                        <button
                                            on:click={() =>
                                                previewTemplate(
                                                    rule.message,
                                                    "事前通知プレビュー",
                                                    rule.minutes_before,
                                                )}
                                            class="text-xs font-bold text-brand-600 dark:text-brand-400 flex items-center gap-1 hover:underline"
                                        >
                                            <Play size={12} /> プレビューを実行
                                        </button>
                                        <button
                                            on:click={() =>
                                                resetAdvanceTemplate(i)}
                                            class="text-xs font-bold text-gray-400 dark:text-gray-500 flex items-center gap-1 hover:text-gray-600 dark:hover:text-gray-400"
                                        >
                                            <RotateCcw size={12} /> デフォルトに戻す
                                        </button>
                                    </div>
                                </div>
                            </div>

                            <div class="space-y-3">
                                <label
                                    for="adv-days-{i}"
                                    class="block text-xs font-bold text-brand-700 dark:text-brand-300 uppercase tracking-wider"
                                    >実行する曜日</label
                                >
                                <div
                                    id="adv-days-{i}"
                                    class="flex flex-wrap gap-2"
                                >
                                    {#each dayValues as day, idx}
                                        <button
                                            on:click={() =>
                                                (rule.conditions.days_of_week =
                                                    toggleDay(
                                                        rule
                                                            .conditions
                                                            .days_of_week,
                                                        day,
                                                    ))}
                                            class="w-10 h-10 rounded-xl flex items-center justify-center text-sm font-bold transition-all {rule.conditions.days_of_week.includes(
                                                day,
                                            )
                                                ? 'bg-brand-600 dark:bg-brand-500 text-white shadow-lg shadow-brand-100 dark:shadow-brand-900 scale-105'
                                                : 'bg-gray-50 dark:bg-gray-700 text-gray-400 dark:text-gray-500 border border-gray-100 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-600'}"
                                            aria-label={`事前通知 ${i + 1} の実行曜日 ${daysLabels[idx]}`}
                                            aria-pressed={rule.conditions.days_of_week.includes(day)}
                                        >
                                            {daysLabels[idx]}
                                        </button>
                                    {/each}
                                </div>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>

            <!-- 定期通知カラム -->
            <div class="space-y-4">
                <div class="flex items-center justify-between">
                    <h2 class="text-xl font-bold text-gray-800 dark:text-gray-200">
                        定期通知ルール
                    </h2>
                    <button
                        on:click={addPeriodicRule}
                        class="text-brand-600 dark:text-brand-400 flex items-center gap-1 text-sm font-bold hover:underline"
                    >
                        <Plus size={16} /> ルールを追加
                    </button>
                </div>

                {#each config.notifications.periodic as rule, i}
                    <div
                        class="bg-white dark:bg-gray-800 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm overflow-hidden transition-all hover:border-brand-200 dark:hover:border-brand-300"
                    >
                        <div
                            class="p-5 border-b border-gray-50 dark:border-gray-600 flex items-center justify-between bg-gray-50/50 dark:bg-gray-700/50"
                        >
                            <div class="flex items-center gap-3">
                                <div
                                    class="w-8 h-8 rounded-lg bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 flex items-center justify-center font-bold text-gray-400 dark:text-gray-500"
                                >
                                    {i + 1}
                                </div>
                                <input
                                    type="checkbox"
                                    bind:checked={rule.enabled}
                                    class="w-5 h-5 accent-brand-600 rounded"
                                />
                                <span class="font-bold text-gray-900 dark:text-gray-100"
                                    >定期通知 {i + 1}</span
                                >
                            </div>
                            <button
                                on:click={() => removePeriodicRule(i)}
                                class="text-gray-400 dark:text-gray-500 hover:text-red-500 dark:hover:text-red-400 transition-colors"
                                aria-label={`定期通知 ${i + 1} を削除`}
                            >
                                <Trash2 size={18} />
                            </button>
                        </div>

                        <div
                            class="p-6 grid grid-cols-1 gap-8"
                        >
                            <div class="space-y-6">
                                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                                    <div>
                                        <label
                                            for="per-time-{i}"
                                            class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                                            >通知時刻</label
                                        >
                                        <input
                                            id="per-time-{i}"
                                            type="time"
                                            bind:value={rule.time}
                                            class="w-full px-4 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all font-medium"
                                        />
                                    </div>
                                    <div>
                                        <label
                                            for="per-days-ahead-{i}"
                                            class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                                            >参照範囲 (何日分)</label
                                        >
                                        <input
                                            id="per-days-ahead-{i}"
                                            type="number"
                                            bind:value={rule.days_ahead}
                                            class="w-full px-4 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all font-medium"
                                        />
                                    </div>
                                </div>

                                <div>
                                    <label
                                        for="per-weekdays-{i}"
                                        class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                                        >実行する曜日</label
                                    >
                                    <div
                                        id="per-weekdays-{i}"
                                        class="flex flex-wrap gap-2"
                                    >
                                        {#each dayValues as day, idx}
                                            <button
                                                on:click={() =>
                                                    (rule.days_of_week =
                                                        toggleDay(
                                                            rule.days_of_week,
                                                            day,
                                                        ))}
                                                class="w-10 h-10 rounded-xl flex items-center justify-center text-sm font-bold transition-all {rule.days_of_week.includes(
                                                    day,
                                                )
                                                    ? 'bg-brand-600 dark:bg-brand-500 text-white shadow-lg shadow-brand-100 dark:shadow-brand-900 scale-105'
                                                    : 'bg-gray-50 dark:bg-gray-700 text-gray-400 dark:text-gray-500 border border-gray-100 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-600'}"
                                                aria-label={`定期通知 ${i + 1} の実行曜日 ${daysLabels[idx]}`}
                                                aria-pressed={rule.days_of_week.includes(day)}
                                            >
                                                {daysLabels[idx]}
                                            </button>
                                        {/each}
                                    </div>
                                </div>
                            </div>

                            <div class="space-y-4">
                                <div>
                                    <label
                                        for="per-message-{i}"
                                        class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                                        >メッセージテンプレート</label
                                    >
                                    <textarea
                                        id="per-message-{i}"
                                        bind:value={rule.message}
                                        placeholder="空欄の場合はデフォルトが使用されます"
                                        class="w-full p-4 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 focus:border-brand-500 dark:focus:border-brand-400 transition-all font-mono text-sm min-h-[150px]"
                                    ></textarea>
                                    <div
                                        class="mt-2 flex items-center gap-3"
                                    >
                                        <button
                                            on:click={() =>
                                                previewTemplate(
                                                    rule.message,
                                                    "定期通知プレビュー",
                                                    undefined,
                                                    rule.days_ahead,
                                                )}
                                            class="text-xs font-bold text-brand-600 dark:text-brand-400 flex items-center gap-1 hover:underline"
                                        >
                                            <Play size={12} /> プレビューを実行
                                        </button>
                                        <button
                                            on:click={() =>
                                                resetPeriodicTemplate(i)}
                                            class="text-xs font-bold text-gray-400 dark:text-gray-500 flex items-center gap-1 hover:text-gray-600 dark:hover:text-gray-400"
                                        >
                                            <RotateCcw size={12} /> デフォルトに戻す
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        </div>
    {/if}
</div>

<PreviewModal
    open={previewOpen}
    title={previewTitle}
    content={previewContent}
    mode="webhook"
    on:close={() => (previewOpen = false)}
/>
