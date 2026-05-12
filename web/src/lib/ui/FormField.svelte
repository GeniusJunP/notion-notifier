<script lang="ts">
    import type { Snippet } from "svelte";
    import { tv } from "tailwind-variants";

    const formFieldRecipe = tv({
        slots: {
            label: "block font-semibold tracking-tight",
            hintText: "mt-1 text-[11px] leading-5 text-gray-500 dark:text-gray-400",
        },
        variants: {
            variant: {
                default: { label: "mb-2 text-sm text-gray-700 dark:text-gray-300" },
                eyebrow: { label: "mb-2 text-xs uppercase tracking-[0.16em] text-gray-500 dark:text-gray-400" },
            },
        },
        defaultVariants: {
            variant: "default",
        },
    });

    type Variant = "default" | "eyebrow";

    interface Props {
        label?: string;
        forId?: string;
        hint?: string;
        variant?: Variant;
        class?: string;
        children?: Snippet;
    }

    let {
        label = "",
        forId = "",
        hint = "",
        variant = "default",
        class: className = "",
        children,
    }: Props = $props();

    let styles = $derived(formFieldRecipe({ variant }));
</script>

<div class={className}>
    {#if label}
        <label
            for={forId}
            class={styles.label()}
        >
            {label}
        </label>
    {/if}

    {@render children?.()}

    {#if hint}
        <p class={styles.hintText()}>
            {hint}
        </p>
    {/if}
</div>
