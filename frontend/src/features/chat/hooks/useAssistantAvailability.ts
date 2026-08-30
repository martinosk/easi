import { useUserStore } from '../../../store/userStore';
import { useAssistantStatus } from './useAssistantStatus';

export interface AssistantAvailability {
  assistantAvailable: boolean;
  assistantWriteAvailable: boolean;
}

export function useAssistantAvailability(): AssistantAvailability {
  const sessionLinks = useUserStore((state) => state.sessionLinks);
  const mayUseAssistant = Boolean(sessionLinks?.['x-assistant']);
  const { data: status } = useAssistantStatus(mayUseAssistant);
  const configured = status?.configured === true;

  return {
    assistantAvailable: mayUseAssistant && configured,
    assistantWriteAvailable: Boolean(sessionLinks?.['x-assistant-write']) && configured,
  };
}
