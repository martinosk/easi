import { useReactFlow } from '@xyflow/react';
import React, { useCallback, useState } from 'react';
import { getEntityId, toNodeId } from '../../../constants/entityIdentifiers';
import type { Position } from '../../../api/types';
import type { RelatedLink } from '../../../utils/xRelated';
import { CreateCapabilityDialog, type CapabilityLevel } from '../../capabilities/components/CreateCapabilityDialog';
import { CreateComponentDialog } from '../../components/components/CreateComponentDialog';
import { CreateAcquiredEntityDialog } from '../../origin-entities/components/CreateAcquiredEntityDialog';
import { CreateInternalTeamDialog } from '../../origin-entities/components/CreateInternalTeamDialog';
import { CreateVendorDialog } from '../../origin-entities/components/CreateVendorDialog';
import { useCreateRelatedEntity } from '../hooks/useCreateRelatedEntity';
import { useHandleClickDetection } from '../hooks/useHandleClickDetection';
import { useSourceEntityRelated } from '../hooks/useSourceEntityRelated';
import type { HandleSide } from '../utils/handleClick';
import { HandleCreatePicker } from './HandleCreatePicker';

interface PickerState {
  sourceNodeId: string;
  side: HandleSide;
  x: number;
  y: number;
}

const CAPABILITY_LEVELS: CapabilityLevel[] = ['L1', 'L2', 'L3', 'L4'];

function nextLevel(current: string | undefined): CapabilityLevel | undefined {
  if (!current) return undefined;
  const idx = CAPABILITY_LEVELS.indexOf(current as CapabilityLevel);
  if (idx < 0 || idx >= CAPABILITY_LEVELS.length - 1) return undefined;
  return CAPABILITY_LEVELS[idx + 1];
}

export const HandleCreateController: React.FC = () => {
  const flow = useReactFlow();
  const [picker, setPicker] = useState<PickerState | null>(null);
  const orchestrator = useCreateRelatedEntity();

  useHandleClickDetection(null, ({ nodeId, side, clientX, clientY }) => {
    setPicker({ sourceNodeId: nodeId, side, x: clientX, y: clientY });
  });

  const entries = useSourceEntityRelated(picker?.sourceNodeId ?? null);

  const lookupSourcePosition = useCallback(
    (nodeId: string): Position => {
      const node = flow.getNode(nodeId);
      return node?.position ?? { x: 0, y: 0 };
    },
    [flow],
  );

  const handleSelect = (entry: RelatedLink) => {
    if (!picker) return;
    const sourcePosition = lookupSourcePosition(picker.sourceNodeId);
    const sourceEntityId = getEntityId(toNodeId(picker.sourceNodeId));
    const prefill = buildPrefill(entry, picker.sourceNodeId, flow);
    orchestrator.start({
      entry,
      sourceEntityId,
      side: picker.side,
      sourcePosition,
      prefill,
    });
    setPicker(null);
  };

  const closePicker = () => setPicker(null);

  return (
    <>
      {picker && entries.length > 0 && (
        <HandleCreatePicker
          x={picker.x}
          y={picker.y}
          entries={entries}
          onSelect={handleSelect}
          onClose={closePicker}
        />
      )}
      <PendingDialogs orchestrator={orchestrator} />
    </>
  );
};

function buildPrefill(
  entry: RelatedLink,
  sourceNodeId: string,
  flow: ReturnType<typeof useReactFlow>,
): { capabilityLevel?: CapabilityLevel } | undefined {
  if (entry.relationType !== 'capability-parent') return undefined;
  const node = flow.getNode(sourceNodeId);
  const sourceLevel = (node?.data as { level?: string } | undefined)?.level;
  const childLevel = nextLevel(sourceLevel);
  return childLevel ? { capabilityLevel: childLevel } : undefined;
}

interface PendingDialogsProps {
  orchestrator: ReturnType<typeof useCreateRelatedEntity>;
}

const PendingDialogs: React.FC<PendingDialogsProps> = ({ orchestrator }) => {
  const { pending, cancel, handleEntityCreated } = orchestrator;
  const targetType = pending?.entry.targetType ?? null;

  const onClose = () => cancel();
  const onCreated = (entity: { id: string }) => handleEntityCreated(entity.id);

  return (
    <>
      <CreateComponentDialog
        isOpen={targetType === 'component'}
        onClose={onClose}
        onCreated={targetType === 'component' ? onCreated : undefined}
      />
      <CreateCapabilityDialog
        isOpen={targetType === 'capability'}
        onClose={onClose}
        onCreated={targetType === 'capability' ? onCreated : undefined}
        prefill={
          targetType === 'capability' && pending?.prefill?.capabilityLevel
            ? { level: pending.prefill.capabilityLevel }
            : undefined
        }
      />
      <CreateAcquiredEntityDialog
        isOpen={targetType === 'acquiredEntity'}
        onClose={onClose}
        onCreated={targetType === 'acquiredEntity' ? onCreated : undefined}
      />
      <CreateVendorDialog
        isOpen={targetType === 'vendor'}
        onClose={onClose}
        onCreated={targetType === 'vendor' ? onCreated : undefined}
      />
      <CreateInternalTeamDialog
        isOpen={targetType === 'internalTeam'}
        onClose={onClose}
        onCreated={targetType === 'internalTeam' ? onCreated : undefined}
      />
    </>
  );
};
