import { ActionIcon, Group, Stack, Text, UnstyledButton } from '@mantine/core';
import type { Conversation } from '../api/types';
import classes from './ConversationList.module.css';

interface ConversationListProps {
  conversations: Conversation[];
  activeConversationId: string | null;
  onSelect: (id: string) => void;
  onDelete: (id: string) => void;
  onNewConversation: () => void;
}

function formatRelativeTime(dateStr: string): string {
  const now = Date.now();
  const then = new Date(dateStr).getTime();
  const diffMs = now - then;
  const diffMin = Math.floor(diffMs / 60_000);

  if (diffMin < 1) return 'just now';
  if (diffMin < 60) return `${diffMin}m ago`;

  const diffHours = Math.floor(diffMin / 60);
  if (diffHours < 24) return `${diffHours}h ago`;

  const diffDays = Math.floor(diffHours / 24);
  if (diffDays < 30) return `${diffDays}d ago`;

  return new Date(dateStr).toLocaleDateString();
}

export function ConversationList({
  conversations,
  activeConversationId,
  onSelect,
  onDelete,
  onNewConversation,
}: ConversationListProps) {
  return (
    <div className={classes.list}>
      <Group justify="flex-end" px="sm" py="xs" className={classes.header}>
        <ActionIcon variant="default" onClick={onNewConversation} aria-label="New conversation" size="sm">
          <svg viewBox="0 0 24 24" fill="none" width="16" height="16">
            <path d="M12 5v14M5 12h14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          </svg>
        </ActionIcon>
      </Group>
      {conversations.length === 0 ? (
        <Text size="sm" ta="center" p="md" className={classes.empty}>
          No conversations yet
        </Text>
      ) : (
        <Stack gap={0}>
          {conversations.map((conv) => (
            <Group
              key={conv.id}
              gap="xs"
              px="md"
              py="sm"
              wrap="nowrap"
              className={classes.item}
              data-active={activeConversationId === conv.id || undefined}
              data-testid="conversation-item"
            >
              <UnstyledButton
                component="button"
                type="button"
                className={classes.itemContent}
                onClick={() => onSelect(conv.id)}
              >
                <span className={classes.itemTitle}>{conv.title}</span>
                <span className={classes.itemTime}>{formatRelativeTime(conv.createdAt)}</span>
              </UnstyledButton>
              <ActionIcon
                variant="subtle"
                color="gray"
                size="sm"
                className={classes.deleteButton}
                onClick={(e) => {
                  e.stopPropagation();
                  onDelete(conv.id);
                }}
                aria-label="Delete conversation"
              >
                <svg viewBox="0 0 24 24" fill="none" width="14" height="14">
                  <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
                </svg>
              </ActionIcon>
            </Group>
          ))}
        </Stack>
      )}
    </div>
  );
}
