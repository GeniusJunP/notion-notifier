<script lang="ts">
    import type { Snippet } from "svelte";
    import type { HTMLAttributes } from "svelte/elements";
    import { tv } from "tailwind-variants";

    import { cn } from "../utils";

    const typographyRecipe = tv({
        base: "m-0",
        variants: {
            variant: {
                "page-title": "text-2xl font-bold tracking-tight text-gray-900 dark:text-gray-100",
                "page-subtitle": "text-sm text-gray-500 dark:text-gray-400",
                "block-title": "text-xl font-bold tracking-tight text-gray-800 dark:text-gray-100",
                "section-title": "flex items-center gap-2 text-lg font-bold text-gray-800 dark:text-gray-200",
                "support-text": "text-sm leading-relaxed text-gray-600 dark:text-gray-300",
                "hint": "mt-1 text-[11px] leading-5 text-gray-500 dark:text-gray-400",
                "inline-code": "rounded bg-white/70 px-1 py-0.5 font-mono dark:bg-gray-900/60",
                "label-caps": "text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400",
                "label-caps-wide": "text-xs font-semibold uppercase tracking-[0.16em] text-gray-500 dark:text-gray-400",
                "meta": "text-xs text-gray-500 dark:text-gray-400",
                "strong": "text-sm font-semibold text-gray-900 dark:text-gray-100",
                "default": "text-sm text-gray-900 dark:text-gray-100",
            },
            block: {
                true: "block",
            }
        },
        defaultVariants: {
            variant: "default",
            block: false,
        },
    });

    type Variant = 
        | "page-title" 
        | "page-subtitle" 
        | "block-title" 
        | "section-title" 
        | "support-text" 
        | "hint" 
        | "inline-code" 
        | "label-caps" 
        | "label-caps-wide" 
        | "meta" 
        | "strong" 
        | "default";

    interface Props extends HTMLAttributes<HTMLElement> {
        variant?: Variant | "default";
        block?: boolean;
        as?: string;
        children?: Snippet;
    }

    let { 
        variant = "default", 
        block = false,
        as = "span", 
        class: className = "", 
        children, ...rest 
    }: Props = $props();

</script>

<svelte:element this={as} class={cn(typographyRecipe({ variant, block }), className)} {...rest}>
    {@render children?.()}
</svelte:element>
