import type { MantineColorsTuple } from '@mantine/core';
import { createTheme } from '@mantine/core';

const cssVar = (name: string) => `var(--${name})`;

function paletteTuple(prefix: string): MantineColorsTuple {
  return [0, 1, 2, 3, 4, 5, 6, 7, 8, 9].map((i) => cssVar(`${prefix}-${i}`)) as unknown as MantineColorsTuple;
}

const gray: MantineColorsTuple = [
  cssVar('color-gray-50'),
  cssVar('color-gray-100'),
  cssVar('color-gray-200'),
  cssVar('color-gray-300'),
  cssVar('color-gray-400'),
  cssVar('color-gray-500'),
  cssVar('color-gray-600'),
  cssVar('color-gray-700'),
  cssVar('color-gray-800'),
  cssVar('color-gray-900'),
];

export const theme = createTheme({
  primaryColor: 'blue',
  defaultRadius: 'md',
  colors: {
    blue: paletteTuple('color-blue'),
    purple: paletteTuple('color-purple'),
    gray,
  },
  spacing: {
    xs: cssVar('spacing-xs'),
    sm: cssVar('spacing-sm'),
    md: cssVar('spacing-md'),
    lg: cssVar('spacing-lg'),
    xl: cssVar('spacing-xl'),
    xxl: cssVar('spacing-2xl'),
  },
  fontSizes: {
    xs: cssVar('font-size-xs'),
    sm: cssVar('font-size-sm'),
    md: cssVar('font-size-base'),
    lg: cssVar('font-size-lg'),
    xl: cssVar('font-size-xl'),
    xxl: cssVar('font-size-2xl'),
  },
  radius: {
    xs: cssVar('radius-xs'),
    sm: cssVar('radius-sm'),
    md: cssVar('radius-md'),
    lg: cssVar('radius-lg'),
    xl: cssVar('radius-xl'),
  },
  shadows: {
    xs: cssVar('shadow-xs'),
    sm: cssVar('shadow-sm'),
    md: cssVar('shadow-md'),
    lg: cssVar('shadow-lg'),
    xl: cssVar('shadow-xl'),
  },
  fontFamily: cssVar('font-family'),
  headings: { fontFamily: cssVar('font-family') },
});
