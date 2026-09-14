import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { HttpResponse, http } from 'msw';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ComponentId, HATEOASLinks } from '../../../api/types';
import { createMantineTestWrapper, seedDb, server } from '../../../test/helpers';
import { buildComponent } from '../../../test/helpers/entityBuilders';
import { CreateRelationDialog } from './CreateRelationDialog';

const API_BASE = 'http://localhost:8080';

Element.prototype.scrollIntoView = vi.fn();

const attachableLinks = (id: string): HATEOASLinks => ({
  self: { href: `/api/v1/components/${id}`, method: 'GET' },
  'x-attach-to': { href: `/api/v1/components/${id}/containment`, method: 'PUT' },
});

const quoting = buildComponent({ id: 'quoting' as ComponentId, name: 'Quoting', _links: attachableLinks('quoting') });
const crm = buildComponent({ id: 'crm' as ComponentId, name: 'CRM Suite', _links: attachableLinks('crm') });
const billing = buildComponent({
  id: 'billing' as ComponentId,
  name: 'Billing',
  partOf: { id: 'crm' as ComponentId, name: 'CRM Suite', kind: 'aggregation' },
});
const readOnly = buildComponent({
  id: 'readonly' as ComponentId,
  name: 'Read Only',
  _links: { self: { href: '/api/v1/components/readonly', method: 'GET' } },
});

function renderDialog(sourceComponentId: string, targetComponentId: string) {
  const { Wrapper } = createMantineTestWrapper();
  const onClose = vi.fn();
  render(
    <CreateRelationDialog
      isOpen
      onClose={onClose}
      sourceComponentId={sourceComponentId}
      targetComponentId={targetComponentId}
    />,
    { wrapper: Wrapper },
  );
  return onClose;
}

async function openKindOptions() {
  await waitFor(() => expect(screen.getByTestId('relation-type-select')).toBeInTheDocument());
  await userEvent.click(screen.getByTestId('relation-type-select'));
}

describe('CreateRelationDialog containment', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    seedDb({ components: [quoting, crm, billing, readOnly] });
  });

  it('offers both containment kinds when the source can attach and the target is not a part', async () => {
    renderDialog('quoting', 'crm');

    await openKindOptions();

    expect(await screen.findByRole('option', { name: 'Part of (composition)', hidden: true })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'Part of (aggregation)', hidden: true })).toBeInTheDocument();
  });

  it('offers no containment kind when the source lacks the attach affordance', async () => {
    renderDialog('readonly', 'crm');

    await openKindOptions();

    expect(await screen.findByRole('option', { name: 'Triggers', hidden: true })).toBeInTheDocument();
    expect(screen.queryByRole('option', { name: 'Part of (composition)', hidden: true })).not.toBeInTheDocument();
  });

  it('offers no containment kind when the target is already a part', async () => {
    renderDialog('quoting', 'billing');

    await openKindOptions();

    expect(await screen.findByRole('option', { name: 'Serves', hidden: true })).toBeInTheDocument();
    expect(screen.queryByRole('option', { name: 'Part of (aggregation)', hidden: true })).not.toBeInTheDocument();
  });

  it('attaches the source as a part of the target through its attach link', async () => {
    const captured: { url: string | null; body: Record<string, unknown> | null } = { url: null, body: null };
    server.use(
      http.put(`${API_BASE}/api/v1/components/:id/containment`, async ({ request }) => {
        captured.url = request.url;
        captured.body = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json({ ...quoting, partOf: { id: 'crm', name: 'CRM Suite', kind: 'composition' } });
      }),
    );
    const onClose = renderDialog('quoting', 'crm');

    await openKindOptions();
    await userEvent.click(await screen.findByRole('option', { name: 'Part of (composition)', hidden: true }));
    expect(screen.queryByTestId('relation-name-input')).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Attach' }));

    await waitFor(() => expect(onClose).toHaveBeenCalled());
    expect(captured.url).toBe(`${API_BASE}/api/v1/components/quoting/containment`);
    expect(captured.body).toEqual({ parentId: 'crm', kind: 'composition' });
  });
});
