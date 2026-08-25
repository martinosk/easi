export type NavDensity = 'full' | 'compact' | 'overflow';

export interface NavLayoutInput {
  availableWidth: number;
  fullWidths: readonly number[];
  compactWidth: number;
  moreWidth: number;
  gap: number;
}

export interface NavLayout {
  mode: NavDensity;
  visibleCount: number;
}

function rowWidth(itemWidths: number, count: number, gap: number): number {
  return count === 0 ? 0 : itemWidths + gap * (count - 1);
}

export function computeNavLayout({
  availableWidth,
  fullWidths,
  compactWidth,
  moreWidth,
  gap,
}: NavLayoutInput): NavLayout {
  const count = fullWidths.length;
  const fullTotal = fullWidths.reduce((sum, w) => sum + w, 0);
  if (rowWidth(fullTotal, count, gap) <= availableWidth) {
    return { mode: 'full', visibleCount: count };
  }
  if (rowWidth(compactWidth * count, count, gap) <= availableWidth) {
    return { mode: 'compact', visibleCount: count };
  }
  const roomForIcons = availableWidth - moreWidth - gap;
  const fits = Math.floor((roomForIcons + gap) / (compactWidth + gap));
  return { mode: 'overflow', visibleCount: Math.max(0, Math.min(count - 1, fits)) };
}
