import type { ArtifactType } from '../edit-grants/types';
import type { OnePagerSubjectType } from '../one-pagers/types';

const SUBJECT_TYPE_ARTIFACT_TYPE: Partial<Record<OnePagerSubjectType, ArtifactType>> = {
  capability: 'capability',
  application: 'component',
  'acquired-entity': 'acquired_entity',
  vendor: 'vendor',
  'internal-team': 'internal_team',
};

export function subjectArtifactType(subjectType: OnePagerSubjectType): ArtifactType | undefined {
  return SUBJECT_TYPE_ARTIFACT_TYPE[subjectType];
}
