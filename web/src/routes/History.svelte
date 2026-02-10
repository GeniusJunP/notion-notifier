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
        Search,
        Filter,
    } from "lucide-svelte";

    let items: HistoryItem[] = [];
    let isLoading = true;
    let filterType = "all";

    async function loadHistory() {
        isLoading = true;
        try {
            items = await api.getHistory();
        } catch (e) {
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
        } catch (e) {
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
    <div
        class="flex flex-col md:flex-row md:items-center justify-between gap-4"
    >
        <div class="flex items-center gap-4">
            <div
                class="w-12 h-12 bg-gray-100 rounded-2xl flex items-center justify-center text-gray-600"
            >
                <History size={24} />
            </div>
            <div>
                <h2 class="text-2xl font-bold text-gray-900">通知履歴</h2>
                <p class="text-sm text-gray-500">
                    最近の通知送信ログ（最新50件）
                </p>
            </div>
        </div>

        <div class="flex items-center gap-2">
            <button
                on:click={loadHistory}
                class="p-2.5 bg-white border border-gray-200 rounded-xl text-gray-600 hover:bg-gray-50 transition-all active:scale-95"
            >
                <RefreshCcw size={20} class={isLoading ? "animate-spin" : ""} />
            </button>
            <button
                on:click={handleClear}
                class="px-4 py-2.5 bg-red-50 text-red-600 rounded-xl text-sm font-bold flex items-center gap-2 hover:bg-red-100 transition-all active:scale-95"
            >
                <Trash2 size={18} />
                履歴をクリア
            </button>
        </div>
    </div>

    <div
        class="bg-white rounded-3xl border border-gray-100 shadow-sm overflow-hidden min-h-[400px]"
    >
        <div
            class="p-4 border-b border-gray-50 bg-gray-50/50 flex flex-wrap items-center gap-4"
        >
            <div
                class="flex items-center gap-2 bg-white px-3 py-1.5 rounded-xl border border-gray-200"
            >
                <Filter size={14} class="text-gray-400" />
                <select
                    bind:value={filterType}
                    class="text-xs font-bold bg-transparent border-none focus:ring-0 cursor-pointer"
                >
                    <option value="all">全ての履歴</option>
                    <option value="advance">事前通知</option>
                    <option value="periodic">定期通知</option>
                    <option value="manual">手動通知</option>
                </select>
            </div>

            <div class="text-xs text-gray-400 ml-auto">
                Showing {filteredItems.length} items
            </div>
        </div>

        {#if isLoading}
            <div class="p-12 space-y-4">
                {#each Array(5) as _}
                    <div
                        class="h-16 bg-gray-50 rounded-2xl animate-pulse"
                    ></div>
                {/each}
            </div>
        {:else if filteredItems.length === 0}
            <div class="p-20 text-center text-gray-400">
                <History size={48} class="mx-auto mb-4 opacity-10" />
                <p class="font-bold tracking-tight">履歴がありません</p>
            </div>
        {:else}
            <div class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                    <thead>
                        <tr
                            class="text-[10px] font-bold text-gray-400 uppercase tracking-widest border-b border-gray-50"
                        >
                            <th class="px-6 py-4">Status</th>
                            <th class="px-6 py-4">Type</th>
                            <th class="px-6 py-4">Message</th>
                            <th class="px-6 py-4">Sent At</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-50">
                        {#each filteredItems as item (item.id)}
                            <tr
                                class="hover:bg-gray-50/50 transition-colors group"
                            >
                                <td class="px-6 py-4">
                                    <div class="flex items-center gap-2">
                                        {#if item.status === "success"}
                                            <CheckCircle2
                                                size={16}
                                                class="text-green-500"
                                            />
                                            <span
                                                class="text-xs font-bold text-green-700"
                                                >Success</span
                                            >
                                        {:else}
                                            <XCircle
                                                size={16}
                                                class="text-red-500"
                                            />
                                            <span
                                                class="text-xs font-bold text-red-700"
                                                >Failed</span
                                            >
                                        {/if}
                                    </div>
                                </td>
                                <td class="px-6 py-4">
                                    <span
                                        class="px-2 py-0.5 bg-gray-100 text-gray-600 rounded-lg text-[10px] font-bold uppercase tracking-wider"
                                    >
                                        {item.type}
                                    </span>
                                </td>
                                <td class="px-6 py-4">
                                    <div class="max-w-md">
                                        <p
                                            class="text-sm font-medium text-gray-900 line-clamp-1"
                                        >
                                            {item.message}
                                        </p>
                                        {#if item.error}
                                            <p
                                                class="text-[10px] text-red-400 mt-1 font-mono italic"
                                            >
                                                {item.error}
                                            </p>
                                        {/if}
                                    </div>
                                </td>
                                <td class="px-6 py-4">
                                    <div
                                        class="flex items-center gap-2 text-xs text-gray-500"
                                    >
                                        <Clock
                                            size={12}
                                            class="text-gray-300"
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
    </div>
</div>
