import { AxiosError } from 'axios';

export function getErrorMessage(error: unknown, defaultMessage: string): string {
  if (error instanceof AxiosError) {
    const status = error.response?.status;
    const serverMessage = error.response?.data?.error || error.response?.data?.message;

    switch (status) {
      case 400:
        return serverMessage || 'Invalid input. Please check your data and try again.';
      case 403:
        return "You don't have permission to perform this action.";
      case 404:
        return 'The requested capability was not found.';
      case 409:
        return serverMessage || 'A capability with this name already exists.';
      case 500:
        return 'An unexpected error occurred. Please try again later.';
      default:
        return serverMessage || error.message || defaultMessage;
    }
  }

  if (error instanceof Error) {
    return error.message || defaultMessage;
  }

  return defaultMessage;
}
