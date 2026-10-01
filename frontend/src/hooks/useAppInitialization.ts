import { useQueryClient } from '@tanstack/react-query';
import { useCallback, useEffect } from 'react';
import toast from 'react-hot-toast';
import { metadataApi } from '../api/metadata';
import type { View, ViewId } from '../api/types';
import { toViewId } from '../api/types';
import { useCreateView, useViews } from '../features/views/hooks/useViews';
import { metadataQueryKeys } from '../lib/appQueryKeys';
import { clearParams, deepLinkParams, readDeepLink } from '../lib/deepLinks';
import { useAppStore } from '../store/appStore';

let initializationInFlight = false;

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

function consumeViewLink(views: View[]): View | undefined {
  const viewIdFromUrl = readDeepLink(deepLinkParams.VIEW);
  if (viewIdFromUrl === null) return undefined;

  clearParams([deepLinkParams.VIEW.param]);
  if (viewIdFromUrl === '') return undefined;

  const linkedView = views.find((v) => v.id === toViewId(viewIdFromUrl));
  if (!linkedView) toast.error('The linked view does not exist');
  return linkedView;
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

function canInitialize(views: View[] | undefined, isLoadingViews: boolean): views is View[] {
  const settled = useAppStore.getState().isInitialized || initializationInFlight;
  return !settled && !isLoadingViews && views !== undefined;
}

function useLinkedViewOnReturn(views: View[] | undefined, isInitialized: boolean): void {
  const setCurrentViewId = useAppStore((state) => state.setCurrentViewId);
  const openView = useAppStore((state) => state.openView);

  useEffect(() => {
    if (!isInitialized || !views) return;
    const linkedView = consumeViewLink(views);
    if (!linkedView) return;
    openView(linkedView.id);
    setCurrentViewId(linkedView.id);
  }, [isInitialized, views, openView, setCurrentViewId]);
}

export function useAppInitialization() {
  const { data: views, isLoading: isLoadingViews, error: viewsError } = useViews();
  const createViewMutation = useCreateView();
  const setCurrentViewId = useAppStore((state) => state.setCurrentViewId);
  const setOpenViewIds = useAppStore((state) => state.setOpenViewIds);
  const setInitialized = useAppStore((state) => state.setInitialized);
  const currentViewId = useAppStore((state) => state.currentViewId);
  const isInitialized = useAppStore((state) => state.isInitialized);

  usePrefetchMetadata();
  useLinkedViewOnReturn(views, isInitialized);

  const createDefaultView = useCallback(async () => {
    const newView = await createViewMutation.mutateAsync({
      name: 'Default View',
      description: 'Main application view',
    });
    selectInitialView(newView.id, { setCurrentViewId, setOpenViewIds });
    toast.success('Created default view');
  }, [createViewMutation, setCurrentViewId, setOpenViewIds]);

  useEffect(() => {
    if (!canInitialize(views, isLoadingViews)) return;

    initializationInFlight = true;

    const initializeView = async () => {
      try {
        if (views.length === 0) {
          await createDefaultView();
        } else {
          const initialView = consumeViewLink(views) ?? findDefaultView(views);
          selectInitialView(initialView.id, { setCurrentViewId, setOpenViewIds });
        }
        setInitialized(true);
        toast.success('Data loaded successfully');
      } catch (error) {
        console.error('Failed to initialize:', error);
        toast.error('Failed to initialize application');
      } finally {
        initializationInFlight = false;
      }
    };

    initializeView();
  }, [views, isLoadingViews, setInitialized, createDefaultView, setCurrentViewId, setOpenViewIds]);

  return {
    isLoading: isLoadingViews || (!isInitialized && !viewsError),
    error: viewsError,
    isInitialized,
    currentViewId,
  };
}
