import type { CSSProperties } from 'react';
import type { CapabilityJourney } from '../types';
import classes from './JourneyProgressBar.module.css';

export function JourneyProgressBar({ journey }: { journey: CapabilityJourney }) {
  const isDone = journey.status === 'done';
  if (!isDone && journey.status !== 'in-flight') return null;
  const value = isDone ? 100 : journey.progress;
  if (value === null) return null;

  return (
    <div className={classes.bar} data-testid="journey-progress-bar">
      <div
        className={isDone ? classes.fillDone : classes.fillFlight}
        data-testid="journey-progress-fill"
        data-fill={isDone ? 'done' : 'in-flight'}
        style={{ '--journey-progress': `${value}%` } as CSSProperties}
      />
    </div>
  );
}
