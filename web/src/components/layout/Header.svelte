<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { Menu, Sun, Moon } from "lucide-svelte";
    import { darkMode } from "../../lib/store";
    import { sidebarOpen } from "../../lib/uiStore";
    import Badge from "../../lib/ui/Badge.svelte";
    import Button from "../../lib/ui/Button.svelte";

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
    class="sticky top-0 z-40 flex h-14 items-center justify-between bg-gray-100 px-4 dark:bg-gray-950 md:px-8"
>
    <div class="flex items-center gap-4">
        <Button
            class="lg:hidden"
            variant="ghost"
            size="icon"
            on:click={() => sidebarOpen.open()}
            aria-expanded={$sidebarOpen}
            aria-controls={mainNavId}
            aria-label="サイドバーを開く"
        >
            <Menu size={20} />
        </Button>
        <h1 class="text-xl font-bold text-gray-800 dark:text-gray-200">
            {activeRouteLabel}
        </h1>
    </div>

    <div class="flex items-center gap-3">
        <span
            class="hidden text-sm font-medium tabular-nums text-gray-500 dark:text-gray-400 sm:inline"
            aria-label="現在日時"
        >
            {currentDate}（{currentWeekday}）{currentTime}
        </span>

        <Badge
            variant={isServiceActive ? "success" : "error"}
            class="hidden sm:inline-flex"
        >
            {isServiceActive ? "ACTIVE" : "OFFLINE"}
        </Badge>

        <Button
            on:click={toggleDarkMode}
            variant="ghost"
            size="icon"
            aria-label={$darkMode
                ? "ライトモードに切り替え"
                : "ダークモードに切り替え"}
        >
            {#if $darkMode}
                <Sun size={18} />
            {:else}
                <Moon size={18} />
            {/if}
        </Button>
    </div>
</header>
