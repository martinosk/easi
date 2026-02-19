import { useEffect, useCallback, useState } from 'react';
import { ChatInput } from './ChatInput';
import { ChatPanelHeader } from './ChatPanelHeader';
import { ConversationList } from './ConversationList';
import { MessageList } from './MessageList';
import { useChat } from '../hooks/useChat';
import { useConversations } from '../hooks/useConversations';
import { chatApi } from '../api/chatApi';
import './ChatPanel.css';

interface ChatPanelProps {
  isOpen: boolean;
  onClose: () => void;
}

export function ChatPanel({ isOpen, onClose }: ChatPanelProps) {
  const [conversationId, setConversationId] = useState<string | null>(null);
  const [yoloEnabled, setYoloEnabled] = useState(false);
  const [showConversationList, setShowConversationList] = useState(false);
  const { conversations, deleteConversation, invalidateList } = useConversations();
  const { messages, toolCalls, isStreaming, error, sendMessage } = useChat({
    onDone: invalidateList,
  });

  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  const handleSend = useCallback(async (content: string) => {
    let convId = conversationId;
    if (!convId) {
      const conversation = await chatApi.createConversation();
      convId = conversation.id;
      setConversationId(convId);
      invalidateList();
    }
    sendMessage(convId, content, yoloEnabled);
  }, [conversationId, sendMessage, yoloEnabled, invalidateList]);

  const handleSelectConversation = useCallback((id: string) => {
    setConversationId(id);
    setShowConversationList(false);
  }, []);

  const handleNewConversation = useCallback(() => {
    setConversationId(null);
    setShowConversationList(false);
  }, []);

  const handleDeleteConversation = useCallback((id: string) => {
    deleteConversation(id);
    if (conversationId === id) {
      setConversationId(null);
    }
  }, [deleteConversation, conversationId]);

  if (!isOpen) return null;

  return (
    <div className="chat-panel" role="complementary" aria-label="Chat panel">
      <ChatPanelHeader
        onToggleHistory={() => setShowConversationList(!showConversationList)}
        onClose={onClose}
      />

      {showConversationList && (
        <ConversationList
          conversations={conversations}
          activeConversationId={conversationId}
          onSelect={handleSelectConversation}
          onDelete={handleDeleteConversation}
          onNewConversation={handleNewConversation}
        />
      )}

      <MessageList
        messages={messages}
        toolCalls={toolCalls}
        isStreaming={isStreaming}
        error={error}
        onSuggestionClick={handleSend}
      />

      <ChatInput
        onSend={handleSend}
        disabled={isStreaming}
        yoloEnabled={yoloEnabled}
        onToggleYolo={() => setYoloEnabled(!yoloEnabled)}
      />
    </div>
  );
}
