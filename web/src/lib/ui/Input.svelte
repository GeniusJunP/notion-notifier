<script lang="ts">
    import type { HTMLInputAttributes } from "svelte/elements";
    import { cn } from "../utils";
    import { fieldRecipe } from "./recipes";

    type UiSize = "sm" | "md";

    export let value: HTMLInputAttributes["value"] = "";
    export let type: HTMLInputAttributes["type"] = "text";
    export let id: HTMLInputAttributes["id"] = undefined;
    export let name: HTMLInputAttributes["name"] = undefined;
    export let placeholder: HTMLInputAttributes["placeholder"] = undefined;
    export let disabled = false;
    export let uiSize: UiSize = "md";
    export let mono = false;
    export let min: HTMLInputAttributes["min"] = undefined;
    export let max: HTMLInputAttributes["max"] = undefined;
    export let step: HTMLInputAttributes["step"] = undefined;

    let className = "";
    export { className as class };

    const sizeClasses = {
        sm: "min-h-10 px-3 py-2 text-sm",
        md: "min-h-11 px-4 py-2.5 text-sm",
    } as const;

    function handleInput(event: Event) {
        const target = event.currentTarget as HTMLInputElement;
        if (type === "number") {
            value = Number.isNaN(target.valueAsNumber)
                ? ""
                : target.valueAsNumber;
            return;
        }
        value = target.value;
    }

    function handleChange(event: Event) {
        if (type !== "number") {
            return;
        }
        const target = event.currentTarget as HTMLInputElement;
        value = Number.isNaN(target.valueAsNumber)
            ? ""
            : target.valueAsNumber;
    }

    $: classes = cn(
        fieldRecipe({ size: uiSize, mono }),
        "dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder:text-gray-500 dark:focus:border-brand-500 dark:focus:ring-brand-900/50",
        className,
    );
</script>

<input
    {id}
    {name}
    {type}
    {placeholder}
    {disabled}
    {min}
    {max}
    {step}
    {value}
    class={classes}
    on:blur
    on:change={handleChange}
    on:focus
    on:input={handleInput}
/>
