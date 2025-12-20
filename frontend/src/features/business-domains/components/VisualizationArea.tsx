import { Box, Group, Stack, Title, Text } from '@mantine/core';
import type { BusinessDomain, Capability, CapabilityId, CapabilityRealization, ComponentId } from '../../../api/types';
import { NestedCapabilityGrid } from './NestedCapabilityGrid';
import { DepthSelector, type DepthLevel } from './DepthSelector';
import { ShowApplicationsToggle } from './ShowApplicationsToggle';
import { useResponsive } from '../../../hooks/useResponsive';

interface VisualizationAreaProps {
  visualizedDomain: BusinessDomain | null;
  capabilities: Capability[];
  capabilitiesLoading: boolean;
  depth: DepthLevel;
  positions: Record<CapabilityId, { x: number; y: number }>;
  onDepthChange: (depth: DepthLevel) => void;
  onCapabilityClick: (capability: Capability, event: React.MouseEvent) => void;
  onContextMenu: (capability: Capability, event: React.MouseEvent) => void;
  selectedCapabilities: Set<CapabilityId>;
  showApplications: boolean;
  onShowApplicationsChange: (value: boolean) => void;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
  onApplicationClick: (componentId: ComponentId) => void;
  isDragOver?: boolean;
  onDragOver?: (e: React.DragEvent) => void;
  onDragLeave?: () => void;
  onDrop?: (e: React.DragEvent) => void;
}

export function VisualizationArea({
  visualizedDomain,
  capabilities,
  capabilitiesLoading,
  depth,
  positions,
  onDepthChange,
  onCapabilityClick,
  onContextMenu,
  selectedCapabilities,
  showApplications,
  onShowApplicationsChange,
  getRealizationsForCapability,
  onApplicationClick,
  isDragOver = false,
  onDragOver,
  onDragLeave,
  onDrop,
}: VisualizationAreaProps) {
  const { isMobile } = useResponsive();

  if (!visualizedDomain) {
    return (
      <Box component="main" className="business-domains-main" style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
        <Stack align="center" mt="xl">
          <Title order={2}>Grid Visualization</Title>
          <Text c="dimmed">
            Click a domain to see its capabilities
          </Text>
        </Stack>
      </Box>
    );
  }

  if (capabilitiesLoading) {
    return (
      <Box component="main" className="business-domains-main" style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
        <div className="loading-message">Loading capabilities...</div>
      </Box>
    );
  }

  return (
    <Box component="main" className="business-domains-main" style={{ flex: 1, padding: isMobile ? '0.5rem' : '1rem', overflow: 'auto' }}>
      <Stack gap="md">
        <Group justify="space-between" align="center" wrap="wrap" gap="sm">
          <Title order={2} size={isMobile ? 'h3' : 'h2'}>{visualizedDomain.name}</Title>
          <Group gap={isMobile ? 'xs' : 'md'} wrap="wrap">
            <ShowApplicationsToggle
              showApplications={showApplications}
              onShowApplicationsChange={onShowApplicationsChange}
            />
            <DepthSelector value={depth} onChange={onDepthChange} />
          </Group>
        </Group>
        <NestedCapabilityGrid
          capabilities={capabilities}
          depth={depth}
          onCapabilityClick={onCapabilityClick}
          onContextMenu={onContextMenu}
          selectedCapabilities={selectedCapabilities}
          positions={positions}
          showApplications={showApplications}
          getRealizationsForCapability={getRealizationsForCapability}
          onApplicationClick={onApplicationClick}
          isDragOver={isDragOver}
          onDragOver={onDragOver}
          onDragLeave={onDragLeave}
          onDrop={onDrop}
        />
      </Stack>
    </Box>
  );
}
