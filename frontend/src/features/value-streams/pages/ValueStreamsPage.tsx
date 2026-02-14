import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useValueStreams } from '../hooks/useValueStreams';
import { useUserStore } from '../../../store/userStore';
import { hasLink } from '../../../utils/hateoas';
import type { ValueStream, HATEOASLinks } from '../../../api/types';
import './ValueStreamsPage.css';

interface ValueStreamFormData {
  name: string;
  description: string;
}

const EMPTY_FORM: ValueStreamFormData = { name: '', description: '' };

interface ValueStreamFormOverlayProps {
  isEditing: boolean;
  formData: ValueStreamFormData;
  onFormDataChange: (data: ValueStreamFormData) => void;
  onSubmit: () => void;
  onCancel: () => void;
}

function ValueStreamFormOverlay({ isEditing, formData, onFormDataChange, onSubmit, onCancel }: ValueStreamFormOverlayProps) {
  return (
    <div className="vs-form-overlay" data-testid="value-stream-form">
      <div className="vs-form">
        <h3>{isEditing ? 'Edit Value Stream' : 'Create Value Stream'}</h3>
        <div className="vs-form-field">
          <label htmlFor="vs-name">Name</label>
          <input
            id="vs-name"
            type="text"
            value={formData.name}
            onChange={(e) => onFormDataChange({ ...formData, name: e.target.value })}
            placeholder="e.g. Customer Onboarding"
            maxLength={100}
            autoFocus
          />
        </div>
        <div className="vs-form-field">
          <label htmlFor="vs-description">Description</label>
          <textarea
            id="vs-description"
            value={formData.description}
            onChange={(e) => onFormDataChange({ ...formData, description: e.target.value })}
            placeholder="Optional description..."
            maxLength={500}
            rows={3}
          />
        </div>
        <div className="vs-form-actions">
          <button type="button" className="btn btn-secondary" onClick={onCancel}>Cancel</button>
          <button type="button" className="btn btn-primary" onClick={onSubmit} disabled={!formData.name.trim()}>
            {isEditing ? 'Save' : 'Create'}
          </button>
        </div>
      </div>
    </div>
  );
}

interface DeleteConfirmOverlayProps {
  streamName: string;
  onConfirm: () => void;
  onCancel: () => void;
}

function DeleteConfirmOverlay({ streamName, onConfirm, onCancel }: DeleteConfirmOverlayProps) {
  return (
    <div className="vs-form-overlay" data-testid="delete-confirmation">
      <div className="vs-form">
        <h3>Delete Value Stream</h3>
        <p className="vs-delete-warning">
          Are you sure you want to delete &ldquo;{streamName}&rdquo;? This will also remove all its stages and mappings.
        </p>
        <div className="vs-form-actions">
          <button type="button" className="btn btn-secondary" onClick={onCancel}>Cancel</button>
          <button type="button" className="btn btn-danger" onClick={onConfirm}>Delete</button>
        </div>
      </div>
    </div>
  );
}

interface ValueStreamCardProps {
  stream: ValueStream;
  canWrite: boolean;
  canDelete: boolean;
  onNavigate: (id: string) => void;
  onEdit: (stream: ValueStream) => void;
  onDelete: (stream: ValueStream) => void;
}

function ValueStreamCard({ stream, canWrite, canDelete, onNavigate, onEdit, onDelete }: ValueStreamCardProps) {
  return (
    <div
      className="vs-card vs-card-clickable"
      data-testid={`value-stream-${stream.id}`}
      onClick={() => onNavigate(stream.id)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => e.key === 'Enter' && onNavigate(stream.id)}
    >
      <div>
        <div className="vs-card-name">{stream.name}</div>
        {stream.description && <div className="vs-card-description">{stream.description}</div>}
        <div className="vs-card-meta">
          <span>{stream.stageCount} stages</span>
          <span>Created {new Date(stream.createdAt).toLocaleDateString()}</span>
        </div>
      </div>
      <div className="vs-card-actions">
        {canWrite && hasLink(stream, 'edit') && (
          <button type="button" className="btn-small" onClick={(e) => { e.stopPropagation(); onEdit(stream); }} data-testid={`edit-${stream.id}`}>
            Edit
          </button>
        )}
        {canDelete && hasLink(stream, 'delete') && (
          <button type="button" className="btn-small btn-danger" onClick={(e) => { e.stopPropagation(); onDelete(stream); }} data-testid={`delete-${stream.id}`}>
            Delete
          </button>
        )}
      </div>
    </div>
  );
}

