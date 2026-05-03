<script lang="ts">
    import { Send, RotateCcw } from "lucide-svelte";
    import { createEventDispatcher } from "svelte";
    import Button from "../../lib/ui/Button.svelte";
    import FormField from "../../lib/ui/FormField.svelte";
    import IconChip from "../../lib/ui/IconChip.svelte";
    import Input from "../../lib/ui/Input.svelte";
    import Panel from "../../lib/ui/Panel.svelte";
    import TemplateEditor from "../TemplateEditor.svelte";

    export let manualFromDate: string;
    export let manualToDate: string;
    export let manualTemplate: string;
    export let isPreviewLoading: boolean;
    export let isSending: boolean;

    const dispatch = createEventDispatcher<{
        loadDefault: void;
        preview: void;
        send: void;
    }>();
</script>

<Panel radius="3xl" bodyClass="space-y-5">
    <svelte:fragment slot="header">
        <div class="flex items-center gap-3">
            <IconChip tone="brand" size="md">
                <Send size={20} />
            </IconChip>
            <div>
                <h2 class="text-lg font-bold text-gray-900 dark:text-gray-100">
                    手動通知
                </h2>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                    テンプレートを使用して即座にWebhook通知を送信
                </p>
            </div>
        </div>
    </svelte:fragment>
    <svelte:fragment slot="actions">
        <Button on:click={() => dispatch("loadDefault")} variant="ghost" size="sm">
            <RotateCcw size={12} />
            デフォルトに戻す
        </Button>
    </svelte:fragment>

    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <FormField label="開始日" forId="manual-from-date" variant="eyebrow">
            <Input
                id="manual-from-date"
                type="date"
                bind:value={manualFromDate}
                uiSize="sm"
            />
        </FormField>

        <FormField label="終了日" forId="manual-to-date" variant="eyebrow">
            <Input
                id="manual-to-date"
                type="date"
                bind:value={manualToDate}
                uiSize="sm"
            />
        </FormField>
    </div>

    <FormField forId="manual-template">
        <TemplateEditor
            id="manual-template"
            label="メッセージテンプレート"
            bind:value={manualTemplate}
            placeholder="Go テンプレート形式で入力..."
            rows={5}
            previewLoading={isPreviewLoading}
            on:preview={() => dispatch("preview")}
            on:reset={() => dispatch("loadDefault")}
        />
    </FormField>

    <div class="flex items-center gap-3">
        <Button
            on:click={() => dispatch("send")}
            disabled={isSending}
            loading={isSending}
            block
            size="md"
        >
            {#if !isSending}
                <Send size={16} />
            {/if}
            通知を送信
        </Button>
    </div>
</Panel>
