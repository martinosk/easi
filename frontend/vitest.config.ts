import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    css: true,

    // Process isolation to prevent memory leaks between test files
    // React Flow components were causing memory leaks in JSDOM
    pool: 'forks',
    poolOptions: {
      forks: {
        singleFork: false,  // Use multiple forks for parallelization
        isolate: true,      // Each test file runs in isolated process
      },
    },

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
