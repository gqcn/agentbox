import { brotliCompressSync, constants, gzipSync } from "node:zlib";
import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { dirname, extname, join, relative, sep } from "node:path";
import { fileURLToPath } from "node:url";

const distDir = fileURLToPath(new URL("../public", import.meta.url));
const args = new Set(process.argv.slice(2));
const shouldCompress = args.has("--compress");
const shouldSkipBudget = args.has("--no-budget");
const maxJavaScriptAssetCount = 24;
const maxTinyJavaScriptAssetCount = 8;
const tinyJavaScriptThreshold = kib(20);
const compressibleExtensions = new Set([
  ".css",
  ".html",
  ".js",
  ".json",
  ".mjs",
  ".otf",
  ".svg",
  ".ttf",
  ".txt",
  ".wasm",
  ".xml",
]);
const reportExtensions = new Set([".css", ".js"]);
const defaultBudget = {
  raw: kib(1536),
  gzip: kib(550),
  brotli: kib(450),
  note: "default JavaScript/CSS budget",
};
const allowlist = [
  {
    pattern: /^assets\/editor\.api2-[A-Za-z0-9_-]+\.js$/,
    raw: kib(2800),
    gzip: kib(750),
    brotli: kib(600),
    note: "Monaco editor ESM runtime; lazy loaded by Files/Git editor paths only",
  },
  {
    pattern: /^assets\/_\.contribution-[A-Za-z0-9_-]+\.js$/,
    raw: kib(1200),
    gzip: kib(320),
    brotli: kib(260),
    note: "Monaco basic language contributions; lazy loaded by editor paths only",
  },
  {
    pattern: /^assets\/monaco-basic-languages-[A-Za-z0-9_-]+\.js$/,
    raw: kib(4200),
    gzip: kib(1100),
    brotli: kib(850),
    note: "Monaco basic language bundle; consolidated lazy editor-only chunk",
  },
  {
    pattern: /^assets\/monaco-basic-languages-[A-Za-z0-9_-]+\.css$/,
    raw: kib(180),
    gzip: kib(32),
    brotli: kib(28),
    note: "Monaco basic language styles; lazy editor-only CSS",
  },
  {
    pattern: /^assets\/vendor-markdown-diagrams-[A-Za-z0-9_-]+\.js$/,
    raw: kib(3300),
    gzip: kib(900),
    brotli: kib(700),
    note: "Markdown Mermaid diagram renderer and diagram dependencies; lazy loaded by Markdown preview only",
  },
  {
    pattern: /^assets\/ts\.worker-[A-Za-z0-9_-]+\.js$/,
    raw: kib(7500),
    gzip: kib(1700),
    brotli: kib(1250),
    note: "Monaco TypeScript/JavaScript language worker; lazy loaded by editor paths only",
  },
  {
    pattern: /^assets\/css\.worker-[A-Za-z0-9_-]+\.js$/,
    raw: kib(1200),
    gzip: kib(280),
    brotli: kib(220),
    note: "Monaco CSS worker; editor-only worker with compressed delivery",
  },
  {
    pattern: /^assets\/html\.worker-[A-Za-z0-9_-]+\.js$/,
    raw: kib(800),
    gzip: kib(230),
    brotli: kib(180),
    note: "Monaco HTML worker; editor-only worker with compressed delivery",
  },
  {
    pattern: /^assets\/json\.worker-[A-Za-z0-9_-]+\.js$/,
    raw: kib(520),
    gzip: kib(150),
    brotli: kib(125),
    note: "Monaco JSON worker; editor-only worker with compressed delivery",
  },
];

if (args.has("--help")) {
  console.log(
    "Usage: node scripts/check-assets.mjs [--compress] [--no-budget]",
  );
  process.exit(0);
}

if (!existsSync(distDir)) {
  console.error(`dist directory does not exist: ${distDir}`);
  process.exit(1);
}

const files = listFiles(distDir)
  .filter((file) => !file.endsWith(".gz") && !file.endsWith(".br"))
  .sort();

if (shouldCompress) {
  for (const file of files) {
    if (!compressibleExtensions.has(extname(file))) {
      continue;
    }
    const source = readFileSync(file);
    writeCompressedFile(`${file}.gz`, gzipSync(source, { level: 9 }));
    writeCompressedFile(
      `${file}.br`,
      brotliCompressSync(source, {
        params: {
          [constants.BROTLI_PARAM_QUALITY]: 11,
        },
      }),
    );
  }
}

