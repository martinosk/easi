import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import type { Capability, CapabilityId } from '../../../api/types';
import { buildCapability, buildExpert, createMantineTestWrapper, seedDb } from '../../../test/helpers';
import { CapabilityDetailsPanel } from './CapabilityDetailsPanel';

function renderPanel(capability: Capability) {
  seedDb({ capabilities: [capability], components: [] });
  const { Wrapper } = createMantineTestWrapper();
  return render(<CapabilityDetailsPanel capabilityId={capability.id} />, {
    wrapper: ({ children }) => (
      <MemoryRouter>
        <Wrapper>{children}</Wrapper>
      </MemoryRouter>
    ),
  });
}

describe('CapabilityDetailsPanel', () => {
  it('renders the details of the capability with the given id', async () => {
    renderPanel(buildCapability({ id: 'cap-1' as CapabilityId, name: 'Order Management' }));

    expect(await screen.findByText('Order Management')).toBeInTheDocument();
    expect(screen.getByText('Capability Details')).toBeInTheDocument();
  });

  it('renders the EA owner display name instead of the stored user id', async () => {
    renderPanel(
      buildCapability({
        id: 'cap-1' as CapabilityId,
        name: 'Order Management',
        eaOwner: '2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b',
        eaOwnerName: 'Alice Smith',
      }),
    );

    expect(await screen.findByText('Alice Smith')).toBeInTheDocument();
    expect(screen.queryByText('2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b')).not.toBeInTheDocument();
  });

  it('falls back to the stored EA owner value when no display name is provided', async () => {
    renderPanel(
      buildCapability({
        id: 'cap-1' as CapabilityId,
        name: 'Order Management',
        eaOwner: 'Bob Jones',
      }),
    );

    expect(await screen.findByText('Bob Jones')).toBeInTheDocument();
  });

  it('shows the Edit affordance when the capability carries an edit link', async () => {
    renderPanel(buildCapability({ id: 'cap-1' as CapabilityId }));

    expect(await screen.findByRole('button', { name: 'Edit' })).toBeInTheDocument();
  });

  it.each([['Edit'], ['+ Add Expert']])(
    'does not offer the %s affordance when the capability carries no corresponding link',
    async (buttonName) => {
      renderPanel(
        buildCapability({
          id: 'cap-1' as CapabilityId,
          name: 'Order Management',
          _links: { self: { href: '/api/v1/capabilities/cap-1', method: 'GET' } },
        }),
      );

      await screen.findByText('Order Management');
      expect(screen.queryByRole('button', { name: buttonName })).not.toBeInTheDocument();
    },
  );

  it('offers Add Expert only when the capability carries the x-add-expert link', async () => {
    renderPanel(
      buildCapability({
        id: 'cap-1' as CapabilityId,
        experts: [buildExpert({ name: 'Jane' })],
        _links: {
          self: { href: '/api/v1/capabilities/cap-1', method: 'GET' },
          'x-add-expert': { href: '/api/v1/capabilities/cap-1/experts', method: 'POST' },
        },
      }),
    );

    expect(await screen.findByRole('button', { name: '+ Add Expert' })).toBeInTheDocument();
  });

  it('does not offer view-scoped actions outside a view context', async () => {
    renderPanel(buildCapability({ id: 'cap-1' as CapabilityId, name: 'Order Management' }));

    await screen.findByText('Order Management');
    expect(screen.queryByRole('button', { name: 'Remove from View' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('color-picker')).not.toBeInTheDocument();
  });
});
