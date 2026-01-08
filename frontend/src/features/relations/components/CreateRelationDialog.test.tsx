import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { CreateRelationDialog } from './CreateRelationDialog';
import { createMantineTestWrapper, seedDb, server } from '../../../test/helpers';
import type { ComponentId } from '../../../api/types';

const API_BASE = 'http://localhost:8080';

Element.prototype.scrollIntoView = vi.fn();

describe('CreateRelationDialog', () => {
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    seedDb({
      components: [
        {
          id: '1' as ComponentId,
          name: 'Component A',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: '/api/v1/components/1' },
        },
        {
          id: '2' as ComponentId,
          name: 'Component B',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: '/api/v1/components/2' },
        },
        {
          id: '3' as ComponentId,
          name: 'Component C',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: '/api/v1/components/3' },
        },
      ],
    });
  });

  it('should render dialog when open', async () => {
    const { Wrapper } = createMantineTestWrapper();
    render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} />, { wrapper: Wrapper });

    await waitFor(() => {
      expect(screen.getAllByText('Create Relation')[0]).toBeInTheDocument();
    });
    expect(screen.getByTestId('relation-source-select')).toBeInTheDocument();
    expect(screen.getByTestId('relation-target-select')).toBeInTheDocument();
    expect(screen.getByTestId('relation-type-select')).toBeInTheDocument();
  });

  it('should display all components in dropdowns', async () => {
    const { Wrapper } = createMantineTestWrapper();
    render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} />, { wrapper: Wrapper });

    await waitFor(() => {
      expect(screen.getByTestId('relation-source-select')).toBeInTheDocument();
    });
    const sourceSelect = screen.getByTestId('relation-source-select');
    const targetSelect = screen.getByTestId('relation-target-select');

    expect(sourceSelect).toBeInTheDocument();
    expect(targetSelect).toBeInTheDocument();
  });

  it('should disable submit button when required fields are empty', async () => {
    const { Wrapper } = createMantineTestWrapper();
    render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} />, { wrapper: Wrapper });

    await waitFor(() => {
      expect(screen.getByTestId('create-relation-submit')).toBeInTheDocument();
    });
    const submitButton = screen.getByTestId('create-relation-submit') as HTMLButtonElement;
    expect(submitButton.disabled).toBe(true);
  });

  it('should enable submit button when source and target are pre-filled', async () => {
    const { Wrapper } = createMantineTestWrapper();
    render(
      <CreateRelationDialog
        isOpen={true}
        onClose={mockOnClose}
        sourceComponentId="1"
        targetComponentId="2"
      />,
      { wrapper: Wrapper }
    );

    await waitFor(() => {
      const submitButton = screen.getByTestId('create-relation-submit') as HTMLButtonElement;
      expect(submitButton.disabled).toBe(false);
    });
  });

  it('should pre-fill source and target when provided', async () => {
    const { Wrapper } = createMantineTestWrapper();
    render(
      <CreateRelationDialog
        isOpen={true}
        onClose={mockOnClose}
        sourceComponentId="1"
        targetComponentId="2"
      />,
      { wrapper: Wrapper }
    );

    await waitFor(() => {
      const sourceSelect = screen.getByTestId('relation-source-select') as HTMLInputElement;
      expect(sourceSelect.value).toBe('Component A');
    });
    const targetSelect = screen.getByTestId('relation-target-select') as HTMLInputElement;
    expect(targetSelect.value).toBe('Component B');
  });

  it('should disable source and target when pre-filled', async () => {
    const { Wrapper } = createMantineTestWrapper();
    render(
      <CreateRelationDialog
        isOpen={true}
        onClose={mockOnClose}
        sourceComponentId="1"
        targetComponentId="2"
      />,
      { wrapper: Wrapper }
    );

    await waitFor(() => {
      const sourceSelect = screen.getByTestId('relation-source-select');
      expect(sourceSelect).toBeDisabled();
    });
    const targetSelect = screen.getByTestId('relation-target-select');
    expect(targetSelect).toBeDisabled();
  });

  it('should display error when source and target are the same', async () => {
    const { Wrapper } = createMantineTestWrapper();
    render(
      <CreateRelationDialog
        isOpen={true}
        onClose={mockOnClose}
        sourceComponentId="1"
        targetComponentId="1"
      />,
      { wrapper: Wrapper }
    );

    await waitFor(() => {
      expect(screen.getByTestId('create-relation-submit')).toBeInTheDocument();
    });

    const submitButton = screen.getByTestId('create-relation-submit');
    expect(submitButton).toBeDisabled();

    expect(screen.getByText('Source and target components must be different')).toBeInTheDocument();
  });

  it('should call API with valid data and close dialog', async () => {
    let capturedRequest: Record<string, unknown> | null = null;
    server.use(
      http.post(`${API_BASE}/api/v1/relations`, async ({ request }) => {
        capturedRequest = await request.json() as Record<string, unknown>;
        return HttpResponse.json({
          id: 'rel-1',
          ...capturedRequest,
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: '/api/v1/relations/rel-1' },
        }, { status: 201 });
      })
    );

    const { Wrapper } = createMantineTestWrapper();
    render(
      <CreateRelationDialog
        isOpen={true}
        onClose={mockOnClose}
        sourceComponentId="1"
        targetComponentId="2"
      />,
      { wrapper: Wrapper }
    );

    await waitFor(() => {
      expect(screen.getByTestId('relation-name-input')).toBeInTheDocument();
    });

    const nameInput = screen.getByTestId('relation-name-input');
    const descriptionInput = screen.getByTestId('relation-description-input');

    fireEvent.change(nameInput, { target: { value: 'Test Relation' } });
    fireEvent.change(descriptionInput, { target: { value: 'Test Description' } });

    const submitButton = screen.getByTestId('create-relation-submit');
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(capturedRequest).toEqual({
        sourceComponentId: '1',
        targetComponentId: '2',
        relationType: 'Triggers',
        name: 'Test Relation',
        description: 'Test Description',
      });
    });

    await waitFor(() => {
      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  it('should handle Serves relation type', async () => {
    let capturedRequest: Record<string, unknown> | null = null;
    server.use(
      http.post(`${API_BASE}/api/v1/relations`, async ({ request }) => {
        capturedRequest = await request.json() as Record<string, unknown>;
        return HttpResponse.json({
          id: 'rel-1',
          ...capturedRequest,
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: '/api/v1/relations/rel-1' },
        }, { status: 201 });
      })
    );

    const user = userEvent.setup();
    const { Wrapper } = createMantineTestWrapper();
    render(
      <CreateRelationDialog
        isOpen={true}
        onClose={mockOnClose}
        sourceComponentId="1"
        targetComponentId="2"
      />,
      { wrapper: Wrapper }
    );

    await waitFor(() => {
      expect(screen.getByTestId('relation-type-select')).toBeInTheDocument();
    });

    const relationTypeSelect = screen.getByTestId('relation-type-select');
    await user.click(relationTypeSelect);

    const servesOption = await screen.findByRole('option', { name: 'Serves' });
    await user.click(servesOption);

    const submitButton = screen.getByTestId('create-relation-submit');
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(capturedRequest).toEqual(
        expect.objectContaining({
          relationType: 'Serves',
        })
      );
    });
  });

  it('should handle create relation error', async () => {
    server.use(
      http.post(`${API_BASE}/api/v1/relations`, () => {
        return HttpResponse.json(
          { error: 'Network error', message: 'Network error' },
          { status: 500 }
        );
      })
    );

    const { Wrapper } = createMantineTestWrapper();
    render(
      <CreateRelationDialog
        isOpen={true}
        onClose={mockOnClose}
        sourceComponentId="1"
        targetComponentId="2"
      />,
      { wrapper: Wrapper }
    );

    await waitFor(() => {
      expect(screen.getByTestId('create-relation-submit')).toBeInTheDocument();
    });

    const submitButton = screen.getByTestId('create-relation-submit');
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByTestId('create-relation-error')).toBeInTheDocument();
    });

    expect(mockOnClose).not.toHaveBeenCalled();
  });

  it('should close dialog when cancel is clicked', async () => {
    const { Wrapper } = createMantineTestWrapper();
    render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} />, { wrapper: Wrapper });

    await waitFor(() => {
      expect(screen.getByTestId('create-relation-cancel')).toBeInTheDocument();
    });

    const cancelButton = screen.getByTestId('create-relation-cancel');
    fireEvent.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should select components via dropdown interaction', async () => {
    let capturedRequest: Record<string, unknown> | null = null;
    server.use(
      http.post(`${API_BASE}/api/v1/relations`, async ({ request }) => {
        capturedRequest = await request.json() as Record<string, unknown>;
        return HttpResponse.json({
          id: 'rel-1',
          ...capturedRequest,
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: '/api/v1/relations/rel-1' },
        }, { status: 201 });
      })
    );

    const user = userEvent.setup();
    const { Wrapper } = createMantineTestWrapper();
    render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} />, { wrapper: Wrapper });

    const sourceSelect = await screen.findByTestId('relation-source-select');
    await user.click(sourceSelect);
    const componentA = await screen.findByRole('option', { name: 'Component A' }, { timeout: 3000 });
    await user.click(componentA);

    const targetSelect = screen.getByTestId('relation-target-select');
    await user.click(targetSelect);
    const componentB = await screen.findByRole('option', { name: 'Component B' }, { timeout: 3000 });
    await user.click(componentB);

    const submitButton = screen.getByTestId('create-relation-submit') as HTMLButtonElement;
    expect(submitButton.disabled).toBe(false);

    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(capturedRequest).toEqual(
        expect.objectContaining({
          sourceComponentId: '1',
          targetComponentId: '2',
        })
      );
    });
  });

  it('should disable inputs while creating', async () => {
    server.use(
      http.post(`${API_BASE}/api/v1/relations`, async () => {
        await new Promise((resolve) => setTimeout(resolve, 100));
        return HttpResponse.json({
          id: 'rel-1',
          sourceComponentId: '1',
          targetComponentId: '2',
          relationType: 'Triggers',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: '/api/v1/relations/rel-1' },
        }, { status: 201 });
      })
    );

    const { Wrapper } = createMantineTestWrapper();
    render(
      <CreateRelationDialog
        isOpen={true}
        onClose={mockOnClose}
        sourceComponentId="1"
        targetComponentId="2"
      />,
      { wrapper: Wrapper }
    );

    await waitFor(() => {
      expect(screen.getByTestId('create-relation-submit')).toBeInTheDocument();
    });

    const submitButton = screen.getByTestId('create-relation-submit');
    fireEvent.click(submitButton);

    await waitFor(() => {
      const nameInput = screen.getByTestId('relation-name-input') as HTMLInputElement;
      expect(nameInput.disabled).toBe(true);
    });
  });
});
