<script lang="ts">
    import { cn } from "../lib/utils";
    import type { HTMLButtonAttributes } from "svelte/elements";

    let {
        active = false,
        justifyBetween = false,
        disabled = false,
        ariaCurrent = undefined,
        class: className = "",
        children,
        ...rest
    }: HTMLButtonAttributes & {
        active?: boolean;
        justifyBetween?: boolean;
        ariaCurrent?: "page" | undefined;
    } = $props();

    const variantClass = $derived(
        active
            ? "border-gray-200/80 bg-white/85 text-gray-900 shadow-sm dark:border-gray-800 dark:bg-gray-900/75 dark:text-gray-100"
            : "text-gray-600 dark:text-gray-400"
    );

    const layoutClass = $derived(
        justifyBetween
            ? "items-center justify-between"
            : "items-center gap-3"
    );
</script>

<button
    {...rest}
    {disabled}
    aria-current={ariaCurrent}
    class={cn(
        "ui-interactive-tile group flex w-full px-3 py-2.5 text-sm font-medium tracking-tight outline-none",
        "focus-visible:ring-2 focus-visible:ring-brand-300/70 dark:focus-visible:ring-brand-700/60",
        "active:scale-[0.985] disabled:pointer-events-none disabled:opacity-50",
        layoutClass,
        variantClass,
        className,
    )}
>
    {#if children}
        {@render children()}
    {/if}
</button>
