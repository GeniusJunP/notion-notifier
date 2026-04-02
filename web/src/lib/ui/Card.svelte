<script lang="ts">
    import { cn } from "../utils";

    type Tone = "default" | "muted" | "brand" | "danger";
    type Padding = "none" | "sm" | "md" | "lg";
    type Radius = "xl" | "2xl" | "3xl";

    export let tone: Tone = "default";
    export let padding: Padding = "md";
    export let radius: Radius = "2xl";
    export let interactive = false;
    export let overflowHidden = false;

    let className = "";
    export { className as class };

    const toneClasses = {
        default:
            "border border-gray-200/80 bg-white shadow-sm dark:border-gray-800 dark:bg-gray-900",
        muted:
            "border border-gray-200/60 bg-gray-50/80 shadow-sm dark:border-gray-700 dark:bg-gray-800/80",
        brand:
            "border border-brand-100 bg-brand-50/80 shadow-sm dark:border-brand-900/70 dark:bg-brand-950/60",
        danger:
            "border border-red-200/70 bg-red-50/60 shadow-sm dark:border-red-900/70 dark:bg-red-950/40",
    } as const;

    const paddingClasses = {
        none: "",
        sm: "p-4",
        md: "p-6",
        lg: "p-8",
    } as const;

    const radiusClasses = {
        xl: "rounded-xl",
        "2xl": "rounded-2xl",
        "3xl": "rounded-3xl",
    } as const;
</script>

<div
    class={cn(
        toneClasses[tone],
        paddingClasses[padding],
        radiusClasses[radius],
        interactive &&
            "transition-[background-color,border-color,box-shadow,transform] duration-200 hover:border-gray-300 hover:shadow-md dark:hover:border-gray-700",
        overflowHidden && "overflow-hidden",
        className,
    )}
>
    <slot />
</div>
