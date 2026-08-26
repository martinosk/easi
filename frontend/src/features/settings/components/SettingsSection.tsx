import { Group, Loader, Paper, Stack, Text, type TextProps, Title } from '@mantine/core';
import type { ReactNode } from 'react';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import classes from './SettingsSection.module.css';

export function SettingsSection({ children }: { children: ReactNode }) {
  return (
    <Paper radius="md" p="xl" shadow="sm">
      <Stack gap="lg">{children}</Stack>
    </Paper>
  );
}

interface SettingsSectionHeaderProps {
  title: string;
  description: string;
  help?: string;
  actions?: ReactNode;
}

export function SettingsSectionHeader({ title, description, help, actions }: SettingsSectionHeaderProps) {
  return (
    <Group justify="space-between" align="flex-start" gap="md">
      <Stack gap={0}>
        <Title order={2}>
          {title}
          {help && <HelpTooltip content={help} iconOnly />}
        </Title>
        <Text c="dimmed" size="sm">
          {description}
        </Text>
      </Stack>
      {actions}
    </Group>
  );
}

export function SettingsSectionFooter({ children }: { children: ReactNode }) {
  return (
    <Group justify="flex-end" gap="sm" pt="lg" className={classes.footer}>
      {children}
    </Group>
  );
}

export function SettingsSectionLoading({ message }: { message: string }) {
  return (
    <SettingsSection>
      <Stack align="center" justify="center" gap="md" p="xxl">
        <Loader />
        <Text c="dimmed">{message}</Text>
      </Stack>
    </SettingsSection>
  );
}

export function SettingsErrorNotice({ children, ...textProps }: TextProps & { children: ReactNode }) {
  return (
    <Text className={classes.errorNotice} {...textProps}>
      {children}
    </Text>
  );
}

export function SettingsSectionError({ error, fallback }: { error: unknown; fallback: string }) {
  return (
    <SettingsSection>
      <SettingsErrorNotice>{error instanceof Error ? error.message : fallback}</SettingsErrorNotice>
    </SettingsSection>
  );
}

export function SettingsConflictNotice() {
  return (
    <Text className={classes.conflictNotice}>
      Configuration was modified by another user. Please refresh and try again.
    </Text>
  );
}
