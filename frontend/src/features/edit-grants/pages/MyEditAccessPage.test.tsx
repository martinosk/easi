import { screen } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { server } from '../../../test/mocks/server';
import type { EditGrant } from '../types';
import { MyEditAccessPage } from './MyEditAccessPage';

function grant(overrides: Partial<EditGrant>): EditGrant {
  return {
    id: 'grant-1',
    grantorId: 'grantor-id',
    grantorEmail: 'grantor@example.com',
    granteeEmail: 'grantee@example.com',
    artifactType: 'component',
    artifactId: 'comp-1',
    artifactName: 'CRM',
    scope: 'write',
    status: 'active',
    createdAt: '2026-01-01T00:00:00Z',
    expiresAt: '2026-01-31T00:00:00Z',
    _links: { artifact: { href: '/api/v1/components/comp-1', method: 'GET' } },
    ...overrides,
  };
}

function seedGrants(grants: EditGrant[]): void {
  server.use(http.get('*/api/v1/edit-grants', () => HttpResponse.json({ data: grants })));
}

describe('MyEditAccessPage', () => {
  it('links each artifact to its page in the app, never to its API href', async () => {
    seedGrants([
      grant({}),
      grant({
        id: 'grant-2',
        artifactType: 'view',
        artifactId: 'view-1',
        artifactName: 'Landscape',
        _links: { artifact: { href: '/api/v1/views/view-1', method: 'GET' } },
      }),
    ]);
    renderWithProviders(<MyEditAccessPage />);

    expect(await screen.findByRole('link', { name: 'CRM' })).toHaveAttribute('href', '/one-pagers/application/comp-1');
    expect(screen.getByRole('link', { name: 'Landscape' })).toHaveAttribute('href', '/canvas?view=view-1');
  });

  it('shows the artifact as text when the grant carries no artifact link', async () => {
    seedGrants([grant({ _links: {} })]);
    renderWithProviders(<MyEditAccessPage />);

    expect(await screen.findByText('CRM')).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'CRM' })).not.toBeInTheDocument();
  });
});
