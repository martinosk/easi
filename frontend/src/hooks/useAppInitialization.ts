import { useEffect, useCallback, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useViews, useCreateView } from '../features/views/hooks/useViews';
import { useAppStore } from '../store/appStore';
import type { View, ViewId } from '../api/types';
import { toViewId } from '../api/types';
import toast from 'react-hot-toast';
import { queryKeys } from '../lib/queryClient';
import { metadataApi } from '../api/metadata';
import { getParamValue, clearParams, deepLinkParams } from '../lib/deepLinks';

function findDefaultView(views: View[]): View {
  return views.find(v => v.isDefault) ?? views[0];
}

function resolveViewFromDeepLink(views: View[], setCurrentViewId: (id: ViewId) => void): void {
  const viewIdFromUrl = getParamValue(deepLinkParams.VIEW.param);
  if (!viewIdFromUrl) {
    setCurrentViewId(findDefaultView(views).id);
    return;
  }

  const linkedView = views.find(v => v.id === toViewId(viewIdFromUrl));
  if (linkedView) {
    setCurrentViewId(linkedView.id);
  } else {
    toast.error('The linked view does not exist');
    setCurrentViewId(findDefaultView(views).id);
  }
  clearParams([deepLinkParams.VIEW.param]);
}

function usePrefetchMetadata(): void {
  const queryClient = useQueryClient();
  useEffect(() => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.metadata.maturityScale(),
      queryFn: () => metadataApi.getMaturityScale(),
      staleTime: Infinity,
    });
  }, [queryClient]);
}

function canInitialize(
  isInitialized: boolean,
  isLoadingViews: boolean,
  views: View[] | undefined,
  isInitializing: boolean
): boolean {
  return !isInitialized && !isLoadingViews && !!views && !isInitializing;
}

export function useAppInitialization() {
  const { data: views, isLoading: isLoadingViews, error: viewsError } = useViews();
  const createViewMutation = useCreateView();
  const setCurrentViewId = useAppStore((state) => state.setCurrentViewId);
  const setInitialized = useAppStore((state) => state.setInitialized);
  const currentViewId = useAppStore((state) => state.currentViewId);
  const isInitialized = useAppStore((state) => state.isInitialized);
  const isInitializingRef = useRef(false);

  usePrefetchMetadata();

  const createDefaultView = useCallback(async () => {
    const newView = await createViewMutation.mutateAsync({
      name: 'Default View',
      description: 'Main application view',
    });
    setCurrentViewId(newView.id);
    toast.success('Created default view');
  }, [createViewMutation, setCurrentViewId]);

  useEffect(() => {
    if (!canInitialize(isInitialized, isLoadingViews, views, isInitializingRef.current) || !views) return;

    isInitializingRef.current = true;
    const availableViews = views;

    const initializeView = async () => {
      try {
        if (availableViews.length === 0) {
          await createDefaultView();
        } else {
          resolveViewFromDeepLink(availableViews, setCurrentViewId);
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
  }, [views, isLoadingViews, isInitialized, setInitialized, createDefaultView, setCurrentViewId]);

  return {
    isLoading: isLoadingViews || (!isInitialized && !viewsError),
    error: viewsError,
    isInitialized,
    currentViewId,
  };
}
