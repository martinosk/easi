import { useState, useEffect, useRef, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { httpClient } from '../../../api/core/httpClient';
import { invalidateFor } from '../../../lib/invalidateFor';
import { importsMutationEffects } from '../mutationEffects';
import type { ImportSession, CreateImportSessionRequest, ImportSessionId } from '../types';

interface UseImportSessionReturn {
  session: ImportSession | null;
  isLoading: boolean;
  error: string | null;
  createSession: (request: CreateImportSessionRequest) => Promise<void>;
  confirmSession: () => Promise<void>;
  cancelSession: () => Promise<void>;
  reset: () => void;
}

const POLL_INTERVAL = 2000;

function toErrorMessage(err: unknown, fallback: string): string {
  return err instanceof Error ? err.message : fallback;
}

function buildFormData(request: CreateImportSessionRequest): FormData {
  const formData = new FormData();
  formData.append('file', request.file);
  formData.append('sourceFormat', request.sourceFormat);
  if (request.businessDomainId) formData.append('businessDomainId', request.businessDomainId);
  if (request.capabilityEAOwner) formData.append('capabilityEAOwner', request.capabilityEAOwner);
  return formData;
}

function isStillImporting(session: ImportSession): boolean {
  return session.status === 'importing';
}

interface PollHandlers {
  onImporting: () => void;
  onCompleted: () => void;
  onFinished: () => void;
}

function processPollResult(data: ImportSession, handlers: PollHandlers): void {
  if (isStillImporting(data)) {
    handlers.onImporting();
  } else {
    handlers.onFinished();
    if (data.status === 'completed') handlers.onCompleted();
  }
}

function useSessionPolling(
  setSession: React.Dispatch<React.SetStateAction<ImportSession | null>>,
  setError: React.Dispatch<React.SetStateAction<string | null>>
) {
  const queryClient = useQueryClient();
  const pollTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const isMountedRef = useRef(true);

  const stopPolling = useCallback(() => {
    if (pollTimerRef.current) {
      clearTimeout(pollTimerRef.current);
      pollTimerRef.current = null;
    }
  }, []);

  const pollSession = useCallback(async (sessionId: ImportSessionId) => {
    if (!isMountedRef.current) return;
    try {
      const { data } = await httpClient.get<ImportSession>(`/api/v1/imports/${sessionId}`);
      if (!isMountedRef.current) return;
      setSession(data);
      processPollResult(data, {
        onImporting: () => { pollTimerRef.current = setTimeout(() => pollSession(sessionId), POLL_INTERVAL); },
        onCompleted: () => invalidateFor(queryClient, importsMutationEffects.completed()),
        onFinished: stopPolling,
      });
    } catch (err) {
      if (!isMountedRef.current) return;
      setError(toErrorMessage(err, 'Failed to fetch session status'));
      stopPolling();
    }
  }, [setSession, setError, stopPolling, queryClient]);

  const startPollingIfImporting = useCallback((data: ImportSession) => {
    if (isStillImporting(data)) {
      pollTimerRef.current = setTimeout(() => pollSession(data.id), POLL_INTERVAL);
    }
  }, [pollSession]);

  useEffect(() => {
    isMountedRef.current = true;
    return () => { isMountedRef.current = false; stopPolling(); };
  }, [stopPolling]);

  return { stopPolling, startPollingIfImporting };
}

export function useImportSession(): UseImportSessionReturn {
  const [session, setSession] = useState<ImportSession | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { stopPolling, startPollingIfImporting } = useSessionPolling(setSession, setError);

  const withLoading = useCallback(async (fn: () => Promise<void>, fallbackError: string) => {
    setIsLoading(true);
    setError(null);
    try {
      await fn();
    } catch (err) {
      setError(toErrorMessage(err, fallbackError));
    } finally {
      setIsLoading(false);
    }
  }, []);

  const createSession = (request: CreateImportSessionRequest): Promise<void> =>
    withLoading(async () => {
      const response = await httpClient.post<ImportSession>('/api/v1/imports', buildFormData(request), {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      setSession(response.data);
      startPollingIfImporting(response.data);
    }, 'Failed to create import session');

  const confirmSession = (): Promise<void> => {
    const confirmHref = session?._links.confirm?.href;
    if (!confirmHref) {
      setError('Cannot confirm: session not found or already started');
      return Promise.resolve();
    }
    return withLoading(async () => {
      const response = await httpClient.post<ImportSession>(confirmHref);
      setSession(response.data);
      startPollingIfImporting(response.data);
    }, 'Failed to confirm import');
  };

  const cancelSession = (): Promise<void> => {
    const deleteHref = session?._links.delete?.href;
    if (!deleteHref) {
      setError('Cannot cancel: session not found');
      return Promise.resolve();
    }
    return withLoading(async () => {
      await httpClient.delete(deleteHref);
      setSession(null);
      stopPolling();
    }, 'Failed to cancel import');
  };

  const reset = () => {
    stopPolling();
    setSession(null);
    setError(null);
    setIsLoading(false);
  };

  return { session, isLoading, error, createSession, confirmSession, cancelSession, reset };
}
