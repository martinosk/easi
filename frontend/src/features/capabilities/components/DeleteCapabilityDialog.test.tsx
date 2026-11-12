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
  useDeleteImpacts: vi.fn(),
  useCascadeDeleteCapability: vi.fn(),
}));

import { useCascadeDeleteCapability, useDeleteImpacts } from '../hooks/useCapabilities';

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <MantineTestWrapper>{ui}</MantineTestWrapper>
    </QueryClientProvider>,
  );
}

interface RealizationSpec {
  id: string;
  componentName: string;
  capabilityId: string;
}

const realization = ({ id, componentName, capabilityId }: RealizationSpec): ImpactRealization => ({
  id,
  componentId: toComponentId(`comp-${id}`),
  componentName,
  capabilityId: toCapabilityId(capabilityId),
  realizationLevel: 'Full',
  origin: 'Direct',
});

const impactFor = (capabilityId: string, overrides: Partial<CapabilityDeleteImpact> = {}): CapabilityDeleteImpact => ({
  capabilityId: toCapabilityId(capabilityId),
  capabilityName: capabilityId,
  hasDescendants: false,
  affectedCapabilities: [],
  realizations: [],
  _links: {},
  ...overrides,
});

interface ChildSpec {
  id: string;
  name: string;
  parentId: string;
}

const child = ({ id, name, parentId }: ChildSpec) => ({
  id: toCapabilityId(id),
  name,
  level: 'L2' as const,
  parentId: toCapabilityId(parentId),
});

const leafImpact = impactFor('cap-1');

const cascadeImpact = impactFor('cap-1', {
  hasDescendants: true,
  affectedCapabilities: [
    child({ id: 'cap-2', name: 'Child A', parentId: 'cap-1' }),
    child({ id: 'cap-3', name: 'Child B', parentId: 'cap-1' }),
  ],
});

const mockRetry = vi.fn();

function mockImpacts(impacts: CapabilityDeleteImpact[], state: Partial<DeleteImpacts> = {}) {
  vi.mocked(useDeleteImpacts).mockReturnValue({
    impacts,
    isLoading: false,
    isFetching: false,
    isError: false,
    retry: mockRetry,
    ...state,
  });
}

