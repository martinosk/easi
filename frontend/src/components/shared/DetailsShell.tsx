import { Stack } from '@mantine/core';
import type React from 'react';
import { type DetailGroup, DetailGroups } from './DetailGroups';
import { useDetailGroupLayout } from './useDetailGroupLayout';

export const VIEW_GROUP_ID = 'view';

export interface DetailsShellProps {
  layoutKey: string;
  heading: React.ReactNode;
  groups: DetailGroup[];
  viewMembership?: React.ReactNode;
  footer?: React.ReactNode;
}

function hasContent(group: DetailGroup): boolean {
  return group.content !== null && group.content !== undefined && group.content !== false;
}

export const DetailsShell: React.FC<DetailsShellProps> = ({ layoutKey, heading, groups, viewMembership, footer }) => {
  const allGroups = [...groups, { id: VIEW_GROUP_ID, title: 'In this view', content: viewMembership }];
  const layout = useDetailGroupLayout(
    layoutKey,
    allGroups.map((group) => group.id),
  );

  return (
    <Stack gap="sm">
      {heading}
      <DetailGroups groups={allGroups.filter(hasContent)} layout={layout} />
      {footer}
    </Stack>
  );
};
