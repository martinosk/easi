import { Text } from '@mantine/core';
import classes from './CapabilityDrawer.module.css';

export function DrawerSectionHeader({ children }: { children: React.ReactNode }) {
  return <Text className={classes.sectionHeader}>{children}</Text>;
}
