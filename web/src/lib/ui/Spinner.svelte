<script lang="ts">
    import { cn } from "../utils";

    export let size: "sm" | "md" = "sm";
    export let tone: "current" | "muted" | "inverse" = "current";

    let className = "";
    export { className as class };

    const sizeClasses = {
        sm: "h-4 w-4 border-2",
        md: "h-5 w-5 border-2",
    } as const;

    const toneClasses = {
        current: "border-current/20 border-t-current",
        muted: "border-gray-400/30 border-t-gray-600 dark:border-gray-500/30 dark:border-t-gray-200",
        inverse: "border-white/25 border-t-white",
    } as const;
</script>

<div
    aria-hidden="true"
    class={cn(
        "animate-spin rounded-full",
        sizeClasses[size],
        toneClasses[tone],
        className,
    )}
></div>
