import { vi } from 'vitest';
import type apiClient from '../../api/client';

export type MockedApiClient = {
  [K in keyof typeof apiClient]: ReturnType<typeof vi.fn>;
};

export function createMockApiClient(overrides?: Partial<MockedApiClient>): MockedApiClient {
  return {
    getComponents: vi.fn().mockResolvedValue([]),
    createComponent: vi.fn().mockResolvedValue({ id: '1', name: 'Test', description: '' }),
    updateComponent: vi.fn().mockResolvedValue({ id: '1', name: 'Updated', description: '' }),
    deleteComponent: vi.fn().mockResolvedValue(undefined),
    getRelations: vi.fn().mockResolvedValue([]),
    createRelation: vi.fn().mockResolvedValue({
      id: 'r1',
      sourceComponentId: '1',
      targetComponentId: '2',
      relationType: 'Triggers',
    }),
    updateRelation: vi.fn().mockResolvedValue({
      id: 'r1',
      sourceComponentId: '1',
      targetComponentId: '2',
      relationType: 'Triggers',
    }),
    deleteRelation: vi.fn().mockResolvedValue(undefined),
    getViews: vi.fn().mockResolvedValue([]),
    createView: vi.fn().mockResolvedValue({
      id: 'v1',
      name: 'Test View',
      description: '',
      components: [],
    }),
    getViewById: vi.fn().mockResolvedValue({
      id: 'v1',
      name: 'Test View',
      description: '',
      components: [],
    }),
    updateView: vi.fn().mockResolvedValue({
      id: 'v1',
      name: 'Updated View',
      description: '',
      components: [],
    }),
    deleteView: vi.fn().mockResolvedValue(undefined),
    addComponentToView: vi.fn().mockResolvedValue(undefined),
    updateComponentPosition: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  };
}
