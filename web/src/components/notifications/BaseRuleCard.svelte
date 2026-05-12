<script lang="ts">
    import { Trash2 } from "lucide-svelte";
    import Button from "../../lib/ui/Button.svelte";
    import Panel from "../../lib/ui/Panel.svelte";
    import Toggle from "../../lib/ui/Toggle.svelte";
    import type { Snippet } from "svelte";

    let {
        title,
        index,
        enabled = $bindable(),
        children,
        onremove
    } = $props<{
        title: string;
        index: number;
        enabled: boolean;
        children: Snippet;
        onremove: () => void;
    }>();
</script>

<Panel radius="2xl" interactive bodyClass="grid grid-cols-1 gap-8">
    {#snippet header()}
        <div class="flex items-center gap-3">
            <div
                class="flex h-8 w-8 items-center justify-center rounded-lg border border-gray-200 bg-white text-sm font-semibold text-gray-500 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-400"
            >
                {index + 1}
            </div>
            <Toggle
                bind:checked={enabled}
                ariaLabel={`${title} の有効化`}
                tone="success"
                size="sm"
            />
            <span class="font-semibold text-gray-900 dark:text-gray-100">
                {title}
            </span>
        </div>
    {/snippet}
    {#snippet actions()}
        <Button
            type="button"
            onclick={() => onremove()}
            variant="ghost"
            size="icon"
            aria-label={`${title} を削除`}
        >
            <Trash2 size={18} />
        </Button>
    {/snippet}

    {@render children()}
</Panel>
