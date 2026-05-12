<script lang="ts">
    import { Code2, Eye, Plus, RotateCcw, Trash2 } from "lucide-svelte";
    import Button from "../lib/ui/Button.svelte";
    import Card from "../lib/ui/Card.svelte";
    import FormField from "../lib/ui/FormField.svelte";
    import Input from "../lib/ui/Input.svelte";
    import Select from "../lib/ui/Select.svelte";
    import Textarea from "../lib/ui/Textarea.svelte";
    import Typography from "../lib/ui/Typography.svelte";
    import {
        blocksFromTemplate,
        createTemplateBlock,
        serializeTemplateBlocks,
        templateVariables,
        type TemplateBlock,
        type TemplateBlockKind,
        type TemplateEditorMode,
    } from "../lib/templateEditor";

    let {
        value = $bindable(""),
        id = "",
        label = "テンプレート",
        placeholder = "Go テンプレート形式で入力...",
        rows = 6,
        mode = $bindable("raw"),
        previewLoading = false,
        showPreview = true,
        showReset = true,
        onpreview,
        onreset
    } = $props<{
        value?: string;
        id?: string;
        label?: string;
        placeholder?: string;
        rows?: number;
        mode?: TemplateEditorMode;
        previewLoading?: boolean;
        showPreview?: boolean;
        showReset?: boolean;
        onpreview?: () => void;
        onreset?: () => void;
    }>();

    let blocks = $state<TemplateBlock[]>([]);

    $effect(() => {
        // Sync visual blocks if entering visual mode
        if (mode === "visual" && blocks.length === 0 && value) {
            blocks = blocksFromTemplate(value);
        }
    });

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
        syncFromBlocks();
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
        <Typography variant="label-caps">
            {label}
        </Typography>
        <div class="flex items-center gap-2">
            <Button
                type="button"
                variant={mode === "visual" ? "secondary" : "ghost"}
                size="sm"
                onclick={enterVisualMode}
            >
                <Eye size={14} /> Visual
            </Button>
            <Button
                type="button"
                variant={mode === "raw" ? "secondary" : "ghost"}
                size="sm"
                onclick={enterRawMode}
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
                        <Select bind:value={block.kind} uiSize="sm" onchange={() => updateBlock(index, block)}>
                            {#each kindOptions as option (option.value)}
                                <option value={option.value}>{option.label}</option>
                            {/each}
                        </Select>
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            aria-label="ブロックを削除"
                            onclick={() => removeBlock(index)}
                        >
                            <Trash2 size={14} />
                        </Button>
                    </div>

                    {#if block.kind === "text"}
                        <Textarea
                            bind:value={block.content}
                            rows={3}
                            mono
                            oninput={() => updateBlock(index, block)}
                        />
                    {:else if block.kind === "variable"}
                        <Select bind:value={block.content} onchange={() => updateBlock(index, block)}>
                            {#each variableOptions as option (option.value)}
                                <option value={option.value}>{option.label}</option>
                            {/each}
                        </Select>
                    {:else}
                        <FormField label={`${block.kind} 条件`} variant="eyebrow">
                            <Input
                                bind:value={block.content}
                                mono
                                oninput={() => updateBlock(index, block)}
                            />
                        </FormField>
                        <p class="ui-hint">
                            出力時は <code class="ui-inline-code">{"{{- ... -}}"}</code> を使い、UI整形用の改行を通知本文へ出しません。
                        </p>
                        <Textarea
                            value={block.children?.[0]?.content ?? ""}
                            rows={3}
                            mono
                            oninput={(event) => updateControlBody(index, event)}
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
                        onclick={() => addBlock(option.value)}
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
                onclick={() => onpreview?.()}
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
                onclick={() => onreset?.()}
                variant="ghost"
                size="sm"
            >
                <RotateCcw size={12} /> デフォルトに戻す
            </Button>
        {/if}
    </div>
</div>
