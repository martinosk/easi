import { describe, it, expect } from 'vitest';
import type { Node, Edge } from '@xyflow/react';
import { calculateAutoLayout } from './autoLayout';

describe('calculateAutoLayout', () => {
  it('should layout nodes into stratified layers', () => {
    const nodes: Node[] = [
      { id: 'cap-1', type: 'capability', position: { x: 0, y: 0 }, data: { level: 'L1' } },
      { id: 'cap-2', type: 'capability', position: { x: 0, y: 0 }, data: { level: 'L2' } },
      { id: 'comp-1', type: 'component', position: { x: 0, y: 0 }, data: {} },
      { id: 'orig-1', type: 'originEntity', position: { x: 0, y: 0 }, data: {} },
    ];
    const edges: Edge[] = [
      { id: 'e1-2', source: 'cap-1', target: 'cap-2' },
      { id: 'e2-3', source: 'cap-2', target: 'comp-1' },
      { id: 'e3-4', source: 'comp-1', target: 'orig-1' },
    ];

    const layouted = calculateAutoLayout(nodes, edges);
    const getY = (id: string) => layouted.find((n) => n.id === id)!.position.y;

    expect(getY('cap-1')).toBeLessThan(getY('cap-2'));
    expect(getY('cap-2')).toBeLessThan(getY('comp-1'));
    expect(getY('comp-1')).toBeLessThan(getY('orig-1'));
  });

  it('should throw for unknown node types', () => {
    const nodes: Node[] = [
      { id: 'x', type: 'unknown', position: { x: 0, y: 0 }, data: {} },
    ];

    expect(() => calculateAutoLayout(nodes, [])).toThrow('No layout strategy');
  });
});