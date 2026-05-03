import type { ReactNode } from 'react';

export interface ContextMenuItem {
  label: string;
  onClick: () => void;
  icon?: ReactNode;
  description?: string;
  isDanger?: boolean;
  disabled?: boolean;
  ariaLabel?: string;
}

export type ContextMenuVariant = 'auto' | 'radial' | 'linear';
