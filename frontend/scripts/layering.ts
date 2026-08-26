export interface ScannedFile {
  path: string;
  content: string;
}

export interface LayeringViolation {
  path: string;
  line: number;
  text: string;
}

const Z_INDEX_DECLARATION = /\bz-?index\s*:\s*(?<value>[^;,}\n]+)/gi;
const LAYER_TOKEN = /^['"]?var\(--layer-[a-z]+\)['"]?$/;
const TOKEN_FILE = /(^|\/)src\/theme\/tokens\.css$/;
const TEST_FILE = /\.test\.(ts|tsx)$/;
const SCANNED_EXTENSION = /\.(css|ts|tsx)$/;

function offendingDeclarations(line: string): string[] {
  return Array.from(line.matchAll(Z_INDEX_DECLARATION))
    .filter((match) => !LAYER_TOKEN.test(match.groups?.value.trim() ?? ''))
    .map((match) => match[0].trim());
}

export function findLayeringViolations(files: readonly ScannedFile[]): LayeringViolation[] {
  return files.flatMap(({ path, content }) =>
    content.split('\n').flatMap((line, index) =>
      offendingDeclarations(line).map((text) => ({ path, line: index + 1, text: `${text};` })),
    ),
  );
}

export function isLayeringScanTarget(path: string): boolean {
  return SCANNED_EXTENSION.test(path) && !TOKEN_FILE.test(path) && !TEST_FILE.test(path);
}
