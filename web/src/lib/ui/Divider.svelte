<script lang="ts">
    import type { HTMLAttributes } from "svelte/elements";

    import { cn } from "../utils";

    interface Props extends HTMLAttributes<HTMLDivElement> {
        spacing?: "none" | "sm" | "md" | "lg";
    }

    let { class: className = "", spacing = "md", ...rest }: Props = $props();

    const spacingClasses = {
        none: "m-0",
        sm: "mt-2 pt-2",
        md: "mt-4 pt-4",
        lg: "mt-6 pt-6"
    };
</script>

<div class={cn("border-t border-gray-200/70 dark:border-gray-800", spacingClasses[spacing], className)} {...rest}></div>
