import { beforeEach, describe, expect, it } from 'vitest';
import { applySkin, getSkin, initSkin, SKINS, setSkin } from './skin';

describe('skin', () => {
  beforeEach(() => {
    localStorage.clear();
    delete document.documentElement.dataset.skin;
  });

  it('exposes the shipped skins', () => {
    expect(SKINS).toEqual([
      { value: 'easi', label: 'EASI graphite' },
      { value: 'harbor', label: 'Harbor' },
      { value: 'evergreen', label: 'Evergreen' },
    ]);
  });

  it('defaults to easi when nothing is persisted', () => {
    initSkin();

    expect(getSkin()).toBe('easi');
    expect(document.documentElement.dataset.skin).toBeUndefined();
  });

  it('applies and persists a non-default skin', () => {
    setSkin('harbor');

    expect(document.documentElement.dataset.skin).toBe('harbor');
    expect(getSkin()).toBe('harbor');
    expect(localStorage.getItem('easi.skin')).toBe('harbor');
  });

  it('reads the persisted skin on init', () => {
    localStorage.setItem('easi.skin', 'evergreen');

    initSkin();

    expect(getSkin()).toBe('evergreen');
    expect(document.documentElement.dataset.skin).toBe('evergreen');
  });

  it('falls back to easi for an invalid persisted value', () => {
    localStorage.setItem('easi.skin', 'not-a-real-skin');

    initSkin();

    expect(getSkin()).toBe('easi');
    expect(document.documentElement.dataset.skin).toBeUndefined();
  });

  it('clears the dataset attribute when switching back to easi', () => {
    setSkin('harbor');
    applySkin('easi');

    expect(document.documentElement.dataset.skin).toBeUndefined();
  });
});
