import React from 'react';
import {
  BuildingIcon,
  CapabilityIcon,
  ComponentIcon,
  ContextMenu,
  type ContextMenuItem,
  PackageIcon,
  UsersIcon,
} from '../../../components/shared/ContextMenu';
import type { RelatedLink, RelatedTargetType } from '../../../utils/xRelated';

export interface HandleCreatePickerSelection {
  entry: RelatedLink;
}

interface HandleCreatePickerProps {
  x: number;
  y: number;
  entries: RelatedLink[];
  onSelect: (selection: HandleCreatePickerSelection) => void;
  onClose: () => void;
}

const TARGET_ICONS: Record<RelatedTargetType, React.ReactNode> = {
  component: <ComponentIcon />,
  capability: <CapabilityIcon />,
  acquiredEntity: <PackageIcon />,
  vendor: <BuildingIcon />,
  internalTeam: <UsersIcon />,
};

const TARGET_DESCRIPTIONS: Record<RelatedTargetType, string> = {
  component: 'Create a related component',
  capability: 'Create a related capability',
  acquiredEntity: 'Create a related acquired entity',
  vendor: 'Create a related vendor',
  internalTeam: 'Create a related internal team',
};

export const HandleCreatePicker: React.FC<HandleCreatePickerProps> = ({ x, y, entries, onSelect, onClose }) => {
  if (entries.length === 0) return null;

  const items: ContextMenuItem[] = entries.map((entry) => ({
    label: entry.title,
    description: TARGET_DESCRIPTIONS[entry.targetType],
    icon: TARGET_ICONS[entry.targetType],
    onClick: () => onSelect({ entry }),
  }));

  return <ContextMenu x={x} y={y} items={items} title="Create related" onClose={onClose} />;
};
