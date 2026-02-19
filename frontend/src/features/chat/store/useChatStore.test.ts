import { describe, it, expect, beforeEach } from 'vitest';
import { useChatStore } from './useChatStore';

describe('useChatStore', () => {
  beforeEach(() => {
    useChatStore.setState({
      isOpen: false,
      conversationId: null,
      yoloEnabled: false,
    });
  });

  it('should start with panel closed', () => {
    expect(useChatStore.getState().isOpen).toBe(false);
  });

  it('should start with no conversation', () => {
    expect(useChatStore.getState().conversationId).toBeNull();
  });

  it('should start with yolo disabled', () => {
    expect(useChatStore.getState().yoloEnabled).toBe(false);
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

  it('should set conversation id', () => {
    useChatStore.getState().setConversationId('conv-123');
    expect(useChatStore.getState().conversationId).toBe('conv-123');
  });

  it('should toggle yolo', () => {
    useChatStore.getState().toggleYolo();
    expect(useChatStore.getState().yoloEnabled).toBe(true);
    useChatStore.getState().toggleYolo();
    expect(useChatStore.getState().yoloEnabled).toBe(false);
  });
});
