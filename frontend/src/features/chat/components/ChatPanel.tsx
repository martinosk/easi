import { Drawer } from '@mantine/core';
import { useCallback, useEffect, useState } from 'react';
import { useResponsive } from '../../../hooks/useResponsive';
import { chatApi } from '../api/chatApi';
import { useChat } from '../hooks/useChat';
import { useConversations } from '../hooks/useConversations';
import { ChatInput } from './ChatInput';
import { ChatPanelHeader } from './ChatPanelHeader';
import { ConversationList } from './ConversationList';
import { MessageList } from './MessageList';
import classes from './ChatPanel.module.css';

const PANEL_SIZE = 'var(--chat-panel-width)';
const FULL_WIDTH = '100%';

interface ChatPanelProps {
  isOpen: boolean;
  onClose: () => void;
  writeAvailable: boolean;
}

function useOpenedAfterMount(isOpen: boolean): boolean {
  const [opened, setOpened] = useState(false);
  useEffect(() => {
    setOpened(isOpen);
  }, [isOpen]);
  return opened;
}

export function ChatPanel({ isOpen, onClose, writeAvailable }: ChatPanelProps) {
  const [conversationId, setConversationId] = useState<string | null>(null);
  const [yoloEnabled, setYoloEnabled] = useState(false);
  const [showConversationList, setShowConversationList] = useState(false);
  const opened = useOpenedAfterMount(isOpen);
  const { isMobile } = useResponsive();
  const { conversations, deleteConversation, invalidateList } = useConversations();
  const { messages, toolCalls, isStreaming, error, sendMessage, resetMessages } = useChat({
    onDone: invalidateList,
  });

  const handleSend = useCallback(
    async (content: string) => {
      let convId = conversationId;
      if (!convId) {
        const conversation = await chatApi.createConversation();
        convId = conversation.id;
        setConversationId(convId);
        invalidateList();
      }
      sendMessage(convId, content, writeAvailable && yoloEnabled);
    },
    [conversationId, sendMessage, yoloEnabled, writeAvailable, invalidateList],
  );

  const handleSelectConversation = useCallback(
    async (id: string) => {
      setConversationId(id);
      setShowConversationList(false);
      try {
        const detail = await chatApi.getConversation(id);
        resetMessages(detail.messages.map((m) => ({ id: m.id, role: m.role, content: m.content })));
      } catch {
        resetMessages();
      }
    },
    [resetMessages],
  );

  const handleNewConversation = useCallback(() => {
    setConversationId(null);
    setShowConversationList(false);
    resetMessages();
  }, [resetMessages]);

  const handleDeleteConversation = useCallback(
    (id: string) => {
      deleteConversation(id);
      if (conversationId === id) {
        setConversationId(null);
        resetMessages();
      }
    },
    [deleteConversation, conversationId, resetMessages],
  );

  return (
    <Drawer.Root
      opened={opened}
      onClose={onClose}
      position="right"
      size={isMobile ? FULL_WIDTH : PANEL_SIZE}
      lockScroll={false}
      trapFocus={false}
      closeOnClickOutside={false}
      closeOnEscape
      className={classes.root}
    >
      <Drawer.Content aria-label="Chat panel" className={classes.content}>
        <Drawer.Body className={classes.body}>
          <ChatPanelHeader onToggleHistory={() => setShowConversationList(!showConversationList)} onClose={onClose} />

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
            writeAvailable={writeAvailable}
          />
        </Drawer.Body>
      </Drawer.Content>
    </Drawer.Root>
  );
}
