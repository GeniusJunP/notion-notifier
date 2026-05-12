<script lang="ts">
    import { RefreshCw } from "lucide-svelte";
    import { createEventDispatcher } from "svelte";

    import Button from "../../lib/ui/Button.svelte";
    import Card from "../../lib/ui/Card.svelte";
    import Typography from "../../lib/ui/Typography.svelte";

    export let isSyncing = false;

    const dispatch = createEventDispatcher<{ sync: void }>();

    function handleSync() {
        dispatch("sync");
    }
</script>

<Card tone="brand" padding="lg" radius="3xl">
    <div class="flex flex-col items-center justify-between gap-6 md:flex-row">
        <div>
            <Typography variant="page-title" as="h2" class="mb-2">
                手動同期を実行
            </Typography>
            <Typography variant="support-text" as="p" class="max-w-md">
                Notion
                データベースから最新の変更を即座に取得し、カレンダーや通知スケジュールを更新します。
            </Typography>
        </div>

        <Button
            onclick={handleSync}
            disabled={isSyncing}
            loading={isSyncing}
            variant="secondary"
            size="lg"
        >
            {#if !isSyncing}
                <RefreshCw size={20} />
            {/if}
            {isSyncing ? "同期中..." : "今すぐ同期"}
        </Button>
    </div>
</Card>
