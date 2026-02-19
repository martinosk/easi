interface ChatPanelHeaderProps {
  onToggleHistory: () => void;
  onClose: () => void;
}

export function ChatPanelHeader({ onToggleHistory, onClose }: ChatPanelHeaderProps) {
  return (
    <div className="chat-panel-header">
      <h2 className="chat-panel-title">Architecture Assistant</h2>
      <div className="chat-panel-header-actions">
        <button
          type="button"
          className="chat-panel-history-btn"
          onClick={onToggleHistory}
          aria-label="Conversation history"
        >
          <svg viewBox="0 0 24 24" fill="none" width="18" height="18">
            <path d="M4 6h16M4 12h16M4 18h16" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          </svg>
        </button>
        <button
          type="button"
          className="chat-panel-close"
          onClick={onClose}
          aria-label="Close chat"
        >
          <svg viewBox="0 0 24 24" fill="none" width="20" height="20">
            <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
          </svg>
        </button>
      </div>
    </div>
  );
}
