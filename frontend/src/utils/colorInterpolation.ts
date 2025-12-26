interface HSL {
  h: number;
  s: number;
  l: number;
}

const hexToRgb = (hex: string): { r: number; g: number; b: number } => {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  if (!result) {
    throw new Error(`Invalid hex color: ${hex}`);
  }
  return {
    r: parseInt(result[1], 16) / 255,
    g: parseInt(result[2], 16) / 255,
    b: parseInt(result[3], 16) / 255,
  };
};

const rgbToHsl = (r: number, g: number, b: number): HSL => {
  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);
  let h = 0;
  let s = 0;
  const l = (max + min) / 2;

  if (max !== min) {
    const d = max - min;
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min);

    switch (max) {
      case r:
        h = ((g - b) / d + (g < b ? 6 : 0)) / 6;
        break;
      case g:
        h = ((b - r) / d + 2) / 6;
        break;
      case b:
        h = ((r - g) / d + 4) / 6;
        break;
    }
  }

  return { h: h * 360, s: s * 100, l: l * 100 };
};

const hslToRgb = (h: number, s: number, l: number): { r: number; g: number; b: number } => {
  h = h / 360;
  s = s / 100;
  l = l / 100;

  let r: number, g: number, b: number;

  if (s === 0) {
    r = g = b = l;
  } else {
    const hue2rgb = (p: number, q: number, t: number): number => {
      if (t < 0) t += 1;
      if (t > 1) t -= 1;
      if (t < 1 / 6) return p + (q - p) * 6 * t;
      if (t < 1 / 2) return q;
      if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
      return p;
    };

    const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
    const p = 2 * l - q;

    r = hue2rgb(p, q, h + 1 / 3);
    g = hue2rgb(p, q, h);
    b = hue2rgb(p, q, h - 1 / 3);
  }

  return { r, g, b };
};

const rgbToHex = (r: number, g: number, b: number): string => {
  const toHex = (c: number) => {
    const hex = Math.round(c * 255).toString(16);
    return hex.length === 1 ? '0' + hex : hex;
  };
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`;
};

export const hexToHsl = (hex: string): HSL => {
  const { r, g, b } = hexToRgb(hex);
  return rgbToHsl(r, g, b);
};

export const hslToHex = (hsl: HSL): string => {
  const { r, g, b } = hslToRgb(hsl.h, hsl.s, hsl.l);
  return rgbToHex(r, g, b);
};

export const interpolateHsl = (color1: string, color2: string, ratio: number): string => {
  const clampedRatio = Math.max(0, Math.min(1, ratio));

  const hsl1 = hexToHsl(color1);
  const hsl2 = hexToHsl(color2);

  const hueDiff = hsl2.h - hsl1.h;
  const shortestHueDiff = ((hueDiff + 180) % 360) - 180;

  const interpolatedHsl: HSL = {
    h: (hsl1.h + shortestHueDiff * clampedRatio + 360) % 360,
    s: hsl1.s + (hsl2.s - hsl1.s) * clampedRatio,
    l: hsl1.l + (hsl2.l - hsl1.l) * clampedRatio,
  };

  return hslToHex(interpolatedHsl);
};

export const clampMaturityValue = (value: number): number => {
  return Math.max(0, Math.min(99, value));
};
