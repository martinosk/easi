import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

interface ChatMessageProps {
  sender: 'user' | 'assistant';
  content: string;
  isStreaming?: boolean;
}

export function ChatMessage({ sender, content, isStreaming }: ChatMessageProps) {
  return (
    <div className={`chat-message chat-message-${sender}`}>
      <div className="chat-message-bubble">
        {sender === 'assistant' ? (
          <>
            <Markdown remarkPlugins={[remarkGfm]}>{content}</Markdown>
            {isStreaming && <span className="chat-cursor" />}
          </>
        ) : (
          <p>{content}</p>
        )}
      </div>
    </div>
  );
}
