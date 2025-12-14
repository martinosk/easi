import type { UserRole } from '../auth/types';

export type InvitationStatus = 'pending' | 'accepted' | 'expired' | 'revoked';

export interface Invitation {
  id: string;
  email: string;
  role: UserRole;
  status: InvitationStatus;
  invitedBy?: string;
  createdAt: string;
  expiresAt: string;
  acceptedAt?: string;
  revokedAt?: string;
  _links: {
    self: string;
    revoke?: string;
  };
}

export interface CreateInvitationRequest {
  email: string;
  role: UserRole;
}

export interface InvitationsListResponse {
  data: Invitation[];
  pagination: {
    hasMore: boolean;
    limit: number;
    cursor?: string;
  };
  _links: {
    self: string;
    next?: string;
  };
}
