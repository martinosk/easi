export interface SectionColors {
  lightColor: string;
  saturatedColor: string;
}

export const SECTION_COLORS: Record<number, SectionColors> = {
  1: { lightColor: '#FEE2E2', saturatedColor: '#EF4444' },
  2: { lightColor: '#FFEDD5', saturatedColor: '#F97316' },
  3: { lightColor: '#FEF9C3', saturatedColor: '#EAB308' },
  4: { lightColor: '#D1FAE5', saturatedColor: '#10B981' },
};

export const DEFAULT_MATURITY_COLOR = '#6b7280';
export const CLASSIC_COLOR = '#f9c268';
export const DEFAULT_CUSTOM_COLOR = '#E0E0E0';
export const SELECTED_BORDER_COLOR = '#374151';

export const MATURITY_LEVEL_MIDPOINTS: Record<string, number> = {
  'genesis': 12,
  'custom build': 37,
  'product': 62,
  'commodity': 87,
};

export const deriveMaturityValue = (maturityLevel?: string): number => {
  const level = maturityLevel?.toLowerCase();
  return level && level in MATURITY_LEVEL_MIDPOINTS
    ? MATURITY_LEVEL_MIDPOINTS[level]
    : 0;
};
