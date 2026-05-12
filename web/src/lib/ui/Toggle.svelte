<script lang="ts">
    import { tv } from "tailwind-variants";

    import { cn } from "../utils";

    const toggleRecipe = tv({
        slots: {
            base: "relative inline-flex shrink-0 items-center rounded-full border border-transparent p-1 shadow-inner outline-none transition-[background-color,box-shadow,transform] duration-200 focus-visible:ring-2 focus-visible:ring-brand-300/70 disabled:cursor-not-allowed disabled:opacity-50 dark:focus-visible:ring-brand-700/60",
            thumb: "rounded-full bg-white shadow-sm transition-transform duration-200",
        },
        variants: {
            size: {
                sm: { base: "h-6 w-10", thumb: "size-4" },
                md: { base: "h-8 w-14", thumb: "size-6" },
            },
            tone: {
                brand: "",
                success: "",
                warning: "",
            },
            checked: {
                true: "",
                false: { base: "bg-gray-200 dark:bg-gray-700" },
            },
        },
        compoundVariants: [
            { checked: true, tone: "brand", class: { base: "bg-brand-600 dark:bg-brand-500" } },
            { checked: true, tone: "success", class: { base: "bg-emerald-600 dark:bg-emerald-500" } },
            { checked: true, tone: "warning", class: { base: "bg-amber-500 dark:bg-amber-500" } },
            { checked: true, size: "sm", class: { thumb: "translate-x-4" } },
            { checked: false, size: "sm", class: { thumb: "translate-x-0" } },
            { checked: true, size: "md", class: { thumb: "translate-x-6" } },
            { checked: false, size: "md", class: { thumb: "translate-x-0" } },
        ],
        defaultVariants: {
            size: "md",
            tone: "brand",
            checked: false,
        },
    });

    type Tone = "brand" | "success" | "warning";
    type Size = "sm" | "md";

    interface ToggleProps {
        checked?: boolean;
        disabled?: boolean;
        ariaLabel?: string;
        tone?: Tone;
        size?: Size;
        onchange?: (value: boolean) => void;
        class?: string;
    }

    let {
        checked = $bindable(false),
        disabled = false,
        ariaLabel = "",
        tone = "brand",
        size = "md",
        onchange,
        class: className = "",
    }: ToggleProps = $props();

    function toggle() {
        if (disabled) return;
        checked = !checked;
        onchange?.(checked);
    }

    let styles = $derived(toggleRecipe({ size, tone, checked }));
</script>

<button
    type="button"
    role="switch"
    aria-checked={checked}
    aria-label={ariaLabel}
    {disabled}
    class={cn(styles.base(), className)}
    onclick={toggle}
>
    <span class={styles.thumb()}></span>
</button>
