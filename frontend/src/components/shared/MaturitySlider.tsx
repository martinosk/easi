import React, { useMemo } from 'react';
import { Slider, Box, Text, Group, Stack } from '@mantine/core';
import { useMaturityScale } from '../../hooks/useMaturityScale';
import { getDefaultSections, getSectionForValue } from '../../utils/maturity';
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
}

const SectionTrack: React.FC<SectionTrackProps> = ({ sectionLabels, currentSectionName }) => (
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
          borderRight: section.maxValue !== 99 ? '2px solid var(--mantine-color-gray-3)' : 'none',
        }}
      />
    ))}
  </Box>
);

function useKeyboardNavigation(value: number, disabled: boolean, onChange: (v: number) => void) {
  return (event: React.KeyboardEvent) => {
    if (disabled) return;
    const step = event.shiftKey ? 10 : 1;
    let newValue = value;

    if (event.key === 'ArrowRight' || event.key === 'ArrowUp') {
      newValue = Math.min(99, value + step);
      event.preventDefault();
    } else if (event.key === 'ArrowLeft' || event.key === 'ArrowDown') {
      newValue = Math.max(0, value - step);
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

  const sectionLabels = useMemo(() => calculateSectionLabels(sections), [sections]);
  const currentSection = useMemo(() => getSectionForValue(value, sections), [value, sections]);
  const marks = useMemo(() => [...sections.map((s) => ({ value: s.minValue, label: '' })), { value: 99, label: '' }], [sections]);
  const handleKeyDown = useKeyboardNavigation(value, disabled, onChange);
  const ariaValueText = currentSection ? `${value} - ${currentSection.name}` : `${value}`;

  return (
    <Stack gap="xs">
      <Text size="sm" fw={500}>Maturity Level</Text>
      <Box pos="relative" mb="md">
        <SectionLabelsRow sectionLabels={sectionLabels} currentSectionName={currentSection?.name} />
        <Box pos="relative">
          <SectionTrack sectionLabels={sectionLabels} currentSectionName={currentSection?.name} />
          <Slider
            value={value}
            onChange={onChange}
            onKeyDown={handleKeyDown}
            min={0}
            max={99}
            step={1}
            disabled={disabled}
            marks={marks}
            label={(val) => `${val}`}
            styles={sliderStyles}
            aria-valuemin={0}
            aria-valuemax={99}
            aria-valuenow={value}
            aria-valuetext={ariaValueText}
            data-testid="maturity-slider"
          />
        </Box>
      </Box>
    </Stack>
  );
};