function useValueStreamFormState(
  createFn: (name: string, description?: string) => Promise<unknown>,
  updateFn: (stream: ValueStream, name: string, description?: string) => Promise<unknown>,
  deleteFn: (stream: ValueStream) => Promise<unknown>,
) {
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingStream, setEditingStream] = useState<ValueStream | null>(null);
  const [deletingStream, setDeletingStream] = useState<ValueStream | null>(null);
  const [formData, setFormData] = useState<ValueStreamFormData>(EMPTY_FORM);

  const handleSubmit = useCallback(async () => {
    if (!formData.name.trim()) return;
    const desc = formData.description || undefined;
    if (editingStream) {
      await updateFn(editingStream, formData.name, desc);
      setEditingStream(null);
    } else {
      await createFn(formData.name, desc);
      setShowCreateForm(false);
    }
    setFormData(EMPTY_FORM);
  }, [createFn, updateFn, editingStream, formData]);

  const handleDelete = useCallback(async () => {
    if (!deletingStream) return;
    await deleteFn(deletingStream);
    setDeletingStream(null);
  }, [deleteFn, deletingStream]);

  const startEdit = useCallback((stream: ValueStream) => {
    setEditingStream(stream);
    setFormData({ name: stream.name, description: stream.description || '' });
  }, []);

  const closeForm = useCallback(() => {
    setShowCreateForm(false);
    setEditingStream(null);
    setFormData(EMPTY_FORM);
  }, []);

  const openCreateForm = useCallback(() => {
    setFormData(EMPTY_FORM);
    setShowCreateForm(true);
  }, []);

  return {
    isFormOpen: showCreateForm || editingStream !== null,
    isEditing: editingStream !== null,
    deletingStream, formData, setFormData,
    handleSubmit, handleDelete, startEdit, closeForm, openCreateForm,
    setDeletingStream,
  };
}

function checkCanCreate(canWrite: boolean, collectionLinks: HATEOASLinks | undefined): boolean {
  return canWrite && !!collectionLinks && hasLink({ _links: collectionLinks }, 'create');
}

export function ValueStreamsPage() {
  const { valueStreams, isLoading, error, createValueStream, updateValueStream, deleteValueStream, collectionLinks } = useValueStreams();
  const hasPermission = useUserStore((state) => state.hasPermission);
  const navigate = useNavigate();
  const canWrite = hasPermission('valuestreams:write');
  const canDelete = hasPermission('valuestreams:delete');
  const canCreate = checkCanCreate(canWrite, collectionLinks);

  const form = useValueStreamFormState(createValueStream, updateValueStream, deleteValueStream);

  if (isLoading) {
    return (
      <div className="vs-page">
        <div className="vs-container">
          <div className="vs-loading">Loading value streams...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="vs-page">
        <div className="vs-container">
          <div className="vs-error">Failed to load value streams: {error.message}</div>
        </div>
      </div>
    );
  }

  return (
    <div className="vs-page" data-testid="value-streams-page">
      <div className="vs-container">
        <div className="vs-header">
          <div>
            <h1 className="vs-title">
              Value Streams
              <span className="vs-count">{valueStreams.length}</span>
            </h1>
            <p className="vs-subtitle">Model how your organization delivers value end-to-end.</p>
          </div>
          {canCreate && (
            <button type="button" className="btn btn-primary" onClick={form.openCreateForm} data-testid="create-value-stream-btn">
              <svg className="btn-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 5V19M5 12H19" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
              Create Value Stream
            </button>
          )}
        </div>

        {form.isFormOpen && (
          <ValueStreamFormOverlay
            isEditing={form.isEditing}
            formData={form.formData}
            onFormDataChange={form.setFormData}
            onSubmit={form.handleSubmit}
            onCancel={form.closeForm}
          />
        )}

        {form.deletingStream && (
          <DeleteConfirmOverlay
            streamName={form.deletingStream.name}
            onConfirm={form.handleDelete}
            onCancel={() => form.setDeletingStream(null)}
          />
        )}

        {valueStreams.length === 0 ? (
          <div className="vs-empty" data-testid="empty-state">
            <svg className="vs-empty-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M22 12H18L15 21L9 3L6 12H2" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            <h3>No value streams yet</h3>
            <p>Value streams model how your organization delivers value end-to-end. Create your first value stream to get started.</p>
            {canCreate && (
              <button type="button" className="btn btn-primary" onClick={form.openCreateForm}>
                Create your first value stream
              </button>
            )}
          </div>
        ) : (
          <div className="vs-list" data-testid="value-streams-list">
            {valueStreams.map((stream) => (
              <ValueStreamCard
                key={stream.id}
                stream={stream}
                canWrite={canWrite}
                canDelete={canDelete}
                onNavigate={navigate}
                onEdit={form.startEdit}
                onDelete={form.setDeletingStream}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
