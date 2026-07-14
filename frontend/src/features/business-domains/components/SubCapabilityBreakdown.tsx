import { UnstyledButton } from '@mantine/core';
import { IconChevronRight } from '@tabler/icons-react';
import { useMemo, useState } from 'react';
import type { CapabilityId, CapabilityRealization } from '../../../api/types';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import type { CapabilityJourney } from '../../journeys/types';
import { buildSubCapabilityBreakdown, type SubCapabilityStatus } from '../lens/subCapabilityStatus';
import classes from './SubCapabilityBreakdown.module.css';

const DOT_CLASS: Record<SubCapabilityStatus, string> = {
  done: classes.dotDone,
  'in-flight': classes.dotFlight,
  'not-started': classes.dotIdle,
};

export interface SubCapabilityBreakdownProps {
  node: CapabilityTreeNode;
  journey: CapabilityJourney;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
}

export function SubCapabilityBreakdown({ node, journey, getRealizationsForCapability }: SubCapabilityBreakdownProps) {
  const [open, setOpen] = useState(false);
  const rows = useMemo(
    () => buildSubCapabilityBreakdown(node, journey, getRealizationsForCapability),
    [node, journey, getRealizationsForCapability],
  );

  if (rows.length === 0) return null;

  return (
    <>
      <UnstyledButton
        component="button"
        className={classes.expander}
        aria-expanded={open}
        onClick={(event) => {
          event.stopPropagation();
          setOpen((value) => !value);
        }}
        data-testid={`subcap-expander-${node.capability.id}`}
      >
        <IconChevronRight size={12} className={[classes.chevron, open ? classes.chevronOpen : ''].join(' ')} />
        {rows.length} sub-capabilit{rows.length === 1 ? 'y' : 'ies'}
      </UnstyledButton>
      {open && (
        <div className={classes.children}>
          {rows.map((row) => (
            <div key={row.capability.id} className={classes.child} data-testid={`subcap-${row.capability.id}`}>
              <span className={[classes.dot, DOT_CLASS[row.status]].join(' ')} data-status={row.status} />
              <span className={classes.name}>{row.capability.name}</span>
              <span className={classes.level}>{row.capability.level}</span>
              <span className={classes.app}>{row.appLabel}</span>
            </div>
          ))}
        </div>
      )}
    </>
  );
}
