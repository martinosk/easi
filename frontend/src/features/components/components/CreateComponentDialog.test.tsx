import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CreateComponentDialog } from './CreateComponentDialog';
import { setupDialogTest } from '../../../test/helpers/dialogTestUtils';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';

// Mock the store
vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

describe('CreateComponentDialog', () => {
  const mockOnClose = vi.fn();
  let mocks: ReturnType<typeof setupDialogTest>;

  beforeEach(() => {
    vi.clearAllMocks();

    // Use shared dialog test setup
    mocks = setupDialogTest();
  });

  it('should render dialog when open', () => {
    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />, { wrapper: MantineTestWrapper });

    expect(screen.getAllByText('Create Application')[0]).toBeInTheDocument();
    expect(screen.getByLabelText(/Name/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Description/)).toBeInTheDocument();
  });

  it('should not show modal when isOpen is false', () => {
    render(<CreateComponentDialog isOpen={false} onClose={mockOnClose} />, { wrapper: MantineTestWrapper });

    expect(screen.queryByText('Create Application')).not.toBeInTheDocument();
  });

  it('should show modal when isOpen is true', () => {
    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />, { wrapper: MantineTestWrapper });

    expect(screen.getAllByText('Create Application').length).toBeGreaterThan(0);
  });

  it('should disable submit button when name is empty', () => {
    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />, { wrapper: MantineTestWrapper });

    const buttons = screen.getAllByRole('button');
    const submitButton = buttons.find(btn => btn.textContent === 'Create Application') as HTMLButtonElement;

    expect(submitButton.disabled).toBe(true);
    expect(mocks.mockCreateComponent).not.toHaveBeenCalled();
  });

  it('should call createComponent with valid data', async () => {
    mocks.mockCreateComponent.mockResolvedValueOnce({ id: '1', name: 'Test Component' });

    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />, { wrapper: MantineTestWrapper });

    const nameInput = screen.getByLabelText(/Name/);
    const descriptionInput = screen.getByLabelText(/Description/);
    const buttons = screen.getAllByRole('button');
    const submitButton = buttons.find(btn => btn.textContent === 'Create Application');

    fireEvent.change(nameInput, { target: { value: 'Test Component' } });
    fireEvent.change(descriptionInput, { target: { value: 'Test Description' } });
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(mocks.mockCreateComponent).toHaveBeenCalledWith({
        name: 'Test Component',
        description: 'Test Description',
      });
    });

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should trim whitespace from inputs', async () => {
    mocks.mockCreateComponent.mockResolvedValueOnce({ id: '1', name: 'Test Component' });

    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />, { wrapper: MantineTestWrapper });

    const nameInput = screen.getByLabelText(/Name/);
    const buttons = screen.getAllByRole('button');
    const submitButton = buttons.find(btn => btn.textContent === 'Create Application');

    fireEvent.change(nameInput, { target: { value: '  Test Component  ' } });
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(mocks.mockCreateComponent).toHaveBeenCalledWith({
        name: 'Test Component',
        description: undefined,
      });
    });
  });

  it('should handle create component error', async () => {
    mocks.mockCreateComponent.mockRejectedValueOnce(new Error('Creation failed'));

    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />, { wrapper: MantineTestWrapper });

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

  it('should disable inputs while creating', async () => {
    mocks.mockCreateComponent.mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100))
    );

    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />, { wrapper: MantineTestWrapper });

    const nameInput = screen.getByLabelText(/Name/) as HTMLInputElement;
    const buttons = screen.getAllByRole('button');
    const submitButton = buttons.find(btn => btn.textContent?.includes('Create'));

    fireEvent.change(nameInput, { target: { value: 'Test Component' } });
    fireEvent.click(submitButton!);

    await waitFor(() => {
      const loadingButton = buttons.find(btn => btn.getAttribute('data-loading') === 'true');
      expect(loadingButton).toBeInTheDocument();
    });

    expect(nameInput.disabled).toBe(true);
  });

  it('should reset form on close', async () => {
    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />, { wrapper: MantineTestWrapper });

    const nameInput = screen.getByLabelText(/Name/) as HTMLInputElement;
    const cancelButton = screen.getByText('Cancel');

    fireEvent.change(nameInput, { target: { value: 'Test Component' } });
    fireEvent.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalled();
  });
});
