<script lang="ts">
    import { cn } from "../../lib/utils";

    export let selectedDays: number[] = [];
    export let ariaLabelPrefix = "実行曜日";

    const daysLabels = ["月", "火", "水", "木", "金", "土", "日"];
    const dayValues = [1, 2, 3, 4, 5, 6, 7];

    function toggleDay(day: number) {
        if (selectedDays.includes(day)) {
            selectedDays = selectedDays.filter((d) => d !== day);
        } else {
            selectedDays = [...selectedDays, day].sort();
        }
    }
</script>

<div class="flex flex-wrap gap-2">
    {#each dayValues as day, idx (day)}
        <button
            type="button"
            on:click={() => toggleDay(day)}
            class={cn(
                "flex h-10 w-10 items-center justify-center rounded-xl border text-sm font-semibold outline-none transition-[background-color,border-color,color,box-shadow,transform] duration-200",
                "focus-visible:ring-2 focus-visible:ring-brand-300/70 dark:focus-visible:ring-brand-700/60",
                "active:scale-[0.985]",
                selectedDays.includes(day)
                    ? "border-gray-900 bg-gray-900 text-white shadow-sm dark:border-gray-100 dark:bg-gray-100 dark:text-gray-950"
                    : "border-gray-200/80 bg-white/85 text-gray-500 hover:border-gray-300 hover:bg-white hover:text-gray-900 dark:border-gray-800 dark:bg-gray-900/70 dark:text-gray-400 dark:hover:border-gray-700 dark:hover:bg-gray-900 dark:hover:text-gray-100",
            )}
            aria-label={`${ariaLabelPrefix} ${daysLabels[idx]}`}
            aria-pressed={selectedDays.includes(day)}
        >
            {daysLabels[idx]}
        </button>
    {/each}
</div>
