<script lang="ts">
    import { RefreshCw, CalendarDays, Clock, ArrowRight } from "lucide-svelte";
    import { createEventDispatcher } from "svelte";
    import type { UpcomingEvent } from "../../lib/api";

    export let upcoming: UpcomingEvent[] = [];
    export let isLoading: boolean = false;
    export let isSyncing: boolean = false;

    const dispatch = createEventDispatcher<{ refresh: void }>();

    const calendarStateMeta: Record<
        UpcomingEvent["calendar_state"],
        { className: string; label: string }
    > = {
        disabled: { className: "bg-gray-100 text-gray-600", label: "連携オフ" },
        needs_sync: {
            className: "bg-amber-100 text-amber-700",
            label: "要同期",
        },
        synced: { className: "bg-green-100 text-green-700", label: "反映済み" },
        error: { className: "bg-red-100 text-red-700", label: "連携エラー" },
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
        <h2 class="text-xl font-bold text-gray-800 dark:text-gray-100">
            直近の予定 (14日間)
        </h2>
        <button
            on:click={() => dispatch("refresh")}
            class="text-sm text-brand-600 font-medium hover:underline flex items-center gap-1"
        >
            <RefreshCw size={14} />
            更新
        </button>
    </div>

    {#if isLoading && !isSyncing}
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            {#each Array(4) as _}
                <div
                    class="bg-white dark:bg-gray-800 p-6 rounded-2xl border border-gray-100 dark:border-gray-700 animate-pulse h-32"
                ></div>
            {/each}
        </div>
    {:else if upcoming.length === 0}
        <div
            class="bg-white dark:bg-gray-800 p-12 rounded-3xl border border-dashed border-gray-200 dark:border-gray-600 text-center"
        >
            <div
                class="w-16 h-16 bg-gray-50 dark:bg-gray-700 rounded-2xl flex items-center justify-center text-gray-300 dark:text-gray-500 mx-auto mb-4"
            >
                <CalendarDays size={32} />
            </div>
            <h3 class="text-lg font-bold text-gray-900 dark:text-gray-100 mb-1">
                予定が見つかりません
            </h3>
            <p class="text-gray-500 dark:text-gray-400 max-w-sm mx-auto">
                同期された直近の予定はありません。Notion
                データベースを確認してください。
            </p>
        </div>
    {:else}
        <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {#each upcoming as event}
                <div
                    class="bg-white dark:bg-gray-800 p-5 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm dark:shadow-md hover:shadow-md dark:hover:shadow-lg transition-shadow group flex flex-col justify-between min-h-[140px]"
                >
                    <div>
                        <div
                            class="flex items-start justify-between gap-3 mb-2"
                        >
                            <h3
                                class="font-bold text-gray-900 dark:text-gray-100 line-clamp-2 leading-tight group-hover:text-brand-600 dark:group-hover:text-brand-300 transition-colors"
                            >
                                {event.title}
                            </h3>
                            <div class="flex flex-col items-end gap-1">
                                <span
                                    class={`px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider ${calendarStateMeta[event.calendar_state].className}`}
                                >
                                    {calendarStateMeta[event.calendar_state]
                                        .label}
                                </span>
                            </div>
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
                                        class="w-3.5 flex items-center justify-center"
                                    >
                                        <div
                                            class="w-1 h-3.5 bg-brand-400 dark:bg-brand-500 rounded-full"
                                        ></div>
                                    </div>
                                    <span class="truncate"
                                        >{event.location}</span
                                    >
                                </div>
                            {/if}
                        </div>
                    </div>

                    <div
                        class="mt-4 pt-3 border-t border-gray-50 flex items-center justify-between"
                    >
                        <a
                            href={event.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            class="text-xs text-brand-600 font-bold flex items-center gap-1 hover:underline"
                        >
                            Notion で開く
                            <ArrowRight size={12} />
                        </a>
                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>
