import { Badge } from '@mantine/core';
import React from 'react';
import { DetailField } from '../../../components/shared/DetailField';

interface OriginBadgeProps {
  origin: string;
  isInherited: boolean;
}

export const OriginBadge: React.FC<OriginBadgeProps> = ({ origin, isInherited }) => (
  <DetailField label="Origin">
    <Badge color={isInherited ? 'gray' : 'blue'} variant="light" size="sm">
      {origin}
    </Badge>
  </DetailField>
);
