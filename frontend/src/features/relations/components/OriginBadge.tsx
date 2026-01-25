import React from 'react';
import { DetailField } from '../../../components/shared/DetailField';

interface OriginBadgeProps {
  origin: string;
  isInherited: boolean;
}

export const OriginBadge: React.FC<OriginBadgeProps> = ({ origin, isInherited }) => (
  <DetailField label="Origin">
    <span className={`origin-badge ${isInherited ? 'origin-inherited' : 'origin-direct'}`}>
      {origin}
    </span>
  </DetailField>
);
