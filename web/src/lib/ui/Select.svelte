<script lang="ts">
    import type { HTMLSelectAttributes } from "svelte/elements";
    import { cn } from "../utils";
    import { tv } from "tailwind-variants";

    const selectRecipe = tv({
        base: "w-full appearance-none outline-none transition-[background-color,border-color,box-shadow,color] duration-200 focus:ring-2 focus:ring-brand-200 disabled:cursor-not-allowed disabled:opacity-60 dark:focus:ring-brand-900/50",
        variants: {
            uiSize: {
                sm: "min-h-10 px-3 py-2 text-sm",
                md: "min-h-11 px-4 py-2.5 text-sm",
            },
            variant: {
                default: "rounded-xl border border-gray-200 bg-gray-50 text-gray-900 shadow-sm dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100",
                ghost: "rounded-md border-none bg-transparent text-gray-900 shadow-none dark:text-gray-100",
            },
        },
        defaultVariants: {
            uiSize: "md",
            variant: "default",
        },
    });

    type UiSize = "sm" | "md";
    type Variant = "default" | "ghost";

    interface Props extends Partial<Omit<HTMLSelectAttributes, "size">> {
        value?: any;
        uiSize?: UiSize;
        variant?: Variant;
        children?: import('svelte').Snippet;
    }

    let {
        value = $bindable(""),
        id,
        name,
        disabled = false,
        uiSize = "md",
        variant = "default",
        class: className = "",
        children,
        onchange,
        oninput,
        onblur,
        onfocus,
        ...rest
    }: Props = $props();

    let classes = $derived(cn(selectRecipe({ uiSize, variant }), className));
</script>

<select
    {id}
    {name}
    disabled={Boolean(disabled)}
    bind:value
    class={classes}
    {onchange}
    {oninput}
    {onblur}
    {onfocus}
    {...rest}
>
    {@render children?.()}
</select>
