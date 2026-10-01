import { generatePath } from 'react-router-dom';
import { generateViewPath } from '../../../lib/deepLinks';
import { ROUTES } from '../../../routes/routePaths';
import type { OnePagerSubjectType } from '../../one-pagers/types';
import type { ArtifactType, EditGrant } from '../types';

const ONE_PAGER_SUBJECT: Partial<Record<ArtifactType, OnePagerSubjectType>> = {
  capability: 'capability',
  component: 'application',
  vendor: 'vendor',
  internal_team: 'internal-team',
  acquired_entity: 'acquired-entity',
};

export function artifactPath({ artifactType, artifactId }: Pick<EditGrant, 'artifactType' | 'artifactId'>): string {
  const subjectType = ONE_PAGER_SUBJECT[artifactType];
  if (subjectType) return generatePath(ROUTES.ONE_PAGER_DETAIL, { subjectType, subjectId: artifactId });
  if (artifactType === 'view') return generateViewPath(artifactId);
  if (artifactType === 'domain') return generatePath(ROUTES.BUSINESS_DOMAIN_DETAIL, { domainId: artifactId });
  return ROUTES.CANVAS;
}
