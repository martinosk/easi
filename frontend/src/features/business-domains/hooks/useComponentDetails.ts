import { useState, useEffect } from 'react';
import { apiClient } from '../../../api/client';
import type { Component, ComponentId } from '../../../api/types';

export function useComponentDetails(componentId: ComponentId | null) {
  const [component, setComponent] = useState<Component | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!componentId) {
      setComponent(null);
      return;
    }

    let cancelled = false;
    setIsLoading(true);
    setError(null);

    apiClient
      .getComponentById(componentId)
      .then((data) => {
        if (!cancelled) {
          setComponent(data);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err);
        }
      })
      .finally(() => {
        if (!cancelled) {
          setIsLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [componentId]);

  return { component, isLoading, error };
}
