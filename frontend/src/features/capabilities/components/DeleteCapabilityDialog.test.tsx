import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { CapabilityDeleteImpact, ImpactRealization } from '../../../api/types';
import { toCapabilityId, toComponentId } from '../../../api/types';
import { buildCapability } from '../../../test/helpers/entityBuilders';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import type { DeleteImpacts } from '../hooks/useCapabilities';
import { DeleteCapabilityDialog } from './DeleteCapabilityDialog';

vi.mock('../hooks/useCapabilities', () => ({
  useDeleteCapability: vi.fn(),
  useDeleteImpacts: vi.fn(),
  useCascadeDeleteCapability: vi.fn(),
}));

import { useCascadeDeleteCapability, useDeleteCapability, useDeleteImpacts } from '../hooks/useCapabilities';

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
}

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <MantineTestWrapper>{ui}</MantineTestWrapper>
    </QueryClientProvider>,
  );
}

const realization = (id: string, componentName: string, capabilityId: string): ImpactRealization => ({
  id,
  componentId: toComponentId(`comp-${id}`),
  componentName,
  capabilityId: toCapabilityId(capabilityId),
  realizationLevel: 'Full',
  origin: 'Direct',
});

const leafImpact: CapabilityDeleteImpact = {
  capabilityId: toCapabilityId('cap-1'),
  capabilityName: 'Test Capability',
  hasDescendants: false,
  affectedCapabilities: [],
  realizations: [],
  _links: {},
};

const cascadeImpact: CapabilityDeleteImpact = {
  capabilityId: toCapabilityId('cap-1'),
  capabilityName: 'Parent Capability',
  hasDescendants: true,
  affectedCapabilities: [
    { id: toCapabilityId('cap-2'), name: 'Child A', level: 'L2' as const, parentId: toCapabilityId('cap-1') },
    { id: toCapabilityId('cap-3'), name: 'Child B', level: 'L2' as const, parentId: toCapabilityId('cap-1') },
  ],
  realizations: [],
  _links: {},
};

const mockRetry = vi.fn();

function mockImpacts(impacts: CapabilityDeleteImpact[], state: Partial<DeleteImpacts> = {}) {
  vi.mocked(useDeleteImpacts).mockReturnValue({
    impacts,
    isLoading: false,
    isError: false,
    retry: mockRetry,
    ...state,
  });
}

