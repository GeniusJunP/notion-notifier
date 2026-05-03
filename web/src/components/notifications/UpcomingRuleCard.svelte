<script lang="ts">
    import { Trash2, Clock } from "lucide-svelte";
    import DayPicker from "./DayPicker.svelte";
    import type { UpcomingNotification } from "../../lib/api";
    import { createEventDispatcher } from "svelte";
    import Button from "../../lib/ui/Button.svelte";
    import FormField from "../../lib/ui/FormField.svelte";
    import Input from "../../lib/ui/Input.svelte";
    import Panel from "../../lib/ui/Panel.svelte";
    import Toggle from "../../lib/ui/Toggle.svelte";
    import TemplateEditor from "../TemplateEditor.svelte";

    export let rule: UpcomingNotification;
    export let index: number;

    const dispatch = createEventDispatcher<{
        remove: number;
        preview: { template: string; title: string; minutes_before: number };
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
                ariaLabel={`事前通知 ${index + 1} の有効化`}
                tone="success"
                size="sm"
            />
            <span class="font-semibold text-gray-900 dark:text-gray-100">
                事前通知 {index + 1}
            </span>
        </div>
    </svelte:fragment>
    <svelte:fragment slot="actions">
        <Button
            type="button"
            on:click={() => dispatch("remove", index)}
            variant="ghost"
            size="icon"
            aria-label={`事前通知 ${index + 1} を削除`}
        >
            <Trash2 size={18} />
        </Button>
    </svelte:fragment>

    <div class="space-y-4">
        <FormField label="通知タイミング (分前)" forId={`adv-minutes-${index}`}>
            <div class="relative">
                <Clock
                    class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
                    size={16}
                />
                <Input
                    id={`adv-minutes-${index}`}
                    type="number"
                    bind:value={rule.minutes_before}
                    class="pl-10"
                />
            </div>
        </FormField>

        <FormField
            label="終日予定の基準時刻"
            forId={`adv-allday-base-${index}`}
            hint="終日予定の通知基準として使用（デフォルト: 09:00）"
        >
            <div class="relative">
                <Clock
                    class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
                    size={16}
                />
                <Input
                    id={`adv-allday-base-${index}`}
                    type="time"
                    bind:value={rule.allday_base_time}
                    class="pl-10"
                />
            </div>
        </FormField>

        <FormField forId={`adv-message-${index}`}>
            <TemplateEditor
                id={`adv-message-${index}`}
                label="メッセージテンプレート"
                bind:value={rule.message}
                placeholder="空欄の場合はデフォルトが使用されます"
                rows={5}
                on:preview={() =>
                    dispatch("preview", {
                        template: rule.message,
                        title: "事前通知プレビュー",
                        minutes_before: rule.minutes_before,
                    })}
                on:reset={() => dispatch("reset", index)}
            />
        </FormField>
    </div>

    <div class="space-y-3">
        <span
            class="block text-xs font-semibold uppercase tracking-[0.16em] text-gray-500 dark:text-gray-400"
        >
            実行する曜日
        </span>
        <DayPicker
            bind:selectedDays={rule.conditions.days_of_week}
            ariaLabelPrefix={`事前通知 ${index + 1} の実行曜日`}
        />
    </div>
</Panel>
