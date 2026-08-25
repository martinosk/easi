import { type RefObject, useLayoutEffect, useState } from 'react';
import { computeNavLayout, type NavLayout } from './navLayout';

const MEASURE_FULL = '[data-measure="full"]';
const MEASURE_COMPACT = '[data-measure="compact"]';
const MEASURE_MORE = '[data-measure="more"]';
const NAV_ROW = '[data-nav-row]';
const MEASURE_TWIN = '[data-measure-twin]';

function widthsOf(root: HTMLElement, selector: string): number[] {
  return Array.from(root.querySelectorAll<HTMLElement>(selector)).map((el) => el.offsetWidth);
}

function rowGap(root: HTMLElement): number {
  const row = root.querySelector<HTMLElement>(NAV_ROW);
  return row ? Number.parseFloat(getComputedStyle(row).columnGap) || 0 : 0;
}

export function measureNav(nav: HTMLElement): NavLayout {
  const [compactWidth = 0] = widthsOf(nav, MEASURE_COMPACT);
  const [moreWidth = 0] = widthsOf(nav, MEASURE_MORE);
  return computeNavLayout({
    availableWidth: nav.offsetWidth,
    fullWidths: widthsOf(nav, MEASURE_FULL),
    compactWidth,
    moreWidth,
    gap: rowGap(nav),
  });
}

export function useNavLayout(navRef: RefObject<HTMLElement | null>, entryCount: number): NavLayout {
  const [layout, setLayout] = useState<NavLayout>({ mode: 'full', visibleCount: entryCount });

  useLayoutEffect(() => {
    const nav = navRef.current;
    if (!nav) return;
    const relayout = () => setLayout(measureNav(nav));
    relayout();
    const observer = new ResizeObserver(relayout);
    observer.observe(nav);
    const twin = nav.querySelector<HTMLElement>(MEASURE_TWIN);
    if (twin) observer.observe(twin);
    return () => observer.disconnect();
  }, [navRef]);

  return { mode: layout.mode, visibleCount: Math.min(layout.visibleCount, entryCount) };
}
