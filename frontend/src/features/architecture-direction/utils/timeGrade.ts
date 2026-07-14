import type { TimeGrade } from '../types';

export const TIME_GRADES: readonly TimeGrade[] = ['Invest', 'Tolerate', 'Migrate', 'Eliminate'];

export function normalizeTimeGrade(raw: string | null | undefined): TimeGrade | null {
  if (!raw) return null;
  const candidate = raw.charAt(0).toUpperCase() + raw.slice(1).toLowerCase();
  return (TIME_GRADES as string[]).includes(candidate) ? (candidate as TimeGrade) : null;
}
