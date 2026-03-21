// .opencode/plugins/formatter.js
// Auto-formats TypeScript/JavaScript files with prettier + eslint after every write/edit.
// Go files are handled by gofmt (install separately if needed).

export const FormatterPlugin = async ({ $ }) => {
  return {
    "tool.execute.after": async (input) => {
      const file = input.args?.filePath;
      if (!file) return;
      if (input.tool !== "write" && input.tool !== "edit") return;

      // TypeScript / JavaScript — prettier then eslint
      if (/\.(ts|tsx|js|jsx|mjs)$/.test(file)) {
        try {
          await $`npx prettier --write ${file}`;
        } catch (_) {
          // prettier not installed or formatting failed — skip silently
        }
        try {
          await $`npx eslint --fix ${file}`;
        } catch (_) {
          // eslint not installed or no fixable issues — skip silently
        }
      }

      // Go — gofmt
      // NOTE: gofmt was not detected in the current environment.
      // Install Go toolchain to enable auto-formatting.
      // if (file.endsWith(".go")) {
      //   await $`gofmt -w ${file}`;
      // }
    },
  };
};
