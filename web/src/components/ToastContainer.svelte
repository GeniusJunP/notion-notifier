<script lang="ts">
    import { AlertCircle, CheckCircle2, Info, X } from "lucide-svelte";
    import { toastStore } from "../lib/store";
    import Button from "../lib/ui/Button.svelte";
    import Card from "../lib/ui/Card.svelte";

    const iconClassByType = {
        error: "text-red-500",
        success: "text-emerald-500",
        info: "text-brand-500",
    } as const;
</script>

<div
    class="pointer-events-none fixed bottom-6 right-6 z-[100] flex w-full max-w-sm flex-col gap-3"
>
    {#each $toastStore as toast (toast.id)}
        <Card
            class="pointer-events-auto animate-toast-in"
            radius="2xl"
            padding="sm"
        >
            <div class="flex items-start gap-3">
                <div class="mt-0.5">
                    {#if toast.type === "error"}
                        <AlertCircle size={20} class={iconClassByType.error} />
                    {:else if toast.type === "success"}
                        <CheckCircle2 size={20} class={iconClassByType.success} />
                    {:else}
                        <Info size={20} class={iconClassByType.info} />
                    {/if}
                </div>
                <div class="flex-1">
                    <p class="text-sm font-medium leading-tight text-gray-900 dark:text-gray-100">
                        {toast.message}
                    </p>
                </div>
                <Button
                    on:click={() =>
                        toastStore.update((toasts) =>
                            toasts.filter((t) => t.id !== toast.id),
                        )}
                    variant="ghost"
                    size="icon"
                    aria-label="通知を閉じる"
                >
                    <X size={16} />
                </Button>
            </div>
        </Card>
    {/each}
</div>
