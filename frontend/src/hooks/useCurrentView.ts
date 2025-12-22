import { useAppStore } from '../store/appStore';
import { useView } from '../features/views/hooks/useViews';
import type { View, ViewId } from '../api/types';

export interface UseCurrentViewResult {
  currentView: View | null;
  currentViewId: ViewId | null;
  isLoading: boolean;
  error: Error | null;
}

export function useCurrentView(): UseCurrentViewResult {
  const currentViewId = useAppStore((state) => state.currentViewId);
  const { data: currentView, isLoading, error } = useView(currentViewId ?? undefined);

  return {
    currentView: currentView ?? null,
    currentViewId,
    isLoading,
    error: error ?? null,
  };
}
