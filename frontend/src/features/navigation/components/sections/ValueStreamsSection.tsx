import React, { useState, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import type { ValueStream } from "../../../../api/types";
import { TreeSection } from "../TreeSection";
import { TreeSearchInput } from "../shared/TreeSearchInput";
import { TreeItemList } from "../shared/TreeItemList";

interface ValueStreamsSectionProps {
  valueStreams: ValueStream[];
  isExpanded: boolean;
  onToggle: () => void;
  onAddValueStream?: () => void;
  onValueStreamContextMenu: (
    e: React.MouseEvent,
    valueStream: ValueStream,
  ) => void;
}

function filterValueStreams(
  valueStreams: ValueStream[],
  search: string,
): ValueStream[] {
  if (!search.trim()) return valueStreams;
  const searchLower = search.toLowerCase();
  return valueStreams.filter((v) => v.name.toLowerCase().includes(searchLower));
}

function sortValueStreams(valueStreams: ValueStream[]): ValueStream[] {
  return [...valueStreams].sort((a, b) => a.name.localeCompare(b.name));
}

export const ValueStreamsSection: React.FC<ValueStreamsSectionProps> = ({
  valueStreams,
  isExpanded,
  onToggle,
  onAddValueStream,
  onValueStreamContextMenu,
}) => {
  const [search, setSearch] = useState("");
  const navigate = useNavigate();

  const sortedValueStreams = useMemo(
    () => sortValueStreams(valueStreams),
    [valueStreams],
  );

  const filteredValueStreams = useMemo(
    () => filterValueStreams(sortedValueStreams, search),
    [sortedValueStreams, search],
  );

  const hasNoValueStreams = valueStreams.length === 0;
  const emptyMessage = hasNoValueStreams ? "No value streams" : "No matches";

  const handleSelect = (vs: ValueStream) => {
    navigate(`/value-streams/${vs.id}`);
  };

  const handleContextMenu = (e: React.MouseEvent, vs: ValueStream) => {
    onValueStreamContextMenu(e, vs);
  };

  return (
    <TreeSection
      label="Value Streams"
      count={valueStreams.length}
      isExpanded={isExpanded}
      onToggle={onToggle}
      onAdd={onAddValueStream}
      addTitle="Manage value streams"
      addTestId="navigate-value-streams-button"
    >
      <TreeSearchInput
        value={search}
        onChange={setSearch}
        placeholder="Search value streams..."
      />
      <div className="tree-items">
        <TreeItemList
          items={filteredValueStreams}
          emptyMessage={emptyMessage}
          icon="🔄"
          dragDataKey="valueStreamId"
          isSelected={() => false}
          isInView={() => true}
          getTitle={(vs) => vs.name}
          renderLabel={(vs) => vs.name}
          onSelect={(vs) => handleSelect(vs)}
          onContextMenu={handleContextMenu}
        />
      </div>
    </TreeSection>
  );
};
