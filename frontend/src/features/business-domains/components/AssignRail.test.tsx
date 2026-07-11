import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { CapabilityId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { AssignRail } from './AssignRail';

describe('AssignRail', () => {
  const capabilities = [cap('l1-a', 'Alpha', 'L1'), cap('l1-b', 'Bravo', 'L1')];

  it('renders the shared capability explorer inside the rail container', () => {
    renderWithProviders(
      <AssignRail
        allCapabilities={capabilities}
        isLoading={false}
        globalAssignedCapabilityIds={new Set(['l1-a' as CapabilityId])}
        onDragStart={vi.fn()}
        onDragEnd={vi.fn()}
      />,
      { withRouter: false },
    );

    expect(screen.getByTestId('assign-rail')).toBeInTheDocument();
    expect(screen.getByTestId('assigned-indicator-l1-a')).toBeInTheDocument();
    expect(screen.queryByTestId('assigned-indicator-l1-b')).not.toBeInTheDocument();
  });

  it('forwards drag-start to the caller with the dragged capability', () => {
    const onDragStart = vi.fn();
    renderWithProviders(
      <AssignRail
        allCapabilities={capabilities}
        isLoading={false}
        globalAssignedCapabilityIds={new Set()}
        onDragStart={onDragStart}
        onDragEnd={vi.fn()}
      />,
      { withRouter: false },
    );

    fireEvent.dragStart(screen.getByTestId('draggable-l1-a'), {
      dataTransfer: { setData: vi.fn(), effectAllowed: '' },
    });

    expect(onDragStart).toHaveBeenCalledWith(capabilities[0]);
  });
});
