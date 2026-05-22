import { Badge, Box, Button, Code, Divider, Group, Stack, Text, Title } from '@mantine/core';
import React, { useMemo, useState } from 'react';
import toast from 'react-hot-toast';
import type {
  Capability,
  CapabilityId,
  CapabilityRealization,
  Component,
  StrategyImportance,
  View,
  ViewCapability,
} from '../../../api/types';
import { ColorPicker } from '../../../components/shared/ColorPicker';
import { DetailField } from '../../../components/shared/DetailField';
import { useMaturityScale } from '../../../hooks/useMaturityScale';
import { useAppStore } from '../../../store/appStore';
import { hasLink } from '../../../utils/hateoas';
import { deriveLegacyMaturityValue, getDefaultSections } from '../../../utils/maturity';
import { AuditHistorySection } from '../../audit';
import { useStrategyImportanceByCapability } from '../../business-domains/hooks/useStrategyImportance';
import { useComponents } from '../../components/hooks/useComponents';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useUpdateCapabilityColor } from '../../views/hooks/useViews';
import { useCapabilities, useCapabilityRealizations } from '../hooks/useCapabilities';
import { AddExpertDialog } from './AddExpertDialog';
import { CapabilityExpertsList } from './CapabilityExpertsList';
import { EditCapabilityDialog } from './EditCapabilityDialog';
import { RealizationFitContext } from './RealizationFitContext';

interface CapabilityDetailsProps {
  onRemoveFromView: () => void;
}

type MaturityColor = 'red' | 'orange' | 'green' | 'blue' | 'gray';

const getMaturityBadgeColor = (maturityLevel?: string): MaturityColor => {
  const level = maturityLevel?.toLowerCase();
  const colors: Record<string, MaturityColor> = {
    genesis: 'red',
    'custom build': 'orange',
    'custom built': 'orange',
    product: 'green',
    commodity: 'blue',
  };
  return colors[level || ''] || 'gray';
};

const getLevelBadge = (level: string): string => {
  const badges: Record<string, string> = {
    Full: '100%',
    Partial: 'Partial',
    Planned: 'Planned',
  };
  return badges[level] || level;
};

const getComponentName = (components: Component[], componentId: string): string => {
  const comp = components.find((c) => c.id === componentId);
  return comp?.name || 'Unknown';
};

const TagList: React.FC<{ tags: string[] }> = ({ tags }) => (
  <Group gap="xs">
    {tags.map((tag) => (
      <Badge key={tag} variant="light">
        {tag}
      </Badge>
    ))}
  </Group>
);

interface RealizingComponentsProps {
  realizations: CapabilityRealization[];
  components: Component[];
  capabilityId: CapabilityId;
  importanceRatings: StrategyImportance[];
}

const RealizingComponentsList: React.FC<RealizingComponentsProps> = ({
  realizations,
  components,
  capabilityId,
  importanceRatings,
}) => {
  const uniqueDomainIds = useMemo(() => {
    const ids = new Set(importanceRatings.map((r) => r.businessDomainId));
    return Array.from(ids);
  }, [importanceRatings]);

  return (
    <Stack gap={0}>
      {realizations.map((r, index) => (
        <React.Fragment key={r.id}>
          {index > 0 && <Divider />}
          <Stack gap="xs" py="xs">
            <Group justify="space-between" wrap="nowrap">
              <Text size="sm">{getComponentName(components, r.componentId)}</Text>
              <Badge color="green" variant="filled" size="sm">
                {getLevelBadge(r.realizationLevel)}
              </Badge>
            </Group>
            {uniqueDomainIds.map((domainId) => (
              <RealizationFitContext
                key={`${r.componentId}-${domainId}`}
                componentId={r.componentId}
                capabilityId={capabilityId}
                businessDomainId={domainId}
              />
            ))}
          </Stack>
        </React.Fragment>
      ))}
    </Stack>
  );
};

interface OptionalFieldProps<T> {
  value: T | undefined | null;
  label: string;
  render: (value: T) => React.ReactNode;
}

