import { Group, SegmentedControl, Select, Stack, Text } from '@mantine/core';
import { useMemo } from 'react';
import type { TargetPeriodInput } from '../types';

export const QUARTER_OPTIONS = [
  { value: '', label: '—' },
  { value: '1', label: 'Q1' },
  { value: '2', label: 'Q2' },
  { value: '3', label: 'Q3' },
  { value: '4', label: 'Q4' },
];

export function yearOptions(): { value: string; label: string }[] {
  const current = new Date().getFullYear();
  const options = [{ value: '', label: '—' }];
  for (let year = current - 1; year <= current + 10; year++) {
    options.push({ value: String(year), label: String(year) });
  }
  return options;
}

export interface PeriodValue {
  year?: number;
  quarter?: number;
}

export function isPeriodPaired(value: PeriodValue): boolean {
  return (value.year === undefined) === (value.quarter === undefined);
}

export function toTargetPeriod(value: PeriodValue): TargetPeriodInput | null {
  return value.year !== undefined && value.quarter !== undefined
    ? { year: value.year, quarter: value.quarter }
    : null;
}

export function PeriodFields({ value, onChange }: { value: PeriodValue; onChange: (value: PeriodValue) => void }) {
  const years = useMemo(yearOptions, []);
  return (
    <Group grow align="flex-start">
      <Select
        label="Target year"
        data={years}
        value={value.year !== undefined ? String(value.year) : ''}
        onChange={(year) => onChange({ ...value, year: year ? Number(year) : undefined })}
        data-testid="period-year"
      />
      <Stack gap="xs">
        <Text size="sm" fw={500}>
          Target quarter
        </Text>
        <SegmentedControl
          data={QUARTER_OPTIONS}
          value={value.quarter !== undefined ? String(value.quarter) : ''}
          onChange={(quarter) => onChange({ ...value, quarter: quarter ? Number(quarter) : undefined })}
          data-testid="period-quarter"
        />
      </Stack>
    </Group>
  );
}
