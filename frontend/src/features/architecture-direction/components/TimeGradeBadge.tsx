import type { TimeGrade } from '../types';
import classes from './TimeGradeBadge.module.css';

const GRADE_CLASS: Record<TimeGrade, string> = {
  Invest: classes.invest,
  Tolerate: classes.tolerate,
  Migrate: classes.migrate,
  Eliminate: classes.eliminate,
};

const GRADE_LETTER: Record<TimeGrade, string> = {
  Invest: 'I',
  Tolerate: 'T',
  Migrate: 'M',
  Eliminate: 'E',
};

interface TimeGradeBadgeProps {
  grade: TimeGrade;
  title?: string;
  labelled?: boolean;
  testId?: string;
}

export function TimeGradeBadge({ grade, title, labelled = false, testId }: TimeGradeBadgeProps) {
  const className = `${classes.badge} ${GRADE_CLASS[grade]}`;
  if (labelled) {
    return (
      <span className={className} title={title} role="img" aria-label={grade} data-testid={testId}>
        {GRADE_LETTER[grade]}
      </span>
    );
  }
  return (
    <span className={className} title={title} data-testid={testId}>
      {GRADE_LETTER[grade]}
    </span>
  );
}
