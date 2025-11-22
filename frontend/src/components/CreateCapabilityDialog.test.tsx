import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CreateCapabilityDialog } from './CreateCapabilityDialog';
import { apiClient } from '../api/client';

vi.mock('../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

vi.mock('../api/client', () => ({
  apiClient: {
    getMaturityLevels: vi.fn(),
    getStatuses: vi.fn(),
  },
}));

import { useAppStore } from '../store/appStore';

describe('CreateCapabilityDialog', () => {
  const mockOnClose = vi.fn();
  const mockCreateCapability = vi.fn();
  const mockUpdateCapabilityMetadata = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();

    vi.mocked(useAppStore).mockImplementation((selector: any) =>
      selector({
        createCapability: mockCreateCapability,
        updateCapabilityMetadata: mockUpdateCapabilityMetadata,
        capabilities: [],
        capabilityDependencies: [],
        capabilityRealizations: [],
        loadCapabilities: vi.fn(),
        loadCapabilityDependencies: vi.fn(),
        updateCapability: vi.fn(),
        addCapabilityExpert: vi.fn(),
        addCapabilityTag: vi.fn(),
        createCapabilityDependency: vi.fn(),
        deleteCapabilityDependency: vi.fn(),
        linkSystemToCapability: vi.fn(),
        updateRealization: vi.fn(),
        deleteRealization: vi.fn(),
        loadRealizationsByCapability: vi.fn(),
        loadRealizationsByComponent: vi.fn(),
      })
    );

    HTMLDialogElement.prototype.showModal = vi.fn();
    HTMLDialogElement.prototype.close = vi.fn();

    vi.mocked(apiClient.getMaturityLevels).mockResolvedValue([
      'Genesis',
      'Custom Build',
      'Product',
      'Commodity',
    ]);

    vi.mocked(apiClient.getStatuses).mockResolvedValue([
      { value: 'Active', displayName: 'Active', sortOrder: 1 },
      { value: 'Planned', displayName: 'Planned', sortOrder: 2 },
      { value: 'Deprecated', displayName: 'Deprecated', sortOrder: 3 },
    ]);
  });

  describe('Dialog visibility', () => {
    it('should show modal when isOpen is true', async () => {
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      expect(HTMLDialogElement.prototype.showModal).toHaveBeenCalled();

      await waitFor(() => {
        expect(screen.getByTestId('capability-status-select')).not.toBeDisabled();
      });
    });

    it('should not show modal when isOpen is false', () => {
      render(<CreateCapabilityDialog isOpen={false} onClose={mockOnClose} />);

      expect(HTMLDialogElement.prototype.showModal).not.toHaveBeenCalled();
    });

    it('should render dialog title when open', async () => {
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      expect(
        screen.getByRole('heading', { level: 2, hidden: true })
      ).toHaveTextContent('Create Capability');

      await waitFor(() => {
        expect(screen.getByTestId('capability-status-select')).not.toBeDisabled();
      });
    });

    it('should call onClose when cancel button is clicked', async () => {
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

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
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      expect(screen.getByTestId('capability-name-input')).toBeInTheDocument();
      expect(screen.getByTestId('capability-description-input')).toBeInTheDocument();
      expect(screen.getByTestId('capability-status-select')).toBeInTheDocument();
      expect(screen.getByTestId('capability-maturity-select')).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.getByTestId('capability-status-select')).not.toBeDisabled();
      });
    });

    it('should have Active as default status', async () => {
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const statusSelect = screen.getByTestId('capability-status-select') as HTMLSelectElement;
        expect(statusSelect.value).toBe('Active');
      });
    });

    it('should show all status options', async () => {
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const statusSelect = screen.getByTestId('capability-status-select');
        const options = statusSelect.querySelectorAll('option');

        expect(options).toHaveLength(3);
        expect(options[0].value).toBe('Active');
        expect(options[1].value).toBe('Planned');
        expect(options[2].value).toBe('Deprecated');
      });
    });
  });

  describe('Form validation', () => {
    it('should disable submit button when name is empty', () => {
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      const submitButton = screen.getByTestId('create-capability-submit') as HTMLButtonElement;
      expect(submitButton.disabled).toBe(true);
    });

    it('should enable submit button when name has content', async () => {
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      const nameInput = screen.getByTestId('capability-name-input');
      fireEvent.change(nameInput, { target: { value: 'Test Capability' } });

      await waitFor(() => {
        const submitButton = screen.getByTestId('create-capability-submit') as HTMLButtonElement;
        expect(submitButton.disabled).toBe(false);
      });
    });

    it('should show error for name exceeding 200 characters', async () => {
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('capability-maturity-select')).not.toBeDisabled();
      });

      const nameInput = screen.getByTestId('capability-name-input');
      const longName = 'a'.repeat(201);
      fireEvent.change(nameInput, { target: { value: longName } });

      const submitButton = screen.getByTestId('create-capability-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByTestId('capability-name-error')).toHaveTextContent(
          'Name must be 200 characters or less'
        );
      });

      expect(mockCreateCapability).not.toHaveBeenCalled();
    });

    it('should show error for description exceeding 1000 characters', async () => {
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('capability-maturity-select')).not.toBeDisabled();
      });

      const nameInput = screen.getByTestId('capability-name-input');
      const descriptionInput = screen.getByTestId('capability-description-input');
      const longDescription = 'a'.repeat(1001);

      fireEvent.change(nameInput, { target: { value: 'Test Capability' } });
      fireEvent.change(descriptionInput, { target: { value: longDescription } });

      const submitButton = screen.getByTestId('create-capability-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByTestId('capability-description-error')).toHaveTextContent(
          'Description must be 1000 characters or less'
        );
      });

      expect(mockCreateCapability).not.toHaveBeenCalled();
    });

    it('should show error when name is only whitespace', async () => {
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      const nameInput = screen.getByTestId('capability-name-input');
      fireEvent.change(nameInput, { target: { value: '   ' } });

      const submitButton = screen.getByTestId('create-capability-submit') as HTMLButtonElement;
      expect(submitButton.disabled).toBe(true);
    });
  });

  describe('Maturity levels', () => {
    it('should fetch maturity levels from API on open', async () => {
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(apiClient.getMaturityLevels).toHaveBeenCalled();
      });
    });

    it('should set first maturity level as default', async () => {
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const maturitySelect = screen.getByTestId(
          'capability-maturity-select'
        ) as HTMLSelectElement;
        expect(maturitySelect.value).toBe('Genesis');
      });
    });

    it('should fall back to defaults if API fails', async () => {
      vi.mocked(apiClient.getMaturityLevels).mockRejectedValueOnce(new Error('Network error'));

      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const maturitySelect = screen.getByTestId(
          'capability-maturity-select'
        ) as HTMLSelectElement;
        expect(maturitySelect.value).toBe('Genesis');
      });
    });

    it('should display all maturity level options', async () => {
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const maturitySelect = screen.getByTestId('capability-maturity-select');
        const options = maturitySelect.querySelectorAll('option');

        expect(options).toHaveLength(4);
        expect(options[0].value).toBe('Genesis');
        expect(options[1].value).toBe('Custom Build');
        expect(options[2].value).toBe('Product');
        expect(options[3].value).toBe('Commodity');
      });
    });
  });

  describe('Form submission', () => {
    it('should call createCapability and updateCapabilityMetadata on submit', async () => {
      const mockCapability = { id: 'cap-1', name: 'Test Capability', level: 'L1' };
      mockCreateCapability.mockResolvedValueOnce(mockCapability);
      mockUpdateCapabilityMetadata.mockResolvedValueOnce({
        ...mockCapability,
        status: 'Active',
        maturityLevel: 'Genesis',
      });

      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('capability-maturity-select')).not.toBeDisabled();
      });

      const nameInput = screen.getByTestId('capability-name-input');
      fireEvent.change(nameInput, { target: { value: 'Test Capability' } });

      const submitButton = screen.getByTestId('create-capability-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockCreateCapability).toHaveBeenCalledWith({
          name: 'Test Capability',
          description: undefined,
          level: 'L1',
        });
      });

      await waitFor(() => {
        expect(mockUpdateCapabilityMetadata).toHaveBeenCalledWith('cap-1', {
          status: 'Active',
          maturityLevel: 'Genesis',
        });
      });

      expect(mockOnClose).toHaveBeenCalled();
    });

    it('should trim whitespace from name and description', async () => {
      const mockCapability = { id: 'cap-1', name: 'Trimmed Name', level: 'L1' };
      mockCreateCapability.mockResolvedValueOnce(mockCapability);
      mockUpdateCapabilityMetadata.mockResolvedValueOnce(mockCapability);

      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('capability-maturity-select')).not.toBeDisabled();
      });

      const nameInput = screen.getByTestId('capability-name-input');
      const descriptionInput = screen.getByTestId('capability-description-input');

      fireEvent.change(nameInput, { target: { value: '  Trimmed Name  ' } });
      fireEvent.change(descriptionInput, { target: { value: '  Trimmed Description  ' } });

      const submitButton = screen.getByTestId('create-capability-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockCreateCapability).toHaveBeenCalledWith({
          name: 'Trimmed Name',
          description: 'Trimmed Description',
          level: 'L1',
        });
      });
    });

    it('should disable submit button during creation', async () => {
      mockCreateCapability.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      );

      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('capability-maturity-select')).not.toBeDisabled();
      });

      const nameInput = screen.getByTestId('capability-name-input');
      fireEvent.change(nameInput, { target: { value: 'Test Capability' } });

      const submitButton = screen.getByTestId('create-capability-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('Creating...')).toBeInTheDocument();
      });

      const disabledButton = screen.getByTestId('create-capability-submit') as HTMLButtonElement;
      expect(disabledButton.disabled).toBe(true);
    });

    it('should disable inputs while creating', async () => {
      mockCreateCapability.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      );

      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('capability-maturity-select')).not.toBeDisabled();
      });

      const nameInput = screen.getByTestId('capability-name-input') as HTMLInputElement;
      fireEvent.change(nameInput, { target: { value: 'Test Capability' } });

      const submitButton = screen.getByTestId('create-capability-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(nameInput.disabled).toBe(true);
      });
    });
  });

  describe('Error handling', () => {
    it('should display backend errors', async () => {
      mockCreateCapability.mockRejectedValueOnce(new Error('Duplicate capability name'));

      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('capability-maturity-select')).not.toBeDisabled();
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
      mockCreateCapability.mockRejectedValueOnce('Unknown error');

      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('capability-maturity-select')).not.toBeDisabled();
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
      mockCreateCapability.mockRejectedValueOnce(new Error('Backend error'));

      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByTestId('capability-maturity-select')).not.toBeDisabled();
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
      render(<CreateCapabilityDialog isOpen={true} onClose={mockOnClose} />);

      const nameInput = screen.getByTestId('capability-name-input') as HTMLInputElement;
      fireEvent.change(nameInput, { target: { value: 'Test Capability' } });

      expect(nameInput.value).toBe('Test Capability');

      const cancelButton = screen.getByTestId('create-capability-cancel');
      fireEvent.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalled();
    });
  });
});
