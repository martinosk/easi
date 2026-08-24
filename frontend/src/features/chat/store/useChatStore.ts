import { create } from 'zustand';

interface ChatState {
  isOpen: boolean;
}

interface ChatActions {
  openPanel: () => void;
  closePanel: () => void;
  togglePanel: () => void;
}

export const useChatStore = create<ChatState & ChatActions>()((set) => ({
  isOpen: false,

  openPanel: () => set({ isOpen: true }),
  closePanel: () => set({ isOpen: false }),
  togglePanel: () => set((state) => ({ isOpen: !state.isOpen })),
}));
