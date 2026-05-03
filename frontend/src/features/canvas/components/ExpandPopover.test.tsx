import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { DynamicFilters, UnexpandedByEdgeType } from '../utils/dynamicMode';
import { ExpandPopover } from './ExpandPopover';

const allEnabled: DynamicFilters['edges'] = {
  relation: true,
  realization: true,
  parentage: true,
  origin: true,
};

const emptyBreakdown: UnexpandedByEdgeType = {
  relation: [],
  realization: [],
  parentage: [],
  origin: [],
};

function renderPopover(props: Partial<React.ComponentProps<typeof ExpandPopover>> = {}) {
  return render(
    <ExpandPopover
      entityName="Order Service"
      breakdown={emptyBreakdown}
      enabledEdgeTypes={allEnabled}
      opened
      onClose={vi.fn()}
      onExpandEdgeType={vi.fn()}
      onExpandAll={vi.fn()}
      {...props}
    >
      <button type="button">Trigger</button>
    </ExpandPopover>,
  );
}

describe('ExpandPopover', () => {
  it('shows the entity name as the menu title when opened', () => {
    renderPopover();
    expect(screen.getByText(/Expand from Order Service/i)).toBeTruthy();
  });

  it('renders one menuitem per enabled edge type, with the unexpanded count in its accessible name', () => {
    renderPopover({
      breakdown: {
        relation: ['B', 'C', 'D'],
        realization: ['cap-1'],
        parentage: [],
        origin: ['vendor-1', 'vendor-2'],
      },
    });

    expect(screen.getByRole('menuitem', { name: /Triggers \/ Serves \+3/i })).toBeTruthy();
    expect(screen.getByRole('menuitem', { name: /Realization \+1/i })).toBeTruthy();
    expect(screen.getByRole('menuitem', { name: /Origin \+2/i })).toBeTruthy();
    expect(screen.getByRole('menuitem', { name: /Capability parentage \+0/i })).toBeTruthy();
  });

  it('hides edge types disabled in filters', () => {
    renderPopover({
      enabledEdgeTypes: { ...allEnabled, realization: false, origin: false },
      breakdown: {
        relation: ['B'],
        realization: ['cap-1'],
        parentage: [],
        origin: ['vendor-1'],
      },
    });

    expect(screen.getByRole('menuitem', { name: /Triggers \/ Serves/i })).toBeTruthy();
    expect(screen.queryByRole('menuitem', { name: /^Realization/i })).toBeNull();
    expect(screen.queryByRole('menuitem', { name: /^Origin/i })).toBeNull();
  });

  it('clicking an edge-type petal calls onExpandEdgeType with that type', () => {
    const onExpandEdgeType = vi.fn();
    renderPopover({
      breakdown: { ...emptyBreakdown, relation: ['B', 'C'] },
      onExpandEdgeType,
    });

    fireEvent.click(screen.getByRole('menuitem', { name: /Triggers \/ Serves \+2/i }));
    expect(onExpandEdgeType).toHaveBeenCalledWith('relation');
  });

  it('petals with zero unexpanded neighbors are disabled and do not invoke the handler', () => {
    const onExpandEdgeType = vi.fn();
    renderPopover({ breakdown: emptyBreakdown, onExpandEdgeType });

    const petal = screen.getByRole('menuitem', { name: /Triggers \/ Serves \+0/i }) as HTMLButtonElement;
    expect(petal.disabled).toBe(true);
    fireEvent.click(petal);
    expect(onExpandEdgeType).not.toHaveBeenCalled();
  });

  it('renders an Expand all petal when total > 0', () => {
    const onExpandAll = vi.fn();
    renderPopover({
      breakdown: { ...emptyBreakdown, relation: ['B'], realization: ['cap-1'] },
      onExpandAll,
    });

    const expandAll = screen.getByRole('menuitem', { name: /Expand all \+2/i });
    fireEvent.click(expandAll);
    expect(onExpandAll).toHaveBeenCalled();
  });

  it('hides the Expand all petal when total = 0', () => {
    renderPopover({ breakdown: emptyBreakdown });
    expect(screen.queryByRole('menuitem', { name: /Expand all/i })).toBeNull();
  });

  it('does not render the menu when not opened', () => {
    renderPopover({ opened: false });
    expect(screen.queryByRole('menu')).toBeNull();
  });
});
