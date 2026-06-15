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
          if (id.includes('node_modules/react') || id.includes('node_modules/react-dom') || id.includes('node_modules/react-router-dom')) return 'react-vendor';
          if (id.includes('node_modules/@mantine')) return 'mantine';
          if (id.includes('node_modules/@xyflow') || id.includes('node_modules/dagre')) return 'reactflow';
          if (id.includes('node_modules/dockview')) return 'dockview';
          if (id.includes('node_modules/@tanstack')) return 'query';
          if (id.includes('node_modules/react-hook-form') || id.includes('node_modules/@hookform') || id.includes('node_modules/zod')) return 'forms';
          if (id.includes('node_modules/zustand')) return 'state';
          if (id.includes('node_modules/react-markdown') || id.includes('node_modules/remark-gfm')) return 'markdown';
          if (id.includes('node_modules/axios') || id.includes('node_modules/react-colorful') || id.includes('node_modules/react-hot-toast')) return 'utils';
        },
      },
    },
  },
});
