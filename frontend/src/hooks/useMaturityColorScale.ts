import { useMemo } from 'react';
import { DEFAULT_MATURITY_COLOR, SECTION_COLORS } from '../constants/maturityColors';
import { interpolateHsl, clampMaturityValue } from '../utils/colorInterpolation';
import { useMaturityScale } from './useMaturityScale';
import { getMaturityBounds, getDefaultSections } from '../utils/maturity';
import type { MaturityScaleSection } from '../api/types';

interface MaturityColorScale {
  getColorForValue: (maturityValue: number) => string;
  getSectionNameForValue: (maturityValue: number) => string;
  getBaseSectionColor: (order: number) => string;
  bounds: { min: number; max: number };
}

interface ColoredSection extends MaturityScaleSection {
  lightColor: string;
  saturatedColor: string;
}

const addColorsToSections = (sections: MaturityScaleSection[]): ColoredSection[] => {
  return sections.map(section => {
    const colors = SECTION_COLORS[section.order] || { lightColor: '#E5E7EB', saturatedColor: '#6B7280' };
    return { ...section, ...colors };
  });
};

const findSectionForValue = (value: number, sections: ColoredSection[]): ColoredSection | undefined => {
  return sections.find(section => value >= section.minValue && value <= section.maxValue);
};

const calculatePositionInSection = (value: number, section: ColoredSection): number => {
  const range = section.maxValue - section.minValue;
  if (range === 0) return 0;
  return (value - section.minValue) / range;
};

const buildColorLookupTable = (sections: ColoredSection[], bounds: { min: number; max: number }): Map<number, string> => {
  const lookupTable = new Map<number, string>();

  for (let value = bounds.min; value <= bounds.max; value++) {
    const section = findSectionForValue(value, sections);
    if (!section) {
      lookupTable.set(value, DEFAULT_MATURITY_COLOR);
      continue;
    }

    const position = calculatePositionInSection(value, section);
    const color = interpolateHsl(section.lightColor, section.saturatedColor, position);
    lookupTable.set(value, color);
  }

  return lookupTable;
};

export const useMaturityColorScale = (): MaturityColorScale => {
  const { data: maturityScale } = useMaturityScale();

  const apiSections = maturityScale?.sections ?? getDefaultSections();
  const sections = useMemo(() => addColorsToSections(apiSections), [apiSections]);
  const bounds = useMemo(() => getMaturityBounds(apiSections), [apiSections]);

  const colorLookupTable = useMemo(() => buildColorLookupTable(sections, bounds), [sections, bounds]);

  const sectionByOrder = useMemo(() => {
    const map = new Map<number, ColoredSection>();
    sections.forEach(section => map.set(section.order, section));
    return map;
  }, [sections]);

  const getColorForValue = (maturityValue: number): string => {
    const clampedValue = clampMaturityValue(maturityValue, bounds.min, bounds.max);
    return colorLookupTable.get(clampedValue) || DEFAULT_MATURITY_COLOR;
  };

  const getSectionNameForValue = (maturityValue: number): string => {
    const clampedValue = clampMaturityValue(maturityValue, bounds.min, bounds.max);
    const section = findSectionForValue(clampedValue, sections);
    return section?.name || 'Unknown';
  };

  const getBaseSectionColor = (order: number): string => {
    const section = sectionByOrder.get(order);
    return section?.saturatedColor || DEFAULT_MATURITY_COLOR;
  };

  return {
    getColorForValue,
    getSectionNameForValue,
    getBaseSectionColor,
    bounds,
  };
};
