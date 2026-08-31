import { hasLink } from '../../../utils/hateoas';
import { useUserStore } from '../../../store/userStore';
import { useAssistantStatus } from './useAssistantStatus';

export interface AssistantAvailability {
  assistantAvailable: boolean;
  assistantWriteAvailable: boolean;
}

export function useAssistantAvailability(): AssistantAvailability {
  const sessionLinks = useUserStore((state) => state.sessionLinks);
  const mayCheckAssistantStatus = Boolean(sessionLinks?.['x-assistant-status']);
  const { data: status } = useAssistantStatus(mayCheckAssistantStatus);

  return {
    assistantAvailable: hasLink(status, 'x-conversations'),
    assistantWriteAvailable: hasLink(status, 'x-conversations-write'),
  };
}
