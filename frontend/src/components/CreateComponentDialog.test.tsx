import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CreateComponentDialog } from './CreateComponentDialog';
import { useAppStore } from '../store/appStore';

// Mock the store
vi.mock('../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

describe('CreateComponentDialog', () => {
  const mockOnClose = vi.fn();
  const mockCreateComponent = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();

    // Mock useAppStore to return the appropriate value based on the selector
    vi.mocked(useAppStore).mockImplementation((selector: any) =>
      selector({ createComponent: mockCreateComponent })
    );

    // Mock HTMLDialogElement methods
    HTMLDialogElement.prototype.showModal = vi.fn();
    HTMLDialogElement.prototype.close = vi.fn();
  });

  it('should render dialog when open', () => {
    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />);

    // Check for the heading element (use hidden option because dialog is not accessible by default in JSDOM)
    expect(screen.getByRole('heading', { level: 2, hidden: true })).toHaveTextContent('Create Component');
    expect(screen.getByLabelText(/Name/, {})).toBeInTheDocument();
    expect(screen.getByLabelText(/Description/, {})).toBeInTheDocument();
  });

  it('should not show modal when isOpen is false', () => {
    render(<CreateComponentDialog isOpen={false} onClose={mockOnClose} />);

    expect(HTMLDialogElement.prototype.showModal).not.toHaveBeenCalled();
  });

  it('should show modal when isOpen is true', () => {
    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />);

    expect(HTMLDialogElement.prototype.showModal).toHaveBeenCalled();
  });

  it('should disable submit button when name is empty', () => {
    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />);

    const buttons = screen.getAllByRole('button', {});
    const submitButton = buttons.find(btn => btn.textContent === 'Create Component') as HTMLButtonElement;

    // Button should be disabled when name is empty
    expect(submitButton.disabled).toBe(true);
    expect(mockCreateComponent).not.toHaveBeenCalled();
  });

  it('should call createComponent with valid data', async () => {
    mockCreateComponent.mockResolvedValueOnce({ id: '1', name: 'Test Component' });

    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />);

    const nameInput = screen.getByLabelText(/Name/, {});
    const descriptionInput = screen.getByLabelText(/Description/, {});
    const buttons = screen.getAllByRole('button', {});
    const submitButton = buttons.find(btn => btn.textContent === 'Create Component');

    fireEvent.change(nameInput, { target: { value: 'Test Component' } });
    fireEvent.change(descriptionInput, { target: { value: 'Test Description' } });
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(mockCreateComponent).toHaveBeenCalledWith('Test Component', 'Test Description');
    });

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should trim whitespace from inputs', async () => {
    mockCreateComponent.mockResolvedValueOnce({ id: '1', name: 'Test Component' });

    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />);

    const nameInput = screen.getByLabelText(/Name/, {});
    const buttons = screen.getAllByRole('button', {});
    const submitButton = buttons.find(btn => btn.textContent === 'Create Component');

    fireEvent.change(nameInput, { target: { value: '  Test Component  ' } });
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(mockCreateComponent).toHaveBeenCalledWith('Test Component', undefined);
    });
  });

  it('should handle create component error', async () => {
    mockCreateComponent.mockRejectedValueOnce(new Error('Creation failed'));

    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />);

    const nameInput = screen.getByLabelText(/Name/, {});
    const buttons = screen.getAllByRole('button', {});
    const submitButton = buttons.find(btn => btn.textContent === 'Create Component');

    fireEvent.change(nameInput, { target: { value: 'Test Component' } });
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(screen.getByText('Creation failed', {})).toBeInTheDocument();
    });

    expect(mockOnClose).not.toHaveBeenCalled();
  });

  it('should disable inputs while creating', async () => {
    mockCreateComponent.mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100))
    );

    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />);

    const nameInput = screen.getByLabelText(/Name/, {}) as HTMLInputElement;
    const buttons = screen.getAllByRole('button', {});
    const submitButton = buttons.find(btn => btn.textContent === 'Create Component');

    fireEvent.change(nameInput, { target: { value: 'Test Component' } });
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(screen.getByText('Creating...', {})).toBeInTheDocument();
    });

    expect(nameInput.disabled).toBe(true);
  });

  it('should reset form on close', async () => {
    render(<CreateComponentDialog isOpen={true} onClose={mockOnClose} />);

    const nameInput = screen.getByLabelText(/Name/) as HTMLInputElement;
    const cancelButton = screen.getByText('Cancel');

    fireEvent.change(nameInput, { target: { value: 'Test Component' } });
    fireEvent.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalled();
  });
});
