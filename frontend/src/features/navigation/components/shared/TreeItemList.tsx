import React from 'react';

interface TreeItemProps<T> {
  item: T;
  isSelected: boolean;
  isInView: boolean;
  icon: string;
  label: React.ReactNode;
  title: string;
  dragDataKey: string;
  onSelect: () => void;
  onContextMenu: (e: React.MouseEvent) => void;
}

function TreeItem<T>({
  isSelected,
  isInView,
  icon,
  label,
  title,
  dragDataKey,
  item,
  onSelect,
  onContextMenu,
}: TreeItemProps<T> & { item: T & { id: string } }): React.ReactElement {
  const className = `tree-item${isSelected ? ' selected' : ''}${!isInView ? ' not-in-view' : ''}`;

  return (
    <button
      className={className}
      onClick={onSelect}
      onContextMenu={onContextMenu}
      title={title}
      draggable
      onDragStart={(e) => {
        e.dataTransfer.setData(dragDataKey, item.id);
        e.dataTransfer.effectAllowed = 'copy';
      }}
    >
      <span className="tree-item-icon">{icon}</span>
      <span className="tree-item-label">{label}</span>
    </button>
  );
}

interface TreeItemListProps<T extends { id: string; name: string }> {
  items: T[];
  emptyMessage: string;
  icon: string;
  dragDataKey: string;
  isSelected: (item: T) => boolean;
  isInView: (item: T) => boolean;
  getTitle: (item: T, isInView: boolean) => string;
  renderLabel: (item: T) => React.ReactNode;
  onSelect: (item: T) => void;
  onContextMenu: (e: React.MouseEvent, item: T) => void;
}

export function TreeItemList<T extends { id: string; name: string }>({
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
}: TreeItemListProps<T>): React.ReactElement {
  if (items.length === 0) {
    return <div className="tree-item-empty">{emptyMessage}</div>;
  }

  return (
    <>
      {items.map((item) => {
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
            onSelect={() => onSelect(item)}
            onContextMenu={(e) => onContextMenu(e, item)}
          />
        );
      })}
    </>
  );
}

export { TreeItemList as default };
