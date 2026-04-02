<script lang="ts">
    import type { HTMLButtonAttributes } from "svelte/elements";
    import { cn } from "../utils";
    import Spinner from "./Spinner.svelte";

    type Variant =
        | "primary"
        | "secondary"
        | "ghost"
        | "danger"
        | "text";
    type Size = "sm" | "md" | "lg" | "icon";

    interface $$Props extends HTMLButtonAttributes {
        class?: string;
        variant?: Variant;
        size?: Size;
        loading?: boolean;
        block?: boolean;
    }

    export let variant: Variant = "primary";
    export let size: Size = "md";
    export let loading = false;
    export let block = false;
    export let disabled: HTMLButtonAttributes["disabled"] = false;
    export let type: HTMLButtonAttributes["type"] = "button";

    let className = "";
    export { className as class };

    const variantClasses = {
        primary:
            "bg-brand-600 text-white shadow-sm hover:bg-brand-700 dark:bg-brand-500 dark:hover:bg-brand-400",
        secondary:
            "border border-gray-200 bg-gray-100 text-gray-700 shadow-sm hover:bg-gray-200 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700",
        ghost:
            "text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-gray-100",
        danger:
            "border border-red-200 bg-red-50 text-red-700 shadow-sm hover:bg-red-100 dark:border-red-900/70 dark:bg-red-950/40 dark:text-red-300 dark:hover:bg-red-950/70",
        text: "text-gray-600 hover:text-gray-900 dark:text-gray-300 dark:hover:text-gray-100",
    } as const;

    const sizeClasses = {
        sm: "min-h-9 rounded-lg px-3 py-2 text-xs",
        md: "min-h-10 rounded-xl px-4 py-2.5 text-sm",
        lg: "min-h-12 rounded-2xl px-6 py-3 text-sm",
        icon: "h-10 w-10 rounded-xl p-0",
    } as const;

    let spinnerTone: "current" | "muted" | "inverse" = "muted";

    $: spinnerTone =
        variant === "primary"
            ? "inverse"
            : variant === "text"
              ? "current"
              : "muted";

    $: classes = cn(
        "inline-flex items-center justify-center gap-2 font-semibold tracking-tight outline-none transition-[background-color,border-color,color,box-shadow,transform] duration-200",
        "focus-visible:ring-2 focus-visible:ring-brand-300/70 dark:focus-visible:ring-brand-700/60",
        "active:scale-[0.985] disabled:pointer-events-none disabled:opacity-50",
        sizeClasses[size],
        variantClasses[variant],
        block && "w-full",
        className,
    );
</script>

<button
    {...$$restProps}
    {type}
    disabled={Boolean(disabled || loading)}
    class={classes}
    on:click
>
    {#if loading}
        <Spinner size={size === "lg" ? "md" : "sm"} tone={spinnerTone} />
    {/if}
    <slot />
</button>
