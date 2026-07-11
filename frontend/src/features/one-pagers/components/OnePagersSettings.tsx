import { Stack, Tabs, Text, Title } from '@mantine/core';
import { useState } from 'react';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import { ONE_PAGER_SUBJECT_TYPE_TABS } from '../subjectTypes';
import type { OnePagerSubjectType } from '../types';
import { OnePagerConfigurationPanel } from './OnePagerConfigurationPanel';

export function OnePagersSettings() {
  const [subjectType, setSubjectType] = useState<OnePagerSubjectType>(ONE_PAGER_SUBJECT_TYPE_TABS[0].value);

  return (
    <Stack gap="md" data-testid="one-pagers-settings">
      <div>
        <Title order={2}>
          One-Pagers
          <HelpTooltip
            content="Configure which fields appear on the one-pager fact sheet for each subject type, and add custom fields to capture facts specific to your organization."
            iconOnly
          />
        </Title>
        <Text c="dimmed">Choose which built-in and custom fields appear on each subject type's one-pager.</Text>
      </div>

      <Tabs value={subjectType} onChange={(value) => setSubjectType(value as OnePagerSubjectType)}>
        <Tabs.List>
          {ONE_PAGER_SUBJECT_TYPE_TABS.map((tab) => (
            <Tabs.Tab key={tab.value} value={tab.value} data-testid={`one-pager-tab-${tab.value}`}>
              {tab.label}
            </Tabs.Tab>
          ))}
        </Tabs.List>

        {ONE_PAGER_SUBJECT_TYPE_TABS.map((tab) => (
          <Tabs.Panel key={tab.value} value={tab.value} pt="md">
            {subjectType === tab.value && <OnePagerConfigurationPanel subjectType={tab.value} />}
          </Tabs.Panel>
        ))}
      </Tabs>
    </Stack>
  );
}
