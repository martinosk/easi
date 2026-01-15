export interface InitiateLoginRequest {
  email: string;
  returnUrl?: string;
}

export interface InitiateLoginResponse {
  _links: {
    self: string;
    authorize: string;
  };
}

export type UserRole = 'admin' | 'architect' | 'stakeholder';

export interface SessionUser {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  permissions: string[];
}

export interface SessionTenant {
  id: string;
  name: string;
}

export interface CurrentSessionResponse {
  id: string;
  user: SessionUser;
  tenant: SessionTenant;
  expiresAt: string;
  _links: {
    self: string;
    logout: string;
    user: string;
    tenant: string;
  };
}
