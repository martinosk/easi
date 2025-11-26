import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import React from 'react';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

import { useAppStore } from '../../../store/appStore';
import { EditComponentDialog } from './EditComponentDialog';
import type { Component } from '../../../api/types';

const getDialog = () => {
  const dialog = document.querySelector('dialog');
  if (!dialog) throw new Error('Dialog not found');
  return within(dialog);
};

const mockComponent: Component = {
  id: 'comp-1',
  name: 'Test Component',
  description: 'Test description',
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/components/comp-1' } },
};

const createMockStore = (overrides: Record<string, unknown> = {}) => ({
  updateComponent: vi.fn(),
  ...overrides,
});

describe('EditComponentDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    HTMLDialogElement.prototype.showModal = vi.fn();
    HTMLDialogElement.prototype.close = vi.fn();
  });

  describe('Dialog rendering', () => {
    it('should show modal when isOpen is true', async () => {
      const mockStore = createMockStore();
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      render(<EditComponentDialog isOpen={true} onClose={vi.fn()} component={mockComponent} />);

      await waitFor(() => {
        expect(HTMLDialogElement.prototype.showModal).toHaveBeenCalled();
      });
    });

    it('should not show modal when isOpen is false', () => {
      const mockStore = createMockStore();
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      render(<EditComponentDialog isOpen={false} onClose={vi.fn()} component={mockComponent} />);

      expect(HTMLDialogElement.prototype.showModal).not.toHaveBeenCalled();
    });

    it('should populate form with component data when opened', async () => {
      const mockStore = createMockStore();
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      render(<EditComponentDialog isOpen={true} onClose={vi.fn()} component={mockComponent} />);

      await waitFor(() => {
        const dialog = getDialog();
        expect(dialog.getByDisplayValue('Test Component')).toBeInTheDocument();
        expect(dialog.getByDisplayValue('Test description')).toBeInTheDocument();
      });
    });
  });

  describe('Form submission', () => {
    it('should show error when name is empty', async () => {
      const mockStore = createMockStore();
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      render(<EditComponentDialog isOpen={true} onClose={vi.fn()} component={mockComponent} />);

      const dialog = getDialog();

      await waitFor(() => {
        expect(dialog.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const nameInput = dialog.getByLabelText(/name/i);
      fireEvent.change(nameInput, { target: { value: '' } });

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(dialog.getByText('Application name is required')).toBeInTheDocument();
      });
    });

    it('should show error when component is null on submit', async () => {
      const mockStore = createMockStore();
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      const { rerender } = render(
        <EditComponentDialog isOpen={true} onClose={vi.fn()} component={mockComponent} />
      );

      const dialog = getDialog();

      await waitFor(() => {
        expect(dialog.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      rerender(<EditComponentDialog isOpen={true} onClose={vi.fn()} component={null} />);

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(dialog.getByText('No application selected')).toBeInTheDocument();
      });
    });

    it('should call updateComponent with correct data on successful submit', async () => {
      const mockUpdateComponent = vi.fn().mockResolvedValue(undefined);
      const mockStore = createMockStore({ updateComponent: mockUpdateComponent });
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      const mockOnClose = vi.fn();
      render(<EditComponentDialog isOpen={true} onClose={mockOnClose} component={mockComponent} />);

      const dialog = getDialog();

      await waitFor(() => {
        expect(dialog.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const nameInput = dialog.getByLabelText(/name/i);
      fireEvent.change(nameInput, { target: { value: 'Updated Component' } });

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(mockUpdateComponent).toHaveBeenCalledWith('comp-1', {
          name: 'Updated Component',
          description: 'Test description',
        });
      });
    });

    it('should close dialog after successful update', async () => {
      const mockUpdateComponent = vi.fn().mockResolvedValue(undefined);
      const mockStore = createMockStore({ updateComponent: mockUpdateComponent });
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      const mockOnClose = vi.fn();
      render(<EditComponentDialog isOpen={true} onClose={mockOnClose} component={mockComponent} />);

      const dialog = getDialog();

      await waitFor(() => {
        expect(dialog.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(mockOnClose).toHaveBeenCalled();
      });
    });

    it('should show error message when update fails', async () => {
      const mockUpdateComponent = vi.fn().mockRejectedValue(new Error('Update failed'));
      const mockStore = createMockStore({ updateComponent: mockUpdateComponent });
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      render(<EditComponentDialog isOpen={true} onClose={vi.fn()} component={mockComponent} />);

      const dialog = getDialog();

      await waitFor(() => {
        expect(dialog.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(dialog.getByText('Update failed')).toBeInTheDocument();
      });
    });
  });

  describe('Cancel behavior', () => {
    it('should call onClose when cancel button is clicked', async () => {
      const mockStore = createMockStore();
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      const mockOnClose = vi.fn();
      render(<EditComponentDialog isOpen={true} onClose={mockOnClose} component={mockComponent} />);

      const dialog = getDialog();
      const cancelButton = dialog.getByText('Cancel');
      fireEvent.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalled();
    });

    it('should reset form state when cancelled', async () => {
      const mockStore = createMockStore();
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      const mockOnClose = vi.fn();
      const { rerender } = render(
        <EditComponentDialog isOpen={true} onClose={mockOnClose} component={mockComponent} />
      );

      const dialog = getDialog();

      await waitFor(() => {
        expect(dialog.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const nameInput = dialog.getByLabelText(/name/i);
      fireEvent.change(nameInput, { target: { value: 'Modified Name' } });

      const cancelButton = dialog.getByText('Cancel');
      fireEvent.click(cancelButton);

      const newComponent: Component = {
        id: 'comp-2',
        name: 'Another Component',
        description: 'Another description',
        createdAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/components/comp-2' } },
      };

      rerender(<EditComponentDialog isOpen={true} onClose={mockOnClose} component={newComponent} />);

      await waitFor(() => {
        expect(dialog.getByDisplayValue('Another Component')).toBeInTheDocument();
      });
    });
  });

  describe('Component prop stability (bug fix verification)', () => {
    it('should use the component prop passed at open time, not a derived value', async () => {
      const mockUpdateComponent = vi.fn().mockResolvedValue(undefined);
      const mockStore = createMockStore({ updateComponent: mockUpdateComponent });
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      render(<EditComponentDialog isOpen={true} onClose={vi.fn()} component={mockComponent} />);

      const dialog = getDialog();

      await waitFor(() => {
        expect(dialog.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(mockUpdateComponent).toHaveBeenCalledWith(
          'comp-1',
          expect.objectContaining({ name: 'Test Component' })
        );
      });
    });

    it('should handle component with empty description', async () => {
      const mockStore = createMockStore();
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      const componentWithoutDescription: Component = {
        id: 'comp-2',
        name: 'No Description Component',
        createdAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/components/comp-2' } },
      };

      render(
        <EditComponentDialog isOpen={true} onClose={vi.fn()} component={componentWithoutDescription} />
      );

      const dialog = getDialog();

      await waitFor(() => {
        expect(dialog.getByDisplayValue('No Description Component')).toBeInTheDocument();
      });

      const descriptionTextarea = dialog.getByLabelText(/description/i);
      expect(descriptionTextarea).toHaveValue('');
    });
  });

  describe('Button states', () => {
    it('should disable submit button when name is empty', async () => {
      const mockStore = createMockStore();
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      render(<EditComponentDialog isOpen={true} onClose={vi.fn()} component={mockComponent} />);

      const dialog = getDialog();

      await waitFor(() => {
        expect(dialog.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const nameInput = dialog.getByLabelText(/name/i);
      fireEvent.change(nameInput, { target: { value: '   ' } });

      const submitButton = dialog.getByText('Update Application');
      expect(submitButton).toBeDisabled();
    });

    it('should show loading state during update', async () => {
      const mockUpdateComponent = vi.fn().mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      );
      const mockStore = createMockStore({ updateComponent: mockUpdateComponent });
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      render(<EditComponentDialog isOpen={true} onClose={vi.fn()} component={mockComponent} />);

      const dialog = getDialog();

      await waitFor(() => {
        expect(dialog.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(dialog.getByText('Updating...')).toBeInTheDocument();
      });
    });
  });
});
