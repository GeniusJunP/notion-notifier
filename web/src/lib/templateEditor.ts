export type TemplateEditorMode = "visual" | "raw";
export type TemplateBlockKind = "text" | "variable" | "if" | "with" | "range";

export interface TemplateBlock {
    id: string;
    kind: TemplateBlockKind;
    content: string;
    children?: TemplateBlock[];
    elseChildren?: TemplateBlock[];
}

export const templateVariables = [
    ".Name",
    ".Date",
    ".Time",
    ".EndDate",
    ".EndTime",
    ".IsAllDay",
    ".Location",
    ".URL",
    ".Content",
    ".MinutesBefore",
    ".Custom",
    ".Events",
    ".Type",
    ".Message",
    ".Event",
] as const;

export function createTemplateBlock(kind: TemplateBlockKind): TemplateBlock {
    return {
        id: crypto.randomUUID(),
        kind,
        content:
            kind === "text"
                ? ""
                : kind === "variable"
                  ? ".Name"
                  : kind === "range"
                    ? ".Events"
                    : ".Location",
        children:
            kind === "if" || kind === "with" || kind === "range"
                ? [createTemplateBlock("text")]
                : undefined,
        elseChildren: kind === "if" ? [] : undefined,
    };
}

export function blocksFromTemplate(value: string): TemplateBlock[] {
    return [
        {
            id: crypto.randomUUID(),
            kind: "text",
            content: value,
        },
    ];
}

export function serializeTemplateBlocks(blocks: TemplateBlock[]): string {
    return blocks.map(serializeBlock).join("");
}

function serializeBlock(block: TemplateBlock): string {
    switch (block.kind) {
        case "text":
            return block.content;
        case "variable":
            return `{{${block.content}}}`;
        case "if":
            return serializeControlBlock("if", block);
        case "with":
            return serializeControlBlock("with", block);
        case "range":
            return serializeControlBlock("range", block);
    }
}

function serializeControlBlock(action: "if" | "with" | "range", block: TemplateBlock): string {
    const body = serializeTemplateBlocks(block.children ?? []);
    const elseBody = serializeTemplateBlocks(block.elseChildren ?? []);
    if (action === "if" && elseBody) {
        return `{{- ${action} ${block.content} -}}${body}{{- else -}}${elseBody}{{- end -}}`;
    }
    return `{{- ${action} ${block.content} -}}${body}{{- end -}}`;
}
