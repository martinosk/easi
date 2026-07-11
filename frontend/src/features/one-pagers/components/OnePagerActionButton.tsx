import { Button } from '@mantine/core';
import { generatePath, useNavigate } from 'react-router-dom';
import { FileTextIcon } from '../../../components/shared/ContextMenu/icons';
import { ROUTES } from '../../../routes/routePaths';
import { hasLink, type ResourceWithLinks } from '../../../utils/hateoas';
import type { OnePagerSubjectType } from '../types';

interface OnePagerActionButtonProps {
  subject: ResourceWithLinks;
  subjectType: OnePagerSubjectType;
  subjectId: string;
}

export function OnePagerActionButton({ subject, subjectType, subjectId }: OnePagerActionButtonProps) {
  const navigate = useNavigate();

  if (!hasLink(subject, 'x-one-pager')) return null;

  return (
    <Button
      variant="light"
      size="xs"
      leftSection={<FileTextIcon />}
      onClick={() => navigate(generatePath(ROUTES.ONE_PAGER_DETAIL, { subjectType, subjectId }))}
    >
      One-Pager
    </Button>
  );
}
