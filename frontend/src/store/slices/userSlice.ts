import type { StateCreator } from 'zustand';
import type { SessionUser, SessionTenant } from '../../features/auth/types';
import { authApi } from '../../features/auth/api/authApi';

export interface UserState {
  user: SessionUser | null;
  tenant: SessionTenant | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface UserActions {
  loadSession: () => Promise<void>;
  logout: () => Promise<void>;
  clearUser: () => void;
  hasPermission: (permission: string) => boolean;
}

export const createUserSlice: StateCreator<
  UserState & UserActions,
  [],
  [],
  UserState & UserActions
> = (set, get) => ({
  user: null,
  tenant: null,
  isAuthenticated: false,
  isLoading: true,

  loadSession: async () => {
    try {
      set({ isLoading: true });
      const session = await authApi.getCurrentSession();
      set({
        user: session.user,
        tenant: session.tenant,
        isAuthenticated: true,
        isLoading: false,
      });
    } catch {
      set({
        user: null,
        tenant: null,
        isAuthenticated: false,
        isLoading: false,
      });
    }
  },

  logout: async () => {
    try {
      await authApi.logout();
    } finally {
      set({
        user: null,
        tenant: null,
        isAuthenticated: false,
      });
    }
  },

  clearUser: () => {
    set({
      user: null,
      tenant: null,
      isAuthenticated: false,
    });
  },

  hasPermission: (permission: string) => {
    const { user } = get();
    if (!user) return false;
    return user.permissions.includes(permission);
  },
});
