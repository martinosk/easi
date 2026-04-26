import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
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
    <MantineTestWrapper>
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
      </ExpandPopover>
    </MantineTestWrapper>,
  );
}

describe('ExpandPopover', () => {
  it('shows the entity name in the header when opened', () => {
    renderPopover();
    expect(screen.getByText(/Expand from Order Service/i)).toBeInTheDocument();
  });

  it('renders one row per enabled edge type with its unexpanded count', () => {
    renderPopover({
      breakdown: {
        relation: ['B', 'C', 'D'],
        realization: ['cap-1'],
        parentage: [],
        origin: ['vendor-1', 'vendor-2'],
      },
    });

    expect(screen.getByText(/Triggers \/ Serves/i)).toBeInTheDocument();
    expect(screen.getByText('+3')).toBeInTheDocument();
    expect(screen.getByText(/Realization/i)).toBeInTheDocument();
    expect(screen.getByText('+1')).toBeInTheDocument();
    expect(screen.getByText(/Origin/i)).toBeInTheDocument();
    expect(screen.getByText('+2')).toBeInTheDocument();
    expect(screen.getByText(/Capability parentage/i)).toBeInTheDocument();
    expect(screen.getByText('+0')).toBeInTheDocument();
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

    expect(screen.getByText(/Triggers \/ Serves/i)).toBeInTheDocument();
    expect(screen.queryByText(/Realization/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/Origin/i)).not.toBeInTheDocument();
  });

  it('clicking an edge-type row calls onExpandEdgeType with that type', async () => {
    const onExpandEdgeType = vi.fn();
    renderPopover({
      breakdown: { ...emptyBreakdown, relation: ['B', 'C'] },
      onExpandEdgeType,
    });

    await userEvent.click(screen.getByRole('button', { name: /triggers \/ serves/i }));
    expect(onExpandEdgeType).toHaveBeenCalledWith('relation');
  });

  it('rows with zero unexpanded neighbors are not clickable', async () => {
    const onExpandEdgeType = vi.fn();
    renderPopover({ breakdown: emptyBreakdown, onExpandEdgeType });

    const row = screen.getByRole('button', { name: /triggers \/ serves/i });
    expect(row).toBeDisabled();
  });

  it('renders an Expand all action when total > 0', async () => {
    const onExpandAll = vi.fn();
    renderPopover({
      breakdown: { ...emptyBreakdown, relation: ['B'], realization: ['cap-1'] },
      onExpandAll,
    });

    const expandAll = screen.getByRole('button', { name: /expand all/i });
    expect(expandAll).toHaveTextContent('+2');
    await userEvent.click(expandAll);
    expect(onExpandAll).toHaveBeenCalled();
  });

  it('hides the Expand all action when total = 0', () => {
    renderPopover({ breakdown: emptyBreakdown });
    expect(screen.queryByRole('button', { name: /expand all/i })).not.toBeInTheDocument();
  });
});
