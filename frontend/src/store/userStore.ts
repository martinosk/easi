import { create } from 'zustand';
import { createUserSlice, type UserState, type UserActions } from './slices/userSlice';

export type UserStore = UserState & UserActions;

export const useUserStore = create<UserStore>()((...args) => ({
  ...createUserSlice(...args),
}));

export type { SessionUser, SessionTenant, SessionLinks, UserRole } from '../features/auth/types';