function OptionalField<T>({ value, label, render }: OptionalFieldProps<T>): React.ReactElement | null {
  if (value === undefined || value === null) return null;
  if (Array.isArray(value) && value.length === 0) return null;
  return <DetailField label={label}>{render(value)}</DetailField>;
}

interface CapabilityFieldsProps {
  capability: Capability;
}

const CapabilityOptionalFields: React.FC<CapabilityFieldsProps> = ({ capability }) => (
  <>
    <OptionalField value={capability.description} label="Description" render={(desc) => desc} />
    <OptionalField value={capability.status} label="Status" render={(status) => status} />
    <OptionalField value={capability.ownershipModel} label="Ownership Model" render={(model) => model} />
    <OptionalField value={capability.primaryOwner} label="Primary Owner" render={(owner) => owner} />
    <OptionalField value={capability.eaOwner} label="EA Owner" render={(owner) => owner} />
    <OptionalField value={capability.tags} label="Tags" render={(tags) => <TagList tags={tags} />} />
  </>
);

interface ColorPickerFieldProps {
  capabilityInView: ViewCapability;
  currentView: View;
  onColorChange: (color: string) => void;
}

const ColorPickerField: React.FC<ColorPickerFieldProps> = ({ capabilityInView, currentView, onColorChange }) => {
  const currentColor = capabilityInView.customColor || null;
  const isColorPickerEnabled = currentView.colorScheme === 'custom';

  return (
    <DetailField label="Custom Color">
      <Box data-testid="color-picker">
        <ColorPicker
          color={currentColor}
          onChange={onColorChange}
          disabled={!isColorPickerEnabled}
          disabledTooltip="Switch to custom color scheme to assign colors"
        />
      </Box>
    </DetailField>
  );
};

interface RealizationsFieldProps {
  realizations: CapabilityRealization[];
  components: Component[];
  capabilityId: CapabilityId;
  importanceRatings: StrategyImportance[];
}

const RealizationsField: React.FC<RealizationsFieldProps> = ({
  realizations,
  components,
  capabilityId,
  importanceRatings,
}) => {
  if (realizations.length === 0) return null;
  return (
    <DetailField label="Realized By">
      <RealizingComponentsList
        realizations={realizations}
        components={components}
        capabilityId={capabilityId}
        importanceRatings={importanceRatings}
      />
    </DetailField>
  );
};

interface MaturityDisplay {
  sectionName: string;
  text: string;
}

const DEFAULT_LEGACY_MATURITY = 12;

function computeMaturityDisplay(
  capability: Capability,
  sections: ReturnType<typeof getDefaultSections>,
): MaturityDisplay {
  const value =
    capability.maturityValue ??
    (capability.maturityLevel
      ? deriveLegacyMaturityValue(capability.maturityLevel, sections)
      : DEFAULT_LEGACY_MATURITY);
  const sectionName = sections.find((s) => value >= s.minValue && value <= s.maxValue)?.name ?? 'Unknown';
  return { sectionName, text: `${sectionName} (${value})` };
}

interface CapabilityActionsBarProps {
  canEdit: boolean;
  canRemoveFromView: boolean;
  onEdit: () => void;
  onRemoveFromView: () => void;
}

const CapabilityActionsBar: React.FC<CapabilityActionsBarProps> = ({
  canEdit,
  canRemoveFromView,
  onEdit,
  onRemoveFromView,
}) => {
  if (!canEdit && !canRemoveFromView) return null;
  return (
    <Group gap="sm">
      {canEdit && (
        <Button variant="default" size="xs" onClick={onEdit}>
          Edit
        </Button>
      )}
      {canRemoveFromView && (
        <Button variant="default" size="xs" onClick={onRemoveFromView}>
          Remove from View
        </Button>
      )}
    </Group>
  );
};

interface CapabilityContentProps {
  capability: Capability;
  capabilityInView: ViewCapability | undefined;
  currentView: View | null;
  realizations: CapabilityRealization[];
  components: Component[];
  importanceRatings: StrategyImportance[];
  onColorChange: (color: string) => void;
  onEdit: () => void;
  onRemoveFromView: () => void;
  onAddExpert: () => void;
}

