import { Alert } from '@mantine/core';
import React from 'react';

interface InheritedRealizationInfoProps {
  isInherited: boolean;
}

export const InheritedRealizationInfo: React.FC<InheritedRealizationInfoProps> = ({ isInherited }) => {
  if (!isInherited) return null;

  return (
    <Alert color="gray" variant="light" radius="md">
      This is an inherited realization. It was automatically created when an application was linked to a child
      capability. To edit or delete, modify the original direct realization.
    </Alert>
  );
};
