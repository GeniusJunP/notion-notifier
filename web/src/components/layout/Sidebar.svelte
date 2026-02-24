<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { Bell, X, RefreshCcw, BellOff } from "lucide-svelte";
    import SidebarButton from "../SidebarButton.svelte";
    import TemplateGuideSidebar from "../TemplateGuideSidebar.svelte";
    import { navigate } from "../../lib/store";
    import type { Config, DashboardData } from "../../lib/api";
    import { sidebarOpen, guideModal } from "../../lib/uiStore";
    export let navItems: { path: string; label: string; icon: any }[];
    export let activeRouteValue: string;
    export let isSyncing: boolean;
    export let dashboardData: DashboardData | null;
    export let config: Config | null;
    export let showTemplateGuide: boolean;
    export let mainNavId: string;

    const dispatch = createEventDispatcher<{
        sync: void;
        saveSnooze: void;
        clearSnooze: void;
    }>();
</script>

<aside
    class="fixed inset-y-0 left-0 z-50 w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 transform transition-transform duration-300 lg:relative lg:translate-x-0 {$sidebarOpen
        ? 'translate-x-0'
        : '-translate-x-full'}"
    aria-label="サイドバー"
>
    <div
        class="flex items-center justify-between h-14 px-6 border-b border-gray-100 dark:border-gray-700"
    >
        <div class="flex items-center gap-2">
            <div
                class="w-7 h-7 bg-brand-600 rounded flex items-center justify-center text-white"
            >
                <Bell size={16} />
            </div>
            <span class="font-bold text-base tracking-tight"
                >Notion Notifier</span
            >
        </div>
        <button
            class="lg:hidden text-gray-500"
            on:click={() => sidebarOpen.close()}
            aria-label="サイドバーを閉じる"
        >
            <X size={20} />
        </button>
    </div>

    <nav
        id={mainNavId}
        class="p-4 space-y-1 overflow-y-auto h-[calc(100%-3.5rem)]"
        aria-label="メインナビゲーション"
    >
        {#each navItems as item}
            <SidebarButton
                active={activeRouteValue === item.path}
                ariaCurrent={activeRouteValue === item.path
                    ? "page"
                    : undefined}
                on:click={() => {
                    navigate(item.path);
                    if (window.innerWidth < 1024) sidebarOpen.close();
                }}
            >
                <div
                    class="transition-transform duration-200 group-hover:scale-110"
                >
                    <svelte:component
                        this={item.icon}
                        size={20}
                        strokeWidth={activeRouteValue === item.path ? 2.5 : 2}
                    />
                </div>
                <span>{item.label}</span>
                {#if activeRouteValue === item.path}
                    <div
                        class="ml-auto w-1.5 h-1.5 rounded-full bg-brand-600"
                    ></div>
                {/if}
            </SidebarButton>
        {/each}

        <div
            class="border-t border-gray-100 dark:border-gray-700 mt-4 pt-4 space-y-4"
        >
            <SidebarButton
                justifyBetween
                on:click={() => dispatch("sync")}
                disabled={isSyncing}
            >
                <div class="flex items-center gap-3">
                    <div
                        class={isSyncing
                            ? "animate-spin"
                            : "transition-transform duration-200 group-hover:scale-110"}
                    >
                        <RefreshCcw size={20} />
                    </div>
                    <span>Notion同期</span>
                </div>
                {#if dashboardData}
                    <span
                        class="text-[10px] tabular-nums font-medium opacity-60"
                    >
                        {dashboardData.last_sync
                            ? new Date(
                                  dashboardData.last_sync,
                              ).toLocaleTimeString("ja-JP", {
                                  hour: "2-digit",
                                  minute: "2-digit",
                              })
                            : "--:--"}
                    </span>
                {/if}
            </SidebarButton>
            {#if config}
                <div class="flex flex-col gap-4">
                    <div
                        class="flex-1 p-4 bg-white dark:bg-gray-800 rounded-2xl border border-gray-100 dark:border-gray-700 flex flex-col items-start gap-3"
                    >
                        <div class="flex items-center gap-3 mb-2">
                            <div
                                class="w-9 h-9 bg-amber-50 dark:bg-amber-900 rounded-xl flex items-center justify-center text-amber-600 dark:text-amber-400"
                            >
                                <BellOff size={18} />
                            </div>
                            <div>
                                <span
                                    class="text-sm font-bold text-gray-900 dark:text-gray-100"
                                    >スヌーズ</span
                                >
                                <p
                                    class="text-[10px] text-gray-400 dark:text-gray-500"
                                >
                                    指定日時まで通知を一時停止
                                </p>
                            </div>
                        </div>
                        <div class="flex flex-col w-full gap-2">
                            <div class="flex items-center gap-2 w-full">
                                <input
                                    type="datetime-local"
                                    bind:value={config.snooze_until}
                                    on:change={() => dispatch("saveSnooze")}
                                    class="px-3 py-1.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-xs focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all w-full"
                                />
                                {#if config.snooze_until}
                                    <button
                                        on:click={() => dispatch("clearSnooze")}
                                        class="text-gray-400 dark:text-gray-500 hover:text-red-500 dark:hover:text-red-400 transition-colors p-1"
                                        aria-label="スヌーズ設定をクリア"
                                    >
                                        <X size={14} />
                                    </button>
                                {/if}
                            </div>
                        </div>
                    </div>
                </div>
            {/if}

            {#if showTemplateGuide}
                <TemplateGuideSidebar
                    on:openGuide={(e) =>
                        guideModal.open(e.detail.title, e.detail.content)}
                />
            {/if}
        </div>
    </nav>
</aside>
