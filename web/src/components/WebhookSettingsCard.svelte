<script lang="ts">
    import { Globe } from "lucide-svelte";
    import type { Config } from "../lib/api";

    export let config: Config;
</script>

<div
    class="p-6 bg-brand-50 dark:bg-brand-900 rounded-3xl border border-brand-100 dark:border-brand-700 space-y-5"
>
    <h4
        class="font-bold text-brand-900 dark:text-brand-100 flex items-center gap-2"
    >
        <Globe size={18} class="text-brand-600 dark:text-brand-400" />
        Webhook 設定
    </h4>
    <p
        class="text-xs text-brand-700 dark:text-brand-300 leading-relaxed opacity-80"
    >
        Webhook で送信される JSON ペイロードのテンプレートです。<code
            class="font-mono bg-white/50 dark:bg-gray-800/50 px-1 rounded"
            >{"{{"}.Message{"}}"}</code
        >
        変数が通知内容に置き換わります。
    </p>
    <div
        class="flex items-center justify-between gap-3 p-4 bg-white dark:bg-gray-800 rounded-2xl border border-brand-100 dark:border-brand-700"
    >
        <div class="space-y-1">
            <p class="text-sm font-bold text-gray-700 dark:text-gray-200">
                テストモード
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
                ON にすると、内部通知用のテンプレートと URL で送信します
            </p>
        </div>
        <label class="inline-flex items-center cursor-pointer">
            <input
                type="checkbox"
                class="sr-only peer"
                bind:checked={config.webhook.is_test}
            />
            <span
                class="relative w-11 h-6 bg-gray-200 dark:bg-gray-600 rounded-full peer peer-focus:ring-2 peer-focus:ring-brand-500 dark:peer-focus:ring-brand-400 peer-checked:bg-brand-600 dark:peer-checked:bg-brand-500 after:content-[''] after:absolute after:top-0.5 after:start-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:after:translate-x-full"
            ></span>
        </label>
    </div>

    <div
        class="space-y-3 p-4 bg-white dark:bg-gray-800 rounded-2xl border border-brand-100 dark:border-brand-700"
    >
        <h5
            class="text-xs font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
        >
            Notification Webhook
        </h5>
        <div>
            <label
                for="wh-notification-ct"
                class="block text-xs font-bold text-gray-600 dark:text-gray-300 mb-1"
                >Content-Type</label
            >
            <input
                id="wh-notification-ct"
                type="text"
                bind:value={config.webhook.notification.content_type}
                class="w-full px-3 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
            />
        </div>
        <div>
            <label
                for="wh-notification-pt"
                class="block text-xs font-bold text-gray-600 dark:text-gray-300 mb-1"
                >Payload Template</label
            >
            <textarea
                id="wh-notification-pt"
                bind:value={config.webhook.notification.payload_template}
                class="w-full p-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl font-mono text-xs min-h-[80px] focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
            ></textarea>
        </div>
    </div>

    <div
        class="space-y-3 p-4 bg-white dark:bg-gray-800 rounded-2xl border border-brand-100 dark:border-brand-700"
    >
        <h5
            class="text-xs font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
        >
            Internal Notification Webhook
        </h5>
        <div>
            <label
                for="wh-internal-notif-ct"
                class="block text-xs font-bold text-gray-600 dark:text-gray-300 mb-1"
                >Content-Type</label
            >
            <input
                id="wh-internal-notif-ct"
                type="text"
                bind:value={config.webhook.internal_notification.content_type}
                class="w-full px-3 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl text-sm focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
            />
        </div>
        <div>
            <label
                for="wh-internal-notif-pt"
                class="block text-xs font-bold text-gray-600 dark:text-gray-300 mb-1"
                >Payload Template</label
            >
            <textarea
                id="wh-internal-notif-pt"
                bind:value={
                    config.webhook.internal_notification.payload_template
                }
                class="w-full p-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl font-mono text-xs min-h-[80px] focus:ring-2 focus:ring-brand-500 dark:focus:ring-brand-400 transition-all"
            ></textarea>
        </div>
    </div>
</div>
