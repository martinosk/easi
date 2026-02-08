import { describe, it, expect } from 'vitest';
import {
  isOriginEntityNode,
  getOriginEntityTypeFromNodeId,
  extractOriginEntityId,
  ORIGIN_ENTITY_PREFIXES,
  createOriginEntityNode,
  createAcquiredEntityNode,
  createVendorNode,
  createInternalTeamNode,
} from './nodeFactory';
import type { AcquiredEntity, Vendor, InternalTeam, AcquiredEntityId, VendorId, InternalTeamId, HATEOASLinks, IntegrationStatus } from '../../../api/types';

describe('ORIGIN_ENTITY_PREFIXES', () => {
  it('should have correct prefix for acquired entities', () => {
    expect(ORIGIN_ENTITY_PREFIXES.acquired).toBe('acq-');
  });

  it('should have correct prefix for vendors', () => {
    expect(ORIGIN_ENTITY_PREFIXES.vendor).toBe('vendor-');
  });

  it('should have correct prefix for teams', () => {
    expect(ORIGIN_ENTITY_PREFIXES.team).toBe('team-');
  });
});

describe('isOriginEntityNode', () => {
  it('should return true for acquired entity node IDs', () => {
    expect(isOriginEntityNode('acq-123')).toBe(true);
    expect(isOriginEntityNode('acq-abc-def-456')).toBe(true);
  });

  it('should return true for vendor node IDs', () => {
    expect(isOriginEntityNode('vendor-123')).toBe(true);
    expect(isOriginEntityNode('vendor-abc-def-456')).toBe(true);
  });

  it('should return true for team node IDs', () => {
    expect(isOriginEntityNode('team-123')).toBe(true);
    expect(isOriginEntityNode('team-abc-def-456')).toBe(true);
  });

  it('should return false for component node IDs', () => {
    expect(isOriginEntityNode('comp-123')).toBe(false);
    expect(isOriginEntityNode('component-456')).toBe(false);
  });

  it('should return false for capability node IDs', () => {
    expect(isOriginEntityNode('cap-123')).toBe(false);
  });

  it('should return false for plain IDs without prefix', () => {
    expect(isOriginEntityNode('123')).toBe(false);
    expect(isOriginEntityNode('abc-def')).toBe(false);
  });

  it('should return false for empty string', () => {
    expect(isOriginEntityNode('')).toBe(false);
  });

  it('should return false for partial prefix matches', () => {
    expect(isOriginEntityNode('acquire-123')).toBe(false);
    expect(isOriginEntityNode('vend-123')).toBe(false);
    expect(isOriginEntityNode('teams-123')).toBe(false);
  });
});

describe('getOriginEntityTypeFromNodeId', () => {
  it('should return acquired for acq- prefixed IDs', () => {
    expect(getOriginEntityTypeFromNodeId('acq-123')).toBe('acquired');
    expect(getOriginEntityTypeFromNodeId('acq-uuid-here')).toBe('acquired');
  });

  it('should return vendor for vendor- prefixed IDs', () => {
    expect(getOriginEntityTypeFromNodeId('vendor-123')).toBe('vendor');
    expect(getOriginEntityTypeFromNodeId('vendor-uuid-here')).toBe('vendor');
  });

  it('should return team for team- prefixed IDs', () => {
    expect(getOriginEntityTypeFromNodeId('team-123')).toBe('team');
    expect(getOriginEntityTypeFromNodeId('team-uuid-here')).toBe('team');
  });

  it('should return null for non-origin entity node IDs', () => {
    expect(getOriginEntityTypeFromNodeId('comp-123')).toBeNull();
    expect(getOriginEntityTypeFromNodeId('cap-456')).toBeNull();
    expect(getOriginEntityTypeFromNodeId('plain-id')).toBeNull();
  });

  it('should return null for empty string', () => {
    expect(getOriginEntityTypeFromNodeId('')).toBeNull();
  });

  it('should return null for partial prefix matches', () => {
    expect(getOriginEntityTypeFromNodeId('acquire-123')).toBeNull();
    expect(getOriginEntityTypeFromNodeId('vend-123')).toBeNull();
    expect(getOriginEntityTypeFromNodeId('teams-123')).toBeNull();
  });
});

