import { describe, expect, it } from 'vitest';
import { computeOffsetPosition } from './offsetPosition';

describe('computeOffsetPosition', () => {
  const source = { x: 100, y: 200 };

  it('places the new node above the source for "top"', () => {
    const result = computeOffsetPosition(source, 'top');
    expect(result.x).toBe(100);
    expect(result.y).toBeLessThan(200);
  });

  it('places the new node below the source for "bottom"', () => {
    const result = computeOffsetPosition(source, 'bottom');
    expect(result.x).toBe(100);
    expect(result.y).toBeGreaterThan(200);
  });

  it('places the new node to the right of the source for "right"', () => {
    const result = computeOffsetPosition(source, 'right');
    expect(result.x).toBeGreaterThan(100);
    expect(result.y).toBe(200);
  });

  it('places the new node to the left of the source for "left"', () => {
    const result = computeOffsetPosition(source, 'left');
    expect(result.x).toBeLessThan(100);
    expect(result.y).toBe(200);
  });

  it('produces deterministic results for the same input', () => {
    const a = computeOffsetPosition(source, 'right');
    const b = computeOffsetPosition(source, 'right');
    expect(a).toEqual(b);
  });

  it('does not overlap the source node (offset >= node width)', () => {
    const right = computeOffsetPosition(source, 'right');
    const left = computeOffsetPosition(source, 'left');
    expect(right.x - source.x).toBeGreaterThanOrEqual(180);
    expect(source.x - left.x).toBeGreaterThanOrEqual(180);
  });

  it('does not overlap vertically (offset >= node height)', () => {
    const top = computeOffsetPosition(source, 'top');
    const bottom = computeOffsetPosition(source, 'bottom');
    expect(bottom.y - source.y).toBeGreaterThanOrEqual(100);
    expect(source.y - top.y).toBeGreaterThanOrEqual(100);
  });
});
