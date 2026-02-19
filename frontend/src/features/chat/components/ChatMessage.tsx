import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

interface ChatMessageProps {
  role: 'user' | 'assistant';
  content: string;
  isStreaming?: boolean;
}

export function ChatMessage({ role, content, isStreaming }: ChatMessageProps) {
  return (
    <div className={`chat-message chat-message-${role}`}>
      <div className="chat-message-bubble">
        {role === 'assistant' ? (
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
