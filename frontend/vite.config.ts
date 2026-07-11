import react from '@vitejs/plugin-react';
import { visualizer } from 'rollup-plugin-visualizer';
import { defineConfig } from 'vite';

const vendorChunks: { chunk: string; packages: string[] }[] = [
  { chunk: 'react-vendor', packages: ['react-dom', 'react-router-dom', 'react'] },
  { chunk: 'mantine', packages: ['@mantine/core', '@mantine/hooks'] },
  { chunk: 'reactflow', packages: ['@xyflow/react', 'dagre'] },
  { chunk: 'query', packages: ['@tanstack/react-query', '@tanstack/query-core'] },
  { chunk: 'forms', packages: ['react-hook-form', '@hookform/resolvers', 'zod'] },
  { chunk: 'state', packages: ['zustand'] },
  { chunk: 'markdown', packages: ['react-markdown', 'remark-gfm'] },
  { chunk: 'utils', packages: ['axios', 'react-colorful', 'react-hot-toast'] },
];

function chunkForModule(id: string): string | undefined {
  const path = id.replaceAll('\\', '/');
  return vendorChunks.find(({ packages }) => packages.some((pkg) => path.includes(`/node_modules/${pkg}/`)))?.chunk;
}

export default defineConfig({
  cacheDir: process.env.VITE_CACHE_DIR || undefined,
  plugins: [
    react(),
    visualizer({
      filename: 'dist/stats.html',
      open: false,
      gzipSize: true,
      brotliSize: true,
    }),
  ],
  base: process.env.VITE_BASE_PATH || '/',
  build: {
    rollupOptions: {
      output: {
        manualChunks: chunkForModule,
      },
    },
  },
});
