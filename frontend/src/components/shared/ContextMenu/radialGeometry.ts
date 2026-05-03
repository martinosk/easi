export interface Point {
  x: number;
  y: number;
}

export interface Viewport {
  width: number;
  height: number;
}

const RADIUS_BY_COUNT: Record<number, number> = {
  1: 56,
  2: 64,
  3: 72,
  4: 82,
  5: 94,
  6: 104,
};
const FALLBACK_RADIUS = 110;

export const PETAL_HALF = 22;
export const VIEWPORT_PADDING = 8;

export function radiusFor(count: number): number {
  return RADIUS_BY_COUNT[count] ?? FALLBACK_RADIUS;
}

export function placePetals(count: number, radius: number): Point[] {
  if (count === 0) return [];
  if (count === 1) return [{ x: 0, y: -radius }];
  const step = (Math.PI * 2) / count;
  const startAngle = -Math.PI / 2;
  return Array.from({ length: count }, (_, i) => {
    const a = startAngle + step * i;
    return { x: radius * Math.cos(a), y: radius * Math.sin(a) };
  });
}

export function clampCenter(desired: Point, radius: number, viewport: Viewport): Point {
  const half = radius + PETAL_HALF;
  const maxX = viewport.width - half - VIEWPORT_PADDING;
  const minX = half + VIEWPORT_PADDING;
  const maxY = viewport.height - half - VIEWPORT_PADDING;
  const minY = half + VIEWPORT_PADDING;
  return {
    x: Math.min(Math.max(desired.x, minX), maxX),
    y: Math.min(Math.max(desired.y, minY), maxY),
  };
}
