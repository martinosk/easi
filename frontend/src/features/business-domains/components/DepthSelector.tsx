import { SegmentedControl } from '@mantine/core';

export type DepthLevel = 1 | 2 | 3 | 4;

export interface DepthSelectorProps {
  value: DepthLevel;
  onChange: (depth: DepthLevel) => void;
}

const DEPTH_OPTIONS = [
  { value: '1', label: 'L1' },
  { value: '2', label: 'L1-L2' },
  { value: '3', label: 'L1-L3' },
  { value: '4', label: 'L1-L4' },
];

export function DepthSelector({ value, onChange }: DepthSelectorProps) {
  return (
    <SegmentedControl
      data={DEPTH_OPTIONS}
      value={String(value)}
      onChange={(v) => onChange(Number(v) as DepthLevel)}
      size="sm"
    />
  );
}
