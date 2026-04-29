import { HttpResponse, http } from 'msw';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { HATEOASLinks } from '../../../../api/types';
import { createMantineTestWrapper } from '../../../../test/helpers';
import { server } from '../../../../test/mocks/server';
import {
  CreateConnectedEntityDialog,
  type CreateConnectedEntityDialogProps,
} from './CreateConnectedEntityDialog';

Element.prototype.scrollIntoView = vi.fn();

const NEW_COMPONENT_ID = 'new-comp-123';
const SOURCE_NODE_ID = 'source-node-1';

const baseProps: CreateConnectedEntityDialogProps = {
  isOpen: true,
  sourceNodeId: SOURCE_NODE_ID,
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

describe('CreateConnectedEntityDialog integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('happy path — relation creation', () => {
    it('submits correct data for creating a related component', async () => {
      const onSubmit = vi.fn();
      const links: HATEOASLinks = {
        'x-add-relation': { href: '/api/v1/relations', method: 'POST' },
        'x-relations-from': { href: `/api/v1/relations/from/${SOURCE_NODE_ID}`, method: 'GET' },
        'x-relations-to': { href: `/api/v1/relations/to/${SOURCE_NODE_ID}`, method: 'GET' },
        'x-origins': { href: `/api/v1/components/${SOURCE_NODE_ID}/origins`, method: 'GET' },
      };

      renderDialog({ links, onSubmit });

      const user = userEvent.setup();

      await user.type(screen.getByTestId('connected-entity-name-input'), 'Payment Service');

      await waitFor(() => {
        expect(screen.getByTestId('create-connected-entity-submit')).not.toBeDisabled();
      });

      expect(screen.getByTestId('connected-entity-relation-type-select')).toBeInTheDocument();

      await user.click(screen.getByTestId('create-connected-entity-submit'));

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith(
          expect.objectContaining({
            name: 'Payment Service',
            actionType: 'x-add-relation',
            relationType: 'Triggers',
            actionLink: { href: '/api/v1/relations', method: 'POST' },
          }),
        );
      });
    });

    it('defaults relation type to Triggers when x-add-relation is selected', async () => {
      const onSubmit = vi.fn();
      const links: HATEOASLinks = {
        'x-add-relation': { href: '/api/v1/relations', method: 'POST' },
      };

      renderDialog({ links, onSubmit });
      const user = userEvent.setup();

      await user.type(screen.getByTestId('connected-entity-name-input'), 'Auth Service');

      const relationSelect = screen.getByTestId('connected-entity-relation-type-select');
      expect(relationSelect).toHaveValue('Triggers');

      await waitFor(() => {
        expect(screen.getByTestId('create-connected-entity-submit')).not.toBeDisabled();
      });

      await user.click(screen.getByTestId('create-connected-entity-submit'));

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith(
          expect.objectContaining({
            name: 'Auth Service',
            actionType: 'x-add-relation',
            relationType: 'Triggers',
            actionLink: { href: '/api/v1/relations', method: 'POST' },
          }),
        );
      });
    });

    it('closes dialog after successful submit', async () => {
      const onClose = vi.fn();
      const onSubmit = vi.fn();
      const links: HATEOASLinks = {
        'x-add-relation': { href: '/api/v1/relations', method: 'POST' },
      };

      renderDialog({ links, onSubmit, onClose });
      const user = userEvent.setup();

      await user.type(screen.getByTestId('connected-entity-name-input'), 'New Service');

      await waitFor(() => {
        expect(screen.getByTestId('create-connected-entity-submit')).not.toBeDisabled();
      });

      await user.click(screen.getByTestId('create-connected-entity-submit'));

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalled();
      });
    });
  });

  describe('happy path — origin creation', () => {
    it('shows only Built by option when x-set-origin-built-by is the sole write link', () => {
      const links: HATEOASLinks = {
        'x-set-origin-built-by': {
          href: `/api/v1/components/${SOURCE_NODE_ID}/origins/built-by`,
          method: 'PUT',
        },
        'x-relations-from': { href: `/api/v1/relations/from/${SOURCE_NODE_ID}`, method: 'GET' },
        'x-relations-to': { href: `/api/v1/relations/to/${SOURCE_NODE_ID}`, method: 'GET' },
        'x-origins': { href: `/api/v1/components/${SOURCE_NODE_ID}/origins`, method: 'GET' },
      };

      renderDialog({ links });

      expect(screen.queryByTestId('connected-entity-action-select')).not.toBeInTheDocument();
      expect(screen.queryByTestId('connected-entity-relation-type-select')).not.toBeInTheDocument();
      expect(screen.getByTestId('connected-entity-name-input')).toBeInTheDocument();
    });

    it('submits correct data for origin built-by creation', async () => {
      const onSubmit = vi.fn();
      const links: HATEOASLinks = {
        'x-set-origin-built-by': {
          href: `/api/v1/components/${SOURCE_NODE_ID}/origins/built-by`,
          method: 'PUT',
        },
      };

      renderDialog({ links, onSubmit });
      const user = userEvent.setup();

      await user.type(screen.getByTestId('connected-entity-name-input'), 'Engineering Team');

      await waitFor(() => {
        expect(screen.getByTestId('create-connected-entity-submit')).not.toBeDisabled();
      });

      await user.click(screen.getByTestId('create-connected-entity-submit'));

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith(
          expect.objectContaining({
            name: 'Engineering Team',
            actionType: 'x-set-origin-built-by',
            relationType: undefined,
            actionLink: {
              href: `/api/v1/components/${SOURCE_NODE_ID}/origins/built-by`,
              method: 'PUT',
            },
          }),
        );
      });
    });
  });

  describe('permission-gated — read-only user', () => {
    it('shows no-actions state when only read links are present', () => {
      const links: HATEOASLinks = {
        'x-relations-from': { href: `/api/v1/relations/from/${SOURCE_NODE_ID}`, method: 'GET' },
        'x-relations-to': { href: `/api/v1/relations/to/${SOURCE_NODE_ID}`, method: 'GET' },
        'x-origins': { href: `/api/v1/components/${SOURCE_NODE_ID}/origins`, method: 'GET' },
      };

      renderDialog({ links });

      expect(screen.getByTestId('no-actions-message')).toHaveTextContent('No actions available');
      expect(screen.queryByTestId('connected-entity-name-input')).not.toBeInTheDocument();
      expect(screen.queryByTestId('create-connected-entity-submit')).not.toBeInTheDocument();
    });

    it('shows no-actions state when links object is empty', () => {
      renderDialog({ links: {} });

      expect(screen.getByTestId('no-actions-message')).toHaveTextContent('No actions available');
    });

    it('does not make any API calls for read-only user', () => {
      const onSubmit = vi.fn();
      const links: HATEOASLinks = {
        'x-relations-from': { href: `/api/v1/relations/from/${SOURCE_NODE_ID}`, method: 'GET' },
        'x-relations-to': { href: `/api/v1/relations/to/${SOURCE_NODE_ID}`, method: 'GET' },
        'x-origins': { href: `/api/v1/components/${SOURCE_NODE_ID}/origins`, method: 'GET' },
      };

      renderDialog({ links, onSubmit });

      expect(onSubmit).not.toHaveBeenCalled();
    });
  });
});
