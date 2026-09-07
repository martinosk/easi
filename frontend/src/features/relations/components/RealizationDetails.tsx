import type React from 'react';
import { useRealizationDetails } from '../hooks/useRealizationDetails';
import { RealizationDetailsContent } from './RealizationDetailsContent';

export const RealizationDetails: React.FC = () => {
  const data = useRealizationDetails();
  if (!data) return null;

  return <RealizationDetailsContent data={data} />;
};
