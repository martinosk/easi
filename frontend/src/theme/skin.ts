import { clearTokenCache } from './resolveToken';

export type SkinName = 'easi' | 'harbor' | 'evergreen';

export const SKINS: { value: SkinName; label: string }[] = [
  { value: 'easi', label: 'EASI graphite' },
  { value: 'harbor', label: 'Harbor' },
  { value: 'evergreen', label: 'Evergreen' },
];

const STORAGE_KEY = 'easi.skin';

const isSkinName = (value: string | null): value is SkinName => SKINS.some((skin) => skin.value === value);

export function applySkin(name: SkinName): void {
  clearTokenCache();
  if (name === 'easi') {
    delete document.documentElement.dataset.skin;
    return;
  }
  document.documentElement.dataset.skin = name;
}

export function setSkin(name: SkinName): void {
  applySkin(name);
  localStorage.setItem(STORAGE_KEY, name);
}

export function getSkin(): SkinName {
  const current = document.documentElement.dataset.skin;
  return isSkinName(current ?? null) ? (current as SkinName) : 'easi';
}

export function initSkin(): void {
  const stored = localStorage.getItem(STORAGE_KEY);
  applySkin(isSkinName(stored) ? stored : 'easi');
}
