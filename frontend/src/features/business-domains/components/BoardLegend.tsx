import type { BoardLens } from '../lens/boardLens';
import classes from './BoardLegend.module.css';

interface SwatchItem {
  swatch: keyof typeof classes;
  label: string;
}

interface BadgeItem {
  badge: keyof typeof classes;
  letter: string;
  label: string;
}

const NOW_SWATCHES: SwatchItem[] = [
  { swatch: 'swatchFull', label: 'Full' },
  { swatch: 'swatchPartial', label: 'Partial' },
  { swatch: 'swatchPlanned', label: 'Planned' },
  { swatch: 'swatchInherited', label: 'Inherited' },
];

const STATUS_SWATCHES: SwatchItem[] = [
  { swatch: 'swatchDone', label: 'standard / done' },
  { swatch: 'swatchFlight', label: 'in transition' },
  { swatch: 'swatchIdle', label: 'unchanged' },
];

const MOVE_SWATCHES: SwatchItem[] = [
  { swatch: 'swatchLeaving', label: 'moving out' },
  { swatch: 'swatchIncoming', label: 'arriving' },
];

const TIME_BADGES: BadgeItem[] = [
  { badge: 'gradeInvest', letter: 'I', label: 'invest' },
  { badge: 'gradeTolerate', letter: 'T', label: 'tolerate' },
  { badge: 'gradeMigrate', letter: 'M', label: 'migrate' },
  { badge: 'gradeEliminate', letter: 'E', label: 'eliminate' },
];

function Swatch({ item }: { item: SwatchItem }) {
  return (
    <span className={classes.legendItem}>
      <span className={[classes.swatch, classes[item.swatch]].join(' ')} />
      {item.label}
    </span>
  );
}

function Badge({ item }: { item: BadgeItem }) {
  return (
    <span className={classes.legendItem}>
      <span className={[classes.badge, classes[item.badge]].join(' ')}>{item.letter}</span>
      {item.label}
    </span>
  );
}

function legendSwatches(lens: BoardLens): SwatchItem[] {
  if (lens === 'now') return NOW_SWATCHES;
  if (lens === 'journey') return [...STATUS_SWATCHES, ...MOVE_SWATCHES];
  return STATUS_SWATCHES;
}

export function BoardLegend({ lens }: { lens: BoardLens }) {
  return (
    <div className={classes.legend} data-testid="board-legend">
      {legendSwatches(lens).map((item) => (
        <Swatch key={item.label} item={item} />
      ))}
      {lens === 'now' && (
        <>
          {TIME_BADGES.map((item) => (
            <Badge key={item.letter} item={item} />
          ))}
          <span className={classes.disclaimer}>TIME = per-capability assessment, not commitment</span>
        </>
      )}
    </div>
  );
}