const CapabilityContent: React.FC<CapabilityContentProps> = ({
  capability,
  capabilityInView,
  currentView,
  realizations,
  components,
  importanceRatings,
  onColorChange,
  onEdit,
  onRemoveFromView,
  onAddExpert,
}) => {
  const { data: maturityScale } = useMaturityScale();
  const sections = maturityScale?.sections ?? getDefaultSections();
  const maturity = computeMaturityDisplay(capability, sections);
  const formattedDate = new Date(capability.createdAt).toLocaleString();

  const showColorPicker = capabilityInView !== undefined && currentView !== null;

  return (
    <Stack gap="sm">
      <CapabilityActionsBar
        canEdit={capability._links?.edit !== undefined}
        canRemoveFromView={capabilityInView?._links?.['x-remove'] !== undefined}
        onEdit={onEdit}
        onRemoveFromView={onRemoveFromView}
      />

      <DetailField label="Name">{capability.name}</DetailField>
      <DetailField label="Level">
        <Badge color="dark" variant="filled" size="sm">
          {capability.level}
        </Badge>
      </DetailField>
      <CapabilityOptionalFields capability={capability} />
      <DetailField label="Maturity Level">
        <Badge color={getMaturityBadgeColor(maturity.sectionName)} variant="filled" size="md">
          {maturity.text}
        </Badge>
      </DetailField>
      <DetailField label="Created">
        <Text size="sm" c="dimmed">
          {formattedDate}
        </Text>
      </DetailField>
      <DetailField label="ID">
        <Code>{capability.id}</Code>
      </DetailField>

      <CapabilityExpertsList
        capabilityId={capability.id}
        experts={capability.experts}
        canAddExpert={hasLink(capability, 'x-add-expert')}
        onAddClick={onAddExpert}
      />

      {showColorPicker && (
        <ColorPickerField capabilityInView={capabilityInView} currentView={currentView} onColorChange={onColorChange} />
      )}

      <RealizationsField
        realizations={realizations}
        components={components}
        capabilityId={capability.id}
        importanceRatings={importanceRatings}
      />

      <AuditHistorySection aggregateId={capability.id} />
    </Stack>
  );
};

export const CapabilityDetails: React.FC<CapabilityDetailsProps> = ({ onRemoveFromView }) => {
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const { data: capabilities = [] } = useCapabilities();
  const { data: components = [] } = useComponents();
  const { currentView } = useCurrentView();
  const { data: capabilityRealizationsForThis = [] } = useCapabilityRealizations(selectedCapabilityId ?? undefined);
  const { data: importanceRatings = [] } = useStrategyImportanceByCapability(selectedCapabilityId ?? undefined);
  const updateCapabilityColorMutation = useUpdateCapabilityColor();
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showAddExpertDialog, setShowAddExpertDialog] = useState(false);

  const capability = capabilities.find((c) => c.id === selectedCapabilityId);
  if (!selectedCapabilityId || !capability) return null;

  const capabilityInView = currentView?.capabilities.find((vc) => vc.capabilityId === selectedCapabilityId);

  const handleColorChange = async (color: string) => {
    if (!currentView || !selectedCapabilityId) return;

    try {
      await updateCapabilityColorMutation.mutateAsync({
        viewId: currentView.id,
        capabilityId: selectedCapabilityId,
        color,
      });
    } catch {
      toast.error('Failed to update color');
    }
  };

  return (
    <Stack gap="sm" p="md">
      <Title order={4}>Capability Details</Title>
      <Divider />

      <CapabilityContent
        capability={capability}
        capabilityInView={capabilityInView}
        currentView={currentView}
        realizations={capabilityRealizationsForThis}
        components={components}
        importanceRatings={importanceRatings}
        onColorChange={handleColorChange}
        onEdit={() => setShowEditDialog(true)}
        onRemoveFromView={onRemoveFromView}
        onAddExpert={() => setShowAddExpertDialog(true)}
      />

      <EditCapabilityDialog isOpen={showEditDialog} onClose={() => setShowEditDialog(false)} capability={capability} />
      <AddExpertDialog
        isOpen={showAddExpertDialog}
        onClose={() => setShowAddExpertDialog(false)}
        capabilityId={capability.id}
      />
    </Stack>
  );
};
