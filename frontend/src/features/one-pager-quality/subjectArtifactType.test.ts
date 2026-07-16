import { describe, expect, it } from 'vitest';
import type { OnePagerSubjectType } from '../one-pagers/types';
import { subjectArtifactType } from './subjectArtifactType';

describe('subjectArtifactType', () => {
  it.each([
    ['capability', 'capability'],
    ['application', 'component'],
    ['acquired-entity', 'acquired_entity'],
    ['vendor', 'vendor'],
    ['internal-team', 'internal_team'],
  ] as const)('maps subject type %s to artifact type %s', (subjectType, artifactType) => {
    expect(subjectArtifactType(subjectType)).toBe(artifactType);
  });

  it('has no artifact type for enterprise-capability', () => {
    expect(subjectArtifactType('enterprise-capability' as OnePagerSubjectType)).toBeUndefined();
  });
});