describe('extractOriginEntityId', () => {
  it('should extract ID from acquired entity node ID', () => {
    expect(extractOriginEntityId('acq-123')).toBe('123');
    expect(extractOriginEntityId('acq-uuid-with-dashes')).toBe('uuid-with-dashes');
  });

  it('should extract ID from vendor node ID', () => {
    expect(extractOriginEntityId('vendor-456')).toBe('456');
    expect(extractOriginEntityId('vendor-uuid-with-dashes')).toBe('uuid-with-dashes');
  });

  it('should extract ID from team node ID', () => {
    expect(extractOriginEntityId('team-789')).toBe('789');
    expect(extractOriginEntityId('team-uuid-with-dashes')).toBe('uuid-with-dashes');
  });

  it('should return null for non-origin entity node IDs', () => {
    expect(extractOriginEntityId('comp-123')).toBeNull();
    expect(extractOriginEntityId('cap-456')).toBeNull();
    expect(extractOriginEntityId('plain-id')).toBeNull();
  });

  it('should return null for empty string', () => {
    expect(extractOriginEntityId('')).toBeNull();
  });

  it('should handle edge case of prefix-only ID', () => {
    expect(extractOriginEntityId('acq-')).toBe('');
    expect(extractOriginEntityId('vendor-')).toBe('');
    expect(extractOriginEntityId('team-')).toBe('');
  });
});

describe('createOriginEntityNode', () => {
  const emptyLayoutPositions = {};
  const layoutWithPosition = { '123': { x: 100, y: 200 } };

  const makeParams = (overrides = {}) => ({
    entityId: '123',
    entityType: 'acquired' as const,
    name: 'TechCorp',
    layoutPositions: emptyLayoutPositions,
    selectedOriginEntityId: null as string | null,
    ...overrides,
  });

  it('should create node with correct ID format for acquired entity', () => {
    const node = createOriginEntityNode(makeParams());
    expect(node.id).toBe('acq-123');
    expect(node.type).toBe('originEntity');
    expect(node.data.entityType).toBe('acquired');
  });

  it('should create node with correct ID format for vendor', () => {
    const node = createOriginEntityNode(makeParams({ entityId: '456', entityType: 'vendor', name: 'SAP' }));
    expect(node.id).toBe('vendor-456');
    expect(node.type).toBe('originEntity');
    expect(node.data.entityType).toBe('vendor');
  });

  it('should create node with correct ID format for team', () => {
    const node = createOriginEntityNode(makeParams({ entityId: '789', entityType: 'team', name: 'Platform Eng' }));
    expect(node.id).toBe('team-789');
    expect(node.type).toBe('originEntity');
    expect(node.data.entityType).toBe('team');
  });

  it('should use layout position when available', () => {
    const node = createOriginEntityNode(makeParams({ layoutPositions: layoutWithPosition }));
    expect(node.position).toEqual({ x: 100, y: 200 });
  });

  it('should use default position when not in layout', () => {
    const node = createOriginEntityNode(makeParams({ entityId: '999' }));
    expect(node.position).toEqual({ x: 400, y: 300 });
  });

  it('should set isSelected when node ID matches selectedOriginEntityId', () => {
    const node = createOriginEntityNode(makeParams({ selectedOriginEntityId: 'acq-123' }));
    expect(node.data.isSelected).toBe(true);
  });

  it('should not set isSelected when node ID does not match', () => {
    const node = createOriginEntityNode(makeParams({ selectedOriginEntityId: '456' }));
    expect(node.data.isSelected).toBe(false);
  });

  it('should include subtitle when provided', () => {
    const node = createOriginEntityNode(makeParams({ subtitle: '2021' }));
    expect(node.data.subtitle).toBe('2021');
  });

  it('should not include subtitle when not provided', () => {
    const node = createOriginEntityNode(makeParams());
    expect(node.data.subtitle).toBeUndefined();
  });
});

describe('createAcquiredEntityNode', () => {
  const mockLinks: HATEOASLinks = { self: { href: '/acquired-entities/123', method: 'GET' } };
  const emptyLayoutPositions = {};

  const createMockAcquiredEntity = (overrides = {}): AcquiredEntity => ({
    id: 'ae-123' as AcquiredEntityId,
    name: 'TechCorp',
    acquisitionDate: '2021-03-15',
    integrationStatus: 'InProgress' as IntegrationStatus,
    notes: 'Some notes',
    componentCount: 5,
    createdAt: '2021-01-01T00:00:00Z',
    _links: mockLinks,
    ...overrides,
  });

  it('should create node with correct entity ID prefix', () => {
    const entity = createMockAcquiredEntity();
    const node = createAcquiredEntityNode(entity, emptyLayoutPositions, null);
    expect(node.id).toBe('acq-ae-123');
    expect(node.type).toBe('originEntity');
  });

  it('should use entity name as label', () => {
    const entity = createMockAcquiredEntity({ name: 'AcmeCo' });
    const node = createAcquiredEntityNode(entity, emptyLayoutPositions, null);
    expect(node.data.label).toBe('AcmeCo');
  });

  it('should extract year from acquisition date for subtitle', () => {
    const entity = createMockAcquiredEntity({ acquisitionDate: '2021-03-15' });
    const node = createAcquiredEntityNode(entity, emptyLayoutPositions, null);
    expect(node.data.subtitle).toBe('2021');
  });

  it('should not include subtitle when acquisition date is undefined', () => {
    const entity = createMockAcquiredEntity({ acquisitionDate: undefined });
    const node = createAcquiredEntityNode(entity, emptyLayoutPositions, null);
    expect(node.data.subtitle).toBeUndefined();
  });

  it('should set entity type to acquired', () => {
    const entity = createMockAcquiredEntity();
    const node = createAcquiredEntityNode(entity, emptyLayoutPositions, null);
    expect(node.data.entityType).toBe('acquired');
  });
});

