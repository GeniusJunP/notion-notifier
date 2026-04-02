<script lang="ts">
    import type { HTMLSelectAttributes } from "svelte/elements";
    import { cn } from "../utils";

    type UiSize = "sm" | "md";
    type Variant = "default" | "ghost";

    export let value: HTMLSelectAttributes["value"] = "";
    export let id: HTMLSelectAttributes["id"] = undefined;
    export let name: HTMLSelectAttributes["name"] = undefined;
    export let disabled: HTMLSelectAttributes["disabled"] = false;
    export let uiSize: UiSize = "md";
    export let variant: Variant = "default";

    let className = "";
    export { className as class };

    const sizeClasses = {
        sm: "min-h-10 px-3 py-2 text-sm",
        md: "min-h-11 px-4 py-2.5 text-sm",
    } as const;

    const variantClasses = {
        default:
            "rounded-xl border border-gray-200 bg-gray-50 text-gray-900 shadow-sm dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100",
        ghost: "rounded-md border-none bg-transparent text-gray-900 shadow-none dark:text-gray-100",
    } as const;

    $: classes = cn(
        "w-full appearance-none outline-none transition-[background-color,border-color,box-shadow,color] duration-200",
        "focus:ring-2 focus:ring-brand-200 dark:focus:ring-brand-900/50",
        "disabled:cursor-not-allowed disabled:opacity-60",
        sizeClasses[uiSize],
        variantClasses[variant],
        className,
    );

    function handleChange(event: Event) {
        value = (event.currentTarget as HTMLSelectElement).value;
    }
</script>

<select
    {id}
    {name}
    disabled={Boolean(disabled)}
    {value}
    class={classes}
    on:change={handleChange}
>
    <slot />
</select>
