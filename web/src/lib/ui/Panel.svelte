<script lang="ts">
    import { cn } from "../utils";
    import Card from "./Card.svelte";

    type Tone = "default" | "muted" | "brand" | "danger";
    type Padding = "none" | "sm" | "md" | "lg";
    type Radius = "xl" | "2xl" | "3xl";

    export let tone: Tone = "default";
    export let padding: Padding = "md";
    export let radius: Radius = "2xl";
    export let interactive = false;

    export let headerClass = "";
    export let bodyClass = "";

    let className = "";
    export { className as class };

    const headerPaddingClasses: Record<Padding, string> = {
        none: "",
        sm: "px-4 py-4",
        md: "px-6 py-5",
        lg: "px-8 py-6",
    } as const;

    const bodyPaddingClasses: Record<Padding, string> = {
        none: "",
        sm: "p-4",
        md: "p-6",
        lg: "p-8",
    } as const;
</script>

<Card
    {tone}
    padding="none"
    {radius}
    {interactive}
    overflowHidden
    class={className}
>
    <div class={cn("ui-panel-header", headerPaddingClasses[padding], headerClass)}>
        <div class="min-w-0 flex-1">
            <slot name="header" />
        </div>
        {#if $$slots.actions}
            <div class="shrink-0">
                <slot name="actions" />
            </div>
        {/if}
    </div>
    <div class={cn(bodyPaddingClasses[padding], bodyClass)}>
        <slot />
    </div>
</Card>

