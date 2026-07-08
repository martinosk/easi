import { Box, Text } from '@mantine/core';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { TreeSearchInput } from '../../../../components/shared';
import type { CapabilityTreeNode } from '../../hooks/useCapabilityTree';
import classes from './CapabilityTree.module.css';
import type { CapabilityTreeRowOptions } from './CapabilityTreeRow';
import { CapabilityTreeRow } from './CapabilityTreeRow';
import { collectExpandableIds, filterTreeByName } from './treeFiltering';

type IsExpandedFn = (node: CapabilityTreeNode) => boolean;

export interface CapabilityTreeProps {
  tree: CapabilityTreeNode[];
  isLoading?: boolean;
  searchPlaceholder?: string;
  emptyText?: string;
  noMatchText?: string;
  getRowProps?: (node: CapabilityTreeNode) => CapabilityTreeRowOptions;
  renderRight?: (node: CapabilityTreeNode) => React.ReactNode;
  expandedIds?: Set<string>;
  onToggleExpanded?: (id: string) => void;
  onVisibleNodesChange?: (nodes: CapabilityTreeNode[]) => void;
  testId?: string;
  searchTestId?: string;
  className?: string;
}

interface ExpansionArgs {
  searchExpandedIds: Set<string> | null;
  expandedIds?: Set<string>;
  onToggleExpanded?: (id: string) => void;
}

function useTreeExpansion({ searchExpandedIds, expandedIds, onToggleExpanded }: ExpansionArgs) {
  const [toggledIds, setToggledIds] = useState<Set<string>>(new Set());

  const isExpanded = useCallback<IsExpandedFn>(
    (node) => {
      if (searchExpandedIds) return searchExpandedIds.has(node.capability.id);
      if (expandedIds) return expandedIds.has(node.capability.id);
      return toggledIds.has(node.capability.id);
    },
    [searchExpandedIds, expandedIds, toggledIds],
  );

  const toggle = useCallback(
    (id: string) => {
      if (onToggleExpanded) {
        onToggleExpanded(id);
        return;
      }
      setToggledIds((prev) => {
        const next = new Set(prev);
        if (next.has(id)) {
          next.delete(id);
        } else {
          next.add(id);
        }
        return next;
      });
    },
    [onToggleExpanded],
  );

  return { isExpanded, toggle };
}

function flattenVisibleNodes(nodes: CapabilityTreeNode[], isExpanded: IsExpandedFn): CapabilityTreeNode[] {
  return nodes.flatMap((node) => {
    const rest = node.children.length > 0 && isExpanded(node) ? flattenVisibleNodes(node.children, isExpanded) : [];
    return [node, ...rest];
  });
}

function useVisibleNodesReport(
  filteredTree: CapabilityTreeNode[],
  isExpanded: IsExpandedFn,
  onVisibleNodesChange?: (nodes: CapabilityTreeNode[]) => void,
) {
  const visibleNodes = useMemo(() => flattenVisibleNodes(filteredTree, isExpanded), [filteredTree, isExpanded]);

  useEffect(() => {
    onVisibleNodesChange?.(visibleNodes);
  }, [visibleNodes, onVisibleNodesChange]);
}

interface TreeLevelProps {
  nodes: CapabilityTreeNode[];
  searchQuery: string;
  isExpanded: IsExpandedFn;
  onToggle: (id: string) => void;
  getRowProps?: (node: CapabilityTreeNode) => CapabilityTreeRowOptions;
  renderRight?: (node: CapabilityTreeNode) => React.ReactNode;
}

const TreeLevel: React.FC<TreeLevelProps> = (props) => {
  const { nodes, searchQuery, isExpanded, onToggle, getRowProps, renderRight } = props;

  return (
    <>
      {nodes.map((node) => {
        const expanded = isExpanded(node);
        return (
          <div key={node.capability.id}>
            <CapabilityTreeRow
              node={node}
              isExpanded={expanded}
              onToggle={() => onToggle(node.capability.id)}
              searchQuery={searchQuery}
              options={getRowProps?.(node) ?? {}}
              rightContent={renderRight?.(node)}
            />
            {expanded && node.children.length > 0 && (
              <Box role="group" className={classes.children}>
                <TreeLevel {...props} nodes={node.children} />
              </Box>
            )}
          </div>
        );
      })}
    </>
  );
};

export const CapabilityTree: React.FC<CapabilityTreeProps> = ({
  tree,
  isLoading = false,
  searchPlaceholder = 'Search capabilities...',
  emptyText = 'No capabilities',
  noMatchText = 'No matches',
  getRowProps,
  renderRight,
  expandedIds,
  onToggleExpanded,
  onVisibleNodesChange,
  testId,
  searchTestId,
  className,
}) => {
  const [search, setSearch] = useState('');
  const isSearching = search.trim().length > 0;

  const filteredTree = useMemo(() => filterTreeByName(tree, search), [tree, search]);
  const searchExpandedIds = useMemo(
    () => (isSearching ? collectExpandableIds(filteredTree) : null),
    [isSearching, filteredTree],
  );
  const { isExpanded, toggle } = useTreeExpansion({ searchExpandedIds, expandedIds, onToggleExpanded });
  useVisibleNodesReport(filteredTree, isExpanded, onVisibleNodesChange);

  if (isLoading) {
    return (
      <Box p="md">
        <Text c="dimmed" size="sm">
          Loading capabilities...
        </Text>
      </Box>
    );
  }

  return (
    <Box className={`${classes.root} ${className ?? ''}`} data-testid={testId}>
      <TreeSearchInput value={search} onChange={setSearch} placeholder={searchPlaceholder} testId={searchTestId} />
      <Box role="tree" className={classes.tree}>
        {filteredTree.length === 0 ? (
          <Text c="dimmed" size="sm" p="sm">
            {isSearching ? noMatchText : emptyText}
          </Text>
        ) : (
          <TreeLevel
            nodes={filteredTree}
            searchQuery={search}
            isExpanded={isExpanded}
            onToggle={toggle}
            getRowProps={getRowProps}
            renderRight={renderRight}
          />
        )}
      </Box>
    </Box>
  );
};
