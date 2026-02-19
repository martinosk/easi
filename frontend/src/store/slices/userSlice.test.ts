import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { act } from '@testing-library/react';
import { createUserSlice, type UserState, type UserActions } from './userSlice';
import type { SessionUser, SessionTenant } from '../../features/auth/types';

const mockUser: SessionUser = {
  id: 'user-123',
  email: 'john@acme.com',
  name: 'John Doe',
  role: 'architect',
  permissions: ['components:read', 'components:write', 'views:read', 'views:write'],
};

const mockTenant: SessionTenant = {
  id: 'acme',
  name: 'Acme Corporation',
};

const mockSessionResponse = {
  id: 'session-123',
  user: mockUser,
  tenant: mockTenant,
  expiresAt: '2025-12-02T12:00:00Z',
  _links: {
    self: '/auth/sessions/current',
    logout: '/auth/sessions/current',
    user: '/api/v1/users/user-123',
    tenant: '/api/v1/tenants/current',
  },
};

vi.mock('../../features/auth/api/authApi', () => ({
  authApi: {
    getCurrentSession: vi.fn(),
    logout: vi.fn(),
  },
}));

import { authApi } from '../../features/auth/api/authApi';

function createStore() {
  let state: UserState & UserActions;
  const setState = (partial: Partial<UserState & UserActions> | ((s: UserState & UserActions) => Partial<UserState & UserActions>)) => {
    const update = typeof partial === 'function' ? partial(state) : partial;
    state = { ...state, ...update };
  };
  const getState = () => state;

  state = createUserSlice(setState as never, getState as never, {} as never);
  return { getState, setState };
}

async function loadSession(store: ReturnType<typeof createStore>, overrides?: Record<string, unknown>) {
  vi.mocked(authApi.getCurrentSession).mockResolvedValue({ ...mockSessionResponse, ...overrides });
  await act(async () => { await store.getState().loadSession(); });
}

describe('userSlice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  it('should have initial state with no user', () => {
    const { getState } = createStore();
    expect(getState().user).toBeNull();
    expect(getState().tenant).toBeNull();
    expect(getState().isAuthenticated).toBe(false);
    expect(getState().isLoading).toBe(true);
  });

  it('should load session successfully', async () => {
    const store = createStore();
    await loadSession(store);

    expect(store.getState().user).toEqual(mockUser);
    expect(store.getState().tenant).toEqual(mockTenant);
    expect(store.getState().isAuthenticated).toBe(true);
    expect(store.getState().isLoading).toBe(false);
  });

  it('should handle session load failure', async () => {
    vi.mocked(authApi.getCurrentSession).mockRejectedValue(new Error('Unauthorized'));

    const { getState } = createStore();

    await act(async () => {
      await getState().loadSession();
    });

    expect(getState().user).toBeNull();
    expect(getState().tenant).toBeNull();
    expect(getState().isAuthenticated).toBe(false);
    expect(getState().isLoading).toBe(false);
  });

  it('should logout successfully', async () => {
    vi.mocked(authApi.logout).mockResolvedValue();
    const store = createStore();
    await loadSession(store);

    expect(store.getState().isAuthenticated).toBe(true);

    await act(async () => {
      await store.getState().logout();
    });

    expect(authApi.logout).toHaveBeenCalled();
    expect(store.getState().user).toBeNull();
    expect(store.getState().tenant).toBeNull();
    expect(store.getState().isAuthenticated).toBe(false);
  });

  it('should check if user has permission', async () => {
    const store = createStore();
    await loadSession(store);

    expect(store.getState().hasPermission('components:read')).toBe(true);
    expect(store.getState().hasPermission('users:manage')).toBe(false);
  });

  it('should store session links after loading', async () => {
    const store = createStore();
    await loadSession(store, {
      _links: {
        ...mockSessionResponse._links,
        'x-assistant': '/api/v1/assistant/conversations',
      },
    });

    expect(store.getState().sessionLinks).toEqual({
      self: '/auth/sessions/current',
      logout: '/auth/sessions/current',
      user: '/api/v1/users/user-123',
      tenant: '/api/v1/tenants/current',
      'x-assistant': '/api/v1/assistant/conversations',
    });
  });

  it('should have null sessionLinks initially', () => {
    const { getState } = createStore();
    expect(getState().sessionLinks).toBeNull();
  });

  it('should clear sessionLinks on logout', async () => {
    vi.mocked(authApi.logout).mockResolvedValue();
    const store = createStore();
    await loadSession(store, {
      _links: {
        ...mockSessionResponse._links,
        'x-assistant': '/api/v1/assistant/conversations',
      },
    });

    expect(store.getState().sessionLinks).not.toBeNull();

    await act(async () => {
      await store.getState().logout();
    });

    expect(store.getState().sessionLinks).toBeNull();
  });

  it('should clear user state', async () => {
    const store = createStore();
    await loadSession(store);

    expect(store.getState().isAuthenticated).toBe(true);

    store.getState().clearUser();

    expect(store.getState().user).toBeNull();
    expect(store.getState().tenant).toBeNull();
    expect(store.getState().isAuthenticated).toBe(false);
  });
});
