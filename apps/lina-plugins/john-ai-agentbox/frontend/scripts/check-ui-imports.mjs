import { readdirSync, readFileSync, statSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { join, relative, sep } from 'node:path';

const srcDir = fileURLToPath(new URL('../src', import.meta.url));
const forbidden = [
  'ant-design-vue',
  '@ant-design/icons-vue',
  'vue',
  'lucide-vue-next',
];
const offenders = [];
const nativeControlPattern = /<(button|input|select|textarea|label|form|datalist|option)\b/g;
const uiComponentPath = `${sep}components${sep}ui${sep}`;

function lineNumber(content, index) {
  return content.slice(0, index).split('\n').length;
}

function displayPath(path) {
  return relative(srcDir, path);
}

function isUnifiedUiFile(path) {
  return path.includes(uiComponentPath);
}

function walk(dir) {
  for (const entry of readdirSync(dir)) {
    const path = join(dir, entry);
    const stat = statSync(path);
    if (stat.isDirectory()) {
      walk(path);
      continue;
    }
    if (!/\.(ts|tsx)$/.test(path)) {
      continue;
    }
    const content = readFileSync(path, 'utf8');
    for (const moduleName of forbidden) {
      const pattern = new RegExp(`from ['"]${moduleName}['"]|import\\(['"]${moduleName}['"]\\)`);
      if (pattern.test(content)) {
        offenders.push(`${displayPath(path)}: imports ${moduleName}`);
      }
    }
    if (/\.tsx$/.test(path) && !isUnifiedUiFile(path)) {
      for (const match of content.matchAll(nativeControlPattern)) {
        offenders.push(`${displayPath(path)}:${lineNumber(content, match.index ?? 0)} uses native <${match[1]}>; use @/components/ui`);
      }
    }
  }
}

walk(srcDir);

if (offenders.length > 0) {
  console.error('UI consistency violations found:');
  for (const offender of offenders) {
    console.error(`- ${offender}`);
  }
  process.exit(1);
}

console.log('UI import check passed');
