import { Button, Stack, Text } from '@mantine/core';
import { useEffect, useRef } from 'react';
import type { ChatMessage as ChatMessageType } from '../api/types';
import type { ToolCallState } from '../hooks/useChat';
import { ChatMessage } from './ChatMessage';
import { ToolCallIndicator } from './ToolCallIndicator';

const PROMPT_SUGGESTIONS = [
  'What applications are in the Finance domain?',
  'Show me a portfolio summary',
  "Create a new application called 'Payment Gateway'",
];

function EmptyState({ onSuggestionClick }: { onSuggestionClick: (s: string) => void }) {
  return (
    <Stack align="center" gap="lg" p="lg" className="chat-empty-state">
      <Text fw={600}>How can I help with your architecture?</Text>
      <Stack gap="sm" w="100%">
        {PROMPT_SUGGESTIONS.map((suggestion) => (
          <Button
            key={suggestion}
            variant="default"
            justify="flex-start"
            onClick={() => onSuggestionClick(suggestion)}
          >
            {suggestion}
          </Button>
        ))}
      </Stack>
    </Stack>
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

      {toolCalls.map((tc) => (
        <ToolCallIndicator
          key={tc.id}
          status={tc.status}
          name={tc.name}
          resultPreview={tc.resultPreview}
          errorMessage={tc.errorMessage}
        />
      ))}

      {error && <div className="chat-error">{error}</div>}

      <div ref={messagesEndRef} />
    </div>
  );
}
