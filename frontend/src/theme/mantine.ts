import { createTheme } from '@mantine/core';
import type { MantineColorsTuple } from '@mantine/core';

const blue: MantineColorsTuple = [
  '#eff6ff', '#dbeafe', '#bfdbfe', '#93c5fd', '#60a5fa',
  '#3b82f6', '#2563eb', '#1e40af', '#1e3a8a', '#172554',
];

const purple: MantineColorsTuple = [
  '#faf5ff', '#f3e8ff', '#e9d5ff', '#d8b4fe', '#c084fc',
  '#a855f7', '#9333ea', '#7e22ce', '#6b21a8', '#581c87',
];

const gray: MantineColorsTuple = [
  '#f9fafb', '#f3f4f6', '#e5e7eb', '#d1d5db', '#9ca3af',
  '#6b7280', '#4b5563', '#374151', '#1f2937', '#111827',
];

const systemFontStack = '-apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif';

const spacing = { xs: '0.25rem', sm: '0.5rem', md: '1rem', lg: '1.5rem', xl: '2rem', xxl: '3rem' };

const fontSizes = { xs: '0.75rem', sm: '0.875rem', md: '1rem', lg: '1.125rem', xl: '1.25rem', xxl: '1.5rem' };

const radius = { xs: '0.125rem', sm: '0.25rem', md: '0.375rem', lg: '0.5rem', xl: '0.75rem' };

const shadows = {
  xs: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
  sm: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
  md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
  lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
  xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
};

export const theme = createTheme({
  primaryColor: 'blue',
  defaultRadius: 'md',
  colors: { blue, purple, gray },
  spacing,
  fontSizes,
  radius,
  shadows,
  fontFamily: systemFontStack,
  headings: { fontFamily: systemFontStack },
});
