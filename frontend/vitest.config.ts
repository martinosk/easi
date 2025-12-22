import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  define: {
    'import.meta.env.VITE_API_URL': JSON.stringify('http://localhost:8080'),
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    css: true,

    // Use threads pool - better compatibility with MSW interceptors
    // The forks pool causes "read EINVAL" errors with MSW's socket interceptors
    pool: 'threads',
    isolate: true,

    // Timeouts
    testTimeout: 10000,
    hookTimeout: 10000,

    // Exclude problematic and non-unit test files
    exclude: [
      '**/node_modules/**',
      '**/dist/**',
      '**/e2e/**',
      '**/ComponentCanvas*.test.tsx',  // Moved to E2E in Phase 3 (spec 011)
    ],

    // Test reporters for CI/CD
    reporters: process.env.CI ? ['default', 'junit'] : ['default'],
    outputFile: {
      junit: './test-results/junit.xml',
    },
  },
});
