import { HttpResponse, http } from 'msw';
import { describe, expect, it } from 'vitest';
import type { CapabilityId, ComponentId, ViewId } from '../../../api/types';
import { server } from '../../../test/helpers';
import { viewsApi } from '../../views/api';
import { type DraftSaveApi, saveDraft } from './saveDraft';

const API_BASE = 'http://localhost:8080';
const viewId = 'view-draft-real-api' as ViewId;

const api: DraftSaveApi = {
  addComponent: (v, id, x, y) => viewsApi.addComponent(v, { componentId: id as ComponentId, x, y }),
  addCapability: (v, id, x, y) => viewsApi.addCapability(v, { capabilityId: id as CapabilityId, x, y }),
  addOriginEntity: (v, id, x, y) => viewsApi.addOriginEntity(v, { originEntityId: id, x, y }),
  removeComponent: (v, id) => viewsApi.removeComponent(v, id as ComponentId),
  removeCapability: (v, id) => viewsApi.removeCapability(v, id as CapabilityId),
  removeOriginEntity: (v, id) => viewsApi.removeOriginEntity(v, id),
  updateComponentPosition: (v, id, x, y) => viewsApi.updateComponentPosition(v, id as ComponentId, { x, y }),
  updateCapabilityPosition: (v, id, x, y) => viewsApi.updateCapabilityPosition(v, id as CapabilityId, { x, y }),
  updateOriginEntityPosition: (v, id, x, y) => viewsApi.updateOriginEntityPosition(v, id, { x, y }),
};

describe('saveDraft against the real viewsApi', () => {
  it('counts a position 404 as success', async () => {
    server.use(
      http.patch(`${API_BASE}/api/v1/views/${viewId}/components/comp-gone/position`, () =>
        HttpResponse.json({ message: 'not in view' }, { status: 404 }),
      ),
    );

    const result = await saveDraft(api, {
      viewId,
      additions: [],
      removals: [],
      positionDeltas: [{ id: 'comp-gone', type: 'component', x: 1, y: 2 }],
    });

    expect(result).toEqual({ successCount: 1, failures: [] });
  });

  it('still reports a position 500 as a failure', async () => {
    server.use(
      http.patch(`${API_BASE}/api/v1/views/${viewId}/capabilities/cap-1/position`, () =>
        HttpResponse.json({ message: 'boom' }, { status: 500 }),
      ),
    );

    const result = await saveDraft(api, {
      viewId,
      additions: [],
      removals: [],
      positionDeltas: [{ id: 'cap-1', type: 'capability', x: 1, y: 2 }],
    });

    expect(result.failures).toHaveLength(1);
  });

  it('keeps counting add 409 and remove 404 as success', async () => {
    server.use(
      http.post(`${API_BASE}/api/v1/views/${viewId}/components`, () =>
        HttpResponse.json({ message: 'already in view' }, { status: 409 }),
      ),
      http.delete(`${API_BASE}/api/v1/views/${viewId}/capabilities/cap-gone`, () =>
        HttpResponse.json({ message: 'not in view' }, { status: 404 }),
      ),
    );

    const result = await saveDraft(api, {
      viewId,
      additions: [{ id: 'comp-dup', type: 'component', x: 0, y: 0 }],
      removals: [{ id: 'cap-gone', type: 'capability' }],
      positionDeltas: [],
    });

    expect(result).toEqual({ successCount: 2, failures: [] });
  });
});
