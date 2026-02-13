<script lang="ts">
  export let active = false;
  export let justifyBetween = false;
  export let disabled = false;
  export let ariaCurrent: "page" | undefined = undefined;

  $: variantClass =
    active
      ? "bg-brand-50 dark:bg-brand-900/20 text-brand-700 dark:text-brand-300 font-semibold"
      : "text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100";

  $: layoutClass = justifyBetween
    ? "items-center justify-between"
    : "items-center gap-3";
</script>

<button
  on:click
  {disabled}
  aria-current={ariaCurrent}
  class={`w-full flex ${layoutClass} px-3 py-2.5 rounded-xl transition-all duration-200 group disabled:opacity-50 ${variantClass}`}
>
  <slot />
</button>
