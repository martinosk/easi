import { Box, Text } from '@mantine/core';
import type { CSSProperties } from 'react';
import { useMemo } from 'react';
import type { Capability, CapabilityId, ComponentId } from '../../../api/types';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import { nodeMatchesSearch } from '../hooks/boardSearch';
import type { AssessedRealization, DomainBoardViewModel } from '../hooks/domainBoardViewModel';
import { buildMapTree } from '../hooks/mapLayout';
import type { MapDepth } from '../hooks/useMapViewState';
import { AppChip } from './AppChip';
import { activationKeyHandler } from './boardCardKeyboard';
import classes from './CapabilityMap.module.css';

export interface CapabilityMapProps {
  viewModel: DomainBoardViewModel;
  depth: MapDepth;
  searchQuery: string;
  showApps: boolean;
  getColorForValue: (maturityValue: number) => string;
  onCapabilityClick: (capability: Capability, event: React.MouseEvent) => void;
  onChipClick: (componentId: ComponentId) => void;
}

interface CellChrome {
  className: string;
  style?: CSSProperties;
}

interface CellChromeParams {
  capability: Capability;
  dimmed: boolean;
  getColorForValue: (maturityValue: number) => string;
}

function cellChrome({ capability, dimmed, getColorForValue }: CellChromeParams): CellChrome {
  const maturityBorder =
    capability.maturityValue !== undefined ? getColorForValue(capability.maturityValue) : undefined;
  const className = [classes.cell, dimmed ? classes.dimmed : ''].filter(Boolean).join(' ');
  return { className, style: maturityBorder ? { borderLeftColor: maturityBorder } : undefined };
}

interface MapBoxProps {
  node: CapabilityTreeNode;
  searchQuery: string;
  showApps: boolean;
  getColorForValue: (maturityValue: number) => string;
  getRealizationsForCapability: (capabilityId: CapabilityId) => AssessedRealization[];
  onCapabilityClick: (capability: Capability, event: React.MouseEvent) => void;
  onChipClick: (componentId: ComponentId) => void;
}

function MapBoxHeader({ capability }: { capability: Capability }) {
  return (
    <div className={classes.cellHeader}>
      <span className={classes.name}>{capability.name}</span>
      <span className={classes.levelTag}>{capability.level}</span>
    </div>
  );
}

interface MapBoxAppsProps {
  realizations: AssessedRealization[];
  onChipClick: (componentId: ComponentId) => void;
}

function MapBoxApps({ realizations, onChipClick }: MapBoxAppsProps) {
  if (realizations.length === 0) return null;
  return (
    <div className={classes.apps}>
      {realizations.map((realization) => (
        <AppChip key={realization.id} realization={realization} onClick={onChipClick} />
      ))}
    </div>
  );
}

function MapBox({ node, ...boxProps }: MapBoxProps) {
  const { searchQuery, showApps, getColorForValue, getRealizationsForCapability, onCapabilityClick, onChipClick } = boxProps;
  const { capability } = node;
  const dimmed = Boolean(searchQuery) && !nodeMatchesSearch(node, searchQuery, getRealizationsForCapability);
  const { className, style } = cellChrome({ capability, dimmed, getColorForValue });
  const realizations = showApps ? getRealizationsForCapability(capability.id) : [];

  return (
    <Box
      className={className}
      style={style}
      role="button"
      tabIndex={0}
      onClick={(event) => {
        event.stopPropagation();
        onCapabilityClick(capability, event);
      }}
      onKeyDown={(event) => {
        event.stopPropagation();
        activationKeyHandler(capability, onCapabilityClick)(event);
      }}
      data-testid={`map-cell-${capability.id}`}
      data-dimmed={dimmed || undefined}
    >
      <MapBoxHeader capability={capability} />
      <MapBoxApps realizations={realizations} onChipClick={onChipClick} />
      {node.children.length > 0 && (
        <div className={classes.children}>
          {node.children.map((child) => (
            <MapBox key={child.capability.id} {...boxProps} node={child} />
          ))}
        </div>
      )}
    </Box>
  );
}

export function CapabilityMap({
  viewModel,
  depth,
  searchQuery,
  showApps,
  getColorForValue,
  onCapabilityClick,
  onChipClick,
}: CapabilityMapProps) {
  const mapTree = useMemo(
    () =>
      buildMapTree(
        viewModel.l1Groups.map((group) => group.node),
        depth,
      ),
    [viewModel, depth],
  );

  if (viewModel.l1Groups.length === 0) {
    return (
      <Text c="dimmed" data-testid="capability-map-empty">
        This domain has no capabilities yet. Switch to the Board view to assign capabilities to it.
      </Text>
    );
  }

  return (
    <div className={classes.map} data-testid="capability-map">
      {mapTree.map((node) => (
        <MapBox
          key={node.capability.id}
          node={node}
          searchQuery={searchQuery}
          showApps={showApps}
          getColorForValue={getColorForValue}
          getRealizationsForCapability={viewModel.getRealizationsForCapability}
          onCapabilityClick={onCapabilityClick}
          onChipClick={onChipClick}
        />
      ))}
    </div>
  );
}
