<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { X } from "lucide-svelte";
    import { marked } from "marked";
    import DOMPurify from "dompurify";

    export let open = false;
    export let title = "プレビュー";
    export let content = "";

    const dispatch = createEventDispatcher<{ close: void }>();
    const markdownOptions = {
        gfm: true,
        breaks: true,
    } as const;

    function renderMarkdown(source: string): string {
        if (!source.trim()) {
            return "";
        }
        const html = marked.parse(source, markdownOptions) as string;
        return DOMPurify.sanitize(html);
    }

    $: renderedContent = renderMarkdown(content);

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
                {#if renderedContent}
                    <div class="markdown-preview rounded-xl border border-gray-200 bg-gray-50 p-4 text-sm text-gray-800 dark:border-gray-700 dark:bg-gray-900/60 dark:text-gray-100">
                        {@html renderedContent}
                    </div>
                {:else}
                    <div class="rounded-xl border border-dashed border-gray-300 bg-gray-50 p-4 text-sm text-gray-500 dark:border-gray-700 dark:bg-gray-900/40 dark:text-gray-400">
                        プレビュー内容がありません
                    </div>
                {/if}
            </div>
        </div>
    </div>
{/if}

<style>
    .markdown-preview {
        word-break: break-word;
    }

    .markdown-preview :global(h1),
    .markdown-preview :global(h2),
    .markdown-preview :global(h3) {
        margin-top: 1.2rem;
        margin-bottom: 0.5rem;
        font-weight: 700;
        line-height: 1.4;
    }

    .markdown-preview :global(h1) {
        font-size: 1.25rem;
    }

    .markdown-preview :global(h2) {
        font-size: 1.1rem;
    }

    .markdown-preview :global(h3) {
        font-size: 1rem;
    }

    .markdown-preview :global(p) {
        margin: 0.5rem 0;
    }

    .markdown-preview :global(ul),
    .markdown-preview :global(ol) {
        margin: 0.6rem 0;
        padding-left: 1.25rem;
    }

    .markdown-preview :global(li) {
        margin: 0.2rem 0;
    }

    .markdown-preview :global(a) {
        color: rgb(37 99 235);
        text-decoration: underline;
    }

    .markdown-preview :global(strong) {
        font-weight: 700;
    }

    .markdown-preview :global(code) {
        border-radius: 0.375rem;
        background-color: rgb(229 231 235);
        padding: 0.1rem 0.3rem;
        font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
        font-size: 0.82em;
    }

    .markdown-preview :global(pre) {
        overflow-x: auto;
        border-radius: 0.75rem;
        background-color: rgb(17 24 39);
        padding: 0.85rem 1rem;
        color: rgb(243 244 246);
    }

    .markdown-preview :global(pre code) {
        background: transparent;
        padding: 0;
        color: inherit;
        font-size: 0.9em;
    }

    .markdown-preview :global(blockquote) {
        margin: 0.75rem 0;
        border-left: 3px solid rgb(209 213 219);
        padding-left: 0.85rem;
        color: rgb(75 85 99);
    }

    :global(.dark) .markdown-preview :global(code) {
        background-color: rgb(55 65 81);
    }

    :global(.dark) .markdown-preview :global(a) {
        color: rgb(96 165 250);
    }

    :global(.dark) .markdown-preview :global(blockquote) {
        border-left-color: rgb(75 85 99);
        color: rgb(209 213 219);
    }
</style>
