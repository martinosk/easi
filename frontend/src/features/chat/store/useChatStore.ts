import { create } from 'zustand';

interface ChatState {
  isOpen: boolean;
  conversationId: string | null;
  yoloEnabled: boolean;
}

interface ChatActions {
  openPanel: () => void;
  closePanel: () => void;
  togglePanel: () => void;
  setConversationId: (id: string | null) => void;
  toggleYolo: () => void;
}

export const useChatStore = create<ChatState & ChatActions>()((set) => ({
  isOpen: false,
  conversationId: null,
  yoloEnabled: false,

  openPanel: () => set({ isOpen: true }),
  closePanel: () => set({ isOpen: false }),
  togglePanel: () => set((state) => ({ isOpen: !state.isOpen })),
  setConversationId: (id) => set({ conversationId: id }),
  toggleYolo: () => set((state) => ({ yoloEnabled: !state.yoloEnabled })),
}));
