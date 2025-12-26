import type { MaturityScaleSection } from '../api/types';

const DEFAULT_SECTIONS: MaturityScaleSection[] = [
  { name: 'Genesis', order: 1, minValue: 0, maxValue: 24 },
  { name: 'Custom Built', order: 2, minValue: 25, maxValue: 49 },
  { name: 'Product', order: 3, minValue: 50, maxValue: 74 },
  { name: 'Commodity', order: 4, minValue: 75, maxValue: 99 },
];

export function getDefaultSections(): MaturityScaleSection[] {
  return DEFAULT_SECTIONS;
}

export function deriveLegacyMaturityValue(
  maturityLevel: string,
  sections: MaturityScaleSection[]
): number {
  const section = sections.find(
    (s) => s.name.toLowerCase() === maturityLevel.toLowerCase()
  );

  if (!section) {
    return 12;
  }

  return Math.floor((section.minValue + section.maxValue) / 2);
}

export function getSectionForValue(
  value: number,
  sections: MaturityScaleSection[]
): MaturityScaleSection | undefined {
  return sections.find((s) => value >= s.minValue && value <= s.maxValue);
}
