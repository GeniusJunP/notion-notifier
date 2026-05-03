<script lang="ts">
    import type { HTMLButtonAttributes } from "svelte/elements";
    import { cn } from "../utils";
    import { buttonRecipe } from "./recipes";
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

    let spinnerTone: "current" | "muted" | "inverse" = "muted";

    $: spinnerTone =
        variant === "primary"
            ? "inverse"
            : variant === "text"
              ? "current"
              : "muted";

    $: classes = cn(
        buttonRecipe({ variant, size, block }),
        variant === "primary" && "dark:bg-brand-500 dark:hover:bg-brand-400",
        variant === "secondary" &&
            "dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700",
        variant === "ghost" &&
            "dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-gray-100",
        variant === "danger" &&
            "dark:border-red-900/70 dark:bg-red-950/40 dark:text-red-300 dark:hover:bg-red-950/70",
        variant === "text" && "dark:text-gray-300 dark:hover:text-gray-100",
        "dark:focus-visible:ring-brand-700/60",
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
