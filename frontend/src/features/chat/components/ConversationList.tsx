import type { Conversation } from '../api/types';

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
    <div className="conversation-list">
      <div className="conversation-list-header">
        <button
          type="button"
          className="conversation-new-btn"
          onClick={onNewConversation}
          aria-label="New conversation"
        >
          <svg viewBox="0 0 24 24" fill="none" width="16" height="16">
            <path d="M12 5v14M5 12h14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          </svg>
        </button>
      </div>
      {conversations.length === 0 ? (
        <div className="conversation-list-empty">No conversations yet</div>
      ) : (
        <div className="conversation-list-items">
          {conversations.map(conv => (
            <div
              key={conv.id}
              className={`conversation-item${activeConversationId === conv.id ? ' conversation-item-active' : ''}`}
              onClick={() => onSelect(conv.id)}
              role="button"
              tabIndex={0}
              onKeyDown={(e) => { if (e.key === 'Enter') onSelect(conv.id); }}
            >
              <div className="conversation-item-content">
                <span className="conversation-item-title">{conv.title}</span>
                <span className="conversation-item-time">{formatRelativeTime(conv.createdAt)}</span>
              </div>
              <button
                type="button"
                className="conversation-delete-btn"
                onClick={(e) => { e.stopPropagation(); onDelete(conv.id); }}
                aria-label="Delete conversation"
              >
                <svg viewBox="0 0 24 24" fill="none" width="14" height="14">
                  <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
                </svg>
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
