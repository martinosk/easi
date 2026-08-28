import { describe, expect, it } from 'vitest';
import { buildJourneyMilestone } from '../../../test/helpers/entityBuilders';
import { milestoneScheduleConflicts } from './milestoneSchedule';

const dated = (id: string, year: number, quarter: number) =>
  buildJourneyMilestone({ id, targetPeriod: { year, quarter } });
const undated = (id: string) => buildJourneyMilestone({ id, targetPeriod: null });

describe('milestoneScheduleConflicts', () => {
  it('marks a milestone placed below a later-targeted one, naming that period (rule 1)', () => {
    const conflicts = milestoneScheduleConflicts([dated('a', 2027, 1), dated('b', 2026, 4)]);

    expect([...conflicts.entries()]).toEqual([['b', { year: 2027, quarter: 1 }]]);
  });

  it('compares against the latest period above, not the immediate neighbour (decision 3)', () => {
    const conflicts = milestoneScheduleConflicts([dated('a', 2027, 1), dated('b', 2026, 2), dated('c', 2026, 4)]);

    expect(conflicts.get('b')).toEqual({ year: 2027, quarter: 1 });
    expect(conflicts.get('c')).toEqual({ year: 2027, quarter: 1 });
  });

  it('marks nothing for a chronological list', () => {
    expect(milestoneScheduleConflicts([dated('a', 2025, 4), dated('b', 2026, 1), dated('c', 2026, 1)]).size).toBe(0);
  });

  it('ignores undated milestones on both sides (rule 1)', () => {
    const conflicts = milestoneScheduleConflicts([
      undated('x'),
      dated('a', 2027, 1),
      undated('y'),
      dated('b', 2027, 2),
    ]);

    expect(conflicts.size).toBe(0);
  });

  it('marks nothing for empty and single-milestone lists', () => {
    expect(milestoneScheduleConflicts([]).size).toBe(0);
    expect(milestoneScheduleConflicts([dated('a', 2026, 1)]).size).toBe(0);
  });
});
