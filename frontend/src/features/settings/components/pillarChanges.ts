import type { PillarChange } from '../../../api/metadata';
import { ApiError, type FitType, type StrategyPillar } from '../../../api/types';

export const MAX_PILLARS = 20;

export interface EditablePillar {
  id: string;
  name: string;
  description: string;
  active: boolean;
  fitScoringEnabled: boolean;
  fitCriteria: string;
  fitType: FitType;
  isNew: boolean;
  markedForDeletion: boolean;
}

export interface ValidationErrors {
  [index: number]: { name?: string };
}

export function isConflictError(err: unknown): boolean {
  return err instanceof ApiError && (err.statusCode === 409 || err.statusCode === 412);
}

export function toEditable(pillars: StrategyPillar[]): EditablePillar[] {
  return pillars.map((p) => ({
    ...p,
    fitScoringEnabled: p.fitScoringEnabled ?? false,
    fitCriteria: p.fitCriteria ?? '',
    fitType: p.fitType ?? '',
    isNew: false,
    markedForDeletion: false,
  }));
}

export function countActive(pillars: EditablePillar[]): number {
  return pillars.filter((p) => (p.active || p.isNew) && !p.markedForDeletion).length;
}

export function patchPillarAt(pillars: EditablePillar[], index: number, patch: Partial<EditablePillar>): EditablePillar[] {
  return pillars.map((row, i) => (i === index ? { ...row, ...patch } : row));
}

export function deleteOrMarkAt(pillars: EditablePillar[], index: number): EditablePillar[] {
  if (pillars[index].isNew) return pillars.filter((_, i) => i !== index);
  return patchPillarAt(pillars, index, { markedForDeletion: true });
}

export function emptyEditablePillar(): EditablePillar {
  return {
    id: `new-${Date.now()}`,
    name: '',
    description: '',
    active: true,
    fitScoringEnabled: false,
    fitCriteria: '',
    fitType: '',
    isNew: true,
    markedForDeletion: false,
  };
}

export function validatePillars(pillars: EditablePillar[]): ValidationErrors {
  const errors: ValidationErrors = {};
  const seen = new Set<string>();

  pillars.forEach((pillar, index) => {
    if (pillar.markedForDeletion && !pillar.isNew) return;
    if (!pillar.active && !pillar.isNew) return;

    const trimmed = pillar.name.trim();
    if (!trimmed) {
      errors[index] = { name: 'Name cannot be empty' };
      return;
    }
    if (trimmed.length > 100) {
      errors[index] = { name: 'Name must be 100 characters or less' };
      return;
    }
    const key = trimmed.toLowerCase();
    if (seen.has(key)) {
      errors[index] = { name: 'Name must be unique' };
      return;
    }
    seen.add(key);
  });

  return errors;
}

export function buildPillarChanges(edited: EditablePillar[], originals: StrategyPillar[]): PillarChange[] {
  return edited.map((p) => buildSingleChange(p, originals)).filter((c): c is PillarChange => c !== null);
}

function buildSingleChange(pillar: EditablePillar, originals: StrategyPillar[]): PillarChange | null {
  if (pillar.markedForDeletion) {
    return pillar.isNew ? null : { operation: 'remove', id: pillar.id };
  }
  if (pillar.isNew) {
    return { operation: 'add', ...trimmedFields(pillar) };
  }
  if (!pillar.active) return null;

  const original = originals.find((p) => p.id === pillar.id);
  if (!original || !hasChanged(pillar, original)) return null;

  return { operation: 'update', id: pillar.id, ...trimmedFields(pillar) };
}

function trimmedFields(pillar: EditablePillar) {
  return {
    name: pillar.name.trim(),
    description: pillar.description.trim(),
    fitScoringEnabled: pillar.fitScoringEnabled,
    fitCriteria: pillar.fitCriteria.trim(),
    fitType: pillar.fitType,
  };
}

function hasChanged(pillar: EditablePillar, original: StrategyPillar): boolean {
  return (
    original.name !== pillar.name.trim() ||
    original.description !== pillar.description.trim() ||
    original.fitScoringEnabled !== pillar.fitScoringEnabled ||
    original.fitCriteria !== pillar.fitCriteria.trim() ||
    original.fitType !== pillar.fitType
  );
}
