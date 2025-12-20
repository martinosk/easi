import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CreateComponentDialog } from './CreateComponentDialog';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

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

vi.mock('../../../store/appStore', () => ({
  useAppStore: (selector: (state: { currentView: null }) => unknown) => selector({ currentView: null }),
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

  const renderDialog = (isOpen = true) => {
    return render(
      <QueryClientProvider client={queryClient}>
        <CreateComponentDialog isOpen={isOpen} onClose={mockOnClose} />
      </QueryClientProvider>,
      { wrapper: MantineTestWrapper }
    );
  };

  it('should render dialog when open', () => {
    renderDialog();

    expect(screen.getAllByText('Create Application')[0]).toBeInTheDocument();
    expect(screen.getByLabelText(/Name/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Description/)).toBeInTheDocument();
  });

  it('should not show modal when isOpen is false', () => {
    renderDialog(false);

    expect(screen.queryByText('Create Application')).not.toBeInTheDocument();
  });

  it('should show modal when isOpen is true', () => {
    renderDialog();

    expect(screen.getAllByText('Create Application').length).toBeGreaterThan(0);
  });

  it('should disable submit button when name is empty', () => {
    renderDialog();

    const buttons = screen.getAllByRole('button');
    const submitButton = buttons.find(btn => btn.textContent === 'Create Application') as HTMLButtonElement;

    expect(submitButton.disabled).toBe(true);
    expect(mockMutateAsync).not.toHaveBeenCalled();
  });

  it('should call createComponent mutation with valid data', async () => {
    mockMutateAsync.mockResolvedValueOnce({ id: '1', name: 'Test Component' });

    renderDialog();

    const nameInput = screen.getByLabelText(/Name/);
    const descriptionInput = screen.getByLabelText(/Description/);
    const buttons = screen.getAllByRole('button');
    const submitButton = buttons.find(btn => btn.textContent === 'Create Application');

    fireEvent.change(nameInput, { target: { value: 'Test Component' } });
    fireEvent.change(descriptionInput, { target: { value: 'Test Description' } });
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({
        name: 'Test Component',
        description: 'Test Description',
      });
    });

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should trim whitespace from inputs', async () => {
    mockMutateAsync.mockResolvedValueOnce({ id: '1', name: 'Test Component' });

    renderDialog();

    const nameInput = screen.getByLabelText(/Name/);
    const buttons = screen.getAllByRole('button');
    const submitButton = buttons.find(btn => btn.textContent === 'Create Application');

    fireEvent.change(nameInput, { target: { value: '  Test Component  ' } });
    fireEvent.click(submitButton!);

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

    const nameInput = screen.getByLabelText(/Name/);
    const buttons = screen.getAllByRole('button');
    const submitButton = buttons.find(btn => btn.textContent === 'Create Application');

    fireEvent.change(nameInput, { target: { value: 'Test Component' } });
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(screen.getByText('Creation failed')).toBeInTheDocument();
    });

    expect(mockOnClose).not.toHaveBeenCalled();
  });

  it('should reset form on close', async () => {
    renderDialog();

    const nameInput = screen.getByLabelText(/Name/) as HTMLInputElement;
    const cancelButton = screen.getByText('Cancel');

    fireEvent.change(nameInput, { target: { value: 'Test Component' } });
    fireEvent.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalled();
  });
});
