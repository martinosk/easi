import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { HttpResponse, http } from 'msw';
import type React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ComponentId } from '../../../api/types';
import { createMantineTestWrapper, seedDb, server } from '../../../test/helpers';
import { CreateRelationDialog } from './CreateRelationDialog';

const API_BASE = 'http://localhost:8080';

Element.prototype.scrollIntoView = vi.fn();

describe('CreateRelationDialog', () => {
  const mockOnClose = vi.fn();

  const renderDialog = (props: Partial<React.ComponentProps<typeof CreateRelationDialog>> = {}) => {
    const { Wrapper } = createMantineTestWrapper();
    return render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} {...props} />, { wrapper: Wrapper });
  };

  const renderPrefilledDialog = () => renderDialog({ sourceComponentId: '1', targetComponentId: '2' });

  const submitButtonOf = () => screen.getByTestId('create-relation-submit') as HTMLButtonElement;

  const captureRelationPost = () => {
    const captured: { request: Record<string, unknown> | null } = { request: null };
    server.use(
      http.post(`${API_BASE}/api/v1/relations`, async ({ request }) => {
        captured.request = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(
          {
            id: 'rel-1',
            ...captured.request,
            createdAt: '2024-01-01T00:00:00Z',
            _links: { self: '/api/v1/relations/rel-1' },
          },
          { status: 201 },
        );
      }),
    );
    return captured;
  };

  beforeEach(() => {
    vi.clearAllMocks();
    seedDb({
      components: [
        {
          id: '1' as ComponentId,
          name: 'Component A',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/components/1', method: 'GET' } },
        },
        {
          id: '2' as ComponentId,
          name: 'Component B',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/components/2', method: 'GET' } },
        },
        {
          id: '3' as ComponentId,
          name: 'Component C',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/components/3', method: 'GET' } },
        },
      ],
    });
  });

  it('should render dialog when open', async () => {
    renderDialog();

    await waitFor(() => {
      expect(screen.getAllByText('Create Relation')[0]).toBeInTheDocument();
    });
    expect(screen.getByTestId('relation-source-select')).toBeInTheDocument();
    expect(screen.getByTestId('relation-target-select')).toBeInTheDocument();
    expect(screen.getByTestId('relation-type-select')).toBeInTheDocument();
  });

  it('should display all components in dropdowns', async () => {
    renderDialog();

    await waitFor(() => {
      expect(screen.getByTestId('relation-source-select')).toBeInTheDocument();
    });
    const sourceSelect = screen.getByTestId('relation-source-select');
    const targetSelect = screen.getByTestId('relation-target-select');

    expect(sourceSelect).toBeInTheDocument();
    expect(targetSelect).toBeInTheDocument();
  });

  it('should disable submit button when required fields are empty', async () => {
    renderDialog();

    await waitFor(() => {
      expect(screen.getByTestId('create-relation-submit')).toBeInTheDocument();
    });
    expect(submitButtonOf().disabled).toBe(true);
  });

  it('should enable submit button when source and target are pre-filled', async () => {
    renderPrefilledDialog();

    await waitFor(() => {
      expect(submitButtonOf().disabled).toBe(false);
    });
  });

  it('should pre-fill source and target when provided', async () => {
    renderPrefilledDialog();

    await waitFor(() => {
      const sourceSelect = screen.getByTestId('relation-source-select') as HTMLInputElement;
      expect(sourceSelect.value).toBe('Component A');
    });
    const targetSelect = screen.getByTestId('relation-target-select') as HTMLInputElement;
    expect(targetSelect.value).toBe('Component B');
  });

  it('should disable source and target when pre-filled', async () => {
    renderPrefilledDialog();

    await waitFor(() => {
      const sourceSelect = screen.getByTestId('relation-source-select');
      expect(sourceSelect).toBeDisabled();
    });
    const targetSelect = screen.getByTestId('relation-target-select');
    expect(targetSelect).toBeDisabled();
  });

  it('should display error when source and target are the same', async () => {
    renderDialog({ sourceComponentId: '1', targetComponentId: '1' });

    await waitFor(() => {
      expect(screen.getByTestId('create-relation-submit')).toBeInTheDocument();
    });

    const submitButton = screen.getByTestId('create-relation-submit');
    expect(submitButton).toBeDisabled();

    expect(screen.getByText('Source and target components must be different')).toBeInTheDocument();
  });

  it('should call API with valid data and close dialog', async () => {
    const captured = captureRelationPost();

    renderPrefilledDialog();

    await waitFor(() => {
      expect(screen.getByTestId('relation-name-input')).toBeInTheDocument();
    });

    const nameInput = screen.getByTestId('relation-name-input');
    const descriptionInput = screen.getByTestId('relation-description-input');

    fireEvent.change(nameInput, { target: { value: 'Test Relation' } });
    fireEvent.change(descriptionInput, { target: { value: 'Test Description' } });

    fireEvent.click(submitButtonOf());

    await waitFor(() => {
      expect(captured.request).toEqual({
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
    const captured = captureRelationPost();

    const user = userEvent.setup();
    renderPrefilledDialog();

    await waitFor(() => {
      expect(screen.getByTestId('relation-type-select')).toBeInTheDocument();
    });

    const relationTypeSelect = screen.getByTestId('relation-type-select');
    await user.click(relationTypeSelect);

    const servesOption = await screen.findByRole('option', { name: 'Serves', hidden: true });
    await user.click(servesOption);

    fireEvent.click(submitButtonOf());

    await waitFor(() => {
      expect(captured.request).toEqual(
        expect.objectContaining({
          relationType: 'Serves',
        }),
      );
    });
  });

  it('should handle create relation error', async () => {
    server.use(
      http.post(`${API_BASE}/api/v1/relations`, () => {
        return HttpResponse.json({ error: 'Network error', message: 'Network error' }, { status: 500 });
      }),
    );

    renderPrefilledDialog();

    await waitFor(() => {
      expect(screen.getByTestId('create-relation-submit')).toBeInTheDocument();
    });

    fireEvent.click(submitButtonOf());

    await waitFor(() => {
      expect(screen.getByTestId('create-relation-error')).toBeInTheDocument();
    });

    expect(mockOnClose).not.toHaveBeenCalled();
  });

  it('should close dialog when cancel is clicked', async () => {
    renderDialog();

    await waitFor(() => {
      expect(screen.getByTestId('create-relation-cancel')).toBeInTheDocument();
    });

    const cancelButton = screen.getByTestId('create-relation-cancel');
    fireEvent.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should select components via dropdown interaction', async () => {
    const captured = captureRelationPost();

    const user = userEvent.setup();
    renderDialog();

    const sourceSelect = await screen.findByTestId('relation-source-select');
    await user.click(sourceSelect);
    const sourceListboxId = sourceSelect.getAttribute('aria-controls')!;
    const sourceListbox = document.getElementById(sourceListboxId)!;
    const componentA = await within(sourceListbox).findByRole('option', { name: 'Component A', hidden: true });
    await user.click(componentA);

    const targetSelect = screen.getByTestId('relation-target-select');
    await user.click(targetSelect);
    const targetListboxId = targetSelect.getAttribute('aria-controls')!;
    const targetListbox = document.getElementById(targetListboxId)!;
    const componentB = await within(targetListbox).findByRole('option', { name: 'Component B', hidden: true });
    await user.click(componentB);

    expect(submitButtonOf().disabled).toBe(false);

    fireEvent.click(submitButtonOf());

    await waitFor(() => {
      expect(captured.request).toEqual(
        expect.objectContaining({
          sourceComponentId: '1',
          targetComponentId: '2',
        }),
      );
    });
  });

  it('should disable inputs while creating', async () => {
    server.use(
      http.post(`${API_BASE}/api/v1/relations`, async () => {
        await new Promise((resolve) => setTimeout(resolve, 100));
        return HttpResponse.json(
          {
            id: 'rel-1',
            sourceComponentId: '1',
            targetComponentId: '2',
            relationType: 'Triggers',
            createdAt: '2024-01-01T00:00:00Z',
            _links: { self: '/api/v1/relations/rel-1' },
          },
          { status: 201 },
        );
      }),
    );

    renderPrefilledDialog();

    await waitFor(() => {
      expect(screen.getByTestId('create-relation-submit')).toBeInTheDocument();
    });

    fireEvent.click(submitButtonOf());

    await waitFor(() => {
      const nameInput = screen.getByTestId('relation-name-input') as HTMLInputElement;
      expect(nameInput.disabled).toBe(true);
    });
  });
});
