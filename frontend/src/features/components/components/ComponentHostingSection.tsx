import { Badge, Select } from '@mantine/core';
import type React from 'react';
import type { Component, HostingClassification } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { hasLink } from '../../../utils/hateoas';
import { useClassifyComponentHosting } from '../hooks/useComponentHosting';

export const HOSTING_CLASSIFICATION_LABELS: Record<HostingClassification, string> = {
  'on-premises': 'On-premises',
  cloud: 'Cloud',
  saas: 'SaaS',
  'third-party-hosted': 'Third-party hosted',
  unknown: 'Unknown',
};

const HOSTING_COLORS: Record<HostingClassification, string> = {
  'on-premises': 'indigo',
  cloud: 'cyan',
  saas: 'grape',
  'third-party-hosted': 'orange',
  unknown: 'gray',
};

const HOSTING_OPTIONS = (Object.keys(HOSTING_CLASSIFICATION_LABELS) as HostingClassification[]).map((value) => ({
  value,
  label: HOSTING_CLASSIFICATION_LABELS[value],
}));

interface ComponentHostingSectionProps {
  component: Component;
}

export const ComponentHostingSection: React.FC<ComponentHostingSectionProps> = ({ component }) => {
  const classifyMutation = useClassifyComponentHosting();

  return (
    <DetailField label="Hosting">
      {hasLink(component, 'x-classify-hosting') ? (
        <Select
          size="xs"
          data={HOSTING_OPTIONS}
          value={component.hosting}
          allowDeselect={false}
          disabled={classifyMutation.isPending}
          onChange={(value) => {
            if (value && value !== component.hosting) {
              classifyMutation.mutate({ component, hosting: value as HostingClassification });
            }
          }}
          data-testid="hosting-select"
        />
      ) : (
        <Badge color={HOSTING_COLORS[component.hosting]} variant="light" size="sm" data-testid="hosting-badge">
          {HOSTING_CLASSIFICATION_LABELS[component.hosting]}
        </Badge>
      )}
    </DetailField>
  );
};
