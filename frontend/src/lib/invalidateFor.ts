import type { QueryClient } from '@tanstack/react-query';

export function invalidateFor(
  queryClient: QueryClient,
  keys: ReadonlyArray<readonly unknown[]>
): void {
  keys.forEach(key => queryClient.invalidateQueries({ queryKey: key }));
}
