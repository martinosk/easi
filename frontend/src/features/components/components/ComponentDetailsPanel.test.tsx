import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import type { Component, ComponentId } from '../../../api/types';
import { buildComponent, buildExpert, createMantineTestWrapper, seedDb } from '../../../test/helpers';
import { ComponentDetailsPanel } from './ComponentDetailsPanel';

function renderPanel(component: Component) {
  seedDb({ components: [component], capabilities: [] });
  const { Wrapper } = createMantineTestWrapper();
  return render(<ComponentDetailsPanel componentId={component.id} />, {
    wrapper: ({ children }) => (
      <MemoryRouter>
        <Wrapper>{children}</Wrapper>
      </MemoryRouter>
    ),
  });
}

describe('ComponentDetailsPanel', () => {
  it('renders the details of the application with the given id', async () => {
    renderPanel(buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service' }));

    expect(await screen.findByText('Billing Service')).toBeInTheDocument();
    expect(screen.getByText('Application Details')).toBeInTheDocument();
  });

  it('shows the Edit affordance when the application carries an edit link', async () => {
    renderPanel(buildComponent({ id: 'comp-1' as ComponentId }));

    expect(await screen.findByRole('button', { name: 'Edit' })).toBeInTheDocument();
  });

  it('renders read-only when the application carries no edit link', async () => {
    renderPanel(
      buildComponent({
        id: 'comp-1' as ComponentId,
        name: 'Billing Service',
        _links: { self: { href: '/api/v1/components/comp-1', method: 'GET' } },
      }),
    );

    await screen.findByText('Billing Service');
    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument();
  });

  it('offers Add Expert only when the application carries the x-add-expert link', async () => {
    renderPanel(
      buildComponent({
        id: 'comp-1' as ComponentId,
        experts: [buildExpert({ name: 'Jane' })],
        _links: {
          self: { href: '/api/v1/components/comp-1', method: 'GET' },
          'x-add-expert': { href: '/api/v1/components/comp-1/experts', method: 'POST' },
        },
      }),
    );

    expect(await screen.findByRole('button', { name: '+ Add Expert' })).toBeInTheDocument();
  });

  it('does not offer view-scoped actions outside a view context', async () => {
    renderPanel(buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service' }));

    await screen.findByText('Billing Service');
    expect(screen.queryByRole('button', { name: 'Remove from View' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('color-picker')).not.toBeInTheDocument();
  });
});
