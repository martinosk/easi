import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { ComponentId, HATEOASLinks } from '../../../../api/types';
import { renderWithProviders } from '../../../../test/helpers';
import type { EdgeContextMenu as EdgeContextMenuState } from '../../hooks/useEdgeContextMenu';
import { EdgeContextMenu } from './EdgeContextMenu';

const detachable: HATEOASLinks = {
  self: { href: '/api/v1/components/quoting', method: 'GET' },
  'x-detach': { href: '/api/v1/components/quoting/containment', method: 'DELETE' },
};

const containmentMenu = (links: HATEOASLinks | undefined): EdgeContextMenuState => ({
  x: 10,
  y: 10,
  edgeId: 'containment-crm-quoting',
  edgeName: 'CRM Suite contains Quoting',
  edgeType: 'containment',
  componentId: 'quoting' as ComponentId,
  _links: links,
});

describe('EdgeContextMenu for a containment edge', () => {
  it('offers Detach Part when the part carries the detach affordance', async () => {
    const onRequestDelete = vi.fn();
    const onClose = vi.fn();

    renderWithProviders(
      <EdgeContextMenu menu={containmentMenu(detachable)} onClose={onClose} onRequestDelete={onRequestDelete} />,
      { withRouter: false },
    );

    await userEvent.click(screen.getByRole('menuitem', { name: 'Detach part from parent' }));

    expect(onRequestDelete).toHaveBeenCalledWith({
      type: 'containment',
      id: 'containment-crm-quoting',
      name: 'CRM Suite contains Quoting',
      componentId: 'quoting',
    });
    expect(onClose).toHaveBeenCalled();
  });

  it('offers no menu item when the part cannot be detached', () => {
    renderWithProviders(
      <EdgeContextMenu
        menu={containmentMenu({ self: { href: '/api/v1/components/quoting', method: 'GET' } })}
        onClose={vi.fn()}
        onRequestDelete={vi.fn()}
      />,
      { withRouter: false },
    );

    expect(screen.queryByRole('menuitem')).not.toBeInTheDocument();
  });
});
