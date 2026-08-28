import { type DragEvent, type KeyboardEvent, useState } from 'react';
import { hasLink } from '../../../utils/hateoas';
import type { CapabilityJourney } from '../types';
import { moveMilestone } from '../utils/milestoneOrder';
import { useReorderJourneyMilestones } from './useJourneys';

export interface MilestoneReorderControls {
  overIndex: number | null;
  rowProps: (index: number) => {
    draggable: true;
    onDragStart: (event: DragEvent<HTMLElement>) => void;
    onDragOver: (event: DragEvent<HTMLElement>) => void;
    onDrop: (event: DragEvent<HTMLElement>) => void;
    onDragEnd: () => void;
  };
  handleKeyDown: (index: number) => (event: KeyboardEvent<HTMLElement>) => void;
}

const KEY_STEPS: Record<string, number> = { ArrowUp: -1, ArrowDown: 1 };

export function useMilestoneReorder(journey: CapabilityJourney): MilestoneReorderControls | null {
  const reorderMutation = useReorderJourneyMilestones();
  const [dragIndex, setDragIndex] = useState<number | null>(null);
  const [overIndex, setOverIndex] = useState<number | null>(null);

  if (!hasLink(journey, 'x-reorder-milestones')) return null;

  const submitMove = (from: number, to: number) => {
    const milestoneIds = moveMilestone(
      journey.milestones.map((m) => m.id),
      from,
      to,
    );
    if (milestoneIds) reorderMutation.mutate({ journey, request: { milestoneIds } });
  };

  const resetDrag = () => {
    setDragIndex(null);
    setOverIndex(null);
  };

  return {
    overIndex,
    rowProps: (index) => ({
      draggable: true,
      onDragStart: (event) => {
        event.dataTransfer.effectAllowed = 'move';
        event.dataTransfer.setData('text/plain', journey.milestones[index].id);
        setDragIndex(index);
      },
      onDragOver: (event) => {
        event.preventDefault();
        if (overIndex !== index) setOverIndex(index);
      },
      onDrop: (event) => {
        event.preventDefault();
        if (dragIndex !== null) submitMove(dragIndex, index);
        resetDrag();
      },
      onDragEnd: resetDrag,
    }),
    handleKeyDown: (index) => (event) => {
      const step = KEY_STEPS[event.key];
      if (step === undefined) return;
      event.preventDefault();
      submitMove(index, index + step);
    },
  };
}
