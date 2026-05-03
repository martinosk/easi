import React from 'react';
import { ContextMenu, type ContextMenuItem } from '../../../components/shared/ContextMenu';
import type { RelatedLink } from '../../../utils/xRelated';
import type { RelationSubType } from '../utils/relationDispatch';

export interface HandleCreatePickerSelection {
  entry: RelatedLink;
  relationSubType?: RelationSubType;
}

interface HandleCreatePickerProps {
  x: number;
  y: number;
  entries: RelatedLink[];
  onSelect: (selection: HandleCreatePickerSelection) => void;
  onClose: () => void;
}

const COMPONENT_RELATION_VARIANTS: ReadonlyArray<{ subType: RelationSubType; suffix: string }> = [
  { subType: 'Triggers', suffix: '(Triggers)' },
  { subType: 'Serves', suffix: '(Serves)' },
];

function variantsFor(entry: RelatedLink): HandleCreatePickerSelection[] {
  if (entry.relationType === 'component-relation') {
    return COMPONENT_RELATION_VARIANTS.map((v) => ({ entry: variantEntry(entry, v.suffix), relationSubType: v.subType }));
  }
  return [{ entry }];
}

function variantEntry(entry: RelatedLink, suffix: string): RelatedLink {
  return { ...entry, title: `${stripParenthetical(entry.title)} ${suffix}`.trim() };
}

function stripParenthetical(title: string): string {
  return title.replace(/\s*\([^)]*\)\s*$/, '').trim();
}

export const HandleCreatePicker: React.FC<HandleCreatePickerProps> = ({ x, y, entries, onSelect, onClose }) => {
  if (entries.length === 0) return null;

  const items: ContextMenuItem[] = entries.flatMap((entry) =>
    variantsFor(entry).map((selection) => ({
      label: selection.entry.title,
      onClick: () => onSelect(selection),
    })),
  );

  items.push({ label: 'Cancel', onClick: onClose, ariaLabel: 'Cancel' });

  return <ContextMenu x={x} y={y} items={items} onClose={onClose} />;
};
