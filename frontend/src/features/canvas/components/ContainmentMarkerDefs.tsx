import type React from 'react';
import { CONTAINMENT_MARKER_IDS } from '../utils/edgeCreators';
import classes from './ContainmentMarkerDefs.module.css';

const DIAMOND_PATH = 'M0,6 L9,0 L18,6 L9,12 Z';

interface DiamondMarkerProps {
  id: string;
  filled: boolean;
}

const DiamondMarker: React.FC<DiamondMarkerProps> = ({ id, filled }) => (
  <marker
    id={id}
    viewBox="0 0 18 12"
    markerWidth={18}
    markerHeight={12}
    markerUnits="userSpaceOnUse"
    refX={0}
    refY={6}
    orient="auto"
  >
    <path d={DIAMOND_PATH} className={filled ? classes.filledDiamond : classes.hollowDiamond} />
  </marker>
);

export const ContainmentMarkerDefs: React.FC = () => (
  <svg className={classes.defs} aria-hidden="true" data-testid="containment-marker-defs">
    <defs>
      <DiamondMarker id={CONTAINMENT_MARKER_IDS.composition} filled />
      <DiamondMarker id={CONTAINMENT_MARKER_IDS.aggregation} filled={false} />
    </defs>
  </svg>
);
