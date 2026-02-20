import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useChat } from './useChat';

vi.mock('../api/chatApi', () => ({
  chatApi: {
    createConversation: vi.fn(),
    sendMessageStream: vi.fn(),
    listConversations: vi.fn(),
    getConversation: vi.fn(),
    deleteConversation: vi.fn(),
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

  it.each([
    { desc: 'conversationId only', convId: 'my-conv-42', content: 'Hello', writeOps: undefined, expected: { content: 'Hello' } },
    { desc: 'allowWriteOperations true', convId: 'conv-1', content: 'Create app', writeOps: true, expected: { content: 'Create app', allowWriteOperations: true } },
    { desc: 'allowWriteOperations false', convId: 'conv-1', content: 'Read app', writeOps: false, expected: { content: 'Read app', allowWriteOperations: false } },
  ])('should pass $desc to sendMessageStream', async ({ convId, content, writeOps, expected }) => {
    vi.mocked(chatApi.sendMessageStream).mockResolvedValue(createMockSSEResponse('event: done\ndata: {"messageId":"m1","tokensUsed":1}\n\n'));
    const { result } = renderHook(() => useChat());

    await act(async () => {
      await result.current.sendMessage(convId, content, writeOps);
    });

    expect(chatApi.sendMessageStream).toHaveBeenCalledWith(convId, expected);
  });

  it('should call onDone callback when done event is received', async () => {
    const onDone = vi.fn();
    vi.mocked(chatApi.sendMessageStream).mockResolvedValue(
      createMockSSEResponse('event: token\ndata: {"content":"ok"}\n\nevent: done\ndata: {"messageId":"msg-1","tokensUsed":5}\n\n')
    );

    const { result } = renderHook(() => useChat({ onDone }));

    await act(async () => {
      await result.current.sendMessage('conv-1', 'Hello');
    });

    expect(onDone).toHaveBeenCalled();
  });

  it('should track tool calls from SSE events', async () => {
    const sseData = [
      'event: tool_call_start\ndata: {"toolCallId":"tc-1","name":"list_applications","arguments":"{}"}\n\n',
      'event: tool_call_result\ndata: {"toolCallId":"tc-1","name":"list_applications","resultPreview":"Found 3 apps"}\n\n',
      'event: token\ndata: {"content":"Here are the apps"}\n\n',
      'event: done\ndata: {"messageId":"msg-1","tokensUsed":10}\n\n',
    ].join('');

    const result = await sendWithSSE(sseData);

    expect(result.current.toolCalls).toEqual([
      {
        id: 'tc-1',
        name: 'list_applications',
        status: 'completed',
        resultPreview: 'Found 3 apps',
      },
    ]);
  });

  it('should add tool call as running on tool_call_start', async () => {
    const sseData = [
      'event: tool_call_start\ndata: {"toolCallId":"tc-1","name":"list_applications","arguments":"{}"}\n\n',
      'event: token\ndata: {"content":"Working..."}\n\n',
      'event: done\ndata: {"messageId":"msg-1","tokensUsed":5}\n\n',
    ].join('');

    const result = await sendWithSSE(sseData);

    expect(result.current.toolCalls).toEqual([
      { id: 'tc-1', name: 'list_applications', status: 'running' },
    ]);
  });

  it('should handle multi-chunk streaming without duplicate events', async () => {
    const encoder = new TextEncoder();
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(encoder.encode('event: token\ndata: {"content":"Hello"}\n\nevent: tok'));
        controller.enqueue(encoder.encode('en\ndata: {"content":" world"}\n\nevent: done\ndata: {"messageId":"msg-1","tokensUsed":5}\n\n'));
        controller.close();
      },
    });
    const response = new Response(stream, { status: 200, headers: { 'Content-Type': 'text/event-stream' } });
    vi.mocked(chatApi.sendMessageStream).mockResolvedValue(response);

    const { result } = renderHook(() => useChat());
    await act(async () => { await result.current.sendMessage('conv-1', 'Hi'); });

    const assistantMsg = result.current.messages.find(m => m.role === 'assistant');
    expect(assistantMsg).toBeDefined();
    expect(assistantMsg!.content).toBe('Hello world');
  });

  it('should set messages when resetMessages is called with initial messages', () => {
    const { result } = renderHook(() => useChat());
    const initial = [
      { id: 'msg-1', role: 'user' as const, content: 'Hello' },
      { id: 'msg-2', role: 'assistant' as const, content: 'Hi there' },
    ];

    act(() => { result.current.resetMessages(initial); });

    expect(result.current.messages).toEqual(initial);
  });

  it('should clear messages when resetMessages is called without arguments', async () => {
    const result = await sendWithSSE('event: token\ndata: {"content":"Hi"}\n\nevent: done\ndata: {"messageId":"msg-1","tokensUsed":5}\n\n');
    expect(result.current.messages.length).toBeGreaterThan(0);

    act(() => { result.current.resetMessages(); });

    expect(result.current.messages).toEqual([]);
    expect(result.current.toolCalls).toEqual([]);
    expect(result.current.error).toBeNull();
  });

  it('should clear tool calls when sending a new message', async () => {
    const sseDataWithTool = [
      'event: tool_call_start\ndata: {"toolCallId":"tc-1","name":"list_applications","arguments":"{}"}\n\n',
      'event: tool_call_result\ndata: {"toolCallId":"tc-1","name":"list_applications","resultPreview":"Found 3"}\n\n',
      'event: token\ndata: {"content":"Result"}\n\n',
      'event: done\ndata: {"messageId":"msg-1","tokensUsed":5}\n\n',
    ].join('');

    const sseDataSimple = 'event: token\ndata: {"content":"Hello"}\n\nevent: done\ndata: {"messageId":"msg-2","tokensUsed":3}\n\n';

    vi.mocked(chatApi.sendMessageStream)
      .mockResolvedValueOnce(createMockSSEResponse(sseDataWithTool))
      .mockResolvedValueOnce(createMockSSEResponse(sseDataSimple));

    const { result } = renderHook(() => useChat());

    await act(async () => {
      await result.current.sendMessage('conv-1', 'First');
    });
    expect(result.current.toolCalls.length).toBe(1);

    await act(async () => {
      await result.current.sendMessage('conv-1', 'Second');
    });
    expect(result.current.toolCalls).toEqual([]);
  });
});
