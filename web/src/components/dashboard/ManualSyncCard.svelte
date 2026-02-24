<script lang="ts">
    import { RefreshCw } from "lucide-svelte";
    import { createEventDispatcher } from "svelte";

    export let isSyncing: boolean = false;

    const dispatch = createEventDispatcher<{ sync: void }>();

    function handleSync() {
        dispatch("sync");
    }
</script>

<div
    class="bg-brand-600 p-8 rounded-3xl shadow-xl shadow-brand-100 dark:shadow-brand-900 text-white flex flex-col md:flex-row items-center justify-between gap-6 overflow-hidden relative"
>
    <div class="relative z-10">
        <h2 class="text-2xl font-bold mb-2">手動同期を実行</h2>
        <p class="text-brand-100 max-w-md">
            Notion
            データベースから最新の変更を即座に取得し、カレンダーや通知スケジュールを更新します。
        </p>
    </div>
    <button
        on:click={handleSync}
        disabled={isSyncing}
        class="relative z-10 px-8 py-4 bg-white dark:bg-gray-800 text-brand-600 dark:text-brand-300 rounded-2xl font-bold shadow-lg dark:shadow-xl hover:bg-brand-50 dark:hover:bg-brand-900/20 transition-all flex items-center gap-3 active:scale-95 disabled:opacity-70 disabled:pointer-events-none"
    >
        {#if isSyncing}
            <RefreshCw size={20} class="animate-spin" />
            同期中...
        {:else}
            <RefreshCw size={20} />
            今すぐ同期
        {/if}
    </button>

    <!-- Abstract pattern -->
    <div
        class="absolute right-0 top-0 w-64 h-64 bg-white dark:bg-gray-800 opacity-[0.05] rounded-full translate-x-1/2 -translate-y-1/2"
    ></div>
    <div
        class="absolute left-0 bottom-0 w-32 h-32 bg-white dark:bg-gray-800 opacity-[0.05] rounded-full -translate-x-1/2 translate-y-1/2"
    ></div>
</div>
