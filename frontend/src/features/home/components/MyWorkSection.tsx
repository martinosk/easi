import { Button, Group, SimpleGrid, Stack, Text, Title } from '@mantine/core';
import { useState } from 'react';
import type { MyWork } from '../types';
import { MyWorkCard } from './MyWorkCard';

const INITIAL_CARDS = 12;

function MyWorkGrid({ myWork }: { myWork: MyWork }) {
  const [expanded, setExpanded] = useState(false);
  const hasMore = myWork.items.length > INITIAL_CARDS;
  const visible = expanded || !hasMore ? myWork.items : myWork.items.slice(0, INITIAL_CARDS);
  return (
    <Stack gap="md">
      <SimpleGrid cols={{ base: 1, sm: 2, lg: 3, xl: 4 }} spacing="md">
        {visible.map((item) => (
          <MyWorkCard key={`${item.subjectType}-${item.id}`} item={item} />
        ))}
      </SimpleGrid>
      {hasMore && !expanded && (
        <Group justify="flex-start">
          <Button variant="default" onClick={() => setExpanded(true)}>
            {`Show all ${myWork.total}`}
          </Button>
        </Group>
      )}
    </Stack>
  );
}

export function MyWorkSection({ myWork }: { myWork: MyWork }) {
  return (
    <Stack gap="sm">
      <Title order={3}>My Work</Title>
      {myWork.items.length === 0 ? (
        <Text c="dimmed">Nothing is assigned to you personally.</Text>
      ) : (
        <MyWorkGrid myWork={myWork} />
      )}
    </Stack>
  );
}
