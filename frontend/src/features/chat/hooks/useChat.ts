import { useState, useCallback, useRef } from 'react';
import { chatApi } from '../api/chatApi';
import { parseSSEChunk } from '../api/parseSSE';
import type { SSEEvent } from '../api/parseSSE';
import type { ChatMessage } from '../api/types';

interface UseChatReturn {
  messages: ChatMessage[];
  isStreaming: boolean;
  error: string | null;
  sendMessage: (conversationId: string, content: string) => Promise<void>;
}

type MessageUpdater = (fn: (prev: ChatMessage[]) => ChatMessage[]) => void;

interface StreamHandlers {
  msgId: string;
  setMessages: MessageUpdater;
  setError: (e: string | null) => void;
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

function applyEvents(events: SSEEvent[], state: { content: string }, handlers: StreamHandlers) {
  for (const event of events) {
    if (event.type === 'token') {
      state.content += event.content;
      upsertAssistantMessage(handlers.setMessages, handlers.msgId, state.content);
    } else if (event.type === 'error') {
      handlers.setError(event.message);
    }
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

export function useChat(): UseChatReturn {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const messageIdCounter = useRef(0);

  const sendMessage = useCallback(async (conversationId: string, content: string) => {
    setError(null);
    setIsStreaming(true);

    const userMsgId = `user-${++messageIdCounter.current}`;
    setMessages(prev => [...prev, { id: userMsgId, role: 'user', content }]);

    const assistantMsgId = `assistant-${messageIdCounter.current}`;

    try {
      const response = await chatApi.sendMessageStream(conversationId, { content });

      if (!response.ok) {
        setError(`Request failed (${response.status})`);
        return;
      }

      const reader = response.body?.getReader();
      if (!reader) {
        setError('No response stream available');
        return;
      }

      await readStream(reader, { msgId: assistantMsgId, setMessages, setError });
    } catch {
      setError('Connection lost. Click to retry.');
    } finally {
      setIsStreaming(false);
    }
  }, []);

  return { messages, isStreaming, error, sendMessage };
}
