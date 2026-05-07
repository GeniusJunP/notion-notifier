<script lang="ts">
    import { cn } from "../utils";
    import { tv } from "tailwind-variants";
    import Card from "./Card.svelte";
    import type { Snippet } from "svelte";

    const panelRecipe = tv({
        slots: {
            headerWrapper: "ui-panel-header",
            bodyWrapper: "",
        },
        variants: {
            padding: {
                none: { headerWrapper: "", bodyWrapper: "" },
                sm: { headerWrapper: "p-4", bodyWrapper: "p-4" },
                md: { headerWrapper: "px-6 py-5", bodyWrapper: "p-6" },
                lg: { headerWrapper: "px-8 py-6", bodyWrapper: "p-8" },
            },
        },
        defaultVariants: {
            padding: "md",
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
        headerClass?: string;
        bodyClass?: string;
        class?: string;
        header?: Snippet;
        actions?: Snippet;
        children?: Snippet;
    }

    let {
        tone = "default",
        padding = "md",
        radius = "2xl",
        interactive = false,
        headerClass = "",
        bodyClass = "",
        class: className = "",
        header,
        actions,
        children,
    }: Props = $props();

    let styles = $derived(panelRecipe({ padding }));
</script>

<Card
    {tone}
    padding="none"
    {radius}
    {interactive}
    overflowHidden
    class={className}
>
    <div class={cn(styles.headerWrapper(), headerClass)}>
        <div class="min-w-0 flex-1">
            {@render header?.()}
        </div>
        {#if actions}
            <div class="shrink-0">
                {@render actions()}
            </div>
        {/if}
    </div>
    <div class={cn(styles.bodyWrapper(), bodyClass)}>
        {@render children?.()}
    </div>
</Card>

