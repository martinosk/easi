// .opencode/plugins/formatter.js
//
// Generic manifest-driven formatter/linter runner.
// Reads .opencode/project-stack.json and:
//   - on every write/edit: runs matching `format` entries (fail-loud)
//   - on session.idle:     runs matching `lint` entries against files
//                          touched during this turn (fail-loud, batched)
//
// Errors are NEVER suppressed. If a formatter or linter exits non-zero
// the error propagates so the agent sees and fixes it.

import { readFileSync } from "node:fs";
import { resolve, relative, extname, dirname, sep } from "node:path";

export const FormatterPlugin = async ({ $, directory }) => {
  const root = directory;
  const manifestPath = resolve(root, ".opencode/project-stack.json");
  const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
  const formatEntries = manifest.format ?? [];
  const lintEntries = manifest.lint ?? [];

  // Files the agent has touched since the last session.idle. Batched so lint
  // runs once per turn instead of once per edit.
  const touched = new Set();

  const substitute = (arg) => arg.replaceAll("{{ROOT}}", root);

  // Find the first manifest entry that matches a given absolute file path.
  // An entry matches iff (a) the file's extension is listed AND (b) the file
  // lives under the entry's cwd (so `frontend/` rules never fire on backend
  // files and vice versa).
  const findEntry = (entries, absFile) => {
    const ext = extname(absFile).toLowerCase();
    for (const e of entries) {
      const exts = (e.extensions ?? []).map((x) => x.toLowerCase());
      if (!exts.includes(ext)) continue;
      const cwd = resolve(root, e.cwd ?? ".");
      const rel = relative(cwd, absFile);
      if (rel.startsWith("..") || rel.startsWith(sep) || rel === "") continue;
      return { entry: e, cwd, rel };
    }
    return null;
  };

  // Run a single entry against a single file. Bun shell's array interpolation
  // escapes each element as a distinct argv entry, so no shell-quoting is
  // needed. The cwd is set via .cwd() rather than `cd && ...`.
  const runEntry = async (entry, cwd, rel) => {
    const argv = entry.command.map(substitute);
    const target = entry.argMode === "dir" ? (dirname(rel) || ".") : rel;
    argv.push(target);
    await $`${argv}`.cwd(cwd);
  };

  return {
    "tool.execute.after": async (input, output) => {
      if (input.tool !== "write" && input.tool !== "edit") return;
      const file = output.args?.filePath;
      if (!file) return;
      const abs = resolve(root, file);
      const match = findEntry(formatEntries, abs);
      if (!match) return;
      await runEntry(match.entry, match.cwd, match.rel);
      touched.add(abs);
    },

    event: async ({ event }) => {
      if (event.type !== "session.idle") return;
      if (touched.size === 0) return;
      const files = [...touched];
      touched.clear();
      // Deduplicate: a dir-mode entry should only run once per unique dir.
      const seen = new Set();
      for (const abs of files) {
        const match = findEntry(lintEntries, abs);
        if (!match) continue;
        const key = `${lintEntries.indexOf(match.entry)}:${match.entry.argMode === "dir" ? dirname(match.rel) : match.rel}`;
        if (seen.has(key)) continue;
        seen.add(key);
        await runEntry(match.entry, match.cwd, match.rel);
      }
    },
  };
};
