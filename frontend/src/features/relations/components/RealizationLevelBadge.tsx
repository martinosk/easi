import { Badge } from '@mantine/core';
import React from 'react';
import { DetailField } from '../../../components/shared/DetailField';

const LEVEL_DISPLAY_MAP: Record<string, string> = {
  Full: 'Full (100%)',
  Partial: 'Partial',
  Planned: 'Planned',
};

const getLevelDisplay = (level: string): string => LEVEL_DISPLAY_MAP[level] ?? level;

interface RealizationLevelBadgeProps {
  level: string;
}

export const RealizationLevelBadge: React.FC<RealizationLevelBadgeProps> = ({ level }) => (
  <DetailField label="Realization Level">
    <Badge color="gray" variant="filled" size="sm">
      {getLevelDisplay(level)}
    </Badge>
  </DetailField>
);
