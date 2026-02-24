<script lang="ts">
    import { createEventDispatcher, onMount, onDestroy } from "svelte";
    import { Menu, Sun, Moon, Database, BellOff } from "lucide-svelte";
    import { darkMode } from "../../lib/store";
    import { sidebarOpen } from "../../lib/uiStore";
    import type { DashboardData } from "../../lib/api";

    export let activeRouteLabel: string;
    export let dashboardData: DashboardData | null;
    export let isServiceActive: boolean;
    export let mainNavId: string;

    let now = new Date();
    let clockInterval: ReturnType<typeof setInterval>;

    const dateFormatter = new Intl.DateTimeFormat("ja-JP", {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
    });
    const weekdayFormatter = new Intl.DateTimeFormat("ja-JP", {
        weekday: "short",
    });
    const timeFormatter = new Intl.DateTimeFormat("ja-JP", {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        hour12: false,
    });

    $: currentDate = dateFormatter.format(now);
    $: currentWeekday = weekdayFormatter.format(now);
    $: currentTime = timeFormatter.format(now);

    const dispatch = createEventDispatcher();

    function toggleDarkMode() {
        darkMode.update((current) => !current);
    }

    function updateClock() {
        now = new Date();
    }

    onMount(() => {
        clockInterval = setInterval(updateClock, 1000);
        updateClock();
    });

    onDestroy(() => {
        if (clockInterval) clearInterval(clockInterval);
    });
</script>

<header
    class="h-14 bg-white/80 dark:bg-gray-800/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-700 flex items-center px-4 md:px-8 justify-between sticky top-0 z-40"
>
    <div class="flex items-center gap-4">
        <button
            class="lg:hidden p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
            on:click={() => sidebarOpen.open()}
            aria-expanded={$sidebarOpen}
            aria-controls={mainNavId}
            aria-label="サイドバーを開く"
        >
            <Menu size={20} />
        </button>
        <h1 class="text-xl font-bold text-gray-800 dark:text-gray-200">
            {activeRouteLabel}
        </h1>
    </div>

    <div class="flex items-center gap-2 md:gap-4">
        <div
            class="hidden sm:flex items-center gap-4 text-sm font-medium text-gray-500 dark:text-gray-400 tabular-nums"
            aria-label="現在日時"
        >
            <span class="text-gray-700 dark:text-gray-200 font-semibold"
                >{currentDate}（{currentWeekday}）{currentTime}</span
            >
        </div>

        <div
            class="h-4 w-px bg-gray-200 dark:bg-gray-700 hidden lg:block"
        ></div>

        {#if dashboardData}
            <div class="hidden xl:flex items-center gap-4">
                <div
                    class="flex items-center gap-1.5 text-xs font-medium text-gray-500 dark:text-gray-400"
                >
                    <Database
                        size={14}
                        class={dashboardData.last_sync_error
                            ? "text-red-500"
                            : ""}
                    />
                    <span class="tabular-nums">
                        {dashboardData.last_sync
                            ? new Date(
                                  dashboardData.last_sync,
                              ).toLocaleTimeString("ja-JP", {
                                  hour: "2-digit",
                                  minute: "2-digit",
                              })
                            : "--:--"}
                    </span>
                </div>

                {#if dashboardData.snooze_active}
                    <div
                        class="flex items-center gap-1 text-xs font-medium text-amber-600 dark:text-amber-400"
                        title="スヌーズ中"
                    >
                        <BellOff size={14} />
                        <span>SNOOZE</span>
                    </div>
                {/if}
            </div>
            <div
                class="h-4 w-px bg-gray-200 dark:bg-gray-700 hidden xl:block"
            ></div>
        {/if}

        <div class="hidden sm:flex items-center gap-2 px-2 py-1">
            <div
                class="w-2 h-2 rounded-full {isServiceActive
                    ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.4)]'
                    : 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.4)]'}"
            ></div>
            <span
                class="text-[10px] font-bold tracking-wider {isServiceActive
                    ? 'text-green-600 dark:text-green-400'
                    : 'text-red-600 dark:text-red-400'}"
            >
                {isServiceActive ? "SYSTEM ACTIVE" : "SYSTEM OFFLINE"}
            </span>
        </div>

        <div
            class="h-4 w-px bg-gray-200 dark:bg-gray-700 hidden sm:block"
        ></div>

        <button
            on:click={toggleDarkMode}
            class="p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
            aria-label={$darkMode
                ? "ライトモードに切り替え"
                : "ダークモードに切り替え"}
        >
            {#if $darkMode}
                <Sun size={18} />
            {:else}
                <Moon size={18} />
            {/if}
        </button>
    </div>
</header>
