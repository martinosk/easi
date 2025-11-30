import { useState, useEffect, useRef } from 'react';
import axios, { type AxiosInstance, type AxiosError } from 'axios';
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

export function useImportSession(): UseImportSessionReturn {
  const [session, setSession] = useState<ImportSession | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const pollTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const clientRef = useRef<AxiosInstance | null>(null);

  if (!clientRef.current) {
    clientRef.current = axios.create({
      baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    clientRef.current.interceptors.response.use(
      (response) => response,
      (error: AxiosError) => {
        const message = extractErrorMessage(error);
        throw new Error(message);
      }
    );
  }

  const client = clientRef.current;

  const stopPolling = () => {
    if (pollTimerRef.current) {
      clearTimeout(pollTimerRef.current);
      pollTimerRef.current = null;
    }
  };

  const pollSession = async (sessionId: ImportSessionId) => {
    try {
      const response = await client.get<ImportSession>(`/api/v1/imports/${sessionId}`);
      setSession(response.data);

      if (response.data.status === 'importing') {
        pollTimerRef.current = setTimeout(() => pollSession(sessionId), POLL_INTERVAL);
      } else {
        stopPolling();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch session status');
      stopPolling();
    }
  };

  const createSession = async (request: CreateImportSessionRequest): Promise<void> => {
    setIsLoading(true);
    setError(null);

    try {
      const formData = new FormData();
      formData.append('file', request.file);
      formData.append('sourceFormat', request.sourceFormat);
      if (request.businessDomainId) {
        formData.append('businessDomainId', request.businessDomainId);
      }

      const response = await client.post<ImportSession>('/api/v1/imports', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });

      setSession(response.data);

      if (response.data.status === 'importing') {
        pollTimerRef.current = setTimeout(() => pollSession(response.data.id), POLL_INTERVAL);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create import session');
    } finally {
      setIsLoading(false);
    }
  };

  const confirmSession = async (): Promise<void> => {
    if (!session || !session._links.confirm) {
      setError('Cannot confirm: session not found or already started');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await client.post<ImportSession>(session._links.confirm);
      setSession(response.data);

      if (response.data.status === 'importing') {
        pollTimerRef.current = setTimeout(() => pollSession(response.data.id), POLL_INTERVAL);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to confirm import');
    } finally {
      setIsLoading(false);
    }
  };

  const cancelSession = async (): Promise<void> => {
    if (!session || !session._links.delete) {
      setError('Cannot cancel: session not found');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      await client.delete(session._links.delete);
      setSession(null);
      stopPolling();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to cancel import');
    } finally {
      setIsLoading(false);
    }
  };

  const reset = () => {
    stopPolling();
    setSession(null);
    setError(null);
    setIsLoading(false);
  };

  useEffect(() => {
    return () => {
      stopPolling();
    };
  }, []);

  return {
    session,
    isLoading,
    error,
    createSession,
    confirmSession,
    cancelSession,
    reset,
  };
}

function extractErrorMessage(error: AxiosError): string {
  const responseData = error.response?.data;

  if (!responseData || typeof responseData !== 'object') {
    return error.message || 'An unknown error occurred';
  }

  const errorData = responseData as { message?: string; error?: string; details?: Record<string, string> };

  if (errorData.message) return errorData.message;
  if (errorData.error) return errorData.error;

  const detailMessages = errorData.details ? Object.values(errorData.details).filter(Boolean) : [];
  return detailMessages.length > 0 ? detailMessages.join(', ') : 'An error occurred';
}
