import { type ReactNode, useEffect, useLayoutEffect, useRef, useState } from 'react';
import {
  AnchorIcon,
  ContextMenu,
  type ContextMenuItem,
  ExpandIcon,
  GitBranchIcon,
  GitMergeIcon,
  ZapIcon,
} from '../../../components/shared/ContextMenu';
import type { DynamicFilters, EdgeType, UnexpandedByEdgeType } from '../utils/dynamicMode';

interface ExpandPopoverProps {
  entityName: string;
  breakdown: UnexpandedByEdgeType;
  enabledEdgeTypes: DynamicFilters['edges'];
  opened: boolean;
  onClose: () => void;
  onExpandEdgeType: (edge: EdgeType) => void;
  onExpandAll: () => void;
  children: ReactNode;
}

const EDGE_LABEL: Record<EdgeType, string> = {
  relation: 'Triggers / Serves',
  realization: 'Realization',
  parentage: 'Capability parentage',
  origin: 'Origin',
};

const EDGE_ICON: Record<EdgeType, ReactNode> = {
  relation: <ZapIcon />,
  realization: <GitMergeIcon />,
  parentage: <GitBranchIcon />,
  origin: <AnchorIcon />,
};

const EDGE_ORDER: EdgeType[] = ['relation', 'realization', 'parentage', 'origin'];

interface MenuPosition {
  x: number;
  y: number;
}

function buildEdgeItems(
  breakdown: UnexpandedByEdgeType,
  enabled: DynamicFilters['edges'],
  onExpandEdgeType: (edge: EdgeType) => void,
): ContextMenuItem[] {
  return EDGE_ORDER.filter((et) => enabled[et]).map((et) => {
    const count = breakdown[et].length;
    return {
      label: EDGE_LABEL[et],
      description: `+${count}`,
      icon: EDGE_ICON[et],
      disabled: count === 0,
      ariaLabel: `${EDGE_LABEL[et]} +${count}`,
      onClick: () => onExpandEdgeType(et),
    };
  });
}

function totalEnabled(breakdown: UnexpandedByEdgeType, enabled: DynamicFilters['edges']): number {
  return EDGE_ORDER.reduce((acc, et) => (enabled[et] ? acc + breakdown[et].length : acc), 0);
}

function useTriggerCenter(triggerRef: React.RefObject<HTMLElement | null>, opened: boolean): MenuPosition | null {
  const [position, setPosition] = useState<MenuPosition | null>(null);

  useLayoutEffect(() => {
    if (!opened || !triggerRef.current) {
      setPosition(null);
      return;
    }
    const rect = triggerRef.current.getBoundingClientRect();
    setPosition({ x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 });
  }, [opened, triggerRef]);

  useEffect(() => {
    if (!opened) return;
    const onScrollOrResize = () => {
      const el = triggerRef.current;
      if (!el) return;
      const rect = el.getBoundingClientRect();
      setPosition({ x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 });
    };
    window.addEventListener('scroll', onScrollOrResize, true);
    window.addEventListener('resize', onScrollOrResize);
    return () => {
      window.removeEventListener('scroll', onScrollOrResize, true);
      window.removeEventListener('resize', onScrollOrResize);
    };
  }, [opened, triggerRef]);

  return position;
}

export function ExpandPopover({
  entityName,
  breakdown,
  enabledEdgeTypes,
  opened,
  onClose,
  onExpandEdgeType,
  onExpandAll,
  children,
}: ExpandPopoverProps) {
  const triggerRef = useRef<HTMLSpanElement>(null);
  const position = useTriggerCenter(triggerRef, opened);

  const total = totalEnabled(breakdown, enabledEdgeTypes);
  const items: ContextMenuItem[] = buildEdgeItems(breakdown, enabledEdgeTypes, onExpandEdgeType);

  if (total > 0) {
    items.push({
      label: 'Expand all',
      description: `+${total}`,
      icon: <ExpandIcon />,
      ariaLabel: `Expand all +${total}`,
      onClick: onExpandAll,
    });
  }

  return (
    <>
      <span ref={triggerRef} className="ctx-menu-trigger">
        {children}
      </span>
      {opened && position && (
        <ContextMenu
          x={position.x}
          y={position.y}
          items={items}
          title={`Expand from ${entityName}`}
          onClose={onClose}
        />
      )}
    </>
  );
}
