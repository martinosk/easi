import { UnstyledButton } from '@mantine/core';
import React from 'react';
import { OnePagerIncompleteIndicator } from '../../../one-pagers/components/OnePagerIncompleteIndicator';
import classes from './TreeItem.module.css';

interface TreeItemProps<T> {
  item: T;
  isSelected: boolean;
  isInView: boolean;
  icon: React.ReactNode;
  label: React.ReactNode;
  title: string;
  dragDataKey: string;
  onSelect: (e: React.MouseEvent) => void;
  onContextMenu: (e: React.MouseEvent) => void;
  onDragStart?: (e: React.DragEvent) => void;
}

function TreeItem<T extends { onePagerComplete?: boolean }>({
  isSelected,
  isInView,
  icon,
  label,
  title,
  dragDataKey,
  item,
  onSelect,
  onContextMenu,
  onDragStart,
}: TreeItemProps<T> & { item: T & { id: string } }): React.ReactElement {
  return (
    <UnstyledButton
      component="button"
      type="button"
      className={classes.item}
      data-testid="tree-item"
      data-selected={isSelected || undefined}
      data-in-view={isInView}
      onClick={onSelect}
      onContextMenu={onContextMenu}
      title={title}
      draggable
      onDragStart={(e) => {
        if (onDragStart) {
          onDragStart(e);
        } else {
          e.dataTransfer.setData(dragDataKey, item.id);
          e.dataTransfer.effectAllowed = 'copy';
        }
      }}
    >
      <span className={classes.icon}>{icon}</span>
      <span className={classes.label}>{label}</span>
      <OnePagerIncompleteIndicator id={item.id} onePagerComplete={item.onePagerComplete} />
    </UnstyledButton>
  );
}

interface TreeItemListProps<T extends { id: string; name: string; onePagerComplete?: boolean }> {
  items: T[];
  emptyMessage: string;
  icon: React.ReactNode;
  dragDataKey: string;
  isSelected: (item: T) => boolean;
  isInView: (item: T) => boolean;
  getTitle: (item: T, isInView: boolean) => string;
  renderLabel: (item: T) => React.ReactNode;
  onSelect: (item: T, e: React.MouseEvent) => void;
  onContextMenu: (e: React.MouseEvent, item: T) => void;
  onDragStart?: (e: React.DragEvent, item: T) => void;
}

export function TreeItemList<T extends { id: string; name: string; onePagerComplete?: boolean }>({
  items,
  emptyMessage,
  icon,
  dragDataKey,
  isSelected,
  isInView,
  getTitle,
  renderLabel,
  onSelect,
  onContextMenu,
  onDragStart,
}: TreeItemListProps<T>): React.ReactElement {
  return (
    <div className={classes.list}>
      {items.length === 0 ? (
        <div className={classes.empty}>{emptyMessage}</div>
      ) : (
        items.map((item) => {
          const itemIsInView = isInView(item);
          return (
            <TreeItem
              key={item.id}
              item={item}
              isSelected={isSelected(item)}
              isInView={itemIsInView}
              icon={icon}
              label={renderLabel(item)}
              title={getTitle(item, itemIsInView)}
              dragDataKey={dragDataKey}
              onSelect={(e) => onSelect(item, e)}
              onContextMenu={(e) => onContextMenu(e, item)}
              onDragStart={onDragStart ? (e) => onDragStart(e, item) : undefined}
            />
          );
        })
      )}
    </div>
  );
}

export { TreeItemList as default };
