import { findLayeringViolations, isLayeringScanTarget } from './layering.ts';

describe('findLayeringViolations', () => {
  it('flags a numeric z-index in a stylesheet', () => {
    const violations = findLayeringViolations([
      { path: 'src/features/x/X.module.css', content: '.a {\n  color: red;\n  z-index: 1000;\n}\n' },
    ]);
    expect(violations).toEqual([{ path: 'src/features/x/X.module.css', line: 3, text: 'z-index: 1000;' }]);
  });

  it('flags a numeric zIndex in a component', () => {
    const violations = findLayeringViolations([
      { path: 'src/features/x/X.tsx', content: "<Badge style={{ top: -8, zIndex: 5 }} />\n" },
    ]);
    expect(violations).toHaveLength(1);
    expect(violations[0].line).toBe(1);
  });

  it('accepts a layer token', () => {
    const violations = findLayeringViolations([
      { path: 'src/features/x/X.module.css', content: '.a { z-index: var(--layer-raised); }\n' },
      { path: 'src/features/x/X.tsx', content: "style={{ zIndex: 'var(--layer-raised)' }}\n" },
    ]);
    expect(violations).toEqual([]);
  });

  it('flags a Mantine or foreign variable so the EASI scale stays the only one', () => {
    const violations = findLayeringViolations([
      { path: 'src/features/x/X.module.css', content: '.a { z-index: var(--mantine-z-index-modal); }\n' },
    ]);
    expect(violations).toHaveLength(1);
  });

  it('reports every offending line in a file', () => {
    const violations = findLayeringViolations([
      { path: 'src/a.css', content: '.a { z-index: 1; }\n.b { z-index: var(--layer-base); }\n.c { z-index: 2; }\n' },
    ]);
    expect(violations.map((v) => v.line)).toEqual([1, 3]);
  });
});

describe('isLayeringScanTarget', () => {
  it('scans source stylesheets and components', () => {
    expect(isLayeringScanTarget('src/features/x/X.module.css')).toBe(true);
    expect(isLayeringScanTarget('src/features/x/X.tsx')).toBe(true);
    expect(isLayeringScanTarget('src/features/x/x.ts')).toBe(true);
  });

  it('exempts the token file and tests', () => {
    expect(isLayeringScanTarget('src/theme/tokens.css')).toBe(false);
    expect(isLayeringScanTarget('src/features/x/X.test.tsx')).toBe(false);
    expect(isLayeringScanTarget('src/theme/layers.test.ts')).toBe(false);
  });
});
