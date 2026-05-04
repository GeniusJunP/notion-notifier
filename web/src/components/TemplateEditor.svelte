<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { Code2, Eye, Plus, RotateCcw, Trash2 } from "lucide-svelte";
    import Button from "../lib/ui/Button.svelte";
    import Card from "../lib/ui/Card.svelte";
    import FormField from "../lib/ui/FormField.svelte";
    import Input from "../lib/ui/Input.svelte";
    import Select from "../lib/ui/Select.svelte";
    import Textarea from "../lib/ui/Textarea.svelte";
    import {
        blocksFromTemplate,
        createTemplateBlock,
        serializeTemplateBlocks,
        templateVariables,
        type TemplateBlock,
        type TemplateBlockKind,
        type TemplateEditorMode,
    } from "../lib/templateEditor";

    export let value = "";
    export let id = "";
    export let label = "テンプレート";
    export let placeholder = "Go テンプレート形式で入力...";
    export let rows = 6;
    export let mode: TemplateEditorMode = "raw";
    export let previewLoading = false;
    export let showPreview = true;
    export let showReset = true;

    const dispatch = createEventDispatcher<{
        preview: void;
        reset: void;
    }>();

    let blocks: TemplateBlock[] = blocksFromTemplate(value);

    const kindOptions = [
        { value: "text", label: "テキスト" },
        { value: "variable", label: "変数" },
        { value: "if", label: "if" },
        { value: "with", label: "with" },
        { value: "range", label: "range" },
    ] satisfies { value: TemplateBlockKind; label: string }[];
    const variableOptions = templateVariables.map((v) => ({ value: v, label: v }));

    function syncFromBlocks() {
        value = serializeTemplateBlocks(blocks);
    }

    function enterVisualMode() {
        if (mode === "visual") return;
        blocks = blocksFromTemplate(value);
        mode = "visual";
    }

    function enterRawMode() {
        mode = "raw";
    }

    function addBlock(kind: TemplateBlockKind) {
        blocks = [...blocks, createTemplateBlock(kind)];
        syncFromBlocks();
    }

    function removeBlock(index: number) {
        blocks = blocks.filter((_, i) => i !== index);
        syncFromBlocks();
    }

    function updateBlock(index: number, block: TemplateBlock) {
        blocks[index] = block;
        blocks = blocks;
        syncFromBlocks();
    }

    function updateControlBody(index: number, event: Event) {
        const block = blocks[index];
        block.children = block.children?.length
            ? block.children
            : [createTemplateBlock("text")];
        block.children[0].content = (event.currentTarget as HTMLTextAreaElement).value;
        updateBlock(index, block);
    }

</script>

<div class="space-y-3">
    <div class="flex flex-wrap items-center justify-between gap-3">
        <span class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
            {label}
        </span>
        <div class="flex items-center gap-2">
            <Button
                type="button"
                variant={mode === "visual" ? "secondary" : "ghost"}
                size="sm"
                on:click={enterVisualMode}
            >
                <Eye size={14} /> Visual
            </Button>
            <Button
                type="button"
                variant={mode === "raw" ? "secondary" : "ghost"}
                size="sm"
                on:click={enterRawMode}
            >
                <Code2 size={14} /> Raw
            </Button>
        </div>
    </div>

    {#if mode === "raw"}
        <Textarea
            {id}
            bind:value
            {placeholder}
            {rows}
            class="min-h-[120px]"
            mono
        />
    {:else}
        <Card tone="muted" radius="2xl" padding="sm" class="space-y-3">
            {#each blocks as block, index (index)}
                <div class="rounded-xl border border-gray-200/80 bg-white/80 p-3 dark:border-gray-800 dark:bg-gray-950/60">
                    <div class="mb-3 flex items-center gap-2">
                        <Select bind:value={block.kind} uiSize="sm" on:change={() => updateBlock(index, block)}>
                            {#each kindOptions as option (option.value)}
                                <option value={option.value}>{option.label}</option>
                            {/each}
                        </Select>
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            aria-label="ブロックを削除"
                            on:click={() => removeBlock(index)}
                        >
                            <Trash2 size={14} />
                        </Button>
                    </div>

                    {#if block.kind === "text"}
                        <Textarea
                            bind:value={block.content}
                            rows={3}
                            mono
                            on:input={() => updateBlock(index, block)}
                        />
                    {:else if block.kind === "variable"}
                        <Select bind:value={block.content} on:change={() => updateBlock(index, block)}>
                            {#each variableOptions as option (option.value)}
                                <option value={option.value}>{option.label}</option>
                            {/each}
                        </Select>
                    {:else}
                        <FormField label={`${block.kind} 条件`} variant="eyebrow">
                            <Input
                                bind:value={block.content}
                                mono
                                on:input={() => updateBlock(index, block)}
                            />
                        </FormField>
                        <p class="ui-hint">
                            出力時は <code class="ui-inline-code">{"{{- ... -}}"}</code> を使い、UI整形用の改行を通知本文へ出しません。
                        </p>
                        <Textarea
                            value={block.children?.[0]?.content ?? ""}
                            rows={3}
                            mono
                            on:input={(event) => updateControlBody(index, event)}
                        />
                    {/if}
                </div>
            {/each}

            <div class="flex flex-wrap gap-2">
                {#each kindOptions as option (option.value)}
                    <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        on:click={() => addBlock(option.value)}
                    >
                        <Plus size={12} /> {option.label}
                    </Button>
                {/each}
            </div>
        </Card>
    {/if}

    <div class="flex items-center gap-3">
        {#if showPreview}
            <Button
                type="button"
                on:click={() => dispatch("preview")}
                disabled={previewLoading}
                loading={previewLoading}
                variant="text"
                size="sm"
            >
                <Eye size={12} /> プレビューを実行
            </Button>
        {/if}
        {#if showReset}
            <Button
                type="button"
                on:click={() => dispatch("reset")}
                variant="ghost"
                size="sm"
            >
                <RotateCcw size={12} /> デフォルトに戻す
            </Button>
        {/if}
    </div>
</div>
