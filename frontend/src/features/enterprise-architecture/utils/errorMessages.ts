import { AxiosError } from 'axios';

type StatusResolver = (serverMessage: string | undefined, defaultMessage: string) => string;

const statusMessages: Record<number, StatusResolver> = {
  400: (serverMessage) => serverMessage || 'Invalid input. Please check your data and try again.',
  403: () => "You don't have permission to perform this action.",
  404: () => 'The requested capability was not found.',
  409: (serverMessage) => serverMessage || 'A capability with this name already exists.',
  500: () => 'An unexpected error occurred. Please try again later.',
};

function extractServerMessage(error: AxiosError): string | undefined {
  const data = error.response?.data as { error?: string; message?: string } | undefined;
  return data?.error || data?.message;
}

function getAxiosErrorMessage(error: AxiosError, defaultMessage: string): string {
  const status = error.response?.status;
  const serverMessage = extractServerMessage(error);

  const resolver = status !== undefined ? statusMessages[status] : undefined;
  if (resolver) {
    return resolver(serverMessage, defaultMessage);
  }

  return serverMessage || error.message || defaultMessage;
}

export function getErrorMessage(error: unknown, defaultMessage: string): string {
  if (error instanceof AxiosError) {
    return getAxiosErrorMessage(error, defaultMessage);
  }

  if (error instanceof Error) {
    return error.message || defaultMessage;
  }

  return defaultMessage;
}
