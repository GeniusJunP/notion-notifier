<script lang="ts">
    import { cn } from "../utils";

    type Tone = "brand" | "neutral" | "success" | "warning" | "danger";
    type Size = "sm" | "md" | "lg";

    export let tone: Tone = "neutral";
    export let size: Size = "md";

    let className = "";
    export { className as class };

    const sizeClasses = {
        sm: "h-8 w-8 rounded-lg",
        md: "h-10 w-10 rounded-xl",
        lg: "h-16 w-16 rounded-2xl",
    } as const;

    const toneClasses: Record<Tone, string> = {
        brand: "border-brand-200 bg-brand-50 text-brand-700 dark:border-brand-900/70 dark:bg-brand-950/60 dark:text-brand-200",
        neutral:
            "border-gray-200 bg-gray-50 text-gray-600 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300",
        success:
            "border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900/70 dark:bg-emerald-950/60 dark:text-emerald-200",
        warning:
            "border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900/70 dark:bg-amber-950/60 dark:text-amber-200",
        danger:
            "border-red-200 bg-red-50 text-red-700 dark:border-red-900/70 dark:bg-red-950/60 dark:text-red-200",
    };
</script>

<div
    class={cn(
        "flex items-center justify-center border",
        sizeClasses[size],
        toneClasses[tone],
        className,
    )}
>
    <slot />
</div>
