<script lang="ts">
    import { Globe } from "lucide-svelte";
    import type { Config } from "../../lib/api";

    export let config: Config;
</script>

<section
    class="bg-white dark:bg-gray-800 p-6 rounded-3xl border border-gray-100 dark:border-gray-700 shadow-sm space-y-6"
>
    <h3
        class="text-lg font-bold text-gray-800 dark:text-gray-200 flex items-center gap-2"
    >
        <Globe size={20} class="text-blue-500 dark:text-blue-400" />
        基本設定
    </h3>
    <div class="grid grid-cols-1 gap-4">
        <div>
            <label
                for="settings-timezone"
                class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                >タイムゾーン</label
            >
            <input
                id="settings-timezone"
                type="text"
                bind:value={config.timezone}
                class="w-full px-4 py-2.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
            />
            <p class="mt-1 text-[11px] text-gray-400 dark:text-gray-500">
                IANA形式 (例: Asia/Tokyo)
            </p>
        </div>
        <div>
            <label
                for="settings-sync-interval"
                class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-2"
                >Notion 同期間隔 (分)</label
            >
            <input
                id="settings-sync-interval"
                type="number"
                bind:value={config.sync.check_interval}
                class="w-full px-4 py-2.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
            />
        </div>
    </div>
</section>
