<script lang="ts">
    import { cn } from "../utils";

    export let label = "";
    export let forId = "";
    export let hint = "";
    export let variant: "default" | "eyebrow" = "default";

    let className = "";
    export { className as class };
</script>

<div class={className}>
    {#if label}
        <label
            for={forId}
            class={cn(
                "block font-semibold tracking-tight",
                variant === "default"
                    ? "mb-2 text-sm text-gray-700 dark:text-gray-300"
                    : "mb-2 text-xs uppercase tracking-[0.16em] text-gray-500 dark:text-gray-400",
            )}
        >
            {label}
        </label>
    {/if}

    <slot />

    {#if hint}
        <p class="mt-1 text-[11px] leading-5 text-gray-500 dark:text-gray-400">
            {hint}
        </p>
    {/if}
</div>
