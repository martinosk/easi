import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import classes from './ChatMessage.module.css';

interface ChatMessageProps {
  sender: 'user' | 'assistant';
  content: string;
  isStreaming?: boolean;
}

export function ChatMessage({ sender, content, isStreaming }: ChatMessageProps) {
  return (
    <div className={classes.message} data-sender={sender} data-testid="chat-message">
      <div className={classes.bubble}>
        {sender === 'assistant' ? (
          <>
            <Markdown remarkPlugins={[remarkGfm]}>{content}</Markdown>
            {isStreaming && <span className={classes.cursor} data-testid="chat-cursor" />}
          </>
        ) : (
          <p>{content}</p>
        )}
      </div>
    </div>
  );
}
