<script lang="ts">
    import { cn } from "../utils";

    type Variant = "neutral" | "success" | "error" | "warning" | "info";

    export let variant: Variant = "neutral";
    export let caps = true;

    let className = "";
    export { className as class };

    const variantClasses = {
        neutral:
            "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300",
        success:
            "bg-emerald-100 text-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-300",
        error:
            "bg-red-100 text-red-700 dark:bg-red-950/60 dark:text-red-300",
        warning:
            "bg-amber-100 text-amber-700 dark:bg-amber-950/60 dark:text-amber-300",
        info: "bg-brand-100 text-brand-700 dark:bg-brand-950/70 dark:text-brand-300",
    } as const;
</script>

<span
    class={cn(
        "inline-flex items-center rounded-full px-2.5 py-1 font-semibold",
        caps ? "text-[10px] uppercase tracking-[0.16em]" : "text-xs tracking-tight",
        variantClasses[variant],
        className,
    )}
>
    <slot />
</span>
