import { existsSync, readFileSync, readdirSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..');
const outputPath = join(repoRoot, 'docs', 'architecture', 'components.csv');
const repoName = 'easi';

function listDirs(relPath) {
  return readdirSync(join(repoRoot, relPath), { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name)
    .sort();
}

function buildRows() {
  const rows = [];
  for (const context of listDirs('backend/internal')) {
    rows.push([`backend:${context}`, `${repoName}/backend/internal/${context}/**`]);
  }
  rows.push(['backend:cmd', `${repoName}/backend/cmd/**`]);
  for (const feature of listDirs('frontend/src/features')) {
    rows.push([`frontend:${feature}`, `${repoName}/frontend/src/features/${feature}/**`]);
  }
  for (const shared of listDirs('frontend/src').filter((name) => name !== 'features')) {
    rows.push(['frontend:shared', `${repoName}/frontend/src/${shared}/**`]);
  }
  rows.push(['frontend:e2e', `${repoName}/frontend/e2e/**`]);
  rows.push(['deployment', `${repoName}/k8s/**`]);
  rows.push(['deployment', `${repoName}/terraform/**`]);
  return rows.sort(([nameA, patternA], [nameB, patternB]) =>
    nameA === nameB ? patternA.localeCompare(patternB) : nameA.localeCompare(nameB),
  );
}

function check(csv) {
  const normalize = (text) => text.replace(/\r\n/g, '\n');
  const current = existsSync(outputPath) ? normalize(readFileSync(outputPath, 'utf8')) : '';
  if (current !== csv) {
    console.error(
      'docs/architecture/components.csv is out of date. Run: node scripts/generate-architecture-components.js',
    );
    process.exit(1);
  }
  console.log('docs/architecture/components.csv is up to date');
}

function write(rows, csv) {
  writeFileSync(outputPath, csv);
  console.log(`Wrote ${rows.length} component mappings to docs/architecture/components.csv`);
}

function main() {
  const rows = buildRows();
  const csv = `${rows.map((row) => row.join(',')).join('\n')}\n`;
  if (process.argv.includes('--check')) {
    check(csv);
  } else {
    write(rows, csv);
  }
}

main();
