<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { X } from "lucide-svelte";

    export let open = false;
    export let title = "プレビュー";
    export let content = "";

    const dispatch = createEventDispatcher<{ close: void }>();

    function close() {
        dispatch("close");
    }

    function handleKeydown(event: KeyboardEvent) {
        if (event.key === "Escape" && open) {
            close();
        }
    }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if open}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4" role="dialog" aria-modal="true" aria-label={title}>
        <button
            type="button"
            class="absolute inset-0 bg-black/50 backdrop-blur-sm"
            on:click={close}
            aria-label="プレビューを閉じる"
        ></button>
        <div class="relative z-10 w-full max-w-3xl overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-2xl dark:border-gray-700 dark:bg-gray-800">
            <div class="flex items-center justify-between border-b border-gray-100 px-5 py-4 dark:border-gray-700">
                <h3 class="text-sm font-bold text-gray-900 dark:text-gray-100">{title}</h3>
                <button
                    on:click={close}
                    class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-200"
                    aria-label="プレビューを閉じる"
                >
                    <X size={16} />
                </button>
            </div>
            <div class="max-h-[70vh] overflow-auto p-5">
                <pre class="whitespace-pre-wrap rounded-xl bg-gray-900 p-4 font-mono text-xs text-gray-100 dark:bg-gray-950">{content}</pre>
            </div>
        </div>
    </div>
{/if}
