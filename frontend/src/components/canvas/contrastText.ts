const HEX_PATTERN = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i;

const INK_TEXT = 'var(--ink)';
const LIGHT_TEXT = '#ffffff';

const toLinearChannel = (channel: number): number => {
  const normalized = channel / 255;
  return normalized <= 0.03928 ? normalized / 12.92 : ((normalized + 0.055) / 1.055) ** 2.4;
};

const relativeLuminance = (hex: string): number | null => {
  const match = HEX_PATTERN.exec(hex);
  if (!match) return null;

  const [r, g, b] = [match[1], match[2], match[3]].map((part) => toLinearChannel(parseInt(part, 16)));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
};

const INK_LUMINANCE = 0.0175;
const WHITE_LUMINANCE = 1;

const contrastRatio = (a: number, b: number): number => (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);

export const getContrastTextColor = (backgroundHex: string): string => {
  const luminance = relativeLuminance(backgroundHex);
  if (luminance === null) return INK_TEXT;
  return contrastRatio(luminance, INK_LUMINANCE) >= contrastRatio(luminance, WHITE_LUMINANCE)
    ? INK_TEXT
    : LIGHT_TEXT;
};
