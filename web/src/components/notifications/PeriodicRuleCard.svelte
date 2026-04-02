<script lang="ts">
    import { Trash2, Play, RotateCcw } from "lucide-svelte";
    import DayPicker from "./DayPicker.svelte";
    import type { PeriodicNotification } from "../../lib/api";
    import { createEventDispatcher } from "svelte";
    import Button from "../../lib/ui/Button.svelte";
    import FormField from "../../lib/ui/FormField.svelte";
    import Input from "../../lib/ui/Input.svelte";
    import Panel from "../../lib/ui/Panel.svelte";
    import Textarea from "../../lib/ui/Textarea.svelte";
    import Toggle from "../../lib/ui/Toggle.svelte";

    export let rule: PeriodicNotification;
    export let index: number;

    const dispatch = createEventDispatcher<{
        remove: number;
        preview: { template: string; title: string; days_ahead: number };
        reset: number;
    }>();
</script>

<Panel radius="2xl" interactive bodyClass="grid grid-cols-1 gap-8">
    <svelte:fragment slot="header">
        <div class="flex items-center gap-3">
            <div
                class="flex h-8 w-8 items-center justify-center rounded-lg border border-gray-200 bg-white text-sm font-semibold text-gray-500 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-400"
            >
                {index + 1}
            </div>
            <Toggle
                bind:checked={rule.enabled}
                ariaLabel={`定期通知 ${index + 1} の有効化`}
                tone="success"
                size="sm"
            />
            <span class="font-semibold text-gray-900 dark:text-gray-100">
                定期通知 {index + 1}
            </span>
        </div>
    </svelte:fragment>
    <svelte:fragment slot="actions">
        <Button
            type="button"
            on:click={() => dispatch("remove", index)}
            variant="ghost"
            size="icon"
            aria-label={`定期通知 ${index + 1} を削除`}
        >
            <Trash2 size={18} />
        </Button>
    </svelte:fragment>

    <div class="space-y-6">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <FormField label="通知時刻" forId={`per-time-${index}`}>
                <Input id={`per-time-${index}`} type="time" bind:value={rule.time} />
            </FormField>

            <FormField label="参照範囲 (何日分)" forId={`per-days-ahead-${index}`}>
                <Input
                    id={`per-days-ahead-${index}`}
                    type="number"
                    bind:value={rule.days_ahead}
                />
            </FormField>
        </div>

        <div>
            <span
                class="mb-2 block text-sm font-semibold tracking-tight text-gray-700 dark:text-gray-300"
            >
                実行する曜日
            </span>
            <DayPicker
                bind:selectedDays={rule.days_of_week}
                ariaLabelPrefix={`定期通知 ${index + 1} の実行曜日`}
            />
        </div>
    </div>

    <FormField label="メッセージテンプレート" forId={`per-message-${index}`}>
        <Textarea
            id={`per-message-${index}`}
            bind:value={rule.message}
            placeholder="空欄の場合はデフォルトが使用されます"
            class="min-h-[150px]"
            mono
        />
        <div class="mt-3 flex items-center gap-3">
            <Button
                type="button"
                on:click={() =>
                    dispatch("preview", {
                        template: rule.message,
                        title: "定期通知プレビュー",
                        days_ahead: rule.days_ahead,
                    })}
                variant="text"
                size="sm"
            >
                <Play size={12} /> プレビューを実行
            </Button>
            <Button
                type="button"
                on:click={() => dispatch("reset", index)}
                variant="ghost"
                size="sm"
            >
                <RotateCcw size={12} /> デフォルトに戻す
            </Button>
        </div>
    </FormField>
</Panel>
