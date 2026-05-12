<script lang="ts">
    import { Globe } from "lucide-svelte";

    import type { Config } from "../lib/api";
    import Card from "../lib/ui/Card.svelte";
    import FormField from "../lib/ui/FormField.svelte";
    import Input from "../lib/ui/Input.svelte";
    import SectionCard from "../lib/ui/SectionCard.svelte";
    import Toggle from "../lib/ui/Toggle.svelte";
    import Typography from "../lib/ui/Typography.svelte";
    import TemplateEditor from "./TemplateEditor.svelte";

    export let config: Config;
</script>

<SectionCard>
    <h4 class="flex items-center gap-2 text-lg font-bold text-gray-900 dark:text-gray-100">
        <Globe size={18} class="text-brand-600 dark:text-brand-300" />
        Webhook 設定
    </h4>

    <p class="text-xs leading-relaxed text-gray-600 dark:text-gray-300">
        Webhook で送信される JSON ペイロードのテンプレートです。
        <code class="rounded bg-white/70 px-1 py-0.5 font-mono dark:bg-gray-900/60">{"{{.Message}}"}</code>
        変数が通知内容に置き換わります。
    </p>

    <Card tone="muted" radius="2xl" padding="sm" class="flex items-center justify-between gap-3">
        <div class="space-y-1">
            <p class="text-sm font-semibold text-gray-700 dark:text-gray-200">
                テストモード
            </p>
            <Typography variant="meta" block>
                ON にすると、内部通知用のテンプレートと URL で送信します
            </Typography>
        </div>
        <Toggle
            bind:checked={config.webhook.is_test}
            ariaLabel="テストモードを切り替え"
            size="sm"
        />
    </Card>

    <Card tone="muted" radius="2xl" padding="sm" class="space-y-3">
        <Typography variant="label-caps" as="h5">
            通知 Webhook
        </Typography>

        <FormField
            label="Content-Type"
            forId="wh-notification-ct"
            variant="eyebrow"
        >
            <Input
                id="wh-notification-ct"
                type="text"
                bind:value={config.webhook.notification.content_type}
                uiSize="sm"
            />
        </FormField>

        <FormField
            label="ペイロードテンプレート"
            forId="wh-notification-pt"
            variant="eyebrow"
        >
            <TemplateEditor
                id="wh-notification-pt"
                label="ペイロードテンプレート"
                bind:value={config.webhook.notification.payload_template}
                rows={4}
                showPreview={false}
                showReset={false}
            />
        </FormField>
    </Card>

    <Card tone="muted" radius="2xl" padding="sm" class="space-y-3">
        <Typography variant="label-caps" as="h5">
            内部通知 Webhook
        </Typography>

        <FormField
            label="Content-Type"
            forId="wh-internal-notif-ct"
            variant="eyebrow"
        >
            <Input
                id="wh-internal-notif-ct"
                type="text"
                bind:value={config.webhook.internal_notification.content_type}
                uiSize="sm"
            />
        </FormField>

        <FormField
            label="ペイロードテンプレート"
            forId="wh-internal-notif-pt"
            variant="eyebrow"
        >
            <TemplateEditor
                id="wh-internal-notif-pt"
                label="ペイロードテンプレート"
                bind:value={config.webhook.internal_notification.payload_template}
                rows={4}
                showPreview={false}
                showReset={false}
            />
        </FormField>
    </Card>
</SectionCard>
