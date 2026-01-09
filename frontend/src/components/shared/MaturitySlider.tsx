import React, { useMemo } from 'react';
import { Slider, Box, Text, Group, Stack } from '@mantine/core';
import { useMaturityScale } from '../../hooks/useMaturityScale';
import { getDefaultSections, getSectionForValue, getMaturityBounds } from '../../utils/maturity';
import type { MaturityScaleSection } from '../../api/types';

interface MaturitySliderProps {
  value: number;
  onChange: (value: number) => void;
  disabled?: boolean;
}

interface SectionLabel {
  section: MaturityScaleSection;
  width: number;
}

function calculateSectionLabels(sections: MaturityScaleSection[]): SectionLabel[] {
  return sections.map((section) => ({
    section,
    width: section.maxValue - section.minValue + 1,
  }));
}

interface SectionLabelsRowProps {
  sectionLabels: SectionLabel[];
  currentSectionName: string | undefined;
}

const SectionLabelsRow: React.FC<SectionLabelsRowProps> = ({ sectionLabels, currentSectionName }) => (
  <Group gap={0} mb="xs">
    {sectionLabels.map(({ section, width }) => (
      <Box key={section.name} style={{ width: `${width}%`, textAlign: 'center' }}>
        <Text
          size="xs"
          c={currentSectionName === section.name ? 'blue' : 'dimmed'}
          fw={currentSectionName === section.name ? 600 : 400}
        >
          {section.name}
        </Text>
      </Box>
    ))}
  </Group>
);

interface SectionTrackProps {
  sectionLabels: SectionLabel[];
  currentSectionName: string | undefined;
  maxBound: number;
}

const SectionTrack: React.FC<SectionTrackProps> = ({ sectionLabels, currentSectionName, maxBound }) => (
  <Box
    pos="absolute"
    top={0}
    left={0}
    right={0}
    style={{ height: '6px', borderRadius: '3px', overflow: 'hidden', display: 'flex' }}
  >
    {sectionLabels.map(({ section, width }) => (
      <Box
        key={section.name}
        style={{
          width: `${width}%`,
          backgroundColor: currentSectionName === section.name
            ? 'var(--mantine-color-blue-1)'
            : 'var(--mantine-color-gray-1)',
          borderRight: section.maxValue !== maxBound ? '2px solid var(--mantine-color-gray-3)' : 'none',
        }}
      />
    ))}
  </Box>
);

function useKeyboardNavigation(
  value: number,
  disabled: boolean,
  onChange: (v: number) => void,
  bounds: { min: number; max: number }
) {
  return (event: React.KeyboardEvent) => {
    if (disabled) return;
    const step = event.shiftKey ? 10 : 1;
    let newValue = value;

    if (event.key === 'ArrowRight' || event.key === 'ArrowUp') {
      newValue = Math.min(bounds.max, value + step);
      event.preventDefault();
    } else if (event.key === 'ArrowLeft' || event.key === 'ArrowDown') {
      newValue = Math.max(bounds.min, value - step);
      event.preventDefault();
    }

    if (newValue !== value) onChange(newValue);
  };
}

const sliderStyles = {
  track: { backgroundColor: 'transparent' },
  bar: { backgroundColor: 'transparent' },
  markLabel: { display: 'none' as const },
};

export const MaturitySlider: React.FC<MaturitySliderProps> = ({ value, onChange, disabled = false }) => {
  const { data: maturityScale } = useMaturityScale();

  const sections = useMemo(() => {
    if (!maturityScale?.sections || maturityScale.sections.length === 0) {
      return getDefaultSections();
    }
    return maturityScale.sections;
  }, [maturityScale]);

  const bounds = useMemo(() => getMaturityBounds(sections), [sections]);
  const sectionLabels = useMemo(() => calculateSectionLabels(sections), [sections]);
  const currentSection = useMemo(() => getSectionForValue(value, sections), [value, sections]);
  const marks = useMemo(() => [...sections.map((s) => ({ value: s.minValue, label: '' })), { value: bounds.max, label: '' }], [sections, bounds]);
  const handleKeyDown = useKeyboardNavigation(value, disabled, onChange, bounds);
  const ariaValueText = currentSection ? `${value} - ${currentSection.name}` : `${value}`;

  return (
    <Stack gap="xs">
      <Text size="sm" fw={500}>Maturity Level</Text>
      <Box pos="relative" mb="md">
        <SectionLabelsRow sectionLabels={sectionLabels} currentSectionName={currentSection?.name} />
        <Box pos="relative">
          <SectionTrack sectionLabels={sectionLabels} currentSectionName={currentSection?.name} maxBound={bounds.max} />
          <Slider
            value={value}
            onChange={onChange}
            onKeyDown={handleKeyDown}
            min={bounds.min}
            max={bounds.max}
            step={1}
            disabled={disabled}
            marks={marks}
            label={(val) => `${val}`}
            styles={sliderStyles}
            aria-valuemin={bounds.min}
            aria-valuemax={bounds.max}
            aria-valuenow={value}
            aria-valuetext={ariaValueText}
            data-testid="maturity-slider"
          />
        </Box>
      </Box>
    </Stack>
  );
};
