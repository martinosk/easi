import { createContext } from 'react';
import type { NavigationContextValue } from './types';

export const NavigationContext = createContext<NavigationContextValue | null>(null);

NavigationContext.displayName = 'NavigationContext';
