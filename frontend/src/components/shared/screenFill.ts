export type ScreenFill = 'viewport' | 'region';

export function screenFillProps(fill: ScreenFill): { flex: number } | { mih: string } {
  return fill === 'region' ? { flex: 1 } : { mih: '100vh' };
}
