import { describe, it, expect } from 'vitest';
import { calculateDagreLayout } from './layout';
import type { Node, Edge } from '@xyflow/react';

describe('calculateDagreLayout', () => {
  it('should layout nodes in TB direction by default', () => {
    const nodes: Node[] = [
      { id: '1', position: { x: 0, y: 0 }, data: {} },
      { id: '2', position: { x: 0, y: 0 }, data: {} },
    ];
    const edges: Edge[] = [
      { id: 'e1-2', source: '1', target: '2' },
    ];

    const layoutedNodes = calculateDagreLayout(nodes, edges);

    expect(layoutedNodes).toHaveLength(2);
    expect(layoutedNodes[0].id).toBe('1');
    expect(layoutedNodes[1].id).toBe('2');
    expect(layoutedNodes[0].position.y).toBeLessThan(layoutedNodes[1].position.y);
  });

  it('should layout nodes in LR direction', () => {
    const nodes: Node[] = [
      { id: '1', position: { x: 0, y: 0 }, data: {} },
      { id: '2', position: { x: 0, y: 0 }, data: {} },
    ];
    const edges: Edge[] = [
      { id: 'e1-2', source: '1', target: '2' },
    ];

    const layoutedNodes = calculateDagreLayout(nodes, edges, { direction: 'LR' });

    expect(layoutedNodes).toHaveLength(2);
    expect(layoutedNodes[0].position.x).toBeLessThan(layoutedNodes[1].position.x);
  });

  it('should layout nodes in BT direction', () => {
    const nodes: Node[] = [
      { id: '1', position: { x: 0, y: 0 }, data: {} },
      { id: '2', position: { x: 0, y: 0 }, data: {} },
    ];
    const edges: Edge[] = [
      { id: 'e1-2', source: '1', target: '2' },
    ];

    const layoutedNodes = calculateDagreLayout(nodes, edges, { direction: 'BT' });

    expect(layoutedNodes).toHaveLength(2);
    expect(layoutedNodes[0].position.y).toBeGreaterThan(layoutedNodes[1].position.y);
  });

  it('should layout nodes in RL direction', () => {
    const nodes: Node[] = [
      { id: '1', position: { x: 0, y: 0 }, data: {} },
      { id: '2', position: { x: 0, y: 0 }, data: {} },
    ];
    const edges: Edge[] = [
      { id: 'e1-2', source: '1', target: '2' },
    ];

    const layoutedNodes = calculateDagreLayout(nodes, edges, { direction: 'RL' });

    expect(layoutedNodes).toHaveLength(2);
    expect(layoutedNodes[0].position.x).toBeGreaterThan(layoutedNodes[1].position.x);
  });

  it('should handle custom node spacing', () => {
    const nodes: Node[] = [
      { id: '1', position: { x: 0, y: 0 }, data: {} },
      { id: '2', position: { x: 0, y: 0 }, data: {} },
      { id: '3', position: { x: 0, y: 0 }, data: {} },
    ];
    const edges: Edge[] = [
      { id: 'e1-2', source: '1', target: '2' },
      { id: 'e1-3', source: '1', target: '3' },
    ];

    const defaultSpacing = calculateDagreLayout(nodes, edges);
    const widerSpacing = calculateDagreLayout(nodes, edges, { nodeSpacing: 200 });

    const defaultDistance = Math.abs(defaultSpacing[1].position.x - defaultSpacing[2].position.x);
    const widerDistance = Math.abs(widerSpacing[1].position.x - widerSpacing[2].position.x);

    expect(widerDistance).toBeGreaterThan(defaultDistance);
  });

  it('should handle custom rank spacing', () => {
    const nodes: Node[] = [
      { id: '1', position: { x: 0, y: 0 }, data: {} },
      { id: '2', position: { x: 0, y: 0 }, data: {} },
    ];
    const edges: Edge[] = [
      { id: 'e1-2', source: '1', target: '2' },
    ];

    const defaultSpacing = calculateDagreLayout(nodes, edges);
    const widerSpacing = calculateDagreLayout(nodes, edges, { rankSpacing: 300 });

    const defaultDistance = Math.abs(defaultSpacing[0].position.y - defaultSpacing[1].position.y);
    const widerDistance = Math.abs(widerSpacing[0].position.y - widerSpacing[1].position.y);

    expect(widerDistance).toBeGreaterThan(defaultDistance);
  });

  it('should handle empty nodes array', () => {
    const nodes: Node[] = [];
    const edges: Edge[] = [];

    const layoutedNodes = calculateDagreLayout(nodes, edges);

    expect(layoutedNodes).toHaveLength(0);
  });

  it('should handle nodes without edges', () => {
    const nodes: Node[] = [
      { id: '1', position: { x: 0, y: 0 }, data: {} },
      { id: '2', position: { x: 0, y: 0 }, data: {} },
    ];
    const edges: Edge[] = [];

    const layoutedNodes = calculateDagreLayout(nodes, edges);

    expect(layoutedNodes).toHaveLength(2);
    expect(layoutedNodes[0]).toHaveProperty('position');
    expect(layoutedNodes[1]).toHaveProperty('position');
  });

  it('should preserve node data', () => {
    const nodes: Node[] = [
      { id: '1', position: { x: 0, y: 0 }, data: { label: 'Node 1', custom: 'value' } },
      { id: '2', position: { x: 0, y: 0 }, data: { label: 'Node 2' } },
    ];
    const edges: Edge[] = [
      { id: 'e1-2', source: '1', target: '2' },
    ];

    const layoutedNodes = calculateDagreLayout(nodes, edges);

    expect(layoutedNodes[0].data).toEqual({ label: 'Node 1', custom: 'value' });
    expect(layoutedNodes[1].data).toEqual({ label: 'Node 2' });
  });

  it('should handle complex graph with multiple levels', () => {
    const nodes: Node[] = [
      { id: '1', position: { x: 0, y: 0 }, data: {} },
      { id: '2', position: { x: 0, y: 0 }, data: {} },
      { id: '3', position: { x: 0, y: 0 }, data: {} },
      { id: '4', position: { x: 0, y: 0 }, data: {} },
    ];
    const edges: Edge[] = [
      { id: 'e1-2', source: '1', target: '2' },
      { id: 'e2-3', source: '2', target: '3' },
      { id: 'e3-4', source: '3', target: '4' },
    ];

    const layoutedNodes = calculateDagreLayout(nodes, edges);

    expect(layoutedNodes).toHaveLength(4);
    expect(layoutedNodes[0].position.y).toBeLessThan(layoutedNodes[1].position.y);
    expect(layoutedNodes[1].position.y).toBeLessThan(layoutedNodes[2].position.y);
    expect(layoutedNodes[2].position.y).toBeLessThan(layoutedNodes[3].position.y);
  });
});
