import { useContext } from 'react';
import { NavigationContext } from './context';
import type { NavigationContextValue } from './types';

export function useNavigation(): NavigationContextValue {
  const context = useContext(NavigationContext);
  if (!context) {
    throw new Error('useNavigation must be used within a NavigationProvider');
  }
  return context;
}

export function useNavigationActions() {
  const { navigationActions } = useNavigation();
  return navigationActions;
}

export function useDialogActions() {
  const { dialogActions } = useNavigation();
  return dialogActions;
}

export function useViewActions() {
  const { viewActions } = useNavigation();
  return viewActions;
}

export function usePermissions() {
  const { permissions } = useNavigation();
  return permissions;
}

export function useCanvasRef() {
  const { canvasRef } = useNavigation();
  return canvasRef;
}
