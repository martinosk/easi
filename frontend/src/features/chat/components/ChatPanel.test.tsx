import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, act, waitFor } from '@testing-library/react';
import { ChatPanel } from './ChatPanel';

vi.mock('../api/chatApi', () => ({
  chatApi: {
    createConversation: vi.fn(),
    sendMessageStream: vi.fn(),
  },
}));

import { chatApi } from '../api/chatApi';

describe('ChatPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should not render when isOpen is false', () => {
    const { container } = render(<ChatPanel isOpen={false} onClose={vi.fn()} />);
    expect(container.querySelector('.chat-panel')).not.toBeInTheDocument();
  });

  it('should render when isOpen is true', () => {
    const { container } = render(<ChatPanel isOpen={true} onClose={vi.fn()} />);
    expect(container.querySelector('.chat-panel')).toBeInTheDocument();
  });

  it('should render header with title', () => {
    render(<ChatPanel isOpen={true} onClose={vi.fn()} />);
    expect(screen.getByText('Architecture Assistant')).toBeInTheDocument();
  });

  it('should render close button', () => {
    render(<ChatPanel isOpen={true} onClose={vi.fn()} />);
    expect(screen.getByLabelText('Close chat')).toBeInTheDocument();
  });

  it('should call onClose when close button is clicked', () => {
    const onClose = vi.fn();
    render(<ChatPanel isOpen={true} onClose={onClose} />);
    fireEvent.click(screen.getByLabelText('Close chat'));
    expect(onClose).toHaveBeenCalled();
  });

  it('should call onClose when Escape is pressed', () => {
    const onClose = vi.fn();
    render(<ChatPanel isOpen={true} onClose={onClose} />);
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(onClose).toHaveBeenCalled();
  });

  it('should show prompt suggestions in empty state', () => {
    render(<ChatPanel isOpen={true} onClose={vi.fn()} />);
    expect(screen.getByText('What applications are in the Finance domain?')).toBeInTheDocument();
    expect(screen.getByText('Show me a portfolio summary')).toBeInTheDocument();
  });

  it('should render chat input', () => {
    render(<ChatPanel isOpen={true} onClose={vi.fn()} />);
    expect(screen.getByPlaceholderText('Ask about your architecture...')).toBeInTheDocument();
  });

  it('should create conversation on open and send message', async () => {
    vi.mocked(chatApi.createConversation).mockResolvedValue({
      id: 'conv-1',
      title: 'New conversation',
      createdAt: new Date().toISOString(),
      _links: {},
    });

    const encoder = new TextEncoder();
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(encoder.encode('event: token\ndata: {"content":"Hello"}\n\nevent: done\ndata: {"messageId":"msg-1","tokensUsed":5}\n\n'));
        controller.close();
      },
    });
    vi.mocked(chatApi.sendMessageStream).mockResolvedValue(
      new Response(stream, { status: 200, headers: { 'Content-Type': 'text/event-stream' } })
    );

    render(<ChatPanel isOpen={true} onClose={vi.fn()} />);

    const textarea = screen.getByPlaceholderText('Ask about your architecture...');

    await act(async () => {
      fireEvent.change(textarea, { target: { value: 'Hello' } });
      fireEvent.keyDown(textarea, { key: 'Enter' });
    });

    await waitFor(() => {
      expect(chatApi.createConversation).toHaveBeenCalled();
    });
  });

  it('should render YOLO checkbox', () => {
    render(<ChatPanel isOpen={true} onClose={vi.fn()} />);
    expect(screen.getByLabelText('YOLO (allow changes)')).toBeInTheDocument();
  });
});
