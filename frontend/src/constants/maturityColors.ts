export interface SectionColors {
  lightColor: string;
  saturatedColor: string;
}

export const SECTION_COLORS: Record<number, SectionColors> = {
  1: { lightColor: '#F6D9D5', saturatedColor: '#EF4444' },
  2: { lightColor: '#F3E3C9', saturatedColor: '#F97316' },
  3: { lightColor: '#EFE9C8', saturatedColor: '#EAB308' },
  4: { lightColor: '#CFE8DA', saturatedColor: '#10B981' },
};

export const DEFAULT_MATURITY_COLOR = '#6b7280';
export const CLASSIC_COLOR = '#DBE4F5';
export const DEFAULT_CUSTOM_COLOR = '#E2E7EB';

export const MATURITY_LEVEL_MIDPOINTS: Record<string, number> = {
  genesis: 12,
  'custom build': 37,
  product: 62,
  commodity: 87,
};

export const deriveMaturityValue = (maturityLevel?: string): number => {
  const level = maturityLevel?.toLowerCase();
  return level && level in MATURITY_LEVEL_MIDPOINTS ? MATURITY_LEVEL_MIDPOINTS[level] : 0;
};
