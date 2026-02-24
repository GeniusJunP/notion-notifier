<script lang="ts">
    import { Trash2, Play, RotateCcw } from "lucide-svelte";
    import DayPicker from "./DayPicker.svelte";
    import type { PeriodicNotification } from "../../lib/api";
    import { createEventDispatcher } from "svelte";

    export let rule: PeriodicNotification;
    export let index: number;

    const dispatch = createEventDispatcher<{
        remove: number;
        preview: { template: string; title: string; days_ahead: number };
        reset: number;
    }>();
</script>

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
                {index + 1}
            </div>
            <input
                type="checkbox"
                bind:checked={rule.enabled}
                class="w-5 h-5 accent-brand-600 rounded"
            />
            <span class="font-bold text-gray-900 dark:text-gray-100"
                >定期通知 {index + 1}</span
            >
        </div>
        <button
            type="button"
            on:click={() => dispatch("remove", index)}
            class="text-gray-400 dark:text-gray-500 hover:text-red-500 dark:hover:text-red-400 transition-colors"
            aria-label={`定期通知 ${index + 1} を削除`}
        >
            <Trash2 size={18} />
        </button>
    </div>

    <div class="p-6 grid grid-cols-1 gap-8">
        <div class="space-y-6">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                    <label
                        for="per-time-{index}"
                        class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                        >通知時刻</label
                    >
                    <input
                        id="per-time-{index}"
                        type="time"
                        bind:value={rule.time}
                        class="w-full px-4 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all font-medium"
                    />
                </div>
                <div>
                    <label
                        for="per-days-ahead-{index}"
                        class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                        >参照範囲 (何日分)</label
                    >
                    <input
                        id="per-days-ahead-{index}"
                        type="number"
                        bind:value={rule.days_ahead}
                        class="w-full px-4 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all font-medium"
                    />
                </div>
            </div>

            <div>
                <span
                    class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                    >実行する曜日</span
                >
                <DayPicker
                    bind:selectedDays={rule.days_of_week}
                    ariaLabelPrefix="定期通知 {index + 1} の実行曜日"
                />
            </div>
        </div>

        <div class="space-y-4">
            <div>
                <label
                    for="per-message-{index}"
                    class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                    >メッセージテンプレート</label
                >
                <textarea
                    id="per-message-{index}"
                    bind:value={rule.message}
                    placeholder="空欄の場合はデフォルトが使用されます"
                    class="w-full p-4 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all font-mono text-sm min-h-[150px]"
                ></textarea>
                <div class="mt-2 flex items-center gap-3">
                    <button
                        type="button"
                        on:click={() =>
                            dispatch("preview", {
                                template: rule.message,
                                title: "定期通知プレビュー",
                                days_ahead: rule.days_ahead,
                            })}
                        class="text-xs font-bold text-brand-600 dark:text-brand-400 flex items-center gap-1 hover:underline"
                    >
                        <Play size={12} /> プレビューを実行
                    </button>
                    <button
                        type="button"
                        on:click={() => dispatch("reset", index)}
                        class="text-xs font-bold text-gray-400 dark:text-gray-500 flex items-center gap-1 hover:text-gray-600 dark:hover:text-gray-400"
                    >
                        <RotateCcw size={12} /> デフォルトに戻す
                    </button>
                </div>
            </div>
        </div>
    </div>
</div>
