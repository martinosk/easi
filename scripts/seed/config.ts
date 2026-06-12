function getArg(flag: string): string | undefined {
  const idx = process.argv.indexOf(flag);
  if (idx !== -1 && process.argv[idx + 1]) return process.argv[idx + 1];
  return undefined;
}

export const BASE_URL = getArg("--base-url") ?? "http://localhost:8080";
export const TENANT_ID = getArg("--tenant-id") ?? "acme";
export const SESSION_COOKIE = getArg("--cookie");
export const BYPASS_MODE = process.argv.includes("--bypass");
export const API_URL = `${BASE_URL}/api/v1`;

export function buildHeaders(): Record<string, string> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (SESSION_COOKIE) headers["Cookie"] = `easi_session=${SESSION_COOKIE}`;
  else if (BYPASS_MODE) headers["X-Tenant-ID"] = TENANT_ID;
  return headers;
}

export async function apiCall<T>(method: string, path: string, body?: unknown): Promise<T> {
  const url = `${API_URL}${path}`;
  const options: RequestInit = { method, headers: buildHeaders() };
  if (body) options.body = JSON.stringify(body);

  const response = await fetch(url, options);
  if (!response.ok) {
    const text = await response.text();
    throw new Error(`API call failed: ${method} ${path} - ${response.status}: ${text}`);
  }

  if (response.status === 204 || response.status === 201) {
    const text = await response.text();
    if (!text) return {} as T;
    try { return JSON.parse(text); } catch { return {} as T; }
  }

  return response.json();
}

export async function apiCallWithEtag(
  method: "PUT" | "PATCH",
  url: string,
  body: unknown,
  etag: string,
  errorContext: string
): Promise<string> {
  const headers = buildHeaders();
  headers["If-Match"] = etag;
  const response = await fetch(url, { method, headers, body: JSON.stringify(body) });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(`${errorContext}: ${response.status}: ${text}`);
  }
  return response.headers.get("etag") || etag;
}

export async function parallelBatch<T, R>(
  items: T[],
  batchSize: number,
  fn: (item: T) => Promise<R>
): Promise<R[]> {
  const results: R[] = [];
  for (let i = 0; i < items.length; i += batchSize) {
    const batch = items.slice(i, i + batchSize);
    const batchResults = await Promise.all(batch.map(fn));
    results.push(...batchResults);
  }
  return results;
}
