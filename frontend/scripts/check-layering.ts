import { readdirSync, readFileSync } from 'node:fs';
import { join, relative } from 'node:path';
import { findLayeringViolations, isLayeringScanTarget, type ScannedFile } from './layering.ts';

const ROOT = process.cwd();
const SOURCE_DIR = join(ROOT, 'src');

function collectSources(dir: string): ScannedFile[] {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const absolute = join(dir, entry.name);
    if (entry.isDirectory()) return collectSources(absolute);
    const path = relative(ROOT, absolute);
    return isLayeringScanTarget(path) ? [{ path, content: readFileSync(absolute, 'utf8') }] : [];
  });
}

const violations = findLayeringViolations(collectSources(SOURCE_DIR));

if (violations.length > 0) {
  console.error('z-index must use a layer token (var(--layer-*)) defined in src/theme/tokens.css — see spec 202:');
  for (const { path, line, text } of violations) {
    console.error(`  ${path}:${line}  ${text}`);
  }
  process.exit(1);
}

console.log('Layering check passed.');
