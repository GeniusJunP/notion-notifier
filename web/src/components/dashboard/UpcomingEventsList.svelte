<script lang="ts">
    import { RefreshCw, CalendarDays, Clock, ArrowRight } from "lucide-svelte";
    import { createEventDispatcher } from "svelte";
    import type { UpcomingEvent } from "../../lib/api";
    import Badge from "../../lib/ui/Badge.svelte";
    import Button from "../../lib/ui/Button.svelte";
    import Card from "../../lib/ui/Card.svelte";
    import IconChip from "../../lib/ui/IconChip.svelte";

    export let upcoming: UpcomingEvent[] = [];
    export let isLoading = false;
    export let isSyncing = false;

    const dispatch = createEventDispatcher<{ refresh: void }>();

    const calendarStateMeta: Record<
        UpcomingEvent["calendar_state"],
        { variant: "neutral" | "warning" | "success" | "error"; label: string }
    > = {
        disabled: { variant: "neutral", label: "連携オフ" },
        needs_sync: { variant: "warning", label: "要同期" },
        synced: { variant: "success", label: "反映済み" },
        error: { variant: "error", label: "連携エラー" },
    };

    function formatEventDateTime(event: UpcomingEvent): string {
        const startDate = event.start_date;
        const endDate = event.end_date || event.start_date;
        if (event.is_all_day) {
            if (endDate && endDate !== startDate) {
                return `${startDate} - ${endDate} (終日)`;
            }
            return `${startDate} (終日)`;
        }
        const start = event.start_time
            ? `${startDate} ${event.start_time}`
            : startDate;
        if (!event.end_time) {
            return start;
        }
        if (endDate && endDate !== startDate) {
            return `${start} - ${endDate} ${event.end_time}`;
        }
        return `${start} - ${event.end_time}`;
    }
</script>

<div class="space-y-4">
    <div class="flex items-center justify-between">
        <h2 class="ui-block-title">
            直近の予定 (14日間)
        </h2>
        <Button on:click={() => dispatch("refresh")} variant="text" size="sm">
            <RefreshCw size={14} />
            更新
        </Button>
    </div>

    {#if isLoading && !isSyncing}
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            {#each Array(4) as _}
                <Card
                    tone="default"
                    radius="2xl"
                    class="h-32 animate-pulse"
                ></Card>
            {/each}
        </div>
    {:else if upcoming.length === 0}
        <Card
            radius="3xl"
            class="border-dashed p-12 text-center"
        >
            <div class="mx-auto mb-4">
            <IconChip tone="neutral" size="lg">
                <CalendarDays size={32} />
            </IconChip>
            </div>
            <h3 class="mb-1 text-lg font-bold text-gray-900 dark:text-gray-100">
                予定が見つかりません
            </h3>
            <p class="mx-auto max-w-sm text-gray-500 dark:text-gray-400">
                同期された直近の予定はありません。Notion
                データベースを確認してください。
            </p>
        </Card>
    {:else}
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {#each upcoming as event}
                <Card
                    radius="2xl"
                    padding="md"
                    class="group flex min-h-[140px] flex-col justify-between"
                >
                    <div>
                        <div class="mb-2 flex items-start justify-between gap-3">
                            <h3
                                class="line-clamp-2 font-bold leading-tight text-gray-900 transition-colors group-hover:text-brand-700 dark:text-gray-100 dark:group-hover:text-brand-300"
                            >
                                {event.title}
                            </h3>
                            <Badge variant={calendarStateMeta[event.calendar_state].variant}>
                                {calendarStateMeta[event.calendar_state].label}
                            </Badge>
                        </div>
                        <div class="space-y-1.5">
                            <div
                                class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400"
                            >
                                <Clock
                                    size={14}
                                    class="text-gray-400 dark:text-gray-500"
                                />
                                <span>{formatEventDateTime(event)}</span>
                            </div>
                            {#if event.location}
                                <div
                                    class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400"
                                >
                                    <div
                                        class="flex w-3.5 items-center justify-center"
                                    >
                                        <div
                                            class="h-3.5 w-1 rounded-full bg-brand-400 dark:bg-brand-500"
                                        ></div>
                                    </div>
                                    <span class="truncate">{event.location}</span>
                                </div>
                            {/if}
                        </div>
                    </div>

                    <div class="mt-4 border-t border-gray-200/60 pt-3 dark:border-gray-800">
                        <a
                            href={event.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            class="ui-link-button text-xs"
                        >
                            Notion で開く
                            <ArrowRight size={12} />
                        </a>
                    </div>
                </Card>
            {/each}
        </div>
    {/if}
</div>
