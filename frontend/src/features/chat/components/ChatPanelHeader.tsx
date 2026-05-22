import { ActionIcon, Group, Title } from '@mantine/core';

interface ChatPanelHeaderProps {
  onToggleHistory: () => void;
  onClose: () => void;
}

export function ChatPanelHeader({ onToggleHistory, onClose }: ChatPanelHeaderProps) {
  return (
    <Group justify="space-between" px="md" py="sm" className="chat-panel-header">
      <Title order={4}>Architecture Assistant</Title>
      <Group gap="xs">
        <ActionIcon variant="subtle" color="gray" onClick={onToggleHistory} aria-label="Conversation history">
          <svg viewBox="0 0 24 24" fill="none" width="18" height="18">
            <path d="M4 6h16M4 12h16M4 18h16" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          </svg>
        </ActionIcon>
        <ActionIcon variant="subtle" color="gray" onClick={onClose} aria-label="Close chat">
          <svg viewBox="0 0 24 24" fill="none" width="20" height="20">
            <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          </svg>
        </ActionIcon>
      </Group>
    </Group>
  );
}
