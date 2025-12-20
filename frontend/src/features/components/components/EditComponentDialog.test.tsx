import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { EditComponentDialog } from './EditComponentDialog';
import type { Component } from '../../../api/types';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const mockMutateAsync = vi.fn();

vi.mock('../hooks/useComponents', () => ({
  useUpdateComponent: () => ({
    mutateAsync: mockMutateAsync,
    isPending: false,
  }),
}));

const mockComponent: Component = {
  id: 'comp-1',
  name: 'Test Component',
  description: 'Test description',
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/components/comp-1' } },
};

describe('EditComponentDialog', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    mockMutateAsync.mockReset();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
  });

  const renderDialog = (
    component: Component | null = mockComponent,
    isOpen = true,
    onClose = vi.fn()
  ) => {
    return render(
      <QueryClientProvider client={queryClient}>
        <EditComponentDialog isOpen={isOpen} onClose={onClose} component={component} />
      </QueryClientProvider>,
      { wrapper: MantineTestWrapper }
    );
  };

  describe('Dialog rendering', () => {
    it('should show modal when isOpen is true', async () => {
      renderDialog();

      await waitFor(() => {
        expect(screen.getByText('Edit Application')).toBeInTheDocument();
      });
    });

    it('should not show modal when isOpen is false', () => {
      renderDialog(mockComponent, false);

      expect(screen.queryByText('Edit Application')).not.toBeInTheDocument();
    });

    it('should populate form with component data when opened', async () => {
      renderDialog();

      await waitFor(() => {
        expect(screen.getByDisplayValue('Test Component')).toBeInTheDocument();
        expect(screen.getByDisplayValue('Test description')).toBeInTheDocument();
      });
    });
  });

  describe('Form submission', () => {
    it('should show error when name is empty', async () => {
      renderDialog();

      await waitFor(() => {
        expect(screen.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const nameInput = screen.getByLabelText(/name/i);
      fireEvent.change(nameInput, { target: { value: '' } });

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(screen.getByText('Application name is required')).toBeInTheDocument();
      });
    });

    it('should show error when component is null on submit', async () => {
      const { rerender } = renderDialog();

      await waitFor(() => {
        expect(screen.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      rerender(
        <QueryClientProvider client={queryClient}>
          <EditComponentDialog isOpen={true} onClose={vi.fn()} component={null} />
        </QueryClientProvider>
      );

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(screen.getByText('No application selected')).toBeInTheDocument();
      });
    });

    it('should call updateComponent with correct data on successful submit', async () => {
      mockMutateAsync.mockResolvedValue(undefined);

      const mockOnClose = vi.fn();
      renderDialog(mockComponent, true, mockOnClose);

      await waitFor(() => {
        expect(screen.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const nameInput = screen.getByLabelText(/name/i);
      fireEvent.change(nameInput, { target: { value: 'Updated Component' } });

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith({
          id: 'comp-1',
          request: {
            name: 'Updated Component',
            description: 'Test description',
          },
        });
      });
    });

    it('should close dialog after successful update', async () => {
      mockMutateAsync.mockResolvedValue(undefined);

      const mockOnClose = vi.fn();
      renderDialog(mockComponent, true, mockOnClose);

      await waitFor(() => {
        expect(screen.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(mockOnClose).toHaveBeenCalled();
      });
    });

    it('should show error message when update fails', async () => {
      mockMutateAsync.mockRejectedValue(new Error('Update failed'));

      renderDialog();

      await waitFor(() => {
        expect(screen.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(screen.getByText('Update failed')).toBeInTheDocument();
      });
    });
  });

  describe('Cancel behavior', () => {
    it('should call onClose when cancel button is clicked', async () => {
      const mockOnClose = vi.fn();
      renderDialog(mockComponent, true, mockOnClose);

      const cancelButton = screen.getByText('Cancel');
      fireEvent.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalled();
    });

    it('should reset form state when cancelled', async () => {
      const mockOnClose = vi.fn();
      const { rerender } = renderDialog(mockComponent, true, mockOnClose);

      await waitFor(() => {
        expect(screen.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const nameInput = screen.getByLabelText(/name/i);
      fireEvent.change(nameInput, { target: { value: 'Modified Name' } });

      const cancelButton = screen.getByText('Cancel');
      fireEvent.click(cancelButton);

      const newComponent: Component = {
        id: 'comp-2',
        name: 'Another Component',
        description: 'Another description',
        createdAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/components/comp-2' } },
      };

      rerender(
        <QueryClientProvider client={queryClient}>
          <EditComponentDialog isOpen={true} onClose={mockOnClose} component={newComponent} />
        </QueryClientProvider>
      );

      await waitFor(() => {
        expect(screen.getByDisplayValue('Another Component')).toBeInTheDocument();
      });
    });
  });

  describe('Component prop stability (bug fix verification)', () => {
    it('should use the component prop passed at open time, not a derived value', async () => {
      mockMutateAsync.mockResolvedValue(undefined);

      renderDialog();

      await waitFor(() => {
        expect(screen.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith({
          id: 'comp-1',
          request: expect.objectContaining({ name: 'Test Component' }),
        });
      });
    });

    it('should handle component with empty description', async () => {
      const componentWithoutDescription: Component = {
        id: 'comp-2',
        name: 'No Description Component',
        createdAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/components/comp-2' } },
      };

      renderDialog(componentWithoutDescription);

      await waitFor(() => {
        expect(screen.getByDisplayValue('No Description Component')).toBeInTheDocument();
      });

      const descriptionTextarea = screen.getByLabelText(/description/i);
      expect(descriptionTextarea).toHaveValue('');
    });
  });

  describe('Button states', () => {
    it('should disable submit button when name is empty', async () => {
      renderDialog();

      await waitFor(() => {
        expect(screen.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const nameInput = screen.getByLabelText(/name/i);
      fireEvent.change(nameInput, { target: { value: '   ' } });

      const submitButton = screen.getByRole('button', { name: /Save Changes/i });
      expect(submitButton).toBeDisabled();
    });

    it('should call mutation when form is submitted', async () => {
      mockMutateAsync.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      );

      renderDialog();

      await waitFor(() => {
        expect(screen.getByDisplayValue('Test Component')).toBeInTheDocument();
      });

      const form = document.querySelector('form');
      fireEvent.submit(form!);

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalled();
      });
    });
  });
});
