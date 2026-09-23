import { Accordion, ActionIcon, Group } from '@mantine/core';
import { IconChevronDown, IconChevronUp } from '@tabler/icons-react';
import type React from 'react';
import classes from './DetailGroups.module.css';
import type { DetailGroupLayout } from './useDetailGroupLayout';

export interface DetailGroup {
  id: string;
  title: string;
  content: React.ReactNode;
}

interface MoveControlProps {
  title: string;
  direction: 'up' | 'down';
  neighbour: DetailGroup | undefined;
  onMove: (neighbourId: string) => void;
}

const MoveControl: React.FC<MoveControlProps> = ({ title, direction, neighbour, onMove }) => (
  <ActionIcon
    variant="subtle"
    color="gray"
    size="sm"
    aria-label={`Move ${title} ${direction}`}
    disabled={neighbour === undefined}
    onClick={() => neighbour && onMove(neighbour.id)}
  >
    {direction === 'up' ? <IconChevronUp size={16} /> : <IconChevronDown size={16} />}
  </ActionIcon>
);

interface GroupItemProps {
  group: DetailGroup;
  previous: DetailGroup | undefined;
  next: DetailGroup | undefined;
  onSwap: (neighbourId: string) => void;
}

const GroupItem: React.FC<GroupItemProps> = ({ group, previous, next, onSwap }) => (
  <Accordion.Item value={group.id} data-testid={`detail-group-${group.id}`}>
    <Group gap={0} wrap="nowrap" className={classes.header}>
      <Accordion.Control>{group.title}</Accordion.Control>
      <MoveControl title={group.title} direction="up" neighbour={previous} onMove={onSwap} />
      <MoveControl title={group.title} direction="down" neighbour={next} onMove={onSwap} />
    </Group>
    <Accordion.Panel>{group.content}</Accordion.Panel>
  </Accordion.Item>
);

export interface DetailGroupsProps {
  groups: DetailGroup[];
  layout: DetailGroupLayout;
}

function orderedGroups(groups: DetailGroup[], order: readonly string[]): DetailGroup[] {
  return order.map((id) => groups.find((group) => group.id === id)).filter((group) => group !== undefined);
}

export const DetailGroups: React.FC<DetailGroupsProps> = ({ groups, layout }) => {
  const rendered = orderedGroups(groups, layout.order);
  const expanded = rendered.map((group) => group.id).filter((id) => !layout.collapsed.includes(id));

  const handleChange = (open: string[]) => {
    const changed = rendered.find((group) => layout.collapsed.includes(group.id) === open.includes(group.id));
    if (changed) layout.toggle(changed.id);
  };

  return (
    <Accordion multiple value={expanded} onChange={handleChange} chevronPosition="left" classNames={classes}>
      {rendered.map((group, index) => (
        <GroupItem
          key={group.id}
          group={group}
          previous={rendered[index - 1]}
          next={rendered[index + 1]}
          onSwap={(neighbourId) => layout.swap(group.id, neighbourId)}
        />
      ))}
    </Accordion>
  );
};
