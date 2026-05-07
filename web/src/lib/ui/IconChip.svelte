<script lang="ts">
    import { cn } from "../utils";
    import { tv } from "tailwind-variants";
    import type { Snippet } from "svelte";

    const iconChipRecipe = tv({
        base: "flex items-center justify-center border",
        variants: {
            size: {
                sm: "size-8 rounded-lg",
                md: "size-10 rounded-xl",
                lg: "size-16 rounded-2xl",
            },
            tone: {
                brand: "border-brand-200 bg-brand-50 text-brand-700 dark:border-brand-900/70 dark:bg-brand-950/60 dark:text-brand-200",
                neutral: "border-gray-200 bg-gray-50 text-gray-600 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300",
                success: "border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900/70 dark:bg-emerald-950/60 dark:text-emerald-200",
                warning: "border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900/70 dark:bg-amber-950/60 dark:text-amber-200",
                danger: "border-red-200 bg-red-50 text-red-700 dark:border-red-900/70 dark:bg-red-950/60 dark:text-red-200",
            },
        },
        defaultVariants: {
            size: "md",
            tone: "neutral",
        },
    });

    type Tone = "brand" | "neutral" | "success" | "warning" | "danger";
    type Size = "sm" | "md" | "lg";

    interface Props {
        tone?: Tone;
        size?: Size;
        class?: string;
        children?: Snippet;
    }

    let {
        tone = "neutral",
        size = "md",
        class: className = "",
        children,
    }: Props = $props();

    let classes = $derived(cn(iconChipRecipe({ size, tone }), className));
</script>

<div class={classes}>
    {@render children?.()}
</div>
