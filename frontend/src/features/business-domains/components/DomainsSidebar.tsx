import { Box, Button, Stack } from '@mantine/core';
import type { BusinessDomain, BusinessDomainId } from '../../../api/types';
import { DomainList } from './DomainList';

interface DomainsSidebarProps {
  domains: BusinessDomain[];
  canCreateDomain: boolean;
  selectedDomainId: BusinessDomainId | undefined;
  onCreateClick: () => void;
  onVisualize: (domain: BusinessDomain) => void;
  onContextMenu: (e: React.MouseEvent, domain: BusinessDomain) => void;
}

export function DomainsSidebar({
  domains,
  canCreateDomain,
  selectedDomainId,
  onCreateClick,
  onVisualize,
  onContextMenu,
}: DomainsSidebarProps) {
  return (
    <Stack gap="md" p="md" h="100%" style={{ overflow: 'hidden' }}>
      {canCreateDomain && (
        <Button onClick={onCreateClick} data-testid="create-domain-button">
          Create Domain
        </Button>
      )}
      <Box flex={1} style={{ overflow: 'auto' }}>
        <DomainList
          domains={domains}
          onVisualize={onVisualize}
          onContextMenu={onContextMenu}
          selectedDomainId={selectedDomainId}
        />
      </Box>
    </Stack>
  );
}
