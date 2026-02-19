import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ChatInput } from './ChatInput';

describe('ChatInput', () => {
  it('should render textarea with placeholder', () => {
    render(<ChatInput onSend={vi.fn()} disabled={false} yoloEnabled={false} onToggleYolo={vi.fn()} />);
    expect(screen.getByPlaceholderText('Ask about your architecture...')).toBeInTheDocument();
  });

  it('should call onSend with message on Enter', async () => {
    const onSend = vi.fn();
    render(<ChatInput onSend={onSend} disabled={false} yoloEnabled={false} onToggleYolo={vi.fn()} />);

    const textarea = screen.getByPlaceholderText('Ask about your architecture...');
    await userEvent.type(textarea, 'Hello{Enter}');

    expect(onSend).toHaveBeenCalledWith('Hello');
  });

  it('should not call onSend on Shift+Enter (newline)', async () => {
    const onSend = vi.fn();
    render(<ChatInput onSend={onSend} disabled={false} yoloEnabled={false} onToggleYolo={vi.fn()} />);

    const textarea = screen.getByPlaceholderText('Ask about your architecture...');
    await userEvent.type(textarea, 'Hello{Shift>}{Enter}{/Shift}');

    expect(onSend).not.toHaveBeenCalled();
  });

  it('should clear textarea after sending', async () => {
    render(<ChatInput onSend={vi.fn()} disabled={false} yoloEnabled={false} onToggleYolo={vi.fn()} />);

    const textarea = screen.getByPlaceholderText('Ask about your architecture...') as HTMLTextAreaElement;
    await userEvent.type(textarea, 'Hello{Enter}');

    expect(textarea.value).toBe('');
  });

  it('should not send empty message', async () => {
    const onSend = vi.fn();
    render(<ChatInput onSend={onSend} disabled={false} yoloEnabled={false} onToggleYolo={vi.fn()} />);

    const textarea = screen.getByPlaceholderText('Ask about your architecture...');
    await userEvent.type(textarea, '{Enter}');

    expect(onSend).not.toHaveBeenCalled();
  });

  it('should disable textarea when disabled prop is true', () => {
    render(<ChatInput onSend={vi.fn()} disabled={true} yoloEnabled={false} onToggleYolo={vi.fn()} />);
    expect(screen.getByPlaceholderText('Ask about your architecture...')).toBeDisabled();
  });

  it('should render YOLO checkbox', () => {
    render(<ChatInput onSend={vi.fn()} disabled={false} yoloEnabled={false} onToggleYolo={vi.fn()} />);
    expect(screen.getByLabelText('YOLO (allow changes)')).toBeInTheDocument();
  });

  it('should reflect yoloEnabled state in checkbox', () => {
    render(<ChatInput onSend={vi.fn()} disabled={false} yoloEnabled={true} onToggleYolo={vi.fn()} />);
    expect(screen.getByLabelText('YOLO (allow changes)')).toBeChecked();
  });

  it('should call onToggleYolo when checkbox is clicked', () => {
    const onToggleYolo = vi.fn();
    render(<ChatInput onSend={vi.fn()} disabled={false} yoloEnabled={false} onToggleYolo={onToggleYolo} />);

    fireEvent.click(screen.getByLabelText('YOLO (allow changes)'));
    expect(onToggleYolo).toHaveBeenCalled();
  });

  it('should enforce max 2000 character limit', async () => {
    render(<ChatInput onSend={vi.fn()} disabled={false} yoloEnabled={false} onToggleYolo={vi.fn()} />);
    const textarea = screen.getByPlaceholderText('Ask about your architecture...') as HTMLTextAreaElement;
    expect(textarea.maxLength).toBe(2000);
  });
});
