const cache = new Map<string, string>();

export function resolveToken(name: string, fallback: string): string {
  const hit = cache.get(name);
  if (hit) return hit;
  const value = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  if (value) {
    cache.set(name, value);
    return value;
  }
  return fallback;
}

export function clearTokenCache(): void {
  cache.clear();
}
