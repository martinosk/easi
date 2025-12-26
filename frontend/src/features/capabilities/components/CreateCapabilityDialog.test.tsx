import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CreateCapabilityDialog } from './CreateCapabilityDialog';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

vi.mock('../hooks/useCapabilities', () => ({
  useCreateCapability: vi.fn(),
  useUpdateCapabilityMetadata: vi.fn(),
}));

vi.mock('../../../hooks/useMetadata', () => ({
  useMaturityLevels: vi.fn(),
  useStatuses: vi.fn(),
}));

vi.mock('../../../hooks/useMaturityScale', () => ({
  useMaturityScale: vi.fn(),
}));

import { useCreateCapability, useUpdateCapabilityMetadata } from '../hooks/useCapabilities';
import { useMaturityLevels, useStatuses } from '../../../hooks/useMetadata';
import { useMaturityScale } from '../../../hooks/useMaturityScale';

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
      <MantineTestWrapper>
        {ui}
      </MantineTestWrapper>
    </QueryClientProvider>
  );
}

describe('CreateCapabilityDialog', () => {
  const mockOnClose = vi.fn();
  const mockCreateMutateAsync = vi.fn();
  const mockUpdateMetadataMutateAsync = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();

    vi.mocked(useMaturityLevels).mockReturnValue({
      data: ['Genesis', 'Custom Build', 'Product', 'Commodity'],
      isLoading: false,
      error: null,
      isError: false,
      isPending: false,
      isSuccess: true,
      status: 'success',
    } as ReturnType<typeof useMaturityLevels>);

    vi.mocked(useStatuses).mockReturnValue({
      data: [
        { value: 'Active', displayName: 'Active', sortOrder: 1 },
        { value: 'Planned', displayName: 'Planned', sortOrder: 2 },
        { value: 'Deprecated', displayName: 'Deprecated', sortOrder: 3 },
      ],
      isLoading: false,
      error: null,
      isError: false,
      isPending: false,
      isSuccess: true,
      status: 'success',
    } as ReturnType<typeof useStatuses>);

    vi.mocked(useMaturityScale).mockReturnValue({
      data: null,
      isLoading: false,
      error: null,
      isError: false,
      isPending: false,
      isSuccess: true,
      status: 'success',
    } as ReturnType<typeof useMaturityScale>);

    vi.mocked(useCreateCapability).mockReturnValue({
      mutateAsync: mockCreateMutateAsync,
      isPending: false,
      isError: false,
      isSuccess: false,
      error: null,
      data: undefined,
      mutate: vi.fn(),
      reset: vi.fn(),
    } as unknown as ReturnType<typeof useCreateCapability>);

    vi.mocked(useUpdateCapabilityMetadata).mockReturnValue({
      mutateAsync: mockUpdateMetadataMutateAsync,
      isPending: false,
      isError: false,
      isSuccess: false,
      error: null,
      data: undefined,
      mutate: vi.fn(),
      reset: vi.fn(),
    } as unknown as ReturnType<typeof useUpdateCapabilityMetadata>);
  });

  describe('Dialog visibility', () => {
    it('should show modal when isOpen is true', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByText('Create Capability')).toBeInTheDocument();
        expect(screen.getByTestId('capability-status-select')).not.toBeDisabled();
      });
    });

    it('should not show modal when isOpen is false', () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={false} onClose={mockOnClose} />);

      expect(screen.queryByText('Create Capability')).not.toBeInTheDocument();
    });

    it('should render dialog title when open', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByText('Create Capability')).toBeInTheDocument();
        expect(screen.getByTestId('capability-status-select')).not.toBeDisabled();
      });
    });

    it('should call onClose when cancel button is clicked', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('capability-status-select')).not.toBeDisabled();
      });

      const cancelButton = screen.getByTestId('create-capability-cancel');
      fireEvent.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  describe('Form fields', () => {
    it('should render all form fields', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      expect(screen.getByTestId('capability-name-input')).toBeInTheDocument();
      expect(screen.getByTestId('capability-description-input')).toBeInTheDocument();
      expect(screen.getByTestId('capability-status-select')).toBeInTheDocument();
      expect(screen.getByTestId('maturity-slider')).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.getByTestId('capability-status-select')).not.toBeDisabled();
      });
    });

    it('should have Active as default status', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const statusSelect = screen.getByTestId('capability-status-select') as HTMLSelectElement;
        expect(statusSelect.value).toBe('Active');
      });
    });

    it('should show all status options', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const statusSelect = screen.getByTestId('capability-status-select');
        expect(statusSelect).toBeInTheDocument();
        expect(statusSelect).not.toBeDisabled();
      });
    });
  });

  describe('Form validation', () => {
    it('should disable submit button when name is empty', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('capability-status-select')).not.toBeDisabled();
      });

      const submitButton = screen.getByTestId('create-capability-submit') as HTMLButtonElement;
      expect(submitButton.disabled).toBe(true);
    });

    it('should enable submit button when name has content', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      const nameInput = screen.getByTestId('capability-name-input');
      fireEvent.change(nameInput, { target: { value: 'Test Capability' } });

      await waitFor(() => {
        const submitButton = screen.getByTestId('create-capability-submit') as HTMLButtonElement;
        expect(submitButton.disabled).toBe(false);
      });
    });

    it('should show error for name exceeding 200 characters', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('maturity-slider')).toBeInTheDocument();
      });

      const nameInput = screen.getByTestId('capability-name-input');
      const longName = 'a'.repeat(201);
      fireEvent.change(nameInput, { target: { value: longName } });

      const submitButton = screen.getByTestId('create-capability-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('Name must be 200 characters or less')).toBeInTheDocument();
      });

      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it('should show error for description exceeding 1000 characters', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('maturity-slider')).toBeInTheDocument();
      });

      const nameInput = screen.getByTestId('capability-name-input');
      const descriptionInput = screen.getByTestId('capability-description-input');
      const longDescription = 'a'.repeat(1001);

      fireEvent.change(nameInput, { target: { value: 'Test Capability' } });
      fireEvent.change(descriptionInput, { target: { value: longDescription } });

      const submitButton = screen.getByTestId('create-capability-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('Description must be 1000 characters or less')).toBeInTheDocument();
      });

      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it('should show error when name is only whitespace', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('capability-status-select')).not.toBeDisabled();
      });

      const nameInput = screen.getByTestId('capability-name-input');
      fireEvent.change(nameInput, { target: { value: '   ' } });

      const submitButton = screen.getByTestId('create-capability-submit') as HTMLButtonElement;
      expect(submitButton.disabled).toBe(true);
    });
  });

  describe('Maturity slider', () => {
    it('should use maturity scale from hook', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(useMaturityScale).toHaveBeenCalled();
      });
    });

    it('should set default maturity value to 12', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const slider = screen.getByTestId('maturity-slider');
        expect(slider).toHaveAttribute('aria-valuenow', '12');
      });
    });

    it('should fall back to default sections if hook returns empty', async () => {
      vi.mocked(useMaturityScale).mockReturnValue({
        data: null,
        isLoading: false,
        error: null,
        isError: false,
        isPending: false,
        isSuccess: true,
        status: 'success',
      } as ReturnType<typeof useMaturityScale>);

      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const slider = screen.getByTestId('maturity-slider');
        expect(slider).toHaveAttribute('aria-valuenow', '12');
      });
    });

    it('should display maturity slider', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const slider = screen.getByTestId('maturity-slider');
        expect(slider).toBeInTheDocument();
        expect(slider).not.toBeDisabled();
      });
    });
  });

  describe('Form submission', () => {
    it('should call createCapability and updateCapabilityMetadata on submit', async () => {
      const mockCapability = { id: 'cap-1', name: 'Test Capability', level: 'L1' };
      mockCreateMutateAsync.mockResolvedValueOnce(mockCapability);
      mockUpdateMetadataMutateAsync.mockResolvedValueOnce({
        ...mockCapability,
        status: 'Active',
        maturityValue: 12,
      });

      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('maturity-slider')).toBeInTheDocument();
      });

      const nameInput = screen.getByTestId('capability-name-input');
      fireEvent.change(nameInput, { target: { value: 'Test Capability' } });

      const submitButton = screen.getByTestId('create-capability-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockCreateMutateAsync).toHaveBeenCalledWith({
          name: 'Test Capability',
          description: undefined,
          level: 'L1',
        });
      });

      await waitFor(() => {
        expect(mockUpdateMetadataMutateAsync).toHaveBeenCalledWith({
          id: 'cap-1',
          request: {
            status: 'Active',
            maturityValue: 12,
          },
        });
      });

      expect(mockOnClose).toHaveBeenCalled();
    });

    it('should trim whitespace from name and description', async () => {
      const mockCapability = { id: 'cap-1', name: 'Trimmed Name', level: 'L1' };
      mockCreateMutateAsync.mockResolvedValueOnce(mockCapability);
      mockUpdateMetadataMutateAsync.mockResolvedValueOnce(mockCapability);

      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('maturity-slider')).toBeInTheDocument();
      });

      const nameInput = screen.getByTestId('capability-name-input');
      const descriptionInput = screen.getByTestId('capability-description-input');

      fireEvent.change(nameInput, { target: { value: '  Trimmed Name  ' } });
      fireEvent.change(descriptionInput, { target: { value: '  Trimmed Description  ' } });

      const submitButton = screen.getByTestId('create-capability-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockCreateMutateAsync).toHaveBeenCalledWith({
          name: 'Trimmed Name',
          description: 'Trimmed Description',
          level: 'L1',
        });
      });
    });

    it('should disable submit button during creation', async () => {
      vi.mocked(useCreateCapability).mockReturnValue({
        mutateAsync: mockCreateMutateAsync,
        isPending: true,
        isError: false,
        isSuccess: false,
        error: null,
        data: undefined,
        mutate: vi.fn(),
        reset: vi.fn(),
      } as unknown as ReturnType<typeof useCreateCapability>);

      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const submitButton = screen.getByTestId('create-capability-submit') as HTMLButtonElement;
        expect(submitButton).toHaveAttribute('data-loading', 'true');
      });
    });

    it('should disable inputs while creating', async () => {
      vi.mocked(useCreateCapability).mockReturnValue({
        mutateAsync: mockCreateMutateAsync,
        isPending: true,
        isError: false,
        isSuccess: false,
        error: null,
        data: undefined,
        mutate: vi.fn(),
        reset: vi.fn(),
      } as unknown as ReturnType<typeof useCreateCapability>);

      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const nameInput = screen.getByTestId('capability-name-input') as HTMLInputElement;
        expect(nameInput.disabled).toBe(true);
      });
    });
  });

  describe('Error handling', () => {
    it('should display backend errors', async () => {
      mockCreateMutateAsync.mockRejectedValueOnce(new Error('Duplicate capability name'));

      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('maturity-slider')).toBeInTheDocument();
      });

      const nameInput = screen.getByTestId('capability-name-input');
      fireEvent.change(nameInput, { target: { value: 'Test Capability' } });

      const submitButton = screen.getByTestId('create-capability-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByTestId('create-capability-error')).toHaveTextContent(
          'Duplicate capability name'
        );
      });

      expect(mockOnClose).not.toHaveBeenCalled();
    });

    it('should display generic error for non-Error exceptions', async () => {
      mockCreateMutateAsync.mockRejectedValueOnce('Unknown error');

      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('maturity-slider')).toBeInTheDocument();
      });

      const nameInput = screen.getByTestId('capability-name-input');
      fireEvent.change(nameInput, { target: { value: 'Test Capability' } });

      const submitButton = screen.getByTestId('create-capability-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByTestId('create-capability-error')).toHaveTextContent(
          'Failed to create capability'
        );
      });
    });

    it('should clear error when field changes', async () => {
      mockCreateMutateAsync.mockRejectedValueOnce(new Error('Backend error'));

      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('maturity-slider')).toBeInTheDocument();
      });

      const nameInput = screen.getByTestId('capability-name-input');
      fireEvent.change(nameInput, { target: { value: 'Test' } });

      const submitButton = screen.getByTestId('create-capability-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByTestId('create-capability-error')).toBeInTheDocument();
      });

      fireEvent.change(nameInput, { target: { value: 'Changed' } });

      expect(screen.queryByTestId('create-capability-error')).not.toBeInTheDocument();
    });
  });

  describe('Form reset', () => {
    it('should reset form when dialog closes', async () => {
      renderWithProviders(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('capability-status-select')).not.toBeDisabled();
      });

      const nameInput = screen.getByTestId('capability-name-input') as HTMLInputElement;
      fireEvent.change(nameInput, { target: { value: 'Test Capability' } });

      expect(nameInput.value).toBe('Test Capability');

      const cancelButton = screen.getByTestId('create-capability-cancel');
      fireEvent.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalled();
    });
  });
});
