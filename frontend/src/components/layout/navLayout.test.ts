import { describe, expect, it } from 'vitest';
import { computeNavLayout } from './navLayout';

const base = { compactWidth: 40, moreWidth: 40, gap: 4 };

describe('computeNavLayout', () => {
  it('shows all entries with labels when every label fits', () => {
    const result = computeNavLayout({ ...base, availableWidth: 400, fullWidths: [100, 100, 100] });
    expect(result).toEqual({ mode: 'full', visibleCount: 3 });
  });

  it('treats the exact full width (including gaps) as fitting', () => {
    const result = computeNavLayout({ ...base, availableWidth: 308, fullWidths: [100, 100, 100] });
    expect(result).toEqual({ mode: 'full', visibleCount: 3 });
  });

  it('switches to icons only when labels do not fit but icons do', () => {
    const result = computeNavLayout({ ...base, availableWidth: 307, fullWidths: [100, 100, 100] });
    expect(result).toEqual({ mode: 'compact', visibleCount: 3 });
  });

  it('overflows trailing entries behind a More button when icons do not fit', () => {
    const result = computeNavLayout({ ...base, availableWidth: 100, fullWidths: [100, 100, 100, 100] });
    expect(result).toEqual({ mode: 'overflow', visibleCount: 1 });
  });

  it('can overflow every entry when only the More button fits', () => {
    const result = computeNavLayout({ ...base, availableWidth: 50, fullWidths: [100, 100] });
    expect(result).toEqual({ mode: 'overflow', visibleCount: 0 });
  });

  it('never reports a negative visible count', () => {
    const result = computeNavLayout({ ...base, availableWidth: 0, fullWidths: [100, 100] });
    expect(result).toEqual({ mode: 'overflow', visibleCount: 0 });
  });

  it('is full mode with no entries', () => {
    const result = computeNavLayout({ ...base, availableWidth: 0, fullWidths: [] });
    expect(result).toEqual({ mode: 'full', visibleCount: 0 });
  });
});
