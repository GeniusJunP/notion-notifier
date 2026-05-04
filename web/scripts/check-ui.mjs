import fs from "node:fs";
import path from "node:path";

const repoRoot = path.resolve(process.cwd(), "..");
const webRoot = path.resolve(process.cwd());

const scanRoots = [
  path.join(webRoot, "src", "App.svelte"),
  path.join(webRoot, "src", "routes"),
  path.join(webRoot, "src", "components"),
];

// Some components intentionally use raw HTML controls for specialized UX.
const allowRawControlsIn = new Set([
  path.join(webRoot, "src", "App.svelte"),
  path.join(webRoot, "src", "components", "SidebarButton.svelte"),
  path.join(webRoot, "src", "components", "PreviewModal.svelte"),
  path.join(webRoot, "src", "components", "notifications", "DayPicker.svelte"),
]);

const classBudgets = new Map(
  Object.entries({
    "src/App.svelte": 6,
    "src/components/PreviewModal.svelte": 8,
    "src/components/SidebarButton.svelte": 1,
    "src/components/TemplateEditor.svelte": 12,
    "src/components/TemplateGuideSidebar.svelte": 5,
    "src/components/ToastContainer.svelte": 9,
    "src/components/WebhookSettingsCard.svelte": 12,
    "src/components/dashboard/ManualNotificationCard.svelte": 5,
    "src/components/dashboard/ManualSyncCard.svelte": 7,
    "src/components/dashboard/StatCard.svelte": 6,
    "src/components/dashboard/UpcomingEventsList.svelte": 22,
    "src/components/layout/Header.svelte": 7,
    "src/components/layout/Sidebar.svelte": 15,
    "src/components/notifications/DayPicker.svelte": 2,
    "src/components/notifications/PeriodicRuleCard.svelte": 6,
    "src/components/notifications/UpcomingRuleCard.svelte": 12,
    "src/components/settings/ContentRuleSettings.svelte": 11,
    "src/components/settings/GeneralSettings.svelte": 3,
    "src/components/settings/PropertyMappingSettings.svelte": 18,
    "src/routes/Calendar.svelte": 24,
    "src/routes/Dashboard.svelte": 2,
    "src/routes/History.svelte": 40,
    "src/routes/Notifications.svelte": 9,
    "src/routes/Settings.svelte": 9,
  }),
);

const patterns = [
  { tag: "input", re: /<input\b/ },
  { tag: "textarea", re: /<textarea\b/ },
  { tag: "select", re: /<select\b/ },
  { tag: "button", re: /<button\b/ },
];

function listFiles(dir, extensions) {
  if (fs.statSync(dir).isFile()) {
    return extensions.some((ext) => dir.endsWith(ext)) ? [dir] : [];
  }
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  const files = [];
  for (const ent of entries) {
    const full = path.join(dir, ent.name);
    if (ent.isDirectory()) files.push(...listFiles(full, extensions));
    else if (ent.isFile() && extensions.some((ext) => ent.name.endsWith(ext))) {
      files.push(full);
    }
  }
  return files;
}

function listSvelteFiles(dir) {
  return listFiles(dir, [".svelte"]);
}

function lineNumberForIndex(source, index) {
  // 1-based line number
  return source.slice(0, index).split("\n").length;
}

const violations = [];
const classViolations = [];
const pandaViolations = [];
const pandaApiPattern =
  /from\s+["'][^"']*styled-system\/(?:css|patterns)["']|\b(?:css|cva|sva)\s*\(/;

for (const root of [path.join(webRoot, "src")]) {
  for (const file of listSvelteFiles(root)) {
    const source = fs.readFileSync(file, "utf8");
    const match = pandaApiPattern.exec(source);
    if (match) {
      pandaViolations.push({
        file,
        line: lineNumberForIndex(source, match.index),
      });
    }
  }
}

for (const file of listFiles(path.join(webRoot, "src"), [".ts"])) {
  if (file.endsWith(".d.ts")) continue;
  const source = fs.readFileSync(file, "utf8");
  const match = pandaApiPattern.exec(source);
  if (match) {
    pandaViolations.push({
      file,
      line: lineNumberForIndex(source, match.index),
    });
  }
}

for (const root of scanRoots) {
  for (const file of listSvelteFiles(root)) {
    if (allowRawControlsIn.has(file)) continue;
    const source = fs.readFileSync(file, "utf8");
    const rel = path.relative(webRoot, file);
    if (!rel.startsWith("src/lib/ui/")) {
      const classCount = [...source.matchAll(/\bclass(?:\s*=|=)/g)].length;
      const budget = classBudgets.get(rel) ?? 0;
      if (classCount > budget) {
        classViolations.push({ file, count: classCount, budget });
      }
    }
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

if (
  violations.length > 0 ||
  classViolations.length > 0 ||
  pandaViolations.length > 0
) {
  const lines = violations
    .map((v) => {
      const rel = path.relative(repoRoot, v.file);
      return `- ${rel}:${v.line} contains <${v.tag}>`;
    })
    .join("\n");
  const classLines = classViolations
    .map((v) => {
      const rel = path.relative(repoRoot, v.file);
      return `- ${rel} has ${v.count} class bindings; budget is ${v.budget}`;
    })
    .join("\n");
  const pandaLines = pandaViolations
    .map((v) => {
      const rel = path.relative(repoRoot, v.file);
      return `- ${rel}:${v.line} uses Panda styling APIs directly. Use primitives in web/src/lib/ui or define recipes in panda.config.ts instead.`;
    })
    .join("\n");
  console.error(
    [
      "UI policy check failed.",
      "Use primitives in web/src/lib/ui instead of raw HTML controls, and do not increase class bindings outside the UI DSL.",
      violations.length ? "Raw control violations:" : "",
      lines,
      classViolations.length ? "Class budget violations:" : "",
      classLines,
      pandaViolations.length ? "Panda styling API violations:" : "",
      pandaLines,
    ]
      .filter(Boolean)
      .join("\n"),
  );
  process.exit(1);
}

console.log("UI policy check passed.");
