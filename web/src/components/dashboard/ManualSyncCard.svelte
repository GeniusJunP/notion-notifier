<script lang="ts">
    import { RefreshCw } from "lucide-svelte";
    import { createEventDispatcher } from "svelte";
    import Button from "../../lib/ui/Button.svelte";
    import Card from "../../lib/ui/Card.svelte";

    export let isSyncing = false;

    const dispatch = createEventDispatcher<{ sync: void }>();

    function handleSync() {
        dispatch("sync");
    }
</script>

<Card tone="brand" padding="lg" radius="3xl" class="relative overflow-hidden">
    <div
        class="relative z-10 flex flex-col items-center justify-between gap-6 md:flex-row"
    >
        <div>
            <h2 class="ui-page-title mb-2">
                手動同期を実行
            </h2>
            <p class="ui-support-text max-w-md">
                Notion
                データベースから最新の変更を即座に取得し、カレンダーや通知スケジュールを更新します。
            </p>
        </div>

        <Button
            on:click={handleSync}
            disabled={isSyncing}
            loading={isSyncing}
            variant="secondary"
            size="lg"
            class="relative z-10"
        >
            {#if !isSyncing}
                <RefreshCw size={20} />
            {/if}
            {isSyncing ? "同期中..." : "今すぐ同期"}
        </Button>
    </div>

    <div
        class="absolute right-0 top-0 h-64 w-64 rounded-full bg-white/35 translate-x-1/2 -translate-y-1/2 dark:bg-white/5"
    ></div>
    <div
        class="absolute bottom-0 left-0 h-32 w-32 -translate-x-1/2 translate-y-1/2 rounded-full bg-white/25 dark:bg-white/5"
    ></div>
</Card>
