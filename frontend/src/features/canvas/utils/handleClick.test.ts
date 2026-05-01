import { describe, expect, it } from 'vitest';
import { isClickGesture } from './handleClick';

describe('isClickGesture', () => {
  it('returns true when pointer did not move', () => {
    expect(isClickGesture({ x: 100, y: 200 }, { x: 100, y: 200 })).toBe(true);
  });

  it('returns true when movement is exactly the threshold', () => {
    expect(isClickGesture({ x: 0, y: 0 }, { x: 5, y: 0 }, 5)).toBe(true);
  });

  it('returns true when diagonal movement is below the threshold', () => {
    expect(isClickGesture({ x: 0, y: 0 }, { x: 3, y: 3 }, 5)).toBe(true);
  });

  it('returns false when movement exceeds the threshold', () => {
    expect(isClickGesture({ x: 0, y: 0 }, { x: 6, y: 0 }, 5)).toBe(false);
  });

  it('returns false for clear drags far past the threshold', () => {
    expect(isClickGesture({ x: 0, y: 0 }, { x: 200, y: 200 }, 5)).toBe(false);
  });

  it('uses a default threshold of 5px', () => {
    expect(isClickGesture({ x: 0, y: 0 }, { x: 4, y: 0 })).toBe(true);
    expect(isClickGesture({ x: 0, y: 0 }, { x: 6, y: 0 })).toBe(false);
  });
});
