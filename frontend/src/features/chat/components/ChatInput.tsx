import { ActionIcon, Checkbox, Stack, Text, Textarea } from '@mantine/core';
import { type KeyboardEvent, useCallback, useState } from 'react';

interface ChatInputProps {
  onSend: (content: string) => void;
  disabled: boolean;
  yoloEnabled: boolean;
  onToggleYolo: () => void;
}

export function ChatInput({ onSend, disabled, yoloEnabled, onToggleYolo }: ChatInputProps) {
  const [value, setValue] = useState('');

  const handleSend = useCallback(() => {
    const trimmed = value.trim();
    if (!trimmed) return;
    onSend(trimmed);
    setValue('');
  }, [value, onSend]);

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const sendButton = (
    <ActionIcon
      variant="filled"
      onClick={handleSend}
      disabled={disabled || !value.trim()}
      aria-label="Send message"
      size="lg"
    >
      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" width="18" height="18">
        <path d="M22 2L11 13" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        <path
          d="M22 2L15 22L11 13L2 9L22 2Z"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    </ActionIcon>
  );

  return (
    <Stack gap="xs" p="sm" className="chat-input-area">
      <Textarea
        placeholder="Ask about your architecture..."
        value={value}
        onChange={(e) => setValue(e.currentTarget.value)}
        onKeyDown={handleKeyDown}
        disabled={disabled}
        maxLength={2000}
        autosize
        minRows={1}
        maxRows={4}
        rightSection={sendButton}
        rightSectionWidth={48}
      />
      <Stack gap={4}>
        <Checkbox checked={yoloEnabled} onChange={onToggleYolo} label="YOLO (allow changes)" size="xs" />
        <Text size="xs" c="dimmed">
          {yoloEnabled
            ? 'Assistant may apply changes you are already permitted to make.'
            : 'When off, assistant can read only. When on, assistant may apply changes you are already permitted to make.'}
        </Text>
      </Stack>
    </Stack>
  );
}
