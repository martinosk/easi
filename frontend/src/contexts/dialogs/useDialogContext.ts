import { useContext } from 'react';
import { DialogContext } from './context';
import type { DialogContextValue } from './types';

export function useDialogContext(): DialogContextValue {
  const context = useContext(DialogContext);
  if (!context) {
    throw new Error('useDialogContext must be used within a DialogProvider');
  }
  return context;
}
