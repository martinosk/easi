const ISO_DATE_PATTERN = /^(\d{4})-(\d{2})-(\d{2})$/;

export function formatIsoDate(iso: string): string {
  const match = ISO_DATE_PATTERN.exec(iso);
  if (!match) return iso;
  const [, year, month, day] = match;
  const parsed = new Date(Number(year), Number(month) - 1, Number(day));
  return Number.isNaN(parsed.getTime()) ? iso : parsed.toLocaleDateString();
}
