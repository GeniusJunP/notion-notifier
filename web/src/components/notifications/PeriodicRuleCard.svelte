<script lang="ts">
    import DayPicker from "./DayPicker.svelte";
    import type { PeriodicNotification } from "../../lib/api";
    import { createEventDispatcher } from "svelte";
    import BaseRuleCard from "./BaseRuleCard.svelte";
    import FormField from "../../lib/ui/FormField.svelte";
    import Input from "../../lib/ui/Input.svelte";
    import TemplateEditor from "../TemplateEditor.svelte";

    export let rule: PeriodicNotification;
    export let index: number;

    const dispatch = createEventDispatcher<{
        remove: number;
        preview: { template: string; title: string; days_ahead: number };
        reset: number;
    }>();
</script>

<BaseRuleCard
    title={`定期通知 ${index + 1}`}
    {index}
    bind:enabled={rule.enabled}
    onremove={() => dispatch("remove", index)}
>
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

    <FormField forId={`per-message-${index}`}>
        <TemplateEditor
            id={`per-message-${index}`}
            label="メッセージテンプレート"
            bind:value={rule.message}
            placeholder="空欄の場合はデフォルトが使用されます"
            rows={6}
            onpreview={() =>
                dispatch("preview", {
                    template: rule.message,
                    title: "定期通知プレビュー",
                    days_ahead: rule.days_ahead,
                })}
            onreset={() => dispatch("reset", index)}
        />
    </FormField>
</BaseRuleCard>
