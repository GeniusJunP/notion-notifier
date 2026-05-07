<script lang="ts">
    import type { HTMLInputAttributes } from "svelte/elements";
    import { cn } from "../utils";
    import { tv } from "tailwind-variants";
    
    const fieldRecipe = tv({
        base: "w-full rounded-xl border border-gray-200 bg-white text-gray-900 outline-none transition duration-200 placeholder:text-gray-400 focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder:text-gray-500 dark:focus:border-brand-500 dark:focus:ring-brand-900/50",
        variants: {
            size: {
                sm: "min-h-9 px-3 py-2 text-sm",
                md: "min-h-10 px-4 py-2.5 text-sm",
            },
            mono: {
                true: "font-mono text-sm tracking-tight",
            },
        },
        defaultVariants: {
            size: "md",
            mono: false,
        },
    });

    type UiSize = "sm" | "md";

    let {
        value = $bindable(""),
        type = "text",
        id = undefined,
        name = undefined,
        placeholder = undefined,
        disabled = false,
        uiSize = "md" as UiSize,
        mono = false,
        min = undefined,
        max = undefined,
        step = undefined,
        onchange,
        oninput,
        onfocus,
        onblur,
        class: className = "",
        ...rest
    }: HTMLInputAttributes & {
        uiSize?: UiSize;
        mono?: boolean;
    } = $props();

    function handleInput(event: Event & { currentTarget: EventTarget & HTMLInputElement }) {
        const target = event.currentTarget;
        if (type === "number") {
            value = Number.isNaN(target.valueAsNumber)
                ? ""
                : target.valueAsNumber;
            return;
        }
        value = target.value;
        oninput?.(event as any);
    }

    function handleChange(event: Event & { currentTarget: EventTarget & HTMLInputElement }) {
        if (type !== "number") {
            onchange?.(event as any);
            return;
        }
        const target = event.currentTarget;
        value = Number.isNaN(target.valueAsNumber)
            ? ""
            : target.valueAsNumber;
        onchange?.(event as any);
    }

    const classes = $derived(cn(
        fieldRecipe({ size: uiSize, mono }),
        className,
    ));
</script>

<input
    {...rest}
    {id}
    {name}
    {type}
    {placeholder}
    {disabled}
    {min}
    {max}
    {step}
    {value}
    class={classes}
    onblur={(e) => { onblur?.(e); }}
    onchange={handleChange}
    onfocus={(e) => { onfocus?.(e); }}
    oninput={handleInput}
/>
