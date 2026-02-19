import { useEffect, useRef } from 'react';
import { ChatMessage } from './ChatMessage';
import { ToolCallIndicator } from './ToolCallIndicator';
import type { ChatMessage as ChatMessageType } from '../api/types';
import type { ToolCallState } from '../hooks/useChat';

const PROMPT_SUGGESTIONS = [
  'What applications are in the Finance domain?',
  'Show me a portfolio summary',
  "Create a new application called 'Payment Gateway'",
];

function EmptyState({ onSuggestionClick }: { onSuggestionClick: (s: string) => void }) {
  return (
    <div className="chat-empty-state">
      <p className="chat-empty-title">How can I help with your architecture?</p>
      <div className="chat-suggestions">
        {PROMPT_SUGGESTIONS.map((suggestion) => (
          <button
            key={suggestion}
            type="button"
            className="chat-suggestion-btn"
            onClick={() => onSuggestionClick(suggestion)}
          >
            {suggestion}
          </button>
        ))}
      </div>
    </div>
  );
}

function isLastAssistantMessage(index: number, role: string, total: number): boolean {
  return role === 'assistant' && index === total - 1;
}

interface MessageListProps {
  messages: ChatMessageType[];
  toolCalls: ToolCallState[];
  isStreaming: boolean;
  error: string | null;
  onSuggestionClick: (s: string) => void;
}

export function MessageList({ messages, toolCalls, isStreaming, error, onSuggestionClick }: MessageListProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  return (
    <div className="chat-panel-messages">
      {messages.length === 0 && <EmptyState onSuggestionClick={onSuggestionClick} />}

      {messages.map((msg, index) => (
        <ChatMessage
          key={msg.id}
          role={msg.role}
          content={msg.content}
          isStreaming={isStreaming && isLastAssistantMessage(index, msg.role, messages.length)}
        />
      ))}

      {toolCalls.map(tc => (
        <ToolCallIndicator
          key={tc.id}
          status={tc.status}
          name={tc.name}
          resultPreview={tc.resultPreview}
          errorMessage={tc.errorMessage}
        />
      ))}

      {error && (
        <div className="chat-error">
          {error}
        </div>
      )}

      <div ref={messagesEndRef} />
    </div>
  );
}
