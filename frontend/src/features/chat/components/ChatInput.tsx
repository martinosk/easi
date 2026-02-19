import { useState, useCallback, useRef, type KeyboardEvent } from 'react';

interface ChatInputProps {
  onSend: (content: string) => void;
  disabled: boolean;
  yoloEnabled: boolean;
  onToggleYolo: () => void;
}

export function ChatInput({ onSend, disabled, yoloEnabled, onToggleYolo }: ChatInputProps) {
  const [value, setValue] = useState('');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSend = useCallback(() => {
    const trimmed = value.trim();
    if (!trimmed) return;
    onSend(trimmed);
    setValue('');
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
    }
  }, [value, onSend]);

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleInput = () => {
    const el = textareaRef.current;
    if (!el) return;
    el.style.height = 'auto';
    const lineHeight = parseInt(getComputedStyle(el).lineHeight) || 20;
    const maxHeight = lineHeight * 4;
    el.style.height = `${Math.min(el.scrollHeight, maxHeight)}px`;
  };

  return (
    <div className="chat-input-area">
      <div className="chat-input-wrapper">
        <textarea
          ref={textareaRef}
          className="chat-input"
          placeholder="Ask about your architecture..."
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={handleKeyDown}
          onInput={handleInput}
          disabled={disabled}
          maxLength={2000}
          rows={1}
        />
        <button
          type="button"
          className="chat-send-btn"
          onClick={handleSend}
          disabled={disabled || !value.trim()}
          aria-label="Send message"
        >
          <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" width="20" height="20">
            <path d="M22 2L11 13" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M22 2L15 22L11 13L2 9L22 2Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </button>
      </div>
      <div className="chat-input-options">
        <label className="chat-yolo-label">
          <input
            type="checkbox"
            checked={yoloEnabled}
            onChange={onToggleYolo}
          />
          YOLO (allow changes)
        </label>
        <span className="chat-yolo-help">
          {yoloEnabled
            ? 'Assistant may apply changes you are already permitted to make.'
            : 'When off, assistant can read only. When on, assistant may apply changes you are already permitted to make.'}
        </span>
      </div>
    </div>
  );
}
