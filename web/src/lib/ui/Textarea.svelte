<script lang="ts">
    import type { HTMLTextareaAttributes } from "svelte/elements";
    import { cn } from "../utils";
    import { tv } from "tailwind-variants";
    
    const fieldRecipe = tv({
        base: "w-full rounded-xl border border-gray-200 bg-white text-gray-900 outline-none transition duration-200 placeholder:text-gray-400 focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder:text-gray-500 dark:focus:border-brand-500 dark:focus:ring-brand-900/50",
        variants: {
            size: {
                area: "p-4 text-sm",
            },
            mono: {
                true: "font-mono tracking-tight",
            },
        },
        defaultVariants: {
            size: "area",
            mono: false,
        },
    });

    interface Props extends Partial<Omit<HTMLTextareaAttributes, "size">> {
        value?: string | number | string[] | null;
        mono?: boolean;
    }

    let {
        value = $bindable(""),
        id,
        name,
        placeholder,
        rows = 4,
        disabled = false,
        mono = false,
        class: className = "",
        oninput,
        onchange,
        onblur,
        onfocus,
        ...rest
    }: Props = $props();

    let classes = $derived(cn(
        fieldRecipe({ size: "area", mono }),
        className,
    ));
</script>

<textarea
    {id}
    {name}
    {placeholder}
    {rows}
    {disabled}
    bind:value
    class={classes}
    {onblur}
    {onchange}
    {onfocus}
    {oninput}
    {...rest}
></textarea>
