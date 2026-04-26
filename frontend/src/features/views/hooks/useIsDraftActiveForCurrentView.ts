import { useAppStore } from '../../../store/appStore';
import { useCurrentView } from './useCurrentView';

export function useIsDraftActiveForCurrentView(): boolean {
  const { currentViewId } = useCurrentView();
  const dynamicViewId = useAppStore((s) => s.dynamicViewId);
  return dynamicViewId !== null && dynamicViewId === currentViewId;
}
