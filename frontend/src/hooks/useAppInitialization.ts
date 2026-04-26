import { useQueryClient } from '@tanstack/react-query';
import { useCallback, useEffect, useRef } from 'react';
import toast from 'react-hot-toast';
import { metadataApi } from '../api/metadata';
import type { View, ViewId } from '../api/types';
import { toViewId } from '../api/types';
import { useCreateView, useViews } from '../features/views/hooks/useViews';
import { metadataQueryKeys } from '../lib/appQueryKeys';
import { clearParams, deepLinkParams, getParamValue } from '../lib/deepLinks';
import { useAppStore } from '../store/appStore';

function findDefaultView(views: View[]): View {
  return views.find((v) => v.isDefault) ?? views[0];
}

interface InitialViewSelector {
  setCurrentViewId: (id: ViewId | null) => void;
  setOpenViewIds: (ids: ViewId[]) => void;
}

function selectInitialView(viewId: ViewId, selector: InitialViewSelector): void {
  selector.setCurrentViewId(viewId);
  selector.setOpenViewIds([viewId]);
}

function resolveViewFromDeepLink(views: View[], selector: InitialViewSelector): void {
  const viewIdFromUrl = getParamValue(deepLinkParams.VIEW.param);
  if (!viewIdFromUrl) {
    selectInitialView(findDefaultView(views).id, selector);
    return;
  }

  const linkedView = views.find((v) => v.id === toViewId(viewIdFromUrl));
  if (linkedView) {
    selectInitialView(linkedView.id, selector);
  } else {
    toast.error('The linked view does not exist');
    selectInitialView(findDefaultView(views).id, selector);
  }
  clearParams([deepLinkParams.VIEW.param]);
}

function usePrefetchMetadata(): void {
  const queryClient = useQueryClient();
  useEffect(() => {
    queryClient.prefetchQuery({
      queryKey: metadataQueryKeys.maturityScale(),
      queryFn: () => metadataApi.getMaturityScale(),
      staleTime: Infinity,
    });
  }, [queryClient]);
}

function canInitialize(
  isInitialized: boolean,
  isLoadingViews: boolean,
  views: View[] | undefined,
  isInitializing: boolean,
): boolean {
  return !isInitialized && !isLoadingViews && !!views && !isInitializing;
}

export function useAppInitialization() {
  const { data: views, isLoading: isLoadingViews, error: viewsError } = useViews();
  const createViewMutation = useCreateView();
  const setCurrentViewId = useAppStore((state) => state.setCurrentViewId);
  const setOpenViewIds = useAppStore((state) => state.setOpenViewIds);
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
    selectInitialView(newView.id, { setCurrentViewId, setOpenViewIds });
    toast.success('Created default view');
  }, [createViewMutation, setCurrentViewId, setOpenViewIds]);

  useEffect(() => {
    if (!canInitialize(isInitialized, isLoadingViews, views, isInitializingRef.current) || !views) return;

    isInitializingRef.current = true;
    const availableViews = views;

    const initializeView = async () => {
      try {
        if (availableViews.length === 0) {
          await createDefaultView();
        } else {
          resolveViewFromDeepLink(availableViews, { setCurrentViewId, setOpenViewIds });
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
  }, [
    views,
    isLoadingViews,
    isInitialized,
    setInitialized,
    createDefaultView,
    setCurrentViewId,
    setOpenViewIds,
  ]);

  return {
    isLoading: isLoadingViews || (!isInitialized && !viewsError),
    error: viewsError,
    isInitialized,
    currentViewId,
  };
}
