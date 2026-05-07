<script lang="ts">
    import { cn } from "../utils";
    import { tv } from "tailwind-variants";

    const spinnerRecipe = tv({
        base: "animate-spin rounded-full",
        variants: {
            size: {
                sm: "size-4 border-2",
                md: "size-5 border-2",
            },
            tone: {
                current: "border-current/20 border-t-current",
                muted: "border-gray-400/30 border-t-gray-600 dark:border-gray-500/30 dark:border-t-gray-200",
                inverse: "border-white/25 border-t-white",
            },
        },
        defaultVariants: {
            size: "sm",
            tone: "current",
        },
    });

    type Size = "sm" | "md";
    type Tone = "current" | "muted" | "inverse";

    interface Props {
        size?: Size;
        tone?: Tone;
        class?: string;
    }

    let {
        size = "sm",
        tone = "current",
        class: className = "",
    }: Props = $props();

    let classes = $derived(cn(spinnerRecipe({ size, tone }), className));
</script>

<div
    aria-hidden="true"
    class={classes}
></div>
