<script lang="ts">
    import type { Snippet } from "svelte";
    import type { HTMLAttributes } from "svelte/elements";
    import { tv } from "tailwind-variants";

    import { cn } from "../utils";

    const typographyRecipe = tv({
        base: "m-0",
        variants: {
            variant: {
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

    type Variant = "label-caps" | "label-caps-wide" | "meta" | "strong" | "default";

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
