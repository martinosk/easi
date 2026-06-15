import react from '@vitejs/plugin-react';
import { visualizer } from 'rollup-plugin-visualizer';
import { defineConfig } from 'vite';

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
        manualChunks(id) {
          const p = id.replaceAll('\\', '/');
          const seg = (pkg: string) => p.includes(`/node_modules/${pkg}/`);
          if (seg('react-dom') || seg('react-router-dom') || seg('react')) return 'react-vendor';
          if (seg('@mantine/core') || seg('@mantine/hooks')) return 'mantine';
          if (seg('@xyflow/react') || seg('dagre')) return 'reactflow';
          if (seg('dockview')) return 'dockview';
          if (seg('@tanstack/react-query') || seg('@tanstack/query-core')) return 'query';
          if (seg('react-hook-form') || seg('@hookform/resolvers') || seg('zod')) return 'forms';
          if (seg('zustand')) return 'state';
          if (seg('react-markdown') || seg('remark-gfm')) return 'markdown';
          if (seg('axios') || seg('react-colorful') || seg('react-hot-toast')) return 'utils';
        },
      },
    },
  },
});
