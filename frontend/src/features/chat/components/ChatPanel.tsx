import { useEffect, useCallback, useState, useRef } from 'react';
import { ChatMessage } from './ChatMessage';
import { ChatInput } from './ChatInput';
import { useChat } from '../hooks/useChat';
import { chatApi } from '../api/chatApi';
import './ChatPanel.css';

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

interface ChatPanelProps {
  isOpen: boolean;
  onClose: () => void;
}

export function ChatPanel({ isOpen, onClose }: ChatPanelProps) {
  const [conversationId, setConversationId] = useState<string | null>(null);
  const [yoloEnabled, setYoloEnabled] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const { messages, isStreaming, error, sendMessage } = useChat();

  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = useCallback(async (content: string) => {
    let convId = conversationId;
    if (!convId) {
      const conversation = await chatApi.createConversation();
      convId = conversation.id;
      setConversationId(convId);
    }
    sendMessage(convId, content);
  }, [conversationId, sendMessage]);

  if (!isOpen) return null;

  return (
    <div className="chat-panel" role="complementary" aria-label="Chat panel">
      <div className="chat-panel-header">
        <h2 className="chat-panel-title">Architecture Assistant</h2>
        <button
          type="button"
          className="chat-panel-close"
          onClick={onClose}
          aria-label="Close chat"
        >
          <svg viewBox="0 0 24 24" fill="none" width="20" height="20">
            <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
          </svg>
        </button>
      </div>

      <div className="chat-panel-messages">
        {messages.length === 0 && <EmptyState onSuggestionClick={handleSend} />}

        {messages.map((msg, index) => (
          <ChatMessage
            key={msg.id}
            role={msg.role}
            content={msg.content}
            isStreaming={isStreaming && isLastAssistantMessage(index, msg.role, messages.length)}
          />
        ))}

        {error && (
          <div className="chat-error">
            {error}
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      <ChatInput
        onSend={handleSend}
        disabled={isStreaming}
        yoloEnabled={yoloEnabled}
        onToggleYolo={() => setYoloEnabled(!yoloEnabled)}
      />
    </div>
  );
}
