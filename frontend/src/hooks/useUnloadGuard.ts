import { useEffect } from 'react';
import { useAppStore } from '../store/appStore';
import { selectAnyDirty } from '../store/slices/dynamicModeSlice';

function handleBeforeUnload(e: BeforeUnloadEvent): void {
  e.preventDefault();
  e.returnValue = '';
}

export function useUnloadGuard(): void {
  const anyDirty = useAppStore(selectAnyDirty);

  useEffect(() => {
    if (typeof window === 'undefined' || !anyDirty) return;
    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [anyDirty]);
}
