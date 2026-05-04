<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { X } from "lucide-svelte";
    import { marked } from "marked";
    import DOMPurify from "dompurify";
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
            on:click={close}
            aria-label="プレビューを閉じる"
        ></button>

        <Card
            class="relative z-10 w-full max-w-3xl overflow-hidden shadow-2xl"
            padding="none"
            radius="2xl"
            overflowHidden
        >
            <div class="ui-modal-header">
                <h3 class="text-sm font-bold text-gray-900 dark:text-gray-100">
                    {title}
                </h3>
                <Button
                    on:click={close}
                    variant="ghost"
                    size="icon"
                    aria-label="プレビューを閉じる"
                >
                    <X size={16} />
                </Button>
            </div>
            <div class="ui-modal-body ui-scrollbar">
                {#if renderedContent}
                    <div class="ui-markdown-preview" use:setSanitizedHTML={renderedContent}>
                    </div>
                {:else}
                    <div class="ui-empty-preview">
                        プレビュー内容がありません
                    </div>
                {/if}
            </div>
        </Card>
    </div>
{/if}
