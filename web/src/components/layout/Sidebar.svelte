<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { X, RefreshCcw, BellOff } from "lucide-svelte";
    import SidebarButton from "../SidebarButton.svelte";
    import TemplateGuideSidebar from "../TemplateGuideSidebar.svelte";
    import { navigate } from "../../lib/store";
    import type { Config, DashboardData } from "../../lib/api";
    import { sidebarOpen, guideModal } from "../../lib/uiStore";
    import Button from "../../lib/ui/Button.svelte";
    import Card from "../../lib/ui/Card.svelte";
    import IconChip from "../../lib/ui/IconChip.svelte";
    import Input from "../../lib/ui/Input.svelte";
    import Toggle from "../../lib/ui/Toggle.svelte";

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
    class="fixed inset-y-0 left-0 z-50 w-64 border-r border-gray-200 bg-gray-100 shadow-xl transition-transform duration-300 dark:border-gray-800 dark:bg-gray-950 lg:relative lg:translate-x-0 lg:shadow-none {$sidebarOpen
        ? 'translate-x-0'
        : '-translate-x-full'}"
    aria-label="サイドバー"
>
    <div
        class="flex h-14 items-center justify-end px-4 lg:hidden"
    >
        <Button
            variant="ghost"
            size="icon"
            on:click={() => sidebarOpen.close()}
            aria-label="サイドバーを閉じる"
        >
            <X size={20} />
        </Button>
    </div>

    <nav
        id={mainNavId}
        class="h-[calc(100%-3.5rem)] space-y-1 overflow-y-auto p-4 lg:h-full"
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
                <div class="transition-transform duration-200 group-hover:scale-110">
                    <svelte:component
                        this={item.icon}
                        size={20}
                        strokeWidth={activeRouteValue === item.path ? 2.5 : 2}
                    />
                </div>
                <span>{item.label}</span>
                {#if activeRouteValue === item.path}
                    <div class="ml-auto h-1.5 w-1.5 rounded-full bg-brand-600 dark:bg-brand-400"></div>
                {/if}
            </SidebarButton>
        {/each}

        <div class="mt-4 space-y-4 border-t border-gray-200 pt-4 dark:border-gray-800">
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
                    <span class="text-[10px] font-medium tabular-nums opacity-60">
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
                <Card tone="muted" radius="2xl" padding="sm" class="space-y-3">
                    <div class="mb-2 flex items-center gap-3">
                        <IconChip tone="neutral" size="md">
                            <BellOff size={18} />
                        </IconChip>
                        <div>
                            <span class="text-sm font-semibold text-gray-900 dark:text-gray-100">
                                スヌーズ
                            </span>
                            <p class="ui-meta-text text-[10px]">
                                指定日時まで通知を一時停止
                            </p>
                        </div>
                    </div>

                    <div class="flex w-full items-center gap-2">
                        <Input
                            type="datetime-local"
                            bind:value={config.snooze.until}
                            on:change={() => dispatch("saveSnooze")}
                            uiSize="sm"
                            class="w-full text-xs"
                        />
                        {#if config.snooze.until}
                            <Button
                                on:click={() => dispatch("clearSnooze")}
                                variant="ghost"
                                size="icon"
                                aria-label="スヌーズ設定をクリア"
                            >
                                <X size={14} />
                            </Button>
                        {/if}
                    </div>
                    <div class:ui-snooze-targets={true}>
                        <div class:ui-snooze-target-row={true}>
                            <span class:ui-snooze-target-label={true}>
                                事前通知
                            </span>
                            <Toggle
                                bind:checked={config.snooze.mute_upcoming}
                                ariaLabel="スヌーズ対象に事前通知を含める"
                                size="sm"
                                on:change={() => dispatch("saveSnooze")}
                            />
                        </div>
                        <div class:ui-snooze-target-row={true}>
                            <span class:ui-snooze-target-label={true}>
                                定期通知
                            </span>
                            <Toggle
                                bind:checked={config.snooze.mute_periodic}
                                ariaLabel="スヌーズ対象に定期通知を含める"
                                size="sm"
                                on:change={() => dispatch("saveSnooze")}
                            />
                        </div>
                    </div>
                </Card>
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
