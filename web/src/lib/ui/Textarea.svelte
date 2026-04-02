<script lang="ts">
    import type { HTMLTextareaAttributes } from "svelte/elements";
    import { cn } from "../utils";

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
        "w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-900 shadow-sm outline-none transition-[background-color,border-color,box-shadow,color] duration-200",
        "placeholder:text-gray-400 focus:border-brand-300 focus:ring-2 focus:ring-brand-200",
        "dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder:text-gray-500 dark:focus:border-brand-500 dark:focus:ring-brand-900/50",
        "disabled:cursor-not-allowed disabled:opacity-60",
        mono && "font-mono",
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
