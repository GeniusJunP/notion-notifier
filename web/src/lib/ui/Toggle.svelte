<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { cn } from "../utils";

    type Tone = "brand" | "success" | "warning";
    type Size = "sm" | "md";

    export let checked = false;
    export let disabled = false;
    export let ariaLabel = "";
    export let tone: Tone = "brand";
    export let size: Size = "md";

    let className = "";
    export { className as class };

    const dispatch = createEventDispatcher<{ change: boolean }>();

    const sizeClasses = {
        sm: {
            track: "h-6 w-10",
            thumb: "h-4 w-4",
            on: "translate-x-4",
            off: "translate-x-0",
        },
        md: {
            track: "h-8 w-14",
            thumb: "h-6 w-6",
            on: "translate-x-6",
            off: "translate-x-0",
        },
    } as const;

    const activeClasses = {
        brand: "bg-brand-600 dark:bg-brand-500",
        success: "bg-emerald-600 dark:bg-emerald-500",
        warning: "bg-amber-500 dark:bg-amber-500",
    } as const;

    function toggle() {
        if (disabled) return;
        checked = !checked;
        dispatch("change", checked);
    }
</script>

<button
    type="button"
    role="switch"
    aria-checked={checked}
    aria-label={ariaLabel}
    {disabled}
    class={cn(
        "relative inline-flex shrink-0 items-center rounded-full border border-transparent p-1 shadow-inner outline-none transition-[background-color,box-shadow,transform] duration-200",
        "focus-visible:ring-2 focus-visible:ring-brand-300/70 dark:focus-visible:ring-brand-700/60",
        "disabled:cursor-not-allowed disabled:opacity-50",
        sizeClasses[size].track,
        checked
            ? activeClasses[tone]
            : "bg-gray-200 dark:bg-gray-700",
        className,
    )}
    on:click={toggle}
>
    <span
        class={cn(
            "rounded-full bg-white shadow-sm transition-transform duration-200",
            sizeClasses[size].thumb,
            checked ? sizeClasses[size].on : sizeClasses[size].off,
        )}
    ></span>
</button>
