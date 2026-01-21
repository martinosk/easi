import React, { useState, useMemo } from 'react';
import type { InternalTeam } from '../../../../api/types';
import { TreeSection } from '../TreeSection';

interface InternalTeamsSectionProps {
  internalTeams: InternalTeam[];
  selectedTeamId: string | null;
  isExpanded: boolean;
  onToggle: () => void;
  onAddTeam?: () => void;
  onTeamSelect?: (teamId: string) => void;
  onTeamContextMenu: (e: React.MouseEvent, team: InternalTeam) => void;
}

export const InternalTeamsSection: React.FC<InternalTeamsSectionProps> = ({
  internalTeams,
  selectedTeamId,
  isExpanded,
  onToggle,
  onAddTeam,
  onTeamSelect,
  onTeamContextMenu,
}) => {
  const [search, setSearch] = useState('');

  const filteredTeams = useMemo(() => {
    if (!search.trim()) {
      return internalTeams;
    }
    const searchLower = search.toLowerCase();
    return internalTeams.filter(
      (t) =>
        t.name.toLowerCase().includes(searchLower) ||
        (t.department && t.department.toLowerCase().includes(searchLower)) ||
        (t.contactPerson && t.contactPerson.toLowerCase().includes(searchLower)) ||
        (t.notes && t.notes.toLowerCase().includes(searchLower))
    );
  }, [internalTeams, search]);

  const handleTeamClick = (teamId: string) => {
    if (onTeamSelect) {
      onTeamSelect(teamId);
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
      <div className="tree-search">
        <input
          type="text"
          className="tree-search-input"
          placeholder="Search internal teams..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        {search && (
          <button
            className="tree-search-clear"
            onClick={() => setSearch('')}
            aria-label="Clear search"
          >
            x
          </button>
        )}
      </div>
      <div className="tree-items">
        {filteredTeams.length === 0 ? (
          <div className="tree-item-empty">
            {internalTeams.length === 0 ? 'No internal teams' : 'No matches'}
          </div>
        ) : (
          filteredTeams.map((team) => {
            const isSelected = selectedTeamId === team.id;

            return (
              <button
                key={team.id}
                className={`tree-item ${isSelected ? 'selected' : ''}`}
                onClick={() => handleTeamClick(team.id)}
                onContextMenu={(e) => onTeamContextMenu(e, team)}
                title={team.name}
                draggable
                onDragStart={(e) => {
                  e.dataTransfer.setData('internalTeamId', team.id);
                  e.dataTransfer.effectAllowed = 'copy';
                }}
              >
                <span className="tree-item-icon">👥</span>
                <span className="tree-item-label">{team.name}</span>
              </button>
            );
          })
        )}
      </div>
    </TreeSection>
  );
};
