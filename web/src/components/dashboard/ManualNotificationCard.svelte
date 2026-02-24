<script lang="ts">
    import { Send, RotateCcw, Play } from "lucide-svelte";
    import { createEventDispatcher } from "svelte";

    export let manualFromDate: string;
    export let manualToDate: string;
    export let manualTemplate: string;
    export let isPreviewLoading: boolean;
    export let isSending: boolean;

    const dispatch = createEventDispatcher<{
        loadDefault: void;
        preview: void;
        send: void;
    }>();
</script>

<div
    class="bg-white dark:bg-gray-800 rounded-3xl border border-gray-100 dark:border-gray-700 shadow-sm dark:shadow-md overflow-hidden"
>
    <div
        class="p-6 border-b border-gray-50 dark:border-gray-700 flex items-center justify-between bg-gray-50/50 dark:bg-gray-700/50"
    >
        <div class="flex items-center gap-3">
            <div
                class="w-10 h-10 bg-orange-50 dark:bg-orange-900/20 rounded-xl flex items-center justify-center text-orange-600 dark:text-orange-300"
            >
                <Send size={20} />
            </div>
            <div>
                <h2 class="text-lg font-bold text-gray-900 dark:text-gray-100">
                    手動通知
                </h2>
                <p class="text-xs text-gray-400">
                    テンプレートを使用して即座にWebhook通知を送信
                </p>
            </div>
        </div>
        <div class="flex items-center gap-3">
            <button
                on:click={() => dispatch("loadDefault")}
                class="text-xs font-bold text-gray-500 dark:text-gray-400 flex items-center gap-1 hover:text-brand-600 dark:hover:text-brand-300 transition-colors"
            >
                <RotateCcw size={12} />
                デフォルトに戻す
            </button>
        </div>
    </div>

    <div class="p-6 space-y-5">
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
                <label
                    for="manual-from-date"
                    class="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase mb-2"
                >
                    開始日
                </label>
                <input
                    id="manual-from-date"
                    type="date"
                    bind:value={manualFromDate}
                    class="w-full px-4 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
                />
            </div>
            <div>
                <label
                    for="manual-to-date"
                    class="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase mb-2"
                >
                    終了日
                </label>
                <input
                    id="manual-to-date"
                    type="date"
                    bind:value={manualToDate}
                    class="w-full px-4 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
                />
            </div>
        </div>

        <div>
            <label
                for="manual-template"
                class="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase mb-2"
            >
                メッセージテンプレート
            </label>
            <textarea
                id="manual-template"
                bind:value={manualTemplate}
                placeholder="Go テンプレート形式で入力..."
                class="w-full p-4 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all font-mono text-sm min-h-[120px]"
            ></textarea>
        </div>

        <div class="flex items-center gap-3">
            <button
                on:click={() => dispatch("preview")}
                disabled={isPreviewLoading}
                class="flex-1 py-3 bg-gray-100 text-gray-700 rounded-xl font-bold flex items-center justify-center gap-2 hover:bg-gray-200 transition-all active:scale-95 disabled:opacity-50"
            >
                {#if isPreviewLoading}
                    <div
                        class="w-4 h-4 border-2 border-gray-400/30 border-t-gray-600 rounded-full animate-spin"
                    ></div>
                {:else}
                    <Play size={16} />
                {/if}
                プレビュー
            </button>
            <button
                on:click={() => dispatch("send")}
                disabled={isSending}
                class="flex-1 py-3 bg-brand-600 text-white rounded-xl font-bold flex items-center justify-center gap-2 hover:bg-brand-700 transition-all active:scale-95 disabled:opacity-50 shadow-lg dark:shadow-xl shadow-brand-100 dark:shadow-brand-900"
            >
                {#if isSending}
                    <div
                        class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
                    ></div>
                {:else}
                    <Send size={16} />
                {/if}
                通知を送信
            </button>
        </div>
    </div>
</div>
