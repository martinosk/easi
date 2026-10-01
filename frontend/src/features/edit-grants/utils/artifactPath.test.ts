import { describe, expect, it } from 'vitest';
import type { ArtifactType } from '../types';
import { artifactPath } from './artifactPath';

describe('artifactPath', () => {
  it.each([
    ['capability', 'cap-1', '/one-pagers/capability/cap-1'],
    ['component', 'comp-1', '/one-pagers/application/comp-1'],
    ['vendor', 'vendor-1', '/one-pagers/vendor/vendor-1'],
    ['internal_team', 'team-1', '/one-pagers/internal-team/team-1'],
    ['acquired_entity', 'acq-1', '/one-pagers/acquired-entity/acq-1'],
    ['view', 'view-1', '/canvas?view=view-1'],
    ['domain', 'dom-1', '/business-domains/dom-1'],
  ] as const)('leads a %s to its own page', (artifactType, artifactId, expected) => {
    expect(artifactPath({ artifactType, artifactId })).toBe(expected);
  });

  it('leads an artifact type without a page of its own to the canvas', () => {
    expect(artifactPath({ artifactType: 'journey' as ArtifactType, artifactId: 'j-1' })).toBe('/canvas');
  });

  it('never returns an API path', () => {
    expect(artifactPath({ artifactType: 'component', artifactId: 'comp-1' })).not.toContain('/api/');
  });
});
