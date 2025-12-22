import { createContext, useContext, type ReactNode } from 'react';
import { useCanvasLayout, type UseCanvasLayoutResult } from '../hooks/useCanvasLayout';
import { useCurrentView } from '../../../hooks/useCurrentView';

const CanvasLayoutContext = createContext<UseCanvasLayoutResult | null>(null);

interface CanvasLayoutProviderProps {
  children: ReactNode;
}

export function CanvasLayoutProvider({ children }: CanvasLayoutProviderProps) {
  const { currentViewId } = useCurrentView();
  const layoutResult = useCanvasLayout(currentViewId);

  return (
    <CanvasLayoutContext.Provider value={layoutResult}>
      {children}
    </CanvasLayoutContext.Provider>
  );
}

export function useCanvasLayoutContext(): UseCanvasLayoutResult {
  const context = useContext(CanvasLayoutContext);
  if (!context) {
    throw new Error('useCanvasLayoutContext must be used within a CanvasLayoutProvider');
  }
  return context;
}
