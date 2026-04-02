<script lang="ts">
    import { cn } from "../lib/utils";

    export let active = false;
    export let justifyBetween = false;
    export let disabled = false;
    export let ariaCurrent: "page" | undefined = undefined;

    let className = "";
    export { className as class };

    $: variantClass = active
        ? "border-gray-200/80 bg-white/85 text-gray-900 shadow-sm dark:border-gray-800 dark:bg-gray-900/75 dark:text-gray-100"
        : "text-gray-600 dark:text-gray-400";

    $: layoutClass = justifyBetween
        ? "items-center justify-between"
        : "items-center gap-3";
</script>

<button
    on:click
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
    <slot />
</button>
