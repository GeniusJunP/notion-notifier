<script lang="ts">
    import { cn } from "../utils";
    import Card from "./Card.svelte";

    type Tone = "default" | "muted" | "brand" | "danger";
    type Padding = "none" | "sm" | "md" | "lg";
    type Radius = "xl" | "2xl" | "3xl";

    export let tone: Tone = "default";
    export let padding: Padding = "md";
    export let radius: Radius = "3xl";
    export let interactive = false;
    export let overflowHidden = false;

    let className = "";
    export { className as class };
</script>

<Card
    {tone}
    {padding}
    {radius}
    {interactive}
    {overflowHidden}
    class={cn("space-y-6", className)}
>
    <slot />
</Card>

