import { httpClient } from './httpClient';
import type { PaginatedResponse } from '../types';

export interface FetchAllOptions {
  queryParams?: Record<string, string | undefined>;
}

export async function fetchAllPaginated<T>(
  baseUrl: string,
  options: FetchAllOptions = {}
): Promise<T[]> {
  const allItems: T[] = [];
  let cursor: string | undefined;

  do {
    const params = new URLSearchParams();
    if (cursor) {
      params.append('after', cursor);
    }

    if (options.queryParams) {
      Object.entries(options.queryParams).forEach(([key, value]) => {
        if (value !== undefined) {
          params.append(key, value);
        }
      });
    }

    const queryString = params.toString();
    const url = queryString ? `${baseUrl}?${queryString}` : baseUrl;
    const response = await httpClient.get<PaginatedResponse<T>>(url);
    const page = response.data;

    allItems.push(...(page.data || []));

    cursor = page.pagination?.hasMore ? page.pagination.cursor : undefined;
  } while (cursor);

  return allItems;
}
