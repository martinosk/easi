import React, { useState, useMemo } from 'react';
import type { InternalTeam, View } from '../../../../api/types';
import { TreeSection } from '../TreeSection';
import { TreeSearchInput } from '../shared/TreeSearchInput';
import { TreeItemList } from '../shared/TreeItemList';
import { ORIGIN_ENTITY_PREFIXES } from '../../../canvas/utils/nodeFactory';
import type { TreeMultiSelectProps } from '../../types';

interface InternalTeamsSectionProps {
  internalTeams: InternalTeam[];
  currentView: View | null;
  selectedTeamId: string | null;
  isExpanded: boolean;
  onToggle: () => void;
  onAddTeam?: () => void;
  onTeamSelect?: (teamId: string) => void;
  onTeamContextMenu: (e: React.MouseEvent, team: InternalTeam) => void;
  multiSelect: TreeMultiSelectProps;
}

function filterTeams(teams: InternalTeam[], search: string): InternalTeam[] {
  if (!search.trim()) return teams;
  const searchLower = search.toLowerCase();
  return teams.filter(
    (t) =>
      t.name.toLowerCase().includes(searchLower) ||
      (t.department && t.department.toLowerCase().includes(searchLower)) ||
      (t.contactPerson && t.contactPerson.toLowerCase().includes(searchLower)) ||
      (t.notes && t.notes.toLowerCase().includes(searchLower))
  );
}

function buildTeamIdsOnCanvas(teams: InternalTeam[], currentView: View | null): Set<string> {
  const viewOriginEntityIds = new Set(
    (currentView?.originEntities ?? []).map((oe) => oe.originEntityId)
  );
  const onCanvas = new Set<string>();
  for (const team of teams) {
    if (viewOriginEntityIds.has(`${ORIGIN_ENTITY_PREFIXES.team}${team.id}`)) {
      onCanvas.add(team.id);
    }
  }
  return onCanvas;
}

export const InternalTeamsSection: React.FC<InternalTeamsSectionProps> = ({
  internalTeams,
  currentView,
  selectedTeamId,
  isExpanded,
  onToggle,
  onAddTeam,
  onTeamSelect,
  onTeamContextMenu,
  multiSelect,
}) => {
  const [search, setSearch] = useState('');

  const teamIdsOnCanvas = useMemo(
    () => buildTeamIdsOnCanvas(internalTeams, currentView),
    [internalTeams, currentView]
  );

  const filteredTeams = useMemo(
    () => filterTeams(internalTeams, search),
    [internalTeams, search]
  );

  const visibleItems = useMemo(
    () => filteredTeams.map((t) => ({
      id: t.id, name: t.name, type: 'team' as const, links: t._links,
    })),
    [filteredTeams]
  );

  const hasNoTeams = internalTeams.length === 0;
  const emptyMessage = hasNoTeams ? 'No internal teams' : 'No matches';

  const handleSelect = (team: InternalTeam, event: React.MouseEvent) => {
    const result = multiSelect.handleItemClick(
      { id: team.id, name: team.name, type: 'team', links: team._links },
      'teams',
      visibleItems,
      event
    );
    if (result === 'single') {
      onTeamSelect?.(team.id);
    }
  };

  const handleContextMenu = (e: React.MouseEvent, team: InternalTeam) => {
    const handled = multiSelect.handleContextMenu(e, team.id, multiSelect.selectedItems);
    if (!handled) {
      onTeamContextMenu(e, team);
    }
  };

  const handleDragStart = (e: React.DragEvent, team: InternalTeam) => {
    const handled = multiSelect.handleDragStart(e, team.id);
    if (!handled) {
      e.dataTransfer.setData('internalTeamId', team.id);
      e.dataTransfer.effectAllowed = 'copy';
    }
  };

  return (
    <TreeSection
      label="Internal Teams"
      count={internalTeams.length}
      isExpanded={isExpanded}
      onToggle={onToggle}
      onAdd={onAddTeam}
      addTitle="Create new internal team"
      addTestId="create-internal-team-button"
    >
      <TreeSearchInput
        value={search}
        onChange={setSearch}
        placeholder="Search internal teams..."
      />
      <div className="tree-items">
        <TreeItemList
          items={filteredTeams}
          emptyMessage={emptyMessage}
          icon="👥"
          dragDataKey="internalTeamId"
          isSelected={(team) => selectedTeamId === team.id || multiSelect.isMultiSelected(team.id)}
          isInView={(team) => !currentView || teamIdsOnCanvas.has(team.id)}
          getTitle={(team, isInView) =>
            isInView ? team.name : `${team.name} (not on canvas)`
          }
          renderLabel={(team) => team.name}
          onSelect={handleSelect}
          onContextMenu={handleContextMenu}
          onDragStart={handleDragStart}
        />
      </div>
    </TreeSection>
  );
};