describe('DeleteCapabilityDialog', () => {
  const mockOnClose = vi.fn();
  const mockOnConfirm = vi.fn();
  const mockDeleteMutateAsync = vi.fn();
  const mockCascadeMutateAsync = vi.fn();
  const capability = buildCapability({ id: toCapabilityId('cap-1'), name: 'Test Capability' });
  const selected = [
    buildCapability({ id: toCapabilityId('cap-10'), name: 'Cap 10' }),
    buildCapability({ id: toCapabilityId('cap-11'), name: 'Cap 11' }),
  ];

  const renderDialog = (props: Partial<React.ComponentProps<typeof DeleteCapabilityDialog>> = {}) =>
    renderWithProviders(
      <DeleteCapabilityDialog
        isOpen={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
        capability={capability}
        {...props}
      />,
    );

  const renderMultiDelete = () =>
    renderDialog({ capability: selected[0], capabilitiesToDelete: selected, domainId: 'domain-1' });

  const submitButton = () => screen.getByTestId('delete-capability-submit') as HTMLButtonElement;

  beforeEach(() => {
    vi.clearAllMocks();

    mockImpacts([leafImpact]);

    vi.mocked(useDeleteCapability).mockReturnValue({
      mutateAsync: mockDeleteMutateAsync,
      isPending: false,
    } as unknown as ReturnType<typeof useDeleteCapability>);

    vi.mocked(useCascadeDeleteCapability).mockReturnValue({
      mutateAsync: mockCascadeMutateAsync,
      isPending: false,
    } as unknown as ReturnType<typeof useCascadeDeleteCapability>);

    mockDeleteMutateAsync.mockResolvedValue(undefined);
    mockCascadeMutateAsync.mockResolvedValue(undefined);
  });

  it('should not render when capability is null', () => {
    renderDialog({ capability: null });

    expect(screen.queryByTestId('delete-capability-dialog')).not.toBeInTheDocument();
  });

  it('should show loading skeleton when impact is loading', () => {
    mockImpacts([], { isLoading: true });

    renderDialog();

    expect(screen.getByTestId('delete-impact-loading')).toBeInTheDocument();
  });

  it('should show simple delete confirmation for leaf capability', () => {
    renderDialog();

    expect(screen.getByText('Are you sure you want to delete')).toBeInTheDocument();
    expect(screen.getByText('"Test Capability"')).toBeInTheDocument();
  });

  it('should show cascade warning when capability has descendants', () => {
    mockImpacts([cascadeImpact]);

    renderDialog();

    expect(screen.getByTestId('cascade-warning')).toBeInTheDocument();
    expect(screen.getByText('Child A')).toBeInTheDocument();
    expect(screen.getByText('Child B')).toBeInTheDocument();
  });

  it('should list each realization that will be removed with its application and capability', () => {
    mockImpacts([
      {
        ...cascadeImpact,
        realizations: [realization('real-1', 'Billing', 'cap-2'), realization('real-2', 'Ledger', 'cap-1')],
      },
    ]);

    renderDialog();

    const items = screen.getAllByTestId('realization-to-remove');
    expect(items).toHaveLength(2);
    expect(items[0]).toHaveTextContent('Billing');
    expect(items[0]).toHaveTextContent('Child A');
    expect(items[1]).toHaveTextContent('Ledger');
    expect(items[1]).toHaveTextContent('Test Capability');
    expect(screen.getByTestId('realizations-to-remove')).toHaveTextContent(
      '2 application realizations will be removed',
    );
    expect(screen.getByTestId('realizations-to-remove')).toHaveTextContent('The applications themselves are kept');
  });

  it('should list the realization of a realized leaf capability', () => {
    mockImpacts([{ ...leafImpact, realizations: [realization('real-1', 'Billing', 'cap-1')] }]);

    renderDialog();

    expect(screen.getByTestId('realization-to-remove')).toHaveTextContent('Billing');
    expect(screen.getByTestId('realizations-to-remove')).toHaveTextContent('1 application realization will be removed');
  });

  it('should offer no option to delete applications', () => {
    mockImpacts([{ ...cascadeImpact, realizations: [realization('real-1', 'Billing', 'cap-2')] }]);

    renderDialog();

    expect(screen.queryByRole('checkbox')).not.toBeInTheDocument();
  });

  it('should not show a realization section when nothing is realized', () => {
    mockImpacts([cascadeImpact]);

    renderDialog();

    expect(screen.queryByTestId('realizations-to-remove')).not.toBeInTheDocument();
  });

  it('should show cascade text on submit button when has descendants', () => {
    mockImpacts([cascadeImpact]);

    renderDialog();

    expect(submitButton()).toHaveTextContent('Delete Test Capability and 2 children');
  });

  it('should call cascadeDeleteMutation with cascade=false for leaf capability on confirm', async () => {
    renderDialog();

    fireEvent.click(submitButton());

    await waitFor(() => {
      expect(mockCascadeMutateAsync).toHaveBeenCalledWith({
        capability,
        cascade: false,
        parentId: undefined,
        domainId: undefined,
      });
    });
  });

  it('should call cascadeDeleteMutation with cascade=true for cascade delete', async () => {
    mockImpacts([cascadeImpact]);

    renderDialog({ domainId: 'domain-1' });

    fireEvent.click(submitButton());

    await waitFor(() => {
      expect(mockCascadeMutateAsync).toHaveBeenCalledWith({
        capability,
        cascade: true,
        parentId: undefined,
        domainId: 'domain-1',
      });
    });
  });

  it('should show multi-delete message for multiple capabilities', () => {
    mockImpacts([leafImpact, leafImpact]);

    renderMultiDelete();

    expect(screen.getByText(/delete 2 capabilities/)).toBeInTheDocument();
  });

  it('should load the impact of every selected capability in a multi-delete', () => {
    renderMultiDelete();

    expect(useDeleteImpacts).toHaveBeenCalledWith([toCapabilityId('cap-10'), toCapabilityId('cap-11')]);
  });

  it('should list the realizations of every selected capability in a multi-delete', () => {
    mockImpacts([
      { ...leafImpact, realizations: [realization('real-1', 'Billing', 'cap-10')] },
      { ...leafImpact, realizations: [realization('real-2', 'Ledger', 'cap-11')] },
    ]);

    renderMultiDelete();

    const items = screen.getAllByTestId('realization-to-remove');
    expect(items).toHaveLength(2);
    expect(items[0]).toHaveTextContent('Billing stops realising Cap 10');
    expect(items[1]).toHaveTextContent('Ledger stops realising Cap 11');
  });

  it('should disable a multi-delete until every selected impact is loaded', () => {
    mockImpacts([leafImpact]);

    renderMultiDelete();

    expect(submitButton().disabled).toBe(true);
  });

  it('should call deleteCapabilityMutation for each capability in multi-delete', async () => {
    mockImpacts([leafImpact, leafImpact]);

    renderMultiDelete();

    fireEvent.click(submitButton());

    await waitFor(() => {
      expect(mockDeleteMutateAsync).toHaveBeenCalledTimes(2);
      expect(mockDeleteMutateAsync).toHaveBeenCalledWith({
        capability: selected[0],
        parentId: undefined,
        domainId: 'domain-1',
      });
      expect(mockDeleteMutateAsync).toHaveBeenCalledWith({
        capability: selected[1],
        parentId: undefined,
        domainId: 'domain-1',
      });
    });
  });

  it('should keep submit disabled and offer a retry when the impact fails to load', () => {
    mockImpacts([], { isError: true });

    renderDialog();

    expect(submitButton().disabled).toBe(true);
    fireEvent.click(screen.getByTestId('delete-impact-retry'));
    expect(mockRetry).toHaveBeenCalled();
  });

  it('should show error alert on mutation failure', async () => {
    mockCascadeMutateAsync.mockRejectedValueOnce(new Error('Server error'));

    renderDialog();

    fireEvent.click(submitButton());

    await waitFor(() => {
      expect(screen.getByTestId('delete-capability-error')).toHaveTextContent('Server error');
    });

    expect(mockOnClose).not.toHaveBeenCalled();
  });

  it('should call onClose after successful deletion', async () => {
    renderDialog();

    fireEvent.click(submitButton());

    await waitFor(() => {
      expect(mockOnClose).toHaveBeenCalled();
      expect(mockOnConfirm).toHaveBeenCalled();
    });
  });

  it('should call onClose when cancel button is clicked', () => {
    renderDialog();

    fireEvent.click(screen.getByTestId('delete-capability-cancel'));

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should disable submit button while impact is loading', () => {
    mockImpacts([], { isLoading: true });

    renderDialog();

    expect(submitButton().disabled).toBe(true);
  });
});
