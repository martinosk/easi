export const getPersistedBoolean = (key: string, defaultValue: boolean): boolean => {
  const saved = localStorage.getItem(key);
  return saved !== null ? JSON.parse(saved) : defaultValue;
};

export const getPersistedSet = (key: string): Set<string> => {
  const saved = localStorage.getItem(key);
  return saved ? new Set(JSON.parse(saved)) : new Set();
};

export const persistBoolean = (key: string, value: boolean): void => {
  localStorage.setItem(key, JSON.stringify(value));
};

export const persistSet = (key: string, value: Set<string>): void => {
  localStorage.setItem(key, JSON.stringify([...value]));
};

export const getContextMenuPosition = (e: React.MouseEvent) => {
  e.preventDefault();
  e.stopPropagation();
  return { x: e.clientX, y: e.clientY };
};

export const hasCustomColor = (colorScheme: string | undefined, customColor: string | undefined | null): boolean =>
  colorScheme === 'custom' && customColor !== undefined && customColor !== null && customColor !== '';
