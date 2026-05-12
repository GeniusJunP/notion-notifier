<script lang="ts">
    import type { Snippet } from "svelte";
    import { tv } from "tailwind-variants";

    import { cn } from "../utils";

    const badgeRecipe = tv({
        base: "inline-flex items-center rounded-full px-2.5 py-1 font-semibold",
        variants: {
            variant: {
                neutral: "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300",
                success: "bg-emerald-100 text-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-300",
                error: "bg-red-100 text-red-700 dark:bg-red-950/60 dark:text-red-300",
                warning: "bg-amber-100 text-amber-700 dark:bg-amber-950/60 dark:text-amber-300",
                info: "bg-brand-100 text-brand-700 dark:bg-brand-950/70 dark:text-brand-300",
            },
            caps: {
                true: "text-[10px] uppercase tracking-[0.16em]",
                false: "text-xs tracking-tight",
            },
        },
        defaultVariants: {
            variant: "neutral",
            caps: true,
        },
    });

    type Variant = "neutral" | "success" | "error" | "warning" | "info";

    interface Props {
        variant?: Variant;
        caps?: boolean;
        class?: string;
        children?: Snippet;
    }

    let {
        variant = "neutral",
        caps = true,
        class: className = "",
        children,
    }: Props = $props();

    let classes = $derived(cn(badgeRecipe({ variant, caps }), className));
</script>

<span class={classes}>
    {@render children?.()}
</span>
