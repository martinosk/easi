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

export async function optimisticUpdate<T>(
  apiCall: () => Promise<T>,
  onSuccess: (result: T) => void,
  onError: () => void,
  successMessage: string,
  errorMessage: string
): Promise<T> {
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