describe('createVendorNode', () => {
  const mockLinks: HATEOASLinks = { self: { href: '/vendors/456', method: 'GET' } };
  const emptyLayoutPositions = {};

  const createMockVendor = (overrides = {}): Vendor => ({
    id: 'v-456' as VendorId,
    name: 'SAP',
    implementationPartner: 'Accenture',
    notes: 'Some notes',
    componentCount: 3,
    createdAt: '2021-01-01T00:00:00Z',
    _links: mockLinks,
    ...overrides,
  });

  it('should create node with correct entity ID prefix', () => {
    const vendor = createMockVendor();
    const node = createVendorNode(vendor, emptyLayoutPositions, null);
    expect(node.id).toBe('vendor-v-456');
    expect(node.type).toBe('originEntity');
  });

  it('should use vendor name as label', () => {
    const vendor = createMockVendor({ name: 'Microsoft' });
    const node = createVendorNode(vendor, emptyLayoutPositions, null);
    expect(node.data.label).toBe('Microsoft');
  });

  it('should use implementation partner as subtitle', () => {
    const vendor = createMockVendor({ implementationPartner: 'Deloitte' });
    const node = createVendorNode(vendor, emptyLayoutPositions, null);
    expect(node.data.subtitle).toBe('Deloitte');
  });

  it('should not include subtitle when implementation partner is undefined', () => {
    const vendor = createMockVendor({ implementationPartner: undefined });
    const node = createVendorNode(vendor, emptyLayoutPositions, null);
    expect(node.data.subtitle).toBeUndefined();
  });

  it('should set entity type to vendor', () => {
    const vendor = createMockVendor();
    const node = createVendorNode(vendor, emptyLayoutPositions, null);
    expect(node.data.entityType).toBe('vendor');
  });
});

describe('createInternalTeamNode', () => {
  const mockLinks: HATEOASLinks = { self: { href: '/internal-teams/789', method: 'GET' } };
  const emptyLayoutPositions = {};

  const createMockInternalTeam = (overrides = {}): InternalTeam => ({
    id: 'it-789' as InternalTeamId,
    name: 'Platform Engineering',
    department: 'Technology',
    contactPerson: 'John Doe',
    notes: 'Some notes',
    componentCount: 10,
    createdAt: '2021-01-01T00:00:00Z',
    _links: mockLinks,
    ...overrides,
  });

  it('should create node with correct entity ID prefix', () => {
    const team = createMockInternalTeam();
    const node = createInternalTeamNode(team, emptyLayoutPositions, null);
    expect(node.id).toBe('team-it-789');
    expect(node.type).toBe('originEntity');
  });

  it('should use team name as label', () => {
    const team = createMockInternalTeam({ name: 'Data Team' });
    const node = createInternalTeamNode(team, emptyLayoutPositions, null);
    expect(node.data.label).toBe('Data Team');
  });

  it('should use department as subtitle', () => {
    const team = createMockInternalTeam({ department: 'Engineering' });
    const node = createInternalTeamNode(team, emptyLayoutPositions, null);
    expect(node.data.subtitle).toBe('Engineering');
  });

  it('should not include subtitle when department is undefined', () => {
    const team = createMockInternalTeam({ department: undefined });
    const node = createInternalTeamNode(team, emptyLayoutPositions, null);
    expect(node.data.subtitle).toBeUndefined();
  });

  it('should set entity type to team', () => {
    const team = createMockInternalTeam();
    const node = createInternalTeamNode(team, emptyLayoutPositions, null);
    expect(node.data.entityType).toBe('team');
  });
});
