import type { UserRole } from '../auth/types';

export type UserStatus = 'active' | 'disabled';

export interface User {
  id: string;
  email: string;
  name?: string;
  role: UserRole;
  status: UserStatus;
  createdAt: string;
  lastLoginAt?: string;
  invitedBy?: {
    id: string;
    email: string;
  };
  _links: {
    self: string;
    update?: string;
  };
}

export interface UpdateUserRequest {
  role?: UserRole;
  status?: UserStatus;
}

export interface UsersListResponse {
  data: User[];
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
