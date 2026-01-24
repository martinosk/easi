import React, { useState, useMemo } from 'react';
import type { InternalTeam, View } from '../../../../api/types';
import { TreeSection } from '../TreeSection';
import { TreeSearchInput } from '../shared/TreeSearchInput';
import { TreeItemList } from '../shared/TreeItemList';
import { useCanvasLayoutContext } from '../../../canvas/context/CanvasLayoutContext';
import { ORIGIN_ENTITY_PREFIXES } from '../../../canvas/utils/nodeFactory';

interface InternalTeamsSectionProps {
  internalTeams: InternalTeam[];
  currentView: View | null;
  selectedTeamId: string | null;
  isExpanded: boolean;
  onToggle: () => void;
  onAddTeam?: () => void;
  onTeamSelect?: (teamId: string) => void;
  onTeamContextMenu: (e: React.MouseEvent, team: InternalTeam) => void;
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

export const InternalTeamsSection: React.FC<InternalTeamsSectionProps> = ({
  internalTeams,
  currentView,
  selectedTeamId,
  isExpanded,
  onToggle,
  onAddTeam,
  onTeamSelect,
  onTeamContextMenu,
}) => {
  const [search, setSearch] = useState('');
  const { positions: layoutPositions } = useCanvasLayoutContext();

  const teamIdsOnCanvas = useMemo(() => {
    const onCanvas = new Set<string>();
    for (const team of internalTeams) {
      const nodeId = `${ORIGIN_ENTITY_PREFIXES.team}${team.id}`;
      if (layoutPositions[nodeId] !== undefined) {
        onCanvas.add(team.id);
      }
    }
    return onCanvas;
  }, [internalTeams, layoutPositions]);

  const filteredTeams = useMemo(
    () => filterTeams(internalTeams, search),
    [internalTeams, search]
  );

  const hasNoTeams = internalTeams.length === 0;
  const emptyMessage = hasNoTeams ? 'No internal teams' : 'No matches';

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
          isSelected={(team) => selectedTeamId === team.id}
          isInView={(team) => !currentView || teamIdsOnCanvas.has(team.id)}
          getTitle={(team, isInView) =>
            isInView ? team.name : `${team.name} (not on canvas)`
          }
          renderLabel={(team) => team.name}
          onSelect={(team) => onTeamSelect?.(team.id)}
          onContextMenu={onTeamContextMenu}
        />
      </div>
    </TreeSection>
  );
};
