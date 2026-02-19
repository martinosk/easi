import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useChat } from './useChat';

vi.mock('../api/chatApi', () => ({
  chatApi: {
    createConversation: vi.fn(),
    sendMessageStream: vi.fn(),
  },
}));

import { chatApi } from '../api/chatApi';

function createMockSSEResponse(events: string): Response {
  const encoder = new TextEncoder();
  const stream = new ReadableStream({
    start(controller) {
      controller.enqueue(encoder.encode(events));
      controller.close();
    },
  });

  return new Response(stream, {
    status: 200,
    headers: { 'Content-Type': 'text/event-stream' },
  });
}

async function sendWithSSE(sseData: string) {
  vi.mocked(chatApi.sendMessageStream).mockResolvedValue(createMockSSEResponse(sseData));
  const { result } = renderHook(() => useChat());
  await act(async () => { await result.current.sendMessage('conv-1', 'Hi'); });
  return result;
}

async function sendWithError(mockSetup: () => void) {
  mockSetup();
  const { result } = renderHook(() => useChat());
  await act(async () => { await result.current.sendMessage('conv-1', 'Hi'); });
  return result;
}

describe('useChat', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should start with empty messages and not streaming', () => {
    const { result } = renderHook(() => useChat());
    expect(result.current.messages).toEqual([]);
    expect(result.current.isStreaming).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it('should add user message immediately on send', async () => {
    const result = await sendWithSSE('event: token\ndata: {"content":"Hi"}\n\nevent: done\ndata: {"messageId":"msg-1","tokensUsed":5}\n\n');

    const userMsg = result.current.messages.find(m => m.role === 'user');
    expect(userMsg).toBeDefined();
    expect(userMsg!.content).toBe('Hi');
  });

  it('should stream assistant response from SSE tokens', async () => {
    const result = await sendWithSSE('event: token\ndata: {"content":"Hello "}\n\nevent: token\ndata: {"content":"world"}\n\nevent: done\ndata: {"messageId":"msg-1","tokensUsed":10}\n\n');

    const assistantMsg = result.current.messages.find(m => m.role === 'assistant');
    expect(assistantMsg).toBeDefined();
    expect(assistantMsg!.content).toBe('Hello world');
  });

  it('should set error on SSE error event', async () => {
    const result = await sendWithSSE('event: error\ndata: {"code":"llm_error","message":"Service unavailable"}\n\n');
    expect(result.current.error).toBe('Service unavailable');
  });

  it('should set error on non-200 response', async () => {
    const result = await sendWithError(() => {
      vi.mocked(chatApi.sendMessageStream).mockResolvedValue(
        new Response('Too Many Requests', { status: 429 })
      );
    });
    expect(result.current.error).toBeTruthy();
  });

  it('should set error on network failure', async () => {
    const result = await sendWithError(() => {
      vi.mocked(chatApi.sendMessageStream).mockRejectedValue(new Error('Network error'));
    });
    expect(result.current.error).toBe('Connection lost. Click to retry.');
  });

  it('should not be streaming after response completes', async () => {
    const result = await sendWithSSE('event: token\ndata: {"content":"Done"}\n\nevent: done\ndata: {"messageId":"msg-1","tokensUsed":5}\n\n');
    expect(result.current.isStreaming).toBe(false);
  });

  it('should clear error when sending new message', async () => {
    vi.mocked(chatApi.sendMessageStream)
      .mockRejectedValueOnce(new Error('Network error'))
      .mockResolvedValueOnce(createMockSSEResponse('event: token\ndata: {"content":"ok"}\n\nevent: done\ndata: {"messageId":"msg-1","tokensUsed":1}\n\n'));

    const { result } = renderHook(() => useChat());

    await act(async () => {
      await result.current.sendMessage('conv-1', 'Hi');
    });
    expect(result.current.error).toBeTruthy();

    await act(async () => {
      await result.current.sendMessage('conv-1', 'Retry');
    });
    expect(result.current.error).toBeNull();
  });

  it('should pass conversationId to sendMessageStream', async () => {
    vi.mocked(chatApi.sendMessageStream).mockResolvedValue(createMockSSEResponse('event: done\ndata: {"messageId":"m1","tokensUsed":1}\n\n'));
    const { result } = renderHook(() => useChat());

    await act(async () => {
      await result.current.sendMessage('my-conv-42', 'Hello');
    });

    expect(chatApi.sendMessageStream).toHaveBeenCalledWith('my-conv-42', { content: 'Hello' });
  });
});
