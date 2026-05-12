<script lang="ts">
    import type { Snippet } from "svelte";
    import { tv } from "tailwind-variants";

    import { cn } from "../utils";

    const cardRecipe = tv({
        base: "border bg-white shadow-sm dark:bg-gray-900",
        variants: {
            tone: {
                default: "border-gray-200/80 dark:border-gray-800",
                muted: "border-gray-200/60 bg-gray-50/80 dark:border-gray-700 dark:bg-gray-800/80",
                brand: "border-brand-100 bg-brand-50/80 dark:border-brand-900/70 dark:bg-brand-950/60",
                danger: "border-red-200/70 bg-red-50/60 dark:border-red-900/70 dark:bg-red-950/40",
            },
            padding: {
                none: "",
                sm: "p-4",
                md: "p-6",
                lg: "p-8",
            },
            radius: {
                xl: "rounded-xl",
                "2xl": "rounded-2xl",
                "3xl": "rounded-3xl",
            },
            interactive: {
                true: "transition-[background-color,border-color,box-shadow,transform] duration-200 hover:border-gray-300 hover:shadow-md dark:hover:border-gray-700",
            },
            overflowHidden: {
                true: "overflow-hidden",
            },
        },
        defaultVariants: {
            tone: "default",
            padding: "md",
            radius: "2xl",
            interactive: false,
            overflowHidden: false,
        },
    });

    type Tone = "default" | "muted" | "brand" | "danger";
    type Padding = "none" | "sm" | "md" | "lg";
    type Radius = "xl" | "2xl" | "3xl";

    interface Props {
        tone?: Tone;
        padding?: Padding;
        radius?: Radius;
        interactive?: boolean;
        overflowHidden?: boolean;
        class?: string;
        children?: Snippet;
    }

    let {
        tone = "default",
        padding = "md",
        radius = "2xl",
        interactive = false,
        overflowHidden = false,
        class: className = "",
        children,
    }: Props = $props();

    let classes = $derived(cn(cardRecipe({ tone, padding, radius, interactive, overflowHidden }), className));
</script>

<div class={classes}>
    {@render children?.()}
</div>
