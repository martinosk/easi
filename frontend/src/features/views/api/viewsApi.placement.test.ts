import { HttpResponse, http } from 'msw';
import { describe, expect, it } from 'vitest';
import type { CapabilityId, ComponentId, ViewId } from '../../../api/types';
import { server } from '../../../test/helpers';
import { viewsApi } from './viewsApi';

const API_BASE = 'http://localhost:8080';
const position = { x: 10, y: 20 };

let viewCounter = 0;
function freshView(): ViewId {
  viewCounter += 1;
  return `view-placement-${viewCounter}` as ViewId;
}

const positionCases = [
  {
    name: 'component',
    collection: 'components',
    elementId: 'el-component',
    add: (viewId: ViewId) => viewsApi.addComponent(viewId, { componentId: 'el-component' as ComponentId, x: 1, y: 2 }),
    move: (viewId: ViewId) => viewsApi.updateComponentPosition(viewId, 'el-component' as ComponentId, position),
  },
  {
    name: 'capability',
    collection: 'capabilities',
    elementId: 'el-capability',
    add: (viewId: ViewId) =>
      viewsApi.addCapability(viewId, { capabilityId: 'el-capability' as CapabilityId, x: 1, y: 2 }),
    move: (viewId: ViewId) => viewsApi.updateCapabilityPosition(viewId, 'el-capability' as CapabilityId, position),
  },
  {
    name: 'origin entity',
    collection: 'origin-entities',
    elementId: 'el-origin',
    add: (viewId: ViewId) => viewsApi.addOriginEntity(viewId, { originEntityId: 'el-origin', x: 1, y: 2 }),
    move: (viewId: ViewId) => viewsApi.updateOriginEntityPosition(viewId, 'el-origin', position),
  },
];

function positionUrl(viewId: ViewId, collection: string, elementId: string) {
  return `${API_BASE}/api/v1/views/${viewId}/${collection}/${elementId}/position`;
}

function gate() {
  let release!: () => void;
  const opened = new Promise<void>((resolve) => {
    release = resolve;
  });
  return { opened, release };
}

async function until(condition: () => boolean) {
  for (let i = 0; i < 100 && !condition(); i++) {
    await new Promise((r) => setTimeout(r, 5));
  }
}

async function pause() {
  await new Promise((r) => setTimeout(r, 50));
}

describe.each(positionCases)('viewsApi $name position placement sync', ({ collection, elementId, add, move }) => {
  it('treats a 404 as a no-op', async () => {
    const viewId = freshView();
    server.use(
      http.patch(positionUrl(viewId, collection, elementId), () =>
        HttpResponse.json({ message: 'not in view' }, { status: 404 }),
      ),
    );

    await expect(move(viewId)).resolves.toBeUndefined();
  });

  it('throws on a 500', async () => {
    const viewId = freshView();
    server.use(
      http.patch(positionUrl(viewId, collection, elementId), () =>
        HttpResponse.json({ message: 'boom' }, { status: 500 }),
      ),
    );

    await expect(move(viewId)).rejects.toMatchObject({ statusCode: 500 });
  });

  it('sends the position only after the in-flight add of the same element responds', async () => {
    const viewId = freshView();
    const order: string[] = [];
    const addGate = gate();
    server.use(
      http.post(`${API_BASE}/api/v1/views/${viewId}/${collection}`, async () => {
        order.push('add-received');
        await addGate.opened;
        order.push('add-responded');
        return new HttpResponse(null, { status: 201 });
      }),
      http.patch(positionUrl(viewId, collection, elementId), () => {
        order.push('position-received');
        return new HttpResponse(null, { status: 204 });
      }),
    );

    const adding = add(viewId);
    await until(() => order.includes('add-received'));
    const moving = move(viewId);
    await pause();
    expect(order).toEqual(['add-received']);

    addGate.release();
    await Promise.all([adding, moving]);

    expect(order).toEqual(['add-received', 'add-responded', 'position-received']);
  });

  it('still sends the position when the in-flight add fails', async () => {
    const viewId = freshView();
    let positionReceived = false;
    server.use(
      http.post(`${API_BASE}/api/v1/views/${viewId}/${collection}`, () =>
        HttpResponse.json({ message: 'conflict' }, { status: 409 }),
      ),
      http.patch(positionUrl(viewId, collection, elementId), () => {
        positionReceived = true;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    const adding = add(viewId).catch(() => undefined);
    await move(viewId);
    await adding;

    expect(positionReceived).toBe(true);
  });
});

describe('viewsApi.updateMultiplePositions placement sync', () => {
  it('sends the layout only after in-flight component adds respond', async () => {
    const viewId = freshView();
    const order: string[] = [];
    const addGate = gate();
    server.use(
      http.post(`${API_BASE}/api/v1/views/${viewId}/components`, async () => {
        order.push('add-received');
        await addGate.opened;
        order.push('add-responded');
        return new HttpResponse(null, { status: 201 });
      }),
      http.patch(`${API_BASE}/api/v1/views/${viewId}/layout`, () => {
        order.push('layout-received');
        return new HttpResponse(null, { status: 204 });
      }),
    );

    const adding = viewsApi.addComponent(viewId, { componentId: 'layout-comp' as ComponentId, x: 0, y: 0 });
    await until(() => order.includes('add-received'));
    const layouting = viewsApi.updateMultiplePositions(viewId, {
      positions: [{ componentId: 'layout-comp' as ComponentId, x: 5, y: 5 }],
    });
    await pause();
    expect(order).toEqual(['add-received']);

    addGate.release();
    await Promise.all([adding, layouting]);

    expect(order).toEqual(['add-received', 'add-responded', 'layout-received']);
  });

  it('does not swallow a 404 from the layout endpoint', async () => {
    const viewId = freshView();
    server.use(
      http.patch(`${API_BASE}/api/v1/views/${viewId}/layout`, () =>
        HttpResponse.json({ message: 'view missing' }, { status: 404 }),
      ),
    );

    await expect(
      viewsApi.updateMultiplePositions(viewId, { positions: [{ componentId: 'c' as ComponentId, x: 1, y: 1 }] }),
    ).rejects.toMatchObject({ statusCode: 404 });
  });
});
