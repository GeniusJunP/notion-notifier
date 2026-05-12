<script lang="ts">
    import DOMPurify from "dompurify";
    import { X } from "lucide-svelte";
    import { marked } from "marked";
    import { createEventDispatcher } from "svelte";

    import Button from "../lib/ui/Button.svelte";
    import Card from "../lib/ui/Card.svelte";

    export let open = false;
    export let title = "プレビュー";
    export let content = "";
    export let mode: "webhook" | "guide" = "webhook";

    const dispatch = createEventDispatcher<{ close: void }>();

    function setSanitizedHTML(node: HTMLElement, html: string) {
        node.innerHTML = html;
        return {
            update(newHTML: string) {
                node.innerHTML = newHTML;
            }
        };
    }
    const markdownOptions = {
        gfm: true,
        breaks: true,
    } as const;
    const webhookRenderer = new marked.Renderer();
    webhookRenderer.space = (token: { raw?: string }) => {
        const raw = typeof token.raw === "string" ? token.raw : "";
        const newlineCount = raw.split("\n").length - 1;
        if (newlineCount <= 0) {
            return "";
        }
        return "<br>".repeat(newlineCount);
    };

    function renderMarkdown(source: string, currentMode: "webhook" | "guide"): string {
        if (!source.trim()) {
            return "";
        }
        const html =
            currentMode === "webhook"
                ? (marked.parse(source, {
                      ...markdownOptions,
                      renderer: webhookRenderer,
                  }) as string)
                : (marked.parse(source, markdownOptions) as string);
        return DOMPurify.sanitize(html);
    }

    $: renderedContent = renderMarkdown(content, mode);

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
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
        role="dialog"
        aria-modal="true"
        aria-label={title}
    >
        <button
            type="button"
            class="absolute inset-0 bg-black/50 backdrop-blur-sm"
            onclick={close}
            aria-label="プレビューを閉じる"
        ></button>

        <Card
            class="relative z-10 w-full max-w-3xl overflow-hidden shadow-2xl"
            padding="none"
            radius="2xl"
            overflowHidden
        >
            <div class="flex items-center justify-between border-b border-gray-200 px-5 py-3 dark:border-gray-800">
                <h3 class="text-sm font-bold text-gray-900 dark:text-gray-100">
                    {title}
                </h3>
                <Button
                    onclick={close}
                    variant="ghost"
                    size="icon"
                    aria-label="プレビューを閉じる"
                >
                    <X size={16} />
                </Button>
            </div>
            <div class="max-h-[70vh] overflow-auto p-5 custom-scrollbar">
                {#if renderedContent}
                    <div class="markdown-preview" use:setSanitizedHTML={renderedContent}>
                    </div>
                {:else}
                    <div class="rounded-xl border border-dashed border-gray-300 bg-gray-50/70 p-4 text-sm text-gray-500 dark:border-gray-800 dark:bg-gray-900/50 dark:text-gray-400">
                        プレビュー内容がありません
                    </div>
                {/if}
            </div>
        </Card>
    </div>
{/if}
