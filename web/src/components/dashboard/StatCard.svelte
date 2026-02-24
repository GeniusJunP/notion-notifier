<script lang="ts">
    export let label: string;
    export let value: string | number;
    export let subLabel: string;
    export let iconColorClass: string;
</script>

<div
    class="bg-white dark:bg-gray-800 p-6 rounded-2xl shadow-sm dark:shadow-md border border-gray-100 dark:border-gray-700 flex flex-col justify-between"
>
    <div class="flex items-center justify-between mb-4">
        <span class="text-sm font-medium text-gray-500 dark:text-gray-400">
            {label}
        </span>
        <div
            class="w-10 h-10 rounded-xl flex items-center justify-center {iconColorClass}"
        >
            <slot name="icon"></slot>
        </div>
    </div>
    <div>
        <h3 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
            {value}
        </h3>
        <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">
            {subLabel}
        </p>
    </div>
</div>
