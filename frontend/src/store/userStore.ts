import { create } from 'zustand';
import { createUserSlice, type UserActions, type UserState } from './slices/userSlice';

export type UserStore = UserState & UserActions;

export const useUserStore = create<UserStore>()((...args) => ({
  ...createUserSlice(...args),
}));

export type { SessionLinks, SessionTenant, SessionUser, UserRole } from '../features/auth/types';
