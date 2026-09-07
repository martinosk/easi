import { HttpResponse, http } from 'msw';
import { describe, expect, it } from 'vitest';
import type { CapabilityId } from '../../../api/types';
import { buildCapability, server } from '../../../test/helpers';
import { capabilitiesApi } from './capabilitiesApi';

const API_BASE = 'http://localhost:8080';

describe('capabilitiesApi link following', () => {
  it('updates metadata through the x-update-metadata link', async () => {
    const capability = buildCapability({
      id: 'cap-1' as CapabilityId,
      _links: {
        self: { href: '/api/v1/capabilities/cap-1', method: 'GET' },
        'x-update-metadata': { href: '/api/v1/linked/cap-1/metadata', method: 'PUT' },
      },
    });
    let received: unknown;
    server.use(
      http.put(`${API_BASE}/api/v1/linked/cap-1/metadata`, async ({ request }) => {
        received = await request.json();
        return HttpResponse.json({ ...capability, status: 'Retiring' });
      }),
    );

    const updated = await capabilitiesApi.updateMetadata(capability, { status: 'Retiring', maturityValue: 12 });

    expect(received).toEqual({ status: 'Retiring', maturityValue: 12 });
    expect(updated.status).toBe('Retiring');
  });

  it('adds a tag through the x-add-tag link', async () => {
    const capability = buildCapability({
      id: 'cap-1' as CapabilityId,
      _links: {
        self: { href: '/api/v1/capabilities/cap-1', method: 'GET' },
        'x-add-tag': { href: '/api/v1/linked/cap-1/tags', method: 'POST' },
      },
    });
    let received: unknown;
    server.use(
      http.post(`${API_BASE}/api/v1/linked/cap-1/tags`, async ({ request }) => {
        received = await request.json();
        return new HttpResponse(null, { status: 204 });
      }),
    );

    await capabilitiesApi.addTag(capability, { tag: 'core' });

    expect(received).toEqual({ tag: 'core' });
  });
});
