import React from "react";
import { useNavigate } from "react-router-dom";
import { DetailField } from "../../../components/shared/DetailField";
import { useCapabilityValueStreams } from "../../value-streams/hooks/useValueStreams";
import type { CapabilityValueStreamParticipation } from "../../../api/types";

interface CapabilityValueStreamsSectionProps {
  capabilityId: string;
}

interface GroupedParticipation {
  valueStreamId: string;
  valueStreamName: string;
  stages: { stageId: string; stageName: string }[];
}

function groupByValueStream(
  participations: CapabilityValueStreamParticipation[],
): GroupedParticipation[] {
  const map = new Map<string, GroupedParticipation>();

  for (const p of participations) {
    const existing = map.get(p.valueStreamId);
    if (existing) {
      existing.stages.push({ stageId: p.stageId, stageName: p.stageName });
    } else {
      map.set(p.valueStreamId, {
        valueStreamId: p.valueStreamId,
        valueStreamName: p.valueStreamName,
        stages: [{ stageId: p.stageId, stageName: p.stageName }],
      });
    }
  }

  return Array.from(map.values()).sort((a, b) =>
    a.valueStreamName.localeCompare(b.valueStreamName),
  );
}

export const CapabilityValueStreamsSection: React.FC<
  CapabilityValueStreamsSectionProps
> = ({ capabilityId }) => {
  const navigate = useNavigate();
  const { data, isLoading } = useCapabilityValueStreams(capabilityId);

  const participations = data?.data ?? [];

  if (isLoading) return null;
  if (participations.length === 0) return null;

  const grouped = groupByValueStream(participations);

  return (
    <DetailField label="Value Streams">
      <ul className="value-stream-participation-list">
        {grouped.map((group) => (
          <li
            key={group.valueStreamId}
            className="value-stream-participation-item"
          >
            <button
              className="link-button"
              onClick={() => navigate(`/value-streams/${group.valueStreamId}`)}
              title={`Go to ${group.valueStreamName}`}
            >
              {group.valueStreamName}
            </button>
            <span className="value-stream-stages">
              {group.stages.map((s) => s.stageName).join(", ")}
            </span>
          </li>
        ))}
      </ul>
    </DetailField>
  );
};
