import { useMantineTheme } from '@mantine/core';
import { useMediaQuery } from '@mantine/hooks';

export type Breakpoint = 'xs' | 'sm' | 'md' | 'lg' | 'xl';

export interface ResponsiveValue<T> {
  base?: T;
  xs?: T;
  sm?: T;
  md?: T;
  lg?: T;
  xl?: T;
}

export interface ResponsiveBreakpoints {
  isXs: boolean;
  isSm: boolean;
  isMd: boolean;
  isLg: boolean;
  isXl: boolean;
  isMobile: boolean;
  isTablet: boolean;
  isDesktop: boolean;
  currentBreakpoint: Breakpoint;
}

export function useResponsive(): ResponsiveBreakpoints {
  const theme = useMantineTheme();

  const isXs = useMediaQuery(`(max-width: ${theme.breakpoints.sm})`) ?? false;
  const isSm = useMediaQuery(`(min-width: ${theme.breakpoints.sm}) and (max-width: ${theme.breakpoints.md})`) ?? false;
  const isMd = useMediaQuery(`(min-width: ${theme.breakpoints.md}) and (max-width: ${theme.breakpoints.lg})`) ?? false;
  const isLg = useMediaQuery(`(min-width: ${theme.breakpoints.lg}) and (max-width: ${theme.breakpoints.xl})`) ?? false;
  const isXl = useMediaQuery(`(min-width: ${theme.breakpoints.xl})`) ?? false;

  const isMobile = isXs;
  const isTablet = isSm || isMd;
  const isDesktop = isLg || isXl;

  const currentBreakpoint: Breakpoint = isXs ? 'xs' : isSm ? 'sm' : isMd ? 'md' : isLg ? 'lg' : 'xl';

  return {
    isXs,
    isSm,
    isMd,
    isLg,
    isXl,
    isMobile,
    isTablet,
    isDesktop,
    currentBreakpoint,
  };
}

export function getResponsiveValue<T>(value: ResponsiveValue<T>, breakpoint: Breakpoint): T | undefined {
  switch (breakpoint) {
    case 'xl':
      return value.xl ?? value.lg ?? value.md ?? value.sm ?? value.xs ?? value.base;
    case 'lg':
      return value.lg ?? value.md ?? value.sm ?? value.xs ?? value.base;
    case 'md':
      return value.md ?? value.sm ?? value.xs ?? value.base;
    case 'sm':
      return value.sm ?? value.xs ?? value.base;
    case 'xs':
      return value.xs ?? value.base;
    default:
      return value.base;
  }
}

export function useResponsiveValue<T>(value: ResponsiveValue<T>): T | undefined {
  const { currentBreakpoint } = useResponsive();
  return getResponsiveValue(value, currentBreakpoint);
}

export const RESPONSIVE_GRID_COLUMNS = {
  L1: {
    base: 'repeat(auto-fill, minmax(200px, 1fr))',
    xs: 'repeat(auto-fill, minmax(150px, 1fr))',
    sm: 'repeat(auto-fill, minmax(180px, 1fr))',
    md: 'repeat(auto-fill, minmax(200px, 1fr))',
    lg: 'repeat(auto-fill, minmax(250px, 1fr))',
    xl: 'repeat(auto-fill, minmax(300px, 1fr))',
  },
  L2: {
    base: 'repeat(auto-fill, minmax(120px, 1fr))',
    xs: 'repeat(auto-fill, minmax(100px, 1fr))',
    sm: 'repeat(auto-fill, minmax(110px, 1fr))',
    md: 'repeat(auto-fill, minmax(120px, 1fr))',
    lg: 'repeat(auto-fill, minmax(140px, 1fr))',
    xl: 'repeat(auto-fill, minmax(160px, 1fr))',
  },
  L3: {
    base: 'repeat(auto-fill, minmax(100px, 1fr))',
    xs: 'repeat(auto-fill, minmax(80px, 1fr))',
    sm: 'repeat(auto-fill, minmax(90px, 1fr))',
    md: 'repeat(auto-fill, minmax(100px, 1fr))',
    lg: 'repeat(auto-fill, minmax(120px, 1fr))',
    xl: 'repeat(auto-fill, minmax(140px, 1fr))',
  },
  L4: {
    base: 'repeat(auto-fill, minmax(100px, 1fr))',
    xs: 'repeat(auto-fill, minmax(80px, 1fr))',
    sm: 'repeat(auto-fill, minmax(90px, 1fr))',
    md: 'repeat(auto-fill, minmax(100px, 1fr))',
    lg: 'repeat(auto-fill, minmax(110px, 1fr))',
    xl: 'repeat(auto-fill, minmax(120px, 1fr))',
  },
};

export const RESPONSIVE_SPACING = {
  containerPadding: {
    base: '0.5rem',
    xs: '0.5rem',
    sm: '0.75rem',
    md: '1rem',
    lg: '1.25rem',
    xl: '1.5rem',
  },
  gridGap: {
    base: '0.5rem',
    xs: '0.375rem',
    sm: '0.5rem',
    md: '0.75rem',
    lg: '1rem',
    xl: '1rem',
  },
};

export const RESPONSIVE_FONT_SIZES = {
  title: {
    base: '1rem',
    xs: '0.875rem',
    sm: '0.9375rem',
    md: '1rem',
    lg: '1.125rem',
    xl: '1.25rem',
  },
  body: {
    base: '0.875rem',
    xs: '0.75rem',
    sm: '0.8125rem',
    md: '0.875rem',
    lg: '0.9375rem',
    xl: '1rem',
  },
  small: {
    base: '0.75rem',
    xs: '0.625rem',
    sm: '0.6875rem',
    md: '0.75rem',
    lg: '0.8125rem',
    xl: '0.875rem',
  },
};
