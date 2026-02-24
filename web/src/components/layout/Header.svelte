<script lang="ts">
    import { createEventDispatcher, onMount, onDestroy } from "svelte";
    import { Menu, Sun, Moon } from "lucide-svelte";
    import { darkMode } from "../../lib/store";
    import { sidebarOpen } from "../../lib/uiStore";

    export let activeRouteLabel: string;
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

    <div class="flex items-center gap-3">
        <!-- Date & Time -->
        <span
            class="hidden sm:inline text-sm font-medium tabular-nums text-gray-500 dark:text-gray-400"
            aria-label="現在日時"
        >
            {currentDate}（{currentWeekday}）{currentTime}
        </span>

        <!-- Status Indicator -->
        <div
            class="hidden sm:flex items-center gap-1.5 px-2.5 py-1 rounded-full {isServiceActive
                ? 'bg-green-50 dark:bg-green-900/30'
                : 'bg-red-50 dark:bg-red-900/30'}"
        >
            <div
                class="w-1.5 h-1.5 rounded-full {isServiceActive
                    ? 'bg-green-500'
                    : 'bg-red-500'}"
            ></div>
            <span
                class="text-[10px] font-semibold tracking-wide {isServiceActive
                    ? 'text-green-600 dark:text-green-400'
                    : 'text-red-600 dark:text-red-400'}"
            >
                {isServiceActive ? "ACTIVE" : "OFFLINE"}
            </span>
        </div>

        <!-- Dark Mode Toggle -->
        <button
            on:click={toggleDarkMode}
            class="p-2 text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
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
