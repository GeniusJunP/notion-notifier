<script lang="ts">
    import type { Snippet } from "svelte";
    import { tv, type VariantProps } from "tailwind-variants";

    const gridRecipe = tv({
        base: "grid grid-cols-1",
        variants: {
            cols: {
                2: "md:grid-cols-2",
                3: "md:grid-cols-2 xl:grid-cols-3",
                4: "md:grid-cols-2 lg:grid-cols-4",
            },
            gap: {
                sm: "gap-4",
                md: "gap-6",
                lg: "gap-8",
            },
            smCols: {
                2: "sm:grid-cols-2",
            }
        },
        defaultVariants: {
            gap: "sm",
            cols: 2,
        },
    });

    type Variants = VariantProps<typeof gridRecipe>;

    let {
        cols = 2,
        gap = "sm",
        smCols,
        class: className,
        children,
    } = $props<{
        cols?: Variants["cols"];
        gap?: Variants["gap"];
        smCols?: Variants["smCols"];
        class?: string;
        children: Snippet;
    }>();
</script>

<div class={gridRecipe({ cols, gap, smCols, class: className })}>
    {@render children()}
</div>
