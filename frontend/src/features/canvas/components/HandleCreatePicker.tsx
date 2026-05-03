import React from 'react';
import { ContextMenu, type ContextMenuItem } from '../../../components/shared/ContextMenu';
import type { RelatedLink } from '../../../utils/xRelated';

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

export const HandleCreatePicker: React.FC<HandleCreatePickerProps> = ({ x, y, entries, onSelect, onClose }) => {
  if (entries.length === 0) return null;

  const items: ContextMenuItem[] = entries.map((entry) => ({
    label: entry.title,
    onClick: () => onSelect({ entry }),
  }));

  items.push({ label: 'Cancel', onClick: onClose, ariaLabel: 'Cancel' });

  return <ContextMenu x={x} y={y} items={items} onClose={onClose} />;
};
