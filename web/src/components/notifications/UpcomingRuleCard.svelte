<script lang="ts">
    import { Trash2, Clock, Play, RotateCcw } from "lucide-svelte";
    import DayPicker from "./DayPicker.svelte";
    import type { UpcomingNotification } from "../../lib/api";
    import { createEventDispatcher } from "svelte";

    export let rule: UpcomingNotification;
    export let index: number;

    const dispatch = createEventDispatcher<{
        remove: number;
        preview: { template: string; title: string; minutes_before: number };
        reset: number;
    }>();
</script>

<div
    class="bg-white dark:bg-gray-800 rounded-2xl border border-gray-100 dark:border-gray-700 overflow-hidden transition-all hover:border-brand-200 dark:hover:border-brand-300"
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
                >事前通知 {index + 1}</span
            >
        </div>
        <button
            type="button"
            on:click={() => dispatch("remove", index)}
            class="text-gray-400 dark:text-gray-500 hover:text-red-500 dark:hover:text-red-400 transition-colors"
            aria-label={`事前通知 ${index + 1} を削除`}
        >
            <Trash2 size={18} />
        </button>
    </div>

    <div class="p-6 grid grid-cols-1 gap-8">
        <div class="space-y-4">
            <div>
                <label
                    for="adv-minutes-{index}"
                    class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                    >通知タイミング (分前)</label
                >
                <div class="relative">
                    <Clock
                        class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
                        size={16}
                    />
                    <input
                        id="adv-minutes-{index}"
                        type="number"
                        bind:value={rule.minutes_before}
                        class="w-full pl-10 pr-4 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
                    />
                </div>
            </div>

            <div>
                <label
                    for="adv-message-{index}"
                    class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                    >メッセージテンプレート</label
                >
                <textarea
                    id="adv-message-{index}"
                    bind:value={rule.message}
                    placeholder="空欄の場合はデフォルトが使用されます"
                    class="w-full p-4 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all font-mono text-sm min-h-[120px]"
                ></textarea>
                <div class="mt-2 flex items-center gap-3">
                    <button
                        type="button"
                        on:click={() =>
                            dispatch("preview", {
                                template: rule.message,
                                title: "事前通知プレビュー",
                                minutes_before: rule.minutes_before,
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

        <div class="space-y-3">
            <span
                class="block text-xs font-bold text-brand-700 dark:text-brand-300 uppercase tracking-wider"
                >実行する曜日</span
            >
            <DayPicker
                bind:selectedDays={rule.conditions.days_of_week}
                ariaLabelPrefix="事前通知 {index + 1} の実行曜日"
            />
        </div>
    </div>
</div>
