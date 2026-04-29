import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { HATEOASLinks } from '../../../../api/types';
import { createMantineTestWrapper } from '../../../../test/helpers';
import { CreateConnectedEntityDialog, type CreateConnectedEntityDialogProps } from './CreateConnectedEntityDialog';

Element.prototype.scrollIntoView = vi.fn();

const baseProps: CreateConnectedEntityDialogProps = {
  isOpen: true,
  sourceNodeId: 'node-1',
  sourceNodeType: 'component',
  handlePosition: 'right',
  links: {},
  onSubmit: vi.fn(),
  onClose: vi.fn(),
};

function renderDialog(overrides: Partial<CreateConnectedEntityDialogProps> = {}) {
  const props = { ...baseProps, ...overrides };
  const { Wrapper } = createMantineTestWrapper();
  return render(<CreateConnectedEntityDialog {...props} />, { wrapper: Wrapper });
}

describe('CreateConnectedEntityDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows "No actions available" when links is empty', () => {
    renderDialog({ links: {} });
    expect(screen.getByTestId('no-actions-message')).toHaveTextContent('No actions available');
  });

  it('shows relation creation option when x-add-relation link is present', () => {
    const links: HATEOASLinks = {
      'x-add-relation': { href: '/api/v1/relations', method: 'POST' },
    };
    renderDialog({ links });

    expect(screen.queryByTestId('no-actions-message')).not.toBeInTheDocument();
    expect(screen.getByTestId('connected-entity-name-input')).toBeInTheDocument();
  });

  it('shows relation type selector when x-add-relation is the selected action', async () => {
    const links: HATEOASLinks = {
      'x-add-relation': { href: '/api/v1/relations', method: 'POST' },
    };
    renderDialog({ links });

    await waitFor(() => {
      expect(screen.getByTestId('connected-entity-relation-type-select')).toBeInTheDocument();
    });
  });

  it('does NOT show relation type selector when only origin links present', () => {
    const links: HATEOASLinks = {
      'x-set-origin-built-by': { href: '/api/v1/components/node-1/origins/built-by', method: 'POST' },
    };
    renderDialog({ links });

    expect(screen.queryByTestId('connected-entity-relation-type-select')).not.toBeInTheDocument();
  });

  it('shows "Built by" option when x-set-origin-built-by link is present', () => {
    const links: HATEOASLinks = {
      'x-set-origin-built-by': { href: '/api/v1/components/node-1/origins/built-by', method: 'POST' },
      'x-add-relation': { href: '/api/v1/relations', method: 'POST' },
    };
    renderDialog({ links });

    expect(screen.getByTestId('connected-entity-action-select')).toBeInTheDocument();
  });

  it('does NOT show relation option when x-add-relation is absent', () => {
    const links: HATEOASLinks = {
      'x-set-origin-built-by': { href: '/api/v1/components/node-1/origins/built-by', method: 'POST' },
    };
    renderDialog({ links });

    expect(screen.queryByTestId('connected-entity-action-select')).not.toBeInTheDocument();
    expect(screen.getByTestId('connected-entity-name-input')).toBeInTheDocument();
  });

  it('calls onClose when Cancel button is clicked', async () => {
    const onClose = vi.fn();
    const links: HATEOASLinks = {
      'x-add-relation': { href: '/api/v1/relations', method: 'POST' },
    };
    renderDialog({ links, onClose });

    const user = userEvent.setup();
    await user.click(screen.getByTestId('create-connected-entity-cancel'));

    expect(onClose).toHaveBeenCalled();
  });

  it('calls onClose from the empty-state Close button', async () => {
    const onClose = vi.fn();
    renderDialog({ links: {}, onClose });

    const user = userEvent.setup();
    await user.click(screen.getByTestId('create-connected-entity-close'));

    expect(onClose).toHaveBeenCalled();
  });

  it('calls onSubmit with correct payload when form is submitted', async () => {
    const onSubmit = vi.fn();
    const links: HATEOASLinks = {
      'x-add-relation': { href: '/api/v1/relations', method: 'POST' },
    };
    renderDialog({ links, onSubmit });

    const user = userEvent.setup();
    await user.type(screen.getByTestId('connected-entity-name-input'), 'New Component');

    await waitFor(() => {
      expect(screen.getByTestId('create-connected-entity-submit')).not.toBeDisabled();
    });

    await user.click(screen.getByTestId('create-connected-entity-submit'));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'New Component',
          actionType: 'x-add-relation',
          relationType: 'Triggers',
          actionLink: { href: '/api/v1/relations', method: 'POST' },
        }),
      );
    });
  });
});