describe('DeleteCapabilityDialog', () => {
  const mockOnClose = vi.fn();
  const mockOnConfirm = vi.fn();
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

  const renderMultiDelete = (capabilities = selected) =>
    renderDialog({ capability: capabilities[0], capabilitiesToDelete: capabilities, domainId: 'domain-1' });

  const submitButton = () => screen.getByTestId('delete-capability-submit') as HTMLButtonElement;

  const cascadeCall = (target: typeof capability, cascade: boolean, confirmedRealizationIds: string[]) => ({
    capability: target,
    cascade,
    confirmedRealizationIds,
    parentId: undefined,
    domainId: 'domain-1',
  });

  beforeEach(() => {
    vi.clearAllMocks();
    mockImpacts([leafImpact]);
    vi.mocked(useCascadeDeleteCapability).mockReturnValue({
      mutateAsync: mockCascadeMutateAsync,
      isPending: false,
    } as unknown as ReturnType<typeof useCascadeDeleteCapability>);
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
        realizations: [
          realization({ id: 'real-1', componentName: 'Billing', capabilityId: 'cap-2' }),
          realization({ id: 'real-2', componentName: 'Ledger', capabilityId: 'cap-1' }),
        ],
      },
    ]);

    renderDialog();

    const items = screen.getAllByTestId('realization-to-remove');
    expect(items).toHaveLength(2);
    expect(items[0]).toHaveTextContent('Billing stops realising Child A');
    expect(items[1]).toHaveTextContent('Ledger stops realising Test Capability');
    expect(screen.getByTestId('realizations-to-remove')).toHaveTextContent(
      '2 application realizations will be removed',
    );
    expect(screen.getByTestId('realizations-to-remove')).toHaveTextContent('The applications themselves are kept');
  });

  it('should list the realization of a realized leaf capability', () => {
    mockImpacts([
      impactFor('cap-1', {
        realizations: [realization({ id: 'real-1', componentName: 'Billing', capabilityId: 'cap-1' })],
      }),
    ]);

    renderDialog();

    expect(screen.getByTestId('realization-to-remove')).toHaveTextContent('Billing');
    expect(screen.getByTestId('realizations-to-remove')).toHaveTextContent('1 application realization will be removed');
  });

  it('should offer no option to delete applications', () => {
    mockImpacts([
      {
        ...cascadeImpact,
        realizations: [realization({ id: 'real-1', componentName: 'Billing', capabilityId: 'cap-2' })],
      },
    ]);

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

  it('should delete a leaf capability confirming the realizations it showed', async () => {
    mockImpacts([
      impactFor('cap-1', {
        realizations: [realization({ id: 'real-1', componentName: 'Billing', capabilityId: 'cap-1' })],
      }),
    ]);

    renderDialog({ domainId: 'domain-1' });
    fireEvent.click(submitButton());

    await waitFor(() => {
      expect(mockCascadeMutateAsync).toHaveBeenCalledWith(cascadeCall(capability, false, ['real-1']));
    });
  });

  it('should cascade delete a capability with descendants', async () => {
    mockImpacts([cascadeImpact]);

    renderDialog({ domainId: 'domain-1' });
    fireEvent.click(submitButton());

    await waitFor(() => {
      expect(mockCascadeMutateAsync).toHaveBeenCalledWith(cascadeCall(capability, true, []));
    });
  });

  it('should show multi-delete message for multiple capabilities', () => {
    mockImpacts([impactFor('cap-10'), impactFor('cap-11')]);

    renderMultiDelete();

    expect(screen.getByText(/delete 2 capabilities/)).toBeInTheDocument();
  });

  it('should load the impact of every selected capability in a multi-delete', () => {
    renderMultiDelete();

    expect(useDeleteImpacts).toHaveBeenCalledWith([toCapabilityId('cap-10'), toCapabilityId('cap-11')]);
  });

  it('should list the realizations of every selected capability in a multi-delete', () => {
    mockImpacts([
      impactFor('cap-10', {
        realizations: [realization({ id: 'real-1', componentName: 'Billing', capabilityId: 'cap-10' })],
      }),
      impactFor('cap-11', {
        realizations: [realization({ id: 'real-2', componentName: 'Ledger', capabilityId: 'cap-11' })],
      }),
    ]);

    renderMultiDelete();

    const items = screen.getAllByTestId('realization-to-remove');
    expect(items).toHaveLength(2);
    expect(items[0]).toHaveTextContent('Billing stops realising Cap 10');
    expect(items[1]).toHaveTextContent('Ledger stops realising Cap 11');
  });

  it('should delete every selected capability through the cascade with its own confirmed realizations', async () => {
    mockImpacts([
      impactFor('cap-10', {
        hasDescendants: true,
        realizations: [realization({ id: 'real-1', componentName: 'Billing', capabilityId: 'cap-10' })],
      }),
      impactFor('cap-11'),
    ]);

    renderMultiDelete();
    fireEvent.click(submitButton());

    await waitFor(() => {
      expect(mockCascadeMutateAsync).toHaveBeenCalledTimes(2);
    });
    expect(mockCascadeMutateAsync).toHaveBeenCalledWith(cascadeCall(selected[0], true, ['real-1']));
    expect(mockCascadeMutateAsync).toHaveBeenCalledWith(cascadeCall(selected[1], false, []));
  });

  it('should leave a selected child to its selected parent cascade and count its realizations once', async () => {
    const parent = buildCapability({ id: toCapabilityId('cap-20'), name: 'Parent' });
    const selectedChild = buildCapability({ id: toCapabilityId('cap-21'), name: 'Child' });
    const billing = realization({ id: 'real-1', componentName: 'Billing', capabilityId: 'cap-21' });
    mockImpacts([
      impactFor('cap-20', {
        hasDescendants: true,
        affectedCapabilities: [child({ id: 'cap-21', name: 'Child', parentId: 'cap-20' })],
        realizations: [billing],
      }),
      impactFor('cap-21', { realizations: [billing] }),
    ]);

    renderMultiDelete([parent, selectedChild]);

    expect(screen.getAllByTestId('realization-to-remove')).toHaveLength(1);
    fireEvent.click(submitButton());
    await waitFor(() => {
      expect(mockCascadeMutateAsync).toHaveBeenCalledTimes(1);
    });
    expect(mockCascadeMutateAsync).toHaveBeenCalledWith(cascadeCall(parent, true, ['real-1']));
  });

  it('should disable a multi-delete until every selected impact is loaded', () => {
    mockImpacts([impactFor('cap-10')]);

    renderMultiDelete();

    expect(submitButton().disabled).toBe(true);
  });

  it('should keep submit disabled while impacts are refreshing', () => {
    mockImpacts([leafImpact], { isFetching: true });

    renderDialog();

    expect(submitButton().disabled).toBe(true);
  });

  it('should keep submit disabled and offer a retry when the impact fails to load', () => {
    mockImpacts([leafImpact], { isError: true });

    renderDialog();

    expect(submitButton().disabled).toBe(true);
    fireEvent.click(screen.getByTestId('delete-impact-retry'));
    expect(mockRetry).toHaveBeenCalled();
  });

  it('should show the error and refresh the impact when the delete is rejected', async () => {
    mockCascadeMutateAsync.mockRejectedValueOnce(new Error('The realizations this delete removes have changed.'));

    renderDialog();
    fireEvent.click(submitButton());

    await waitFor(() => {
      expect(screen.getByTestId('delete-capability-error')).toHaveTextContent('have changed');
    });
    expect(mockRetry).toHaveBeenCalled();
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
