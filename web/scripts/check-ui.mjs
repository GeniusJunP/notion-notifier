import fs from "node:fs";
import path from "node:path";

const repoRoot = path.resolve(process.cwd(), "..");
const webRoot = path.resolve(process.cwd());

const scanRoots = [
  path.join(webRoot, "src", "routes"),
  path.join(webRoot, "src", "components"),
];

// Some components intentionally use raw HTML controls for specialized UX.
const allowRawControlsIn = new Set([
  path.join(webRoot, "src", "components", "SidebarButton.svelte"),
  path.join(webRoot, "src", "components", "PreviewModal.svelte"),
  path.join(webRoot, "src", "components", "notifications", "DayPicker.svelte"),
]);

const patterns = [
  { tag: "input", re: /<input\b/ },
  { tag: "textarea", re: /<textarea\b/ },
  { tag: "select", re: /<select\b/ },
  { tag: "button", re: /<button\b/ },
];

function listSvelteFiles(dir) {
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  const files = [];
  for (const ent of entries) {
    const full = path.join(dir, ent.name);
    if (ent.isDirectory()) files.push(...listSvelteFiles(full));
    else if (ent.isFile() && ent.name.endsWith(".svelte")) files.push(full);
  }
  return files;
}

function lineNumberForIndex(source, index) {
  // 1-based line number
  return source.slice(0, index).split("\n").length;
}

const violations = [];

for (const root of scanRoots) {
  for (const file of listSvelteFiles(root)) {
    if (allowRawControlsIn.has(file)) continue;
    const source = fs.readFileSync(file, "utf8");
    for (const { tag, re } of patterns) {
      const match = re.exec(source);
      if (!match) continue;
      violations.push({
        file,
        tag,
        line: lineNumberForIndex(source, match.index),
      });
    }
  }
}

if (violations.length > 0) {
  const lines = violations
    .map((v) => {
      const rel = path.relative(repoRoot, v.file);
      return `- ${rel}:${v.line} contains <${v.tag}>`;
    })
    .join("\n");
  console.error(
    [
      "UI policy check failed.",
      "Use primitives in web/src/lib/ui (Button/Input/Textarea/Select/Toggle) instead of raw HTML controls.",
      "Violations:",
      lines,
    ].join("\n"),
  );
  process.exit(1);
}

console.log("UI policy check passed.");
