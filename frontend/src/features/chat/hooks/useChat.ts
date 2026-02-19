import { useState, useCallback, useRef } from 'react';
import { chatApi } from '../api/chatApi';
import { parseSSEChunk } from '../api/parseSSE';
import type { SSEEvent } from '../api/parseSSE';
import type { ChatMessage } from '../api/types';

export interface ToolCallState {
  id: string;
  name: string;
  status: 'running' | 'completed' | 'error';
  resultPreview?: string;
  errorMessage?: string;
}

interface UseChatOptions {
  onDone?: () => void;
}

interface UseChatReturn {
  messages: ChatMessage[];
  toolCalls: ToolCallState[];
  isStreaming: boolean;
  error: string | null;
  sendMessage: (conversationId: string, content: string, allowWriteOperations?: boolean) => Promise<void>;
}

type MessageUpdater = (fn: (prev: ChatMessage[]) => ChatMessage[]) => void;
type ToolCallUpdater = (fn: (prev: ToolCallState[]) => ToolCallState[]) => void;

interface StreamHandlers {
  msgId: string;
  setMessages: MessageUpdater;
  setError: (e: string | null) => void;
  setToolCalls: ToolCallUpdater;
  onDone?: () => void;
}

function upsertAssistantMessage(setMessages: MessageUpdater, id: string, content: string) {
  setMessages(prev => {
    const exists = prev.some(m => m.id === id);
    if (exists) {
      return prev.map(m => m.id === id ? { ...m, content } : m);
    }
    return [...prev, { id, role: 'assistant' as const, content }];
  });
}

function trimProcessedBuffer(buffer: string, hasEvents: boolean): string {
  if (!hasEvents) return buffer;
  const lastDoubleNewline = buffer.lastIndexOf('\n\n');
  return lastDoubleNewline >= 0 ? buffer.slice(lastDoubleNewline + 2) : '';
}

function handleToolCallStart(handlers: StreamHandlers, event: Extract<SSEEvent, { type: 'tool_call_start' }>) {
  handlers.setToolCalls(prev => [
    ...prev,
    { id: event.toolCallId, name: event.name, status: 'running' },
  ]);
}

function handleToolCallResult(handlers: StreamHandlers, event: Extract<SSEEvent, { type: 'tool_call_result' }>) {
  handlers.setToolCalls(prev =>
    prev.map(tc =>
      tc.id === event.toolCallId
        ? { ...tc, status: 'completed' as const, resultPreview: event.resultPreview }
        : tc
    )
  );
}

function applySingleEvent(event: SSEEvent, state: { content: string }, handlers: StreamHandlers) {
  switch (event.type) {
    case 'token':
      state.content += event.content;
      upsertAssistantMessage(handlers.setMessages, handlers.msgId, state.content);
      break;
    case 'tool_call_start': handleToolCallStart(handlers, event); break;
    case 'tool_call_result': handleToolCallResult(handlers, event); break;
    case 'done': handlers.onDone?.(); break;
    case 'error': handlers.setError(event.message); break;
  }
}

function applyEvents(events: SSEEvent[], state: { content: string }, handlers: StreamHandlers) {
  for (const event of events) {
    applySingleEvent(event, state, handlers);
  }
}

async function readStream(reader: ReadableStreamDefaultReader<Uint8Array>, handlers: StreamHandlers) {
  const decoder = new TextDecoder();
  let buffer = '';
  const state = { content: '' };

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const events = parseSSEChunk(buffer);
    buffer = trimProcessedBuffer(buffer, events.length > 0);
    applyEvents(events, state, handlers);
  }
}

export function useChat(options?: UseChatOptions): UseChatReturn {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [toolCalls, setToolCalls] = useState<ToolCallState[]>([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const messageIdCounter = useRef(0);
  const onDoneRef = useRef(options?.onDone);
  onDoneRef.current = options?.onDone;

  const sendMessage = useCallback(async (conversationId: string, content: string, allowWriteOperations?: boolean) => {
    setError(null);
    setIsStreaming(true);
    setToolCalls([]);

    const userMsgId = `user-${++messageIdCounter.current}`;
    setMessages(prev => [...prev, { id: userMsgId, role: 'user', content }]);

    const assistantMsgId = `assistant-${messageIdCounter.current}`;

    try {
      const request = allowWriteOperations !== undefined
        ? { content, allowWriteOperations }
        : { content };
      const response = await chatApi.sendMessageStream(conversationId, request);

      if (!response.ok) {
        setError(`Request failed (${response.status})`);
        return;
      }

      const reader = response.body?.getReader();
      if (!reader) {
        setError('No response stream available');
        return;
      }

      await readStream(reader, {
        msgId: assistantMsgId,
        setMessages,
        setError,
        setToolCalls,
        onDone: () => onDoneRef.current?.(),
      });
    } catch {
      setError('Connection lost. Click to retry.');
    } finally {
      setIsStreaming(false);
    }
  }, []);

  return { messages, toolCalls, isStreaming, error, sendMessage };
}
