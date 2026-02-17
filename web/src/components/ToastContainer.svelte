<script lang="ts">
  import { AlertCircle, CheckCircle2, Info, X } from "lucide-svelte";
  import { toastStore } from "../lib/store";
</script>

<div
  class="fixed bottom-6 right-6 z-[100] flex flex-col gap-3 max-w-sm w-full pointer-events-none"
>
  {#each $toastStore as toast (toast.id)}
    <div
      class="pointer-events-auto flex items-start gap-3 p-4 bg-white dark:bg-gray-800 rounded-2xl shadow-2xl border border-gray-100 dark:border-gray-700 animate-in slide-in-from-right fade-in duration-300"
    >
      <div class="mt-0.5">
        {#if toast.type === "error"}
          <AlertCircle size={20} class="text-red-500" />
        {:else if toast.type === "success"}
          <CheckCircle2 size={20} class="text-green-500" />
        {:else}
          <Info size={20} class="text-brand-500" />
        {/if}
      </div>
      <div class="flex-1">
        <p class="text-sm font-medium text-gray-900 dark:text-gray-100 leading-tight">
          {toast.message}
        </p>
      </div>
      <button
        on:click={() =>
          toastStore.update((toasts) =>
            toasts.filter((t) => t.id !== toast.id),
          )}
        class="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
        aria-label="通知を閉じる"
      >
        <X size={16} />
      </button>
    </div>
  {/each}
</div>
