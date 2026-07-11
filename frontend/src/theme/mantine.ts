import type { MantineColorsTuple } from '@mantine/core';
import { Badge, Button, createTheme, Drawer, Menu, Modal, Tooltip } from '@mantine/core';

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

const componentExtensions = {
  Button: Button.extend({
    styles: { root: { fontWeight: 500 } },
  }),
  Badge: Badge.extend({
    styles: { root: { fontWeight: 600, textTransform: 'none' } },
  }),
  Modal: Modal.extend({
    defaultProps: {
      radius: 'lg',
      shadow: 'xl',
      overlayProps: { opacity: 0.35, color: cssVar('color-gray-900') },
    },
  }),
  Drawer: Drawer.extend({
    defaultProps: { radius: 0, shadow: 'xl' },
  }),
  Menu: Menu.extend({
    defaultProps: { shadow: 'lg', radius: 'md' },
  }),
  Tooltip: Tooltip.extend({
    defaultProps: { fz: 'xs' },
  }),
};

export const theme = createTheme({
  primaryColor: 'accent',
  primaryShade: 8,
  defaultRadius: 'sm',
  colors: {
    accent: paletteTuple('skin-accent'),
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
  fontFamily: cssVar('font-ui'),
  fontFamilyMonospace: cssVar('font-mono'),
  headings: {
    fontFamily: cssVar('font-display'),
    fontWeight: '600',
    sizes: {
      h1: { fontSize: '1.5rem' },
      h2: { fontSize: '1.25rem' },
      h3: { fontSize: '1.0625rem' },
      h4: { fontSize: '0.9375rem' },
    },
  },
  components: componentExtensions,
});
