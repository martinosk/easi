import { useEffect, useCallback, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useViews, useCreateView } from '../features/views/hooks/useViews';
import { useAppStore } from '../store/appStore';
import type { View, ViewId } from '../api/types';
import toast from 'react-hot-toast';
import { queryKeys } from '../lib/queryClient';
import { metadataApi } from '../api/metadata';

function findDefaultView(views: View[]): View {
  return views.find(v => v.isDefault) ?? views[0];
}

function shouldSkipInitialization(
  isInitialized: boolean,
  isLoadingViews: boolean,
  views: View[] | undefined,
  isInitializing: boolean
): boolean {
  return isInitialized || isLoadingViews || !views || isInitializing;
}

export function useAppInitialization() {
  const queryClient = useQueryClient();
  const { data: views, isLoading: isLoadingViews, error: viewsError } = useViews();
  const createViewMutation = useCreateView();
  const setCurrentViewId = useAppStore((state) => state.setCurrentViewId);
  const setInitialized = useAppStore((state) => state.setInitialized);
  const currentViewId = useAppStore((state) => state.currentViewId);
  const isInitialized = useAppStore((state) => state.isInitialized);
  const isInitializingRef = useRef(false);

  useEffect(() => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.metadata.maturityScale(),
      queryFn: () => metadataApi.getMaturityScale(),
      staleTime: Infinity,
    });
  }, [queryClient]);

  const createDefaultView = useCallback(async () => {
    const newView = await createViewMutation.mutateAsync({
      name: 'Default View',
      description: 'Main application view',
    });
    setCurrentViewId(newView.id as ViewId);
    toast.success('Created default view');
  }, [createViewMutation, setCurrentViewId]);

  const selectExistingView = useCallback((availableViews: View[]) => {
    const viewToSelect = findDefaultView(availableViews);
    setCurrentViewId(viewToSelect.id as ViewId);
  }, [setCurrentViewId]);

  useEffect(() => {
    if (shouldSkipInitialization(isInitialized, isLoadingViews, views, isInitializingRef.current)) {
      return;
    }

    isInitializingRef.current = true;

    const initializeView = async () => {
      try {
        if (!views || views.length === 0) {
          await createDefaultView();
        } else {
          selectExistingView(views);
        }
        setInitialized(true);
        toast.success('Data loaded successfully');
      } catch (error) {
        console.error('Failed to initialize:', error);
        toast.error('Failed to initialize application');
        isInitializingRef.current = false;
      }
    };

    initializeView();
  }, [views, isLoadingViews, isInitialized, setInitialized, createDefaultView, selectExistingView]);

  const isLoading = isLoadingViews || (!isInitialized && !viewsError);

  return {
    isLoading,
    error: viewsError,
    isInitialized,
    currentViewId,
  };
}
