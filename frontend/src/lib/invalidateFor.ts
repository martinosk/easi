import type { QueryClient } from '@tanstack/react-query';

export function invalidateFor(queryClient: QueryClient, keys: ReadonlyArray<readonly unknown[]>): void {
  for (const key of keys) {
    queryClient.invalidateQueries({ queryKey: key });
  }
}