const report = files
  .filter((file) => reportExtensions.has(extname(file)))
  .map((file) => {
    const content = readFileSync(file);
    const relPath = displayPath(file);
    return {
      path: relPath,
      raw: content.length,
      gzip: gzipSync(content, { level: 9 }).length,
      brotli: brotliCompressSync(content, {
        params: {
          [constants.BROTLI_PARAM_QUALITY]: 11,
        },
      }).length,
      budget: budgetFor(relPath),
    };
  })
  .sort(
    (left, right) =>
      right.raw - left.raw || left.path.localeCompare(right.path),
  );

printReport(report);

const violations = shouldSkipBudget ? [] : budgetViolations(report);
if (violations.length > 0) {
  console.error("\nAsset budget violations found:");
  for (const violation of violations) {
    console.error(`- ${violation}`);
  }
  process.exit(1);
}

const countViolations = shouldSkipBudget ? [] : assetCountViolations(files);
if (countViolations.length > 0) {
  console.error("\nAsset count violations found:");
  for (const violation of countViolations) {
    console.error(`- ${violation}`);
  }
  process.exit(1);
}

console.log(
  shouldCompress
    ? "Asset report, compression, and budget check passed"
    : "Asset report and budget check passed",
);

function listFiles(dir) {
  const entries = [];
  for (const entry of readdirSync(dir)) {
    const path = join(dir, entry);
    const stat = statSync(path);
    if (stat.isDirectory()) {
      entries.push(...listFiles(path));
      continue;
    }
    if (stat.isFile()) {
      entries.push(path);
    }
  }
  return entries;
}

function writeCompressedFile(path, content) {
  mkdirSync(dirname(path), { recursive: true });
  writeFileSync(path, content);
}

function displayPath(file) {
  return relative(distDir, file).split(sep).join("/");
}

function budgetFor(path) {
  return allowlist.find((item) => item.pattern.test(path)) ?? defaultBudget;
}

function budgetViolations(items) {
  return items.flatMap((item) => {
    const failures = [];
    if (item.raw > item.budget.raw) {
      failures.push(
        `${item.path} raw ${formatKiB(item.raw)} exceeds ${formatKiB(item.budget.raw)} (${item.budget.note})`,
      );
    }
    if (item.gzip > item.budget.gzip) {
      failures.push(
        `${item.path} gzip ${formatKiB(item.gzip)} exceeds ${formatKiB(item.budget.gzip)} (${item.budget.note})`,
      );
    }
    if (item.brotli > item.budget.brotli) {
      failures.push(
        `${item.path} brotli ${formatKiB(item.brotli)} exceeds ${formatKiB(item.budget.brotli)} (${item.budget.note})`,
      );
    }
    return failures;
  });
}

function assetCountViolations(items) {
  const jsFiles = items
    .map(displayPath)
    .filter((path) => extname(path) === ".js");
  const tinyJsFiles = jsFiles.filter((path) => {
    const file = join(distDir, path);
    return statSync(file).size < tinyJavaScriptThreshold;
  });
  const failures = [];
  if (jsFiles.length > maxJavaScriptAssetCount) {
    failures.push(
      `JavaScript asset count ${jsFiles.length} exceeds ${maxJavaScriptAssetCount}; consolidate tiny lazy chunks or document a new budget`,
    );
  }
  if (tinyJsFiles.length > maxTinyJavaScriptAssetCount) {
    failures.push(
      `Tiny JavaScript asset count ${tinyJsFiles.length} exceeds ${maxTinyJavaScriptAssetCount}; files under ${formatKiB(tinyJavaScriptThreshold)} should not be emitted as many separate chunks`,
    );
  }
  return failures;
}

function printReport(items) {
  const rows = [
    ["resource", "raw", "gzip", "brotli", "budget note"],
    ...items.map((item) => [
      item.path,
      formatKiB(item.raw),
      formatKiB(item.gzip),
      formatKiB(item.brotli),
      item.budget.note,
    ]),
  ];
  const widths = rows[0].map((_, index) =>
    Math.max(...rows.map((row) => row[index].length)),
  );
  const lines = rows.map((row) =>
    row.map((value, index) => value.padEnd(widths[index])).join("  "),
  );
  console.log(lines.join("\n"));
}

function kib(value) {
  return value * 1024;
}

function formatKiB(value) {
  return `${(value / 1024).toFixed(1)} KiB`;
}
