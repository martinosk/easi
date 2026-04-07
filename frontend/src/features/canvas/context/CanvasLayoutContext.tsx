/* eslint-disable react-refresh/only-export-components */
import { createContext, type ReactNode, useContext } from 'react';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { type UseCanvasLayoutResult, useCanvasLayout } from '../hooks/useCanvasLayout';

const CanvasLayoutContext = createContext<UseCanvasLayoutResult | null>(null);

interface CanvasLayoutProviderProps {
  children: ReactNode;
}

export function CanvasLayoutProvider({ children }: CanvasLayoutProviderProps) {
  const { currentViewId } = useCurrentView();
  const layoutResult = useCanvasLayout(currentViewId);

  return <CanvasLayoutContext.Provider value={layoutResult}>{children}</CanvasLayoutContext.Provider>;
}

export function useCanvasLayoutContext(): UseCanvasLayoutResult {
  const context = useContext(CanvasLayoutContext);
  if (!context) {
    throw new Error('useCanvasLayoutContext must be used within a CanvasLayoutProvider');
  }
  return context;
}
