import { Box, Menu, Portal, Stack, Text } from '@mantine/core';
import classes from './LinearContextMenu.module.css';
import type { ContextMenuItem } from './types';

interface LinearContextMenuProps {
  x: number;
  y: number;
  items: ContextMenuItem[];
  onClose: () => void;
}

const LinearItem = ({ item }: { item: ContextMenuItem }) => (
  <Menu.Item
    className={classes.item}
    leftSection={item.icon}
    color={item.isDanger ? 'red' : undefined}
    disabled={item.disabled}
    aria-label={item.ariaLabel ?? item.label}
    onClick={item.onClick}
  >
    <Stack gap={0}>
      <Text size="sm" fw={500}>
        {item.label}
      </Text>
      {item.description && (
        <Text size="xs" c="dimmed">
          {item.description}
        </Text>
      )}
    </Stack>
  </Menu.Item>
);

export const LinearContextMenu = ({ x, y, items, onClose }: LinearContextMenuProps) => (
  <Menu
    opened
    onClose={onClose}
    position="bottom-start"
    offset={0}
    shadow="lg"
    withinPortal
    closeOnClickOutside
    closeOnEscape
  >
    <Portal>
      <Menu.Target>
        <Box className={classes.anchor} left={x} top={y} />
      </Menu.Target>
    </Portal>
    <Menu.Dropdown
      className={classes.dropdown}
      data-testid="context-menu"
      data-variant="linear"
      aria-label="Context menu"
    >
      {items.map((item) => (
        <LinearItem key={item.label} item={item} />
      ))}
    </Menu.Dropdown>
  </Menu>
);
