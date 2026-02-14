import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { StageFlowDiagram } from './StageFlowDiagram';
import type { ValueStreamStage, StageCapabilityMapping, StageId, ValueStreamId, HttpMethod } from '../../../api/types';

vi.mock('../hooks/useValueStreamStages', () => ({
  useRemoveStageCapability: () => ({ mutate: vi.fn() }),
}));

function createStage(id: string, position: number, name?: string): ValueStreamStage {
  return {
    id: id as StageId,
    valueStreamId: 'vs-1' as ValueStreamId,
    name: name ?? `Stage ${position}`,
    position,
    _links: {
      edit: { href: `/api/v1/value-streams/vs-1/stages/${id}`, method: 'PUT' as HttpMethod },
      delete: { href: `/api/v1/value-streams/vs-1/stages/${id}`, method: 'DELETE' as HttpMethod },
    },
  };
}

const defaultProps = {
  valueStreamId: 'vs-1',
  stages: [] as ValueStreamStage[],
  stageCapabilities: [] as StageCapabilityMapping[],
  canWrite: true,
  onAddStage: vi.fn(),
  onEditStage: vi.fn(),
  onDeleteStage: vi.fn(),
  onReorder: vi.fn(),
};

describe('StageFlowDiagram', () => {
  describe('Empty state', () => {
    it('should render empty state when no stages exist', () => {
      render(<StageFlowDiagram {...defaultProps} />);

      expect(screen.getByTestId('empty-stages')).toBeInTheDocument();
      expect(screen.getByText('No stages yet')).toBeInTheDocument();
    });

    it('should show add button in empty state when canWrite is true', () => {
      render(<StageFlowDiagram {...defaultProps} canWrite={true} />);

      expect(screen.getByTestId('add-stage-btn')).toBeInTheDocument();
    });

    it('should not show add button in empty state when canWrite is false', () => {
      render(<StageFlowDiagram {...defaultProps} canWrite={false} />);

      expect(screen.queryByTestId('add-stage-btn')).not.toBeInTheDocument();
    });

    it('should call onAddStage without arguments when clicking empty state add button', () => {
      const onAddStage = vi.fn();
      render(<StageFlowDiagram {...defaultProps} onAddStage={onAddStage} />);

      fireEvent.click(screen.getByTestId('add-stage-btn'));

      expect(onAddStage).toHaveBeenCalledTimes(1);
      expect(onAddStage).toHaveBeenCalledWith();
    });

    it('should NOT pass MouseEvent as argument when clicking empty state add button', () => {
      const onAddStage = vi.fn();
      render(<StageFlowDiagram {...defaultProps} onAddStage={onAddStage} />);

      fireEvent.click(screen.getByTestId('add-stage-btn'));

      const callArgs = onAddStage.mock.calls[0];
      expect(callArgs.length).toBe(0);
    });
  });

  describe('With stages', () => {
    const twoStages = [
      createStage('s1', 1, 'Discovery'),
      createStage('s2', 2, 'Delivery'),
    ];

    it('should render the stage flow diagram with stages', () => {
      render(<StageFlowDiagram {...defaultProps} stages={twoStages} />);

      expect(screen.getByTestId('stage-flow-diagram')).toBeInTheDocument();
      expect(screen.getByText('Discovery')).toBeInTheDocument();
      expect(screen.getByText('Delivery')).toBeInTheDocument();
    });

    it('should show trailing add button when canWrite is true', () => {
      render(<StageFlowDiagram {...defaultProps} stages={twoStages} />);

      expect(screen.getByTestId('add-stage-btn')).toBeInTheDocument();
    });

    it('should not show trailing add button when canWrite is false', () => {
      render(<StageFlowDiagram {...defaultProps} stages={twoStages} canWrite={false} />);

      expect(screen.queryByTestId('add-stage-btn')).not.toBeInTheDocument();
    });

    it('should call onAddStage without arguments when clicking trailing add button', () => {
      const onAddStage = vi.fn();
      render(<StageFlowDiagram {...defaultProps} stages={twoStages} onAddStage={onAddStage} />);

      fireEvent.click(screen.getByTestId('add-stage-btn'));

      expect(onAddStage).toHaveBeenCalledTimes(1);
      expect(onAddStage).toHaveBeenCalledWith();
    });

    it('should NOT pass MouseEvent as argument when clicking trailing add button', () => {
      const onAddStage = vi.fn();
      render(<StageFlowDiagram {...defaultProps} stages={twoStages} onAddStage={onAddStage} />);

      fireEvent.click(screen.getByTestId('add-stage-btn'));

      const callArgs = onAddStage.mock.calls[0];
      expect(callArgs.length).toBe(0);
    });

    it('should render connector insert buttons between stages', () => {
      render(<StageFlowDiagram {...defaultProps} stages={twoStages} />);

      expect(screen.getByTestId('insert-stage-btn-1')).toBeInTheDocument();
    });

    it('should call onAddStage with position when clicking connector insert button', () => {
      const onAddStage = vi.fn();
      render(<StageFlowDiagram {...defaultProps} stages={twoStages} onAddStage={onAddStage} />);

      fireEvent.click(screen.getByTestId('insert-stage-btn-1'));

      expect(onAddStage).toHaveBeenCalledTimes(1);
      expect(onAddStage).toHaveBeenCalledWith(2);
    });

    it('should pass a numeric position, not a MouseEvent, when clicking insert button', () => {
      const onAddStage = vi.fn();
      render(<StageFlowDiagram {...defaultProps} stages={twoStages} onAddStage={onAddStage} />);

      fireEvent.click(screen.getByTestId('insert-stage-btn-1'));

      const firstArg = onAddStage.mock.calls[0][0];
      expect(typeof firstArg).toBe('number');
    });

    it('should not show connector insert buttons when canWrite is false', () => {
      render(<StageFlowDiagram {...defaultProps} stages={twoStages} canWrite={false} />);

      expect(screen.queryByTestId('insert-stage-btn-1')).not.toBeInTheDocument();
    });

    it('should sort stages by position', () => {
      const unsorted = [
        createStage('s2', 2, 'Second'),
        createStage('s1', 1, 'First'),
      ];
      render(<StageFlowDiagram {...defaultProps} stages={unsorted} />);

      const stageNames = screen.getAllByRole('heading', { level: 3 }).map(el => el.textContent);
      expect(stageNames).toEqual(['First', 'Second']);
    });
  });

  describe('Capability drag-and-drop', () => {
    const stages = [createStage('s1', 1, 'Discovery')];

    it('should call onAddCapability when dropping a capability on a stage', () => {
      const onAddCapability = vi.fn();
      render(<StageFlowDiagram {...defaultProps} stages={stages} onAddCapability={onAddCapability} />);

      const stageEl = screen.getByTestId('stage-s1');
      const capabilityJson = JSON.stringify({ id: 'cap-1', name: 'Capability 1' });
      fireEvent.drop(stageEl, {
        dataTransfer: {
          getData: () => capabilityJson,
          effectAllowed: 'copy',
          dropEffect: 'copy',
        },
      });

      expect(onAddCapability).toHaveBeenCalledWith('s1', 'cap-1');
    });
  });
});
