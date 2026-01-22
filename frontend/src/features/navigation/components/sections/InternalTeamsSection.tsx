import React, { useState, useMemo } from 'react';
import type { InternalTeam, View, OriginRelationship } from '../../../../api/types';
import { TreeSection } from '../TreeSection';
import { TreeSearchInput } from '../shared/TreeSearchInput';
import { TreeItemList } from '../shared/TreeItemList';

interface InternalTeamsSectionProps {
  internalTeams: InternalTeam[];
  currentView: View | null;
  originRelationships: OriginRelationship[];
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

function buildTeamIdsInView(
  relationships: OriginRelationship[],
  componentIdsInView: Set<string>
): Set<string> {
  const inView = new Set<string>();
  for (const rel of relationships) {
    if (rel.relationshipType === 'BuiltBy' && componentIdsInView.has(rel.componentId)) {
      inView.add(rel.originEntityId);
    }
  }
  return inView;
}

export const InternalTeamsSection: React.FC<InternalTeamsSectionProps> = ({
  internalTeams,
  currentView,
  originRelationships,
  selectedTeamId,
  isExpanded,
  onToggle,
  onAddTeam,
  onTeamSelect,
  onTeamContextMenu,
}) => {
  const [search, setSearch] = useState('');

  const componentIdsInView = useMemo(() => {
    if (!currentView) return new Set<string>();
    return new Set(currentView.components.map(vc => vc.componentId));
  }, [currentView]);

  const teamIdsInView = useMemo(
    () => buildTeamIdsInView(originRelationships, componentIdsInView),
    [originRelationships, componentIdsInView]
  );

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
          isInView={(team) => !currentView || teamIdsInView.has(team.id)}
          getTitle={(team, isInView) =>
            isInView ? team.name : `${team.name} (not linked to components in current view)`
          }
          renderLabel={(team) => team.name}
          onSelect={(team) => onTeamSelect?.(team.id)}
          onContextMenu={onTeamContextMenu}
        />
      </div>
    </TreeSection>
  );
};
