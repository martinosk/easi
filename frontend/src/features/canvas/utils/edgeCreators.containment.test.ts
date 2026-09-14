import type { Node } from '@xyflow/react';
import { describe, expect, it } from 'vitest';
import type { ComponentId, ViewComponent } from '../../../api/types';
import { buildComponent } from '../../../test/helpers/entityBuilders';
import { CONTAINMENT_MARKER_IDS, createContainmentEdges, type EdgeCreationContext } from './edgeCreators';

const node = (id: string): Node => ({ id, position: { x: 0, y: 0 }, data: {} });
const onCanvas = (...ids: string[]): ViewComponent[] => ids.map((id) => ({ componentId: id as ComponentId, x: 0, y: 0 }));

const crm = buildComponent({ id: 'crm' as ComponentId, name: 'CRM Suite' });
const quoting = buildComponent({
  id: 'quoting' as ComponentId,
  name: 'Quoting',
  partOf: { id: 'crm' as ComponentId, name: 'CRM Suite', kind: 'composition' },
});
const billing = buildComponent({
  id: 'billing' as ComponentId,
  name: 'Billing',
  partOf: { id: 'crm' as ComponentId, name: 'CRM Suite', kind: 'aggregation' },
});

const ctx = (overrides: Partial<EdgeCreationContext> = {}): EdgeCreationContext => ({
  nodes: [node('crm'), node('quoting'), node('billing')],
  selectedEdgeId: null,
  edgeType: 'default',
  isClassicScheme: false,
  ...overrides,
});

describe('createContainmentEdges', () => {
  it('joins each part on the canvas to its parent, labelled in UML wording', () => {
    const edges = createContainmentEdges(onCanvas('crm', 'quoting', 'billing'), [crm, quoting, billing], ctx());

    expect(edges.map((e) => [e.id, e.source, e.target, e.label])).toEqual([
      ['containment-crm-quoting', 'crm', 'quoting', 'Composes'],
      ['containment-crm-billing', 'crm', 'billing', 'Aggregates'],
    ]);
  });

  it('puts a UML diamond at the parent end and no arrowhead at the part end', () => {
    const [composition, aggregation] = createContainmentEdges(
      onCanvas('crm', 'quoting', 'billing'),
      [crm, quoting, billing],
      ctx(),
    );

    expect(composition.markerStart).toBe(CONTAINMENT_MARKER_IDS.composition);
    expect(aggregation.markerStart).toBe(CONTAINMENT_MARKER_IDS.aggregation);
    expect(composition.markerEnd).toBeUndefined();
    expect(aggregation.markerEnd).toBeUndefined();
  });

  it('omits the edge when the parent or the part is not on the canvas', () => {
    expect(createContainmentEdges(onCanvas('quoting'), [crm, quoting], ctx())).toEqual([]);
    expect(createContainmentEdges(onCanvas('crm'), [crm, quoting], ctx())).toEqual([]);
  });

  it('marks the selected edge as animated', () => {
    const [edge] = createContainmentEdges(
      onCanvas('crm', 'quoting'),
      [crm, quoting],
      ctx({ selectedEdgeId: 'containment-crm-quoting' }),
    );

    expect(edge.animated).toBe(true);
  });
});
