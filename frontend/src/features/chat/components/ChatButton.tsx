import { AssistantIcon } from '../../../components/layout/AppNavigation.icons';
import { HeaderActionButton } from '../../../components/layout/HeaderActionButton';

interface ChatButtonProps {
  assistantAvailable: boolean;
  onClick: () => void;
  isActive?: boolean;
}

export function ChatButton({ assistantAvailable, onClick, isActive = false }: ChatButtonProps) {
  if (!assistantAvailable) return null;

  return (
    <HeaderActionButton
      icon={AssistantIcon}
      label="Architecture Assistant"
      onClick={onClick}
      active={isActive}
      testId="nav-chat"
    />
  );
}
