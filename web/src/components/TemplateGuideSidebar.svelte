<script lang="ts">
    import { createEventDispatcher } from "svelte";

    const templateVariables = [
        "{{.Name}}",
        "{{.Date}}",
        "{{.Time}}",
        "{{.EndDate}}",
        "{{.EndTime}}",
        "{{.IsAllDay}}",
        "{{.Location}}",
        "{{.URL}}",
        "{{.Content}}",
        "{{.MinutesBefore}}",
        "{{.Custom}}",
        "{{.Events}}",
        "{{.Type}}",
        "{{.Message}}",
        "{{.Event}}",
    ];

    const guideMarkdown = `
### まず覚える基本
- 値の埋め込み: \`{{.Name}}\`
- 条件分岐: \`{{if 条件}} ... {{end}}\`
- 繰り返し: \`{{range .Events}} ... {{end}}\`
- 関数: \`{{json .Message}}\`

### \`end\` とは？
\`end\` は「ブロックの終わり」です。  
\`if\` / \`range\` / \`with\` を使ったら、必ず対応する \`end\` が必要です。

### if / else / end
\`\`\`gotemplate
{{if .IsAllDay}}
終日
{{else}}
{{.Date}} {{.Time}} - {{.EndDate}} {{.EndTime}}
{{end}}

{{if .Location}}
📍 {{.Location}}
{{end}}
\`\`\`

### with / end（値があるときだけ）
\`\`\`gotemplate
{{with .Content}}
---
{{.}}
{{end}}
\`\`\`

### range / end（複数イベント）
\`\`\`gotemplate
{{range .Events}}
- **{{.Name}}** ({{.Date}} {{if .IsAllDay}}終日{{else}}{{.Time}}〜{{.EndDate}} {{.EndTime}}{{end}})
  {{if .Location}}📍 {{.Location}}{{end}}
{{end}}
\`\`\`

### ネスト時の見方（\`end\` が複数になる）
\`\`\`gotemplate
{{range .Events}}
  {{if .Location}}
  - {{.Name}} @ {{.Location}}
  {{end}}
{{end}}
\`\`\`

### プロパティマッピング（設定参照）を増やしたときの使い方

テンプレート側は \`.Custom\` から参照します。  
キー名にハイフンや日本語を使う可能性があるので、\`index\` 形式がおすすめです。

\`\`\`gotemplate
{{with index .Custom "speaker"}}登壇者: {{.}}{{end}}
{{with index .Custom "category"}}カテゴリ: {{.}}{{end}}
\`\`\`

### Webhook payload_template
\`\`\`gotemplate
{"content": {{json .Message}}}
\`\`\`

補足:
- \`{{.Content}}\` は抽出設定や同期結果によって空になる場合があります。
- custom マッピングを追加した後、値が見えない場合は Notion 同期を実行して最新データを反映してください。
`.trim();

    const dispatch = createEventDispatcher<{
        openGuide: { title: string; content: string };
    }>();
    const detailTitle = "Goテンプレートの使い方";
    const detailContent = guideMarkdown;

    function openGuideDetail() {
        dispatch("openGuide", {
            title: detailTitle,
            content: detailContent,
        });
    }
</script>

<div class="rounded-2xl border border-gray-100 bg-white px-3 py-3 dark:border-gray-700 dark:bg-gray-800">
    <div class="mb-2 flex items-center justify-between gap-2">
        <h3 class="text-xs font-bold text-gray-800 dark:text-gray-100">使える変数</h3>
    </div>

    <div class="flex flex-wrap gap-1.5">
        {#each templateVariables as variable}
            <code class="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-[10px] text-gray-700 dark:bg-gray-700 dark:text-gray-100 break-all">
                {variable}
            </code>
        {/each}

        <button
            on:click={openGuideDetail}
            class="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-[10px] transition-colors hover:bg-brand-100 dark:border-brand-900/50 dark:bg-brand-400/20 dark:text-brand-300 dark:hover:bg-brand-500"
        >
            詳細
        </button>
    </div>
</div>
