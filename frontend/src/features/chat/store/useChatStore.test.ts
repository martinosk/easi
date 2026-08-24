import { beforeEach, describe, expect, it } from 'vitest';
import { useChatStore } from './useChatStore';

describe('useChatStore', () => {
  beforeEach(() => {
    useChatStore.setState({
      isOpen: false,
    });
  });

  it('should start with panel closed', () => {
    expect(useChatStore.getState().isOpen).toBe(false);
  });

  it('should open the panel', () => {
    useChatStore.getState().openPanel();
    expect(useChatStore.getState().isOpen).toBe(true);
  });

  it('should close the panel', () => {
    useChatStore.getState().openPanel();
    useChatStore.getState().closePanel();
    expect(useChatStore.getState().isOpen).toBe(false);
  });

  it('should toggle the panel', () => {
    useChatStore.getState().togglePanel();
    expect(useChatStore.getState().isOpen).toBe(true);
    useChatStore.getState().togglePanel();
    expect(useChatStore.getState().isOpen).toBe(false);
  });
});
