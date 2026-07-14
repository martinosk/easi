import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import type { RealizationRoleAssignment } from '../../architecture-direction/types';
import { RealizationRoleControl, type RealizationRoleControlProps } from './RealizationRoleControl';

const mutateAssignAsync = vi.fn();
const mutateClearAsync = vi.fn();

vi.mock('../../architecture-direction/hooks/useRealizationRoles', () => ({
  useAssignRealizationRole: () => ({ mutateAsync: mutateAssignAsync, isPending: false }),
  useClearRealizationRole: () => ({ mutateAsync: mutateClearAsync, isPending: false }),
}));

function buildRole(overrides: Partial<RealizationRoleAssignment> = {}): RealizationRoleAssignment {
  return {
    capabilityId: 'cap-1',
    capabilityName: 'Booking management',
    componentId: 'comp-1',
    componentName: 'Phoenix',
    role: 'standard',
    assignedBy: 'user-1',
    assignedAt: '2026-02-01T00:00:00Z',
    _links: {
      self: { href: '', method: 'GET' },
      edit: { href: '', method: 'PUT' },
      delete: { href: '', method: 'DELETE' },
    },
    ...overrides,
  };
}

function renderControl(overrides: Partial<RealizationRoleControlProps> = {}) {
  const props: RealizationRoleControlProps = {
    capabilityId: 'cap-1',
    componentId: 'comp-1',
    role: undefined,
    canAssign: false,
    ...overrides,
  };
  renderWithProviders(<RealizationRoleControl {...props} />, { withRouter: false });
}

describe('RealizationRoleControl', () => {
  beforeEach(() => {
    mutateAssignAsync.mockReset().mockResolvedValue(undefined);
    mutateClearAsync.mockReset().mockResolvedValue(undefined);
  });

  it('renders nothing for an unclassified realization when the caller cannot assign', () => {
    renderControl();

    expect(screen.queryByTestId('role-comp-1')).not.toBeInTheDocument();
  });

  it('shows assign controls for an unclassified realization when the caller can assign', () => {
    renderControl({ canAssign: true });

    expect(screen.getByTestId('assign-standard-btn-comp-1')).toBeInTheDocument();
    expect(screen.getByTestId('assign-legacy-btn-comp-1')).toBeInTheDocument();
  });

  it('shows the current role badge for a classified realization even when the caller cannot write', () => {
    renderControl({ role: buildRole({ role: 'legacy', _links: { self: { href: '', method: 'GET' } } }) });

    expect(screen.getByTestId('role-badge-comp-1')).toHaveTextContent('legacy');
    expect(screen.queryByTestId('assign-standard-btn-comp-1')).not.toBeInTheDocument();
    expect(screen.queryByTestId('clear-role-btn-comp-1')).not.toBeInTheDocument();
  });

  it('shows switch and clear controls for a classified realization when write links are present', () => {
    renderControl({ role: buildRole({ role: 'standard' }) });

    expect(screen.getByTestId('role-badge-comp-1')).toHaveTextContent('standard');
    expect(screen.getByTestId('assign-legacy-btn-comp-1')).toBeInTheDocument();
    expect(screen.getByTestId('clear-role-btn-comp-1')).toBeInTheDocument();
  });

  it('assigns standard when the assign-standard control is clicked', async () => {
    renderControl({ canAssign: true });

    await userEvent.click(screen.getByTestId('assign-standard-btn-comp-1'));

    expect(mutateAssignAsync).toHaveBeenCalledWith({
      capabilityId: 'cap-1',
      componentId: 'comp-1',
      request: { role: 'standard' },
    });
  });

  it('switches a standard realization to legacy when clicked', async () => {
    renderControl({ role: buildRole({ role: 'standard' }) });

    await userEvent.click(screen.getByTestId('assign-legacy-btn-comp-1'));

    expect(mutateAssignAsync).toHaveBeenCalledWith({
      capabilityId: 'cap-1',
      componentId: 'comp-1',
      request: { role: 'legacy' },
    });
  });

  it('clears the role when the clear control is clicked', async () => {
    const role = buildRole({ role: 'legacy' });
    renderControl({ role });

    await userEvent.click(screen.getByTestId('clear-role-btn-comp-1'));

    expect(mutateClearAsync).toHaveBeenCalledWith({ role });
  });
});
