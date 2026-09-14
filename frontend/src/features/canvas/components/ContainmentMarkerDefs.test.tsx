import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { CONTAINMENT_MARKER_IDS } from '../utils/edgeCreators';
import { ContainmentMarkerDefs } from './ContainmentMarkerDefs';

describe('ContainmentMarkerDefs', () => {
  it('defines one diamond marker per containment kind under the ids the edges reference', () => {
    const { container } = render(<ContainmentMarkerDefs />);

    const composition = container.querySelector(`marker#${CONTAINMENT_MARKER_IDS.composition}`);
    const aggregation = container.querySelector(`marker#${CONTAINMENT_MARKER_IDS.aggregation}`);
    expect(composition?.querySelector('path')?.getAttribute('class')).toMatch(/filledDiamond/);
    expect(aggregation?.querySelector('path')?.getAttribute('class')).toMatch(/hollowDiamond/);
  });
});
