import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { getDefaultZIndex } from '@mantine/core';

const tokens = readFileSync(resolve(process.cwd(), 'src/theme/tokens.css'), 'utf8');

function layer(name: string): number {
  const match = tokens.match(new RegExp(`--layer-${name}:\\s*(\\d+);`));
  if (!match) throw new Error(`--layer-${name} is not defined in tokens.css`);
  return Number(match[1]);
}

describe('layer tokens', () => {
  it('orders in-region layers below the shell', () => {
    expect(layer('base')).toBeLessThan(layer('raised'));
    expect(layer('raised')).toBeLessThan(layer('floating'));
    expect(layer('floating')).toBeLessThan(layer('pinned'));
    expect(layer('pinned')).toBeLessThan(layer('shell'));
  });

  it('orders shell layers below panels and popovers', () => {
    expect(layer('shell')).toBeLessThan(layer('panel'));
    expect(layer('panel')).toBeLessThan(layer('popover'));
  });

  it("matches Mantine's default scale", () => {
    expect(layer('shell')).toBe(getDefaultZIndex('app'));
    expect(layer('panel')).toBe(getDefaultZIndex('modal'));
    expect(layer('popover')).toBe(getDefaultZIndex('popover'));
  });
});
