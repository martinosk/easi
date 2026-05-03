import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import { CreateComponentDialog } from './CreateComponentDialog';

const mockMutateAsync = vi.fn();
const mockAddToViewMutateAsync = vi.fn();

vi.mock('../hooks/useComponents', () => ({
  useCreateComponent: () => ({
    mutateAsync: mockMutateAsync,
    isPending: false,
  }),
}));

vi.mock('../../views/hooks/useViews', () => ({
  useAddComponentToView: () => ({
    mutateAsync: mockAddToViewMutateAsync,
    isPending: false,
  }),
}));

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: () => ({
    currentView: null,
    currentViewId: null,
    isLoading: false,
    error: null,
  }),
}));

describe('CreateComponentDialog', () => {
  const mockOnClose = vi.fn();
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    mockMutateAsync.mockReset();
    mockAddToViewMutateAsync.mockReset();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
  });

  const renderDialog = (props: { isOpen?: boolean; onCreated?: (e: { id: string }) => void | Promise<void> } = {}) => {
    const { isOpen = true, onCreated } = props;
    return render(
      <QueryClientProvider client={queryClient}>
        <CreateComponentDialog isOpen={isOpen} onClose={mockOnClose} onCreated={onCreated} />
      </QueryClientProvider>,
      { wrapper: MantineTestWrapper },
    );
  };

  const findSubmit = (): HTMLElement => {
    const buttons = screen.getAllByRole('button');
    const submit = buttons.find((b) => b.textContent === 'Create Application');
    if (!submit) throw new Error('Create Application button not found');
    return submit;
  };

  const submitWith = async (name: string, description?: string): Promise<void> => {
    fireEvent.change(screen.getByLabelText(/Name/), { target: { value: name } });
    if (description !== undefined) {
      fireEvent.change(screen.getByLabelText(/Description/), { target: { value: description } });
    }
    const submit = findSubmit();
    await waitFor(() => expect(submit).not.toBeDisabled());
    fireEvent.click(submit);
  };

  it('should render dialog when open', () => {
    renderDialog();

    expect(screen.getAllByText('Create Application')[0]).toBeInTheDocument();
    expect(screen.getByLabelText(/Name/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Description/)).toBeInTheDocument();
  });

  it('should not show modal when isOpen is false', () => {
    renderDialog({ isOpen: false });

    expect(screen.queryByText('Create Application')).not.toBeInTheDocument();
  });

  it('should show modal when isOpen is true', () => {
    renderDialog();

    expect(screen.getAllByText('Create Application').length).toBeGreaterThan(0);
  });

  it('should disable submit button when name is empty', () => {
    renderDialog();

    expect((findSubmit() as HTMLButtonElement).disabled).toBe(true);
    expect(mockMutateAsync).not.toHaveBeenCalled();
  });

  it('should call createComponent mutation with valid data', async () => {
    mockMutateAsync.mockResolvedValueOnce({ id: '1', name: 'Test Component' });

    renderDialog();
    await submitWith('Test Component', 'Test Description');

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({
        name: 'Test Component',
        description: 'Test Description',
      });
    });
    await waitFor(() => expect(mockOnClose).toHaveBeenCalled());
  });

  it('should trim whitespace from inputs', async () => {
    mockMutateAsync.mockResolvedValueOnce({ id: '1', name: 'Test Component' });

    renderDialog();
    await submitWith('  Test Component  ');

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({
        name: 'Test Component',
        description: undefined,
      });
    });
  });

  it('should handle create component error', async () => {
    mockMutateAsync.mockRejectedValueOnce(new Error('Creation failed'));

    renderDialog();
    await submitWith('Test Component');

    await waitFor(() => {
      expect(screen.getByText('Creation failed')).toBeInTheDocument();
    });
    expect(mockOnClose).not.toHaveBeenCalled();
  });

  it('should reset form on close', async () => {
    renderDialog();

    const nameInput = screen.getByLabelText(/Name/) as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'Test Component' } });
    await waitFor(() => expect(nameInput.value).toBe('Test Component'));

    fireEvent.click(screen.getByText('Cancel'));
    await waitFor(() => expect(mockOnClose).toHaveBeenCalled());
  });

  describe('onCreated handoff', () => {
    it('calls onCreated with the new entity instead of running default add-to-view', async () => {
      mockMutateAsync.mockResolvedValueOnce({ id: 'new-id', name: 'X' });
      const onCreated = vi.fn();

      renderDialog({ onCreated });
      await submitWith('X');

      await waitFor(() => {
        expect(onCreated).toHaveBeenCalledWith(expect.objectContaining({ id: 'new-id' }));
      });
      expect(mockAddToViewMutateAsync).not.toHaveBeenCalled();
      await waitFor(() => expect(mockOnClose).toHaveBeenCalled());
    });
  });
});
