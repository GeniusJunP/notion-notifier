<script lang="ts">
    import Card from "../../lib/ui/Card.svelte";
    import IconChip from "../../lib/ui/IconChip.svelte";

    type Tone = "brand" | "warning" | "success" | "danger" | "neutral";

    export let label: string;
    export let value: string | number;
    export let subLabel: string;
    export let tone: Tone = "neutral";
</script>

<Card radius="2xl" class="flex flex-col justify-between">
    <div class="mb-4 flex items-center justify-between">
        <span class="text-sm font-medium text-gray-500 dark:text-gray-400">
            {label}
        </span>
        <IconChip tone={tone} size="md" class="border-transparent">
            <slot name="icon"></slot>
        </IconChip>
    </div>
    <div>
        <h3 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
            {value}
        </h3>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {subLabel}
        </p>
    </div>
</Card>
