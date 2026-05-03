<script lang="ts">
    import type { HTMLTextareaAttributes } from "svelte/elements";
    import { cn } from "../utils";
    import { fieldRecipe } from "./recipes";

    export let value: HTMLTextareaAttributes["value"] = "";
    export let id: HTMLTextareaAttributes["id"] = undefined;
    export let name: HTMLTextareaAttributes["name"] = undefined;
    export let placeholder: HTMLTextareaAttributes["placeholder"] = undefined;
    export let rows = 4;
    export let disabled = false;
    export let mono = false;

    let className = "";
    export { className as class };

    $: classes = cn(
        fieldRecipe({ size: "area", mono }),
        "dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder:text-gray-500 dark:focus:border-brand-500 dark:focus:ring-brand-900/50",
        className,
    );

    function handleInput(event: Event) {
        value = (event.currentTarget as HTMLTextAreaElement).value;
    }
</script>

<textarea
    {id}
    {name}
    {placeholder}
    {rows}
    {disabled}
    {value}
    class={classes}
    on:blur
    on:change
    on:focus
    on:input={handleInput}
></textarea>
