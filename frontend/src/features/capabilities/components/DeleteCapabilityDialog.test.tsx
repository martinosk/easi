import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { CapabilityDeleteImpact } from '../../../api/types';
import { toCapabilityId, toComponentId } from '../../../api/types';
import { buildCapability } from '../../../test/helpers/entityBuilders';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import { DeleteCapabilityDialog } from './DeleteCapabilityDialog';

vi.mock('../hooks/useCapabilities', () => ({
  useDeleteCapability: vi.fn(),
  useDeleteImpact: vi.fn(),
  useCascadeDeleteCapability: vi.fn(),
}));

import { useCascadeDeleteCapability, useDeleteCapability, useDeleteImpact } from '../hooks/useCapabilities';

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

const leafImpact: CapabilityDeleteImpact = {
  capabilityId: toCapabilityId('cap-1'),
  capabilityName: 'Test Capability',
  hasDescendants: false,
  affectedCapabilities: [],
  realizationsOnDeletedCapabilities: [],
  realizationsOnRetainedCapabilities: [],
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
  realizationsOnDeletedCapabilities: [],
  realizationsOnRetainedCapabilities: [],
  _links: {},
};

describe('DeleteCapabilityDialog', () => {
  const mockOnClose = vi.fn();
  const mockOnConfirm = vi.fn();
  const mockDeleteMutateAsync = vi.fn();
  const mockCascadeMutateAsync = vi.fn();
  const capability = buildCapability({ id: toCapabilityId('cap-1'), name: 'Test Capability' });

  beforeEach(() => {
    vi.clearAllMocks();

    vi.mocked(useDeleteImpact).mockReturnValue({
      data: leafImpact,
      isLoading: false,
    } as ReturnType<typeof useDeleteImpact>);

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
    renderWithProviders(<DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} capability={null} />);

    expect(screen.queryByTestId('delete-capability-dialog')).not.toBeInTheDocument();
  });

  it('should show loading skeleton when impact is loading', () => {
    vi.mocked(useDeleteImpact).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as ReturnType<typeof useDeleteImpact>);

    renderWithProviders(<DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} capability={capability} />);

    expect(screen.getByTestId('delete-impact-loading')).toBeInTheDocument();
  });

  it('should show simple delete confirmation for leaf capability', () => {
    renderWithProviders(<DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} capability={capability} />);

    expect(screen.getByText('Are you sure you want to delete')).toBeInTheDocument();
    expect(screen.getByText('"Test Capability"')).toBeInTheDocument();
  });

  it('should show cascade warning when capability has descendants', () => {
    vi.mocked(useDeleteImpact).mockReturnValue({
      data: cascadeImpact,
      isLoading: false,
    } as ReturnType<typeof useDeleteImpact>);

    renderWithProviders(<DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} capability={capability} />);

    expect(screen.getByTestId('cascade-warning')).toBeInTheDocument();
    expect(screen.getByText('Child A')).toBeInTheDocument();
    expect(screen.getByText('Child B')).toBeInTheDocument();
  });

  it('should show delete applications checkbox when deletable realizations exist', () => {
    const impactWithRealizations: CapabilityDeleteImpact = {
      ...cascadeImpact,
      realizationsOnDeletedCapabilities: [
        {
          id: 'real-1',
          componentId: toComponentId('comp-1'),
          componentName: 'App 1',
          capabilityId: toCapabilityId('cap-1'),
          realizationLevel: 'Full',
          origin: 'Direct',
        },
      ],
    };

    vi.mocked(useDeleteImpact).mockReturnValue({
      data: impactWithRealizations,
      isLoading: false,
    } as ReturnType<typeof useDeleteImpact>);

    renderWithProviders(<DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} capability={capability} />);

    expect(screen.getByTestId('delete-applications-checkbox')).toBeInTheDocument();
  });

  it('should not show checkbox when no deletable realizations', () => {
    vi.mocked(useDeleteImpact).mockReturnValue({
      data: cascadeImpact,
      isLoading: false,
    } as ReturnType<typeof useDeleteImpact>);

    renderWithProviders(<DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} capability={capability} />);

    expect(screen.queryByTestId('delete-applications-checkbox')).not.toBeInTheDocument();
  });

  it('should show retained realizations info', () => {
    const impactWithRetained: CapabilityDeleteImpact = {
      ...cascadeImpact,
      realizationsOnRetainedCapabilities: [
        {
          id: 'real-2',
          componentId: toComponentId('comp-2'),
          componentName: 'App 2',
          capabilityId: toCapabilityId('cap-2'),
          realizationLevel: 'Partial',
          origin: 'Direct',
        },
      ],
    };

    vi.mocked(useDeleteImpact).mockReturnValue({
      data: impactWithRetained,
      isLoading: false,
    } as ReturnType<typeof useDeleteImpact>);

    renderWithProviders(<DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} capability={capability} />);

    expect(screen.getByText(/1 realization will be retained/)).toBeInTheDocument();
  });

  it('should show cascade text on submit button when has descendants', () => {
    vi.mocked(useDeleteImpact).mockReturnValue({
      data: cascadeImpact,
      isLoading: false,
    } as ReturnType<typeof useDeleteImpact>);

    renderWithProviders(<DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} capability={capability} />);

    const submitButton = screen.getByTestId('delete-capability-submit');
    expect(submitButton).toHaveTextContent('Delete Test Capability and 2 children');
  });

  it('should call cascadeDeleteMutation with cascade=false for leaf capability on confirm', async () => {
    renderWithProviders(
      <DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} onConfirm={mockOnConfirm} capability={capability} />,
    );

    fireEvent.click(screen.getByTestId('delete-capability-submit'));

    await waitFor(() => {
      expect(mockCascadeMutateAsync).toHaveBeenCalledWith({
        capability,
        cascade: false,
        deleteRealisingApplications: false,
        parentId: undefined,
        domainId: undefined,
      });
    });
  });

  it('should call cascadeDeleteMutation with cascade=true for cascade delete', async () => {
    vi.mocked(useDeleteImpact).mockReturnValue({
      data: cascadeImpact,
      isLoading: false,
    } as ReturnType<typeof useDeleteImpact>);

    renderWithProviders(
      <DeleteCapabilityDialog
        isOpen={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
        capability={capability}
        domainId="domain-1"
      />,
    );

    fireEvent.click(screen.getByTestId('delete-capability-submit'));

    await waitFor(() => {
      expect(mockCascadeMutateAsync).toHaveBeenCalledWith({
        capability,
        cascade: true,
        deleteRealisingApplications: false,
        parentId: undefined,
        domainId: 'domain-1',
      });
    });
  });

  it('should pass deleteRealisingApplications=true when checkbox checked', async () => {
    const impactWithRealizations: CapabilityDeleteImpact = {
      ...cascadeImpact,
      realizationsOnDeletedCapabilities: [
        {
          id: 'real-1',
          componentId: toComponentId('comp-1'),
          componentName: 'App 1',
          capabilityId: toCapabilityId('cap-1'),
          realizationLevel: 'Full',
          origin: 'Direct',
        },
      ],
    };

    vi.mocked(useDeleteImpact).mockReturnValue({
      data: impactWithRealizations,
      isLoading: false,
    } as ReturnType<typeof useDeleteImpact>);

    renderWithProviders(
      <DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} onConfirm={mockOnConfirm} capability={capability} />,
    );

    fireEvent.click(screen.getByTestId('delete-applications-checkbox'));
    fireEvent.click(screen.getByTestId('delete-capability-submit'));

    await waitFor(() => {
      expect(mockCascadeMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ deleteRealisingApplications: true }),
      );
    });
  });

  it('should show multi-delete message for multiple capabilities', () => {
    const caps = [
      buildCapability({ id: toCapabilityId('cap-1'), name: 'Cap 1' }),
      buildCapability({ id: toCapabilityId('cap-2'), name: 'Cap 2' }),
    ];

    renderWithProviders(
      <DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} capability={caps[0]} capabilitiesToDelete={caps} />,
    );

    expect(screen.getByText(/delete 2 capabilities/)).toBeInTheDocument();
  });

  it('should call deleteCapabilityMutation for each capability in multi-delete', async () => {
    const caps = [
      buildCapability({ id: toCapabilityId('cap-10'), name: 'Cap 10' }),
      buildCapability({ id: toCapabilityId('cap-11'), name: 'Cap 11' }),
    ];

    renderWithProviders(
      <DeleteCapabilityDialog
        isOpen={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
        capability={caps[0]}
        capabilitiesToDelete={caps}
        domainId="domain-1"
      />,
    );

    fireEvent.click(screen.getByTestId('delete-capability-submit'));

    await waitFor(() => {
      expect(mockDeleteMutateAsync).toHaveBeenCalledTimes(2);
      expect(mockDeleteMutateAsync).toHaveBeenCalledWith({
        capability: caps[0],
        parentId: undefined,
        domainId: 'domain-1',
      });
      expect(mockDeleteMutateAsync).toHaveBeenCalledWith({
        capability: caps[1],
        parentId: undefined,
        domainId: 'domain-1',
      });
    });
  });

  it('should show error alert on mutation failure', async () => {
    mockCascadeMutateAsync.mockRejectedValueOnce(new Error('Server error'));

    renderWithProviders(<DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} capability={capability} />);

    fireEvent.click(screen.getByTestId('delete-capability-submit'));

    await waitFor(() => {
      expect(screen.getByTestId('delete-capability-error')).toHaveTextContent('Server error');
    });

    expect(mockOnClose).not.toHaveBeenCalled();
  });

  it('should call onClose after successful deletion', async () => {
    renderWithProviders(
      <DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} onConfirm={mockOnConfirm} capability={capability} />,
    );

    fireEvent.click(screen.getByTestId('delete-capability-submit'));

    await waitFor(() => {
      expect(mockOnClose).toHaveBeenCalled();
      expect(mockOnConfirm).toHaveBeenCalled();
    });
  });

  it('should call onClose when cancel button is clicked', () => {
    renderWithProviders(<DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} capability={capability} />);

    fireEvent.click(screen.getByTestId('delete-capability-cancel'));

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should disable submit button while impact is loading', () => {
    vi.mocked(useDeleteImpact).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as ReturnType<typeof useDeleteImpact>);

    renderWithProviders(<DeleteCapabilityDialog isOpen={true} onClose={mockOnClose} capability={capability} />);

    const submitButton = screen.getByTestId('delete-capability-submit') as HTMLButtonElement;
    expect(submitButton.disabled).toBe(true);
  });
});
