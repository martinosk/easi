import { Stack, Text } from '@mantine/core';
import type { BusinessDomain, BusinessDomainId } from '../../../api/types';
import { DomainCard } from './DomainCard';

interface DomainListProps {
  domains: BusinessDomain[];
  onVisualize: (domain: BusinessDomain) => void;
  onContextMenu: (e: React.MouseEvent, domain: BusinessDomain) => void;
  selectedDomainId?: BusinessDomainId | null;
}

export function DomainList({ domains, onVisualize, onContextMenu, selectedDomainId }: DomainListProps) {
  if (domains.length === 0) {
    return (
      <Stack gap="xs" align="center" data-testid="domains-empty-state">
        <Text c="dimmed">No business domains yet.</Text>
        <Text c="dimmed">Create your first domain to get started.</Text>
      </Stack>
    );
  }

  return (
    <Stack gap="xs" data-testid="domain-list">
      {domains.map((domain) => (
        <DomainCard
          key={domain.id}
          domain={domain}
          onVisualize={onVisualize}
          onContextMenu={onContextMenu}
          isSelected={domain.id === selectedDomainId}
        />
      ))}
    </Stack>
  );
}
