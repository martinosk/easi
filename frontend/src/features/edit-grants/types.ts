import type { HATEOASLinks } from '../../api/types';

export type EditGrantStatus = 'active' | 'revoked' | 'expired';
export type ArtifactType = 'capability' | 'component' | 'view' | 'domain' | 'vendor' | 'internal_team' | 'acquired_entity';
export type GrantScope = 'write';

export interface EditGrant {
  id: string;
  grantorId: string;
  grantorEmail: string;
  granteeEmail: string;
  artifactType: ArtifactType;
  artifactId: string;
  artifactName?: string;
  scope: GrantScope;
  status: EditGrantStatus;
  reason?: string;
  invitationCreated?: boolean;
  createdAt: string;
  expiresAt: string;
  revokedAt?: string;
  _links: HATEOASLinks;
}

export interface CreateEditGrantRequest {
  granteeEmail: string;
  artifactType: ArtifactType;
  artifactId: string;
  scope?: GrantScope;
  reason?: string;
}
