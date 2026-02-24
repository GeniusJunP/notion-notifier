<script lang="ts">
    export let selectedDays: number[] = [];
    export let ariaLabelPrefix: string = "実行曜日";

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
    {#each dayValues as day, idx}
        <button
            type="button"
            on:click={() => toggleDay(day)}
            class="w-10 h-10 rounded-xl flex items-center justify-center text-sm font-bold transition-all {selectedDays.includes(
                day,
            )
                ? 'bg-brand-600 dark:bg-brand-500 text-white scale-105'
                : 'bg-gray-50 dark:bg-gray-700 text-gray-400 dark:text-gray-500 border border-gray-100 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-600'}"
            aria-label={`${ariaLabelPrefix} ${daysLabels[idx]}`}
            aria-pressed={selectedDays.includes(day)}
        >
            {daysLabels[idx]}
        </button>
    {/each}
</div>
