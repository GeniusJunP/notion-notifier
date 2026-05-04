<script lang="ts">
    import { onMount } from "svelte";
    import { api, type HistoryItem } from "../lib/api";
    import { addToast } from "../lib/store";
    import {
        History,
        Trash2,
        RefreshCcw,
        CheckCircle2,
        XCircle,
        Clock,
        Filter,
    } from "lucide-svelte";
    import Badge from "../lib/ui/Badge.svelte";
    import Button from "../lib/ui/Button.svelte";
    import Card from "../lib/ui/Card.svelte";
    import Select from "../lib/ui/Select.svelte";

    let items: HistoryItem[] = [];
    let isLoading = true;
    let filterType = "all";

    async function loadHistory() {
        isLoading = true;
        try {
            items = await api.getHistory();
        } catch {
            addToast("履歴の取得に失敗しました", "error");
        } finally {
            isLoading = false;
        }
    }

    onMount(loadHistory);

    async function handleClear() {
        if (!confirm("全ての履歴を削除しますか？")) return;
        try {
            await api.clearHistory();
            items = [];
            addToast("履歴を削除しました", "success");
        } catch {
            addToast("削除に失敗しました", "error");
        }
    }

    $: filteredItems = items.filter((item) => {
        if (filterType === "all") return true;
        return item.type === filterType;
    });

    function formatDate(isoString: string) {
        const d = new Date(isoString);
        return d.toLocaleString("ja-JP", {
            month: "2-digit",
            day: "2-digit",
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
        });
    }
</script>

<div class="space-y-6">
    <div class="flex flex-col justify-between gap-4 md:flex-row md:items-center">
        <div class="flex items-center gap-4">
            <div class="ui-icon-chip h-12 w-12 rounded-2xl">
                <History size={24} />
            </div>
            <div>
                <h2 class="ui-page-title">通知履歴</h2>
                <p class="ui-page-subtitle">最近の通知送信ログ（最新50件）</p>
            </div>
        </div>

        <div class="flex items-center gap-2">
            <Button
                on:click={loadHistory}
                variant="secondary"
                size="icon"
                aria-label="履歴を更新"
            >
                <RefreshCcw size={20} class={isLoading ? "animate-spin" : ""} />
            </Button>
            <Button on:click={handleClear} variant="danger" size="md">
                <Trash2 size={18} />
                履歴をクリア
            </Button>
        </div>
    </div>

    <Card radius="3xl" padding="none" overflowHidden class="min-h-[400px] overflow-hidden">
        <div class="ui-filter-bar">
            <div class="ui-filter-control">
                <Filter size={14} class="text-gray-400 dark:text-gray-500" />
                <Select
                    bind:value={filterType}
                    variant="ghost"
                    uiSize="sm"
                    class="min-h-0 px-0 py-0 text-xs font-semibold focus:ring-brand-300 dark:focus:ring-brand-800"
                >
                    <option value="all">全ての履歴</option>
                    <option value="upcoming">事前通知</option>
                    <option value="periodic">定期通知</option>
                    <option value="manual">手動通知</option>
                </Select>
            </div>

            <div class="ui-meta-text ml-auto">
                {filteredItems.length} 件を表示中
            </div>
        </div>

        {#if isLoading}
            <div class="space-y-4 p-12">
                {#each Array(5) as _, index (index)}
                    <div
                        class="h-16 rounded-2xl bg-gray-50 animate-pulse dark:bg-gray-800"
                    ></div>
                {/each}
            </div>
        {:else if filteredItems.length === 0}
            <div class="p-20 text-center text-gray-400 dark:text-gray-500">
                <History size={48} class="mx-auto mb-4 opacity-10" />
                <p class="font-bold tracking-tight">履歴がありません</p>
            </div>
        {:else}
            <div class="overflow-x-auto">
                <table class="w-full border-collapse text-left">
                    <thead>
                        <tr
                            class="border-b border-gray-200/70 text-[10px] font-bold uppercase tracking-widest text-gray-500 dark:border-gray-800 dark:text-gray-400"
                        >
                            <th class="px-6 py-4">Status</th>
                            <th class="px-6 py-4">Type</th>
                            <th class="px-6 py-4">Message</th>
                            <th class="px-6 py-4">Sent At</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-200/70 dark:divide-gray-800">
                        {#each filteredItems as item (item.id)}
                            <tr class="transition-colors hover:bg-gray-50/70 dark:hover:bg-gray-900/70">
                                <td class="px-6 py-4">
                                    <div class="flex items-center gap-2">
                                        {#if item.status === "success"}
                                            <CheckCircle2
                                                size={16}
                                                class="text-emerald-500 dark:text-emerald-400"
                                            />
                                            <Badge variant="success" caps={false}>Success</Badge>
                                        {:else}
                                            <XCircle
                                                size={16}
                                                class="text-red-500 dark:text-red-400"
                                            />
                                            <Badge variant="error" caps={false}>Failed</Badge>
                                        {/if}
                                    </div>
                                </td>
                                <td class="px-6 py-4">
                                    <Badge variant="neutral">{item.type}</Badge>
                                </td>
                                <td class="px-6 py-4">
                                    <div class="max-w-md">
                                        <p
                                            class="line-clamp-1 text-sm font-medium text-gray-900 dark:text-gray-100"
                                        >
                                            {item.message}
                                        </p>
                                        {#if item.error}
                                            <p
                                                class="mt-1 font-mono text-[10px] italic text-red-400 dark:text-red-500"
                                            >
                                                {item.error}
                                            </p>
                                        {/if}
                                    </div>
                                </td>
                                <td class="px-6 py-4">
                                    <div
                                        class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400"
                                    >
                                        <Clock
                                            size={12}
                                            class="text-gray-300 dark:text-gray-600"
                                        />
                                        {formatDate(item.sent_at)}
                                    </div>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {/if}
    </Card>
</div>
