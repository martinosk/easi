import toast from 'react-hot-toast';
import { ApiError } from '../../api/types';

export async function handleApiCall<T>(
  apiCall: () => Promise<T>,
  errorMessage: string
): Promise<T> {
  try {
    return await apiCall();
  } catch (error) {
    const message = error instanceof ApiError ? error.message : errorMessage;
    toast.error(message);
    throw error;
  }
}

interface OptimisticUpdateOptions<T> {
  apiCall: () => Promise<T>;
  onSuccess: (result: T) => void;
  onError: () => void;
  successMessage: string;
  errorMessage: string;
}

export async function optimisticUpdate<T>(
  options: OptimisticUpdateOptions<T>
): Promise<T> {
  const { apiCall, onSuccess, onError, successMessage, errorMessage } = options;

  try {
    const result = await apiCall();
    onSuccess(result);
    toast.success(successMessage);
    return result;
  } catch (error) {
    onError();
    const message = error instanceof ApiError ? error.message : errorMessage;
    toast.error(message);
    throw error;
  }
}
