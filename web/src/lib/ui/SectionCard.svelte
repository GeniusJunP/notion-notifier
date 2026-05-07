<script lang="ts">
    import { cn } from "../utils";
    import Card from "./Card.svelte";
    import type { Snippet } from "svelte";

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
        radius = "3xl",
        interactive = false,
        overflowHidden = false,
        class: className = "",
        children,
    }: Props = $props();
</script>

<Card
    {tone}
    {padding}
    {radius}
    {interactive}
    {overflowHidden}
    class={cn("space-y-6", className)}
>
    {@render children?.()}
</Card>

