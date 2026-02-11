import { useState, useCallback } from 'react';
import { useValueStreams } from '../hooks/useValueStreams';
import { useUserStore } from '../../../store/userStore';
import { hasLink } from '../../../utils/hateoas';
import type { ValueStream } from '../../../api/types';
import './ValueStreamsPage.css';

interface ValueStreamFormData {
  name: string;
  description: string;
}

export function ValueStreamsPage() {
  const { valueStreams, isLoading, error, createValueStream, updateValueStream, deleteValueStream, collectionLinks } = useValueStreams();
  const hasPermission = useUserStore((state) => state.hasPermission);
  const canWrite = hasPermission('valuestreams:write');
  const canDelete = hasPermission('valuestreams:delete');

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingStream, setEditingStream] = useState<ValueStream | null>(null);
  const [deletingStream, setDeletingStream] = useState<ValueStream | null>(null);
  const [formData, setFormData] = useState<ValueStreamFormData>({ name: '', description: '' });

  const handleCreate = useCallback(async () => {
    if (!formData.name.trim()) return;
    await createValueStream(formData.name, formData.description || undefined);
    setFormData({ name: '', description: '' });
    setShowCreateForm(false);
  }, [createValueStream, formData]);

  const handleUpdate = useCallback(async () => {
    if (!editingStream || !formData.name.trim()) return;
    await updateValueStream(editingStream, formData.name, formData.description || undefined);
    setEditingStream(null);
    setFormData({ name: '', description: '' });
  }, [updateValueStream, editingStream, formData]);

  const handleDelete = useCallback(async () => {
    if (!deletingStream) return;
    await deleteValueStream(deletingStream);
    setDeletingStream(null);
  }, [deleteValueStream, deletingStream]);

  const startEdit = useCallback((stream: ValueStream) => {
    setEditingStream(stream);
    setFormData({ name: stream.name, description: stream.description || '' });
  }, []);

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

  const canCreate = canWrite && collectionLinks && hasLink({ _links: collectionLinks }, 'x-create');

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
            <button
              type="button"
              className="btn btn-primary"
              onClick={() => {
                setFormData({ name: '', description: '' });
                setShowCreateForm(true);
              }}
              data-testid="create-value-stream-btn"
            >
              <svg className="btn-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 5V19M5 12H19" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
              Create Value Stream
            </button>
          )}
        </div>

        {(showCreateForm || editingStream) && (
          <div className="vs-form-overlay" data-testid="value-stream-form">
            <div className="vs-form">
              <h3>{editingStream ? 'Edit Value Stream' : 'Create Value Stream'}</h3>
              <div className="vs-form-field">
                <label htmlFor="vs-name">Name</label>
                <input
                  id="vs-name"
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
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
                  onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                  placeholder="Optional description..."
                  maxLength={500}
                  rows={3}
                />
              </div>
              <div className="vs-form-actions">
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => {
                    setShowCreateForm(false);
                    setEditingStream(null);
                    setFormData({ name: '', description: '' });
                  }}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={editingStream ? handleUpdate : handleCreate}
                  disabled={!formData.name.trim()}
                >
                  {editingStream ? 'Save' : 'Create'}
                </button>
              </div>
            </div>
          </div>
        )}

        {deletingStream && (
          <div className="vs-form-overlay" data-testid="delete-confirmation">
            <div className="vs-form">
              <h3>Delete Value Stream</h3>
              <p className="vs-delete-warning">
                Are you sure you want to delete &ldquo;{deletingStream.name}&rdquo;? This will also remove all its stages and mappings.
              </p>
              <div className="vs-form-actions">
                <button type="button" className="btn btn-secondary" onClick={() => setDeletingStream(null)}>Cancel</button>
                <button type="button" className="btn btn-danger" onClick={handleDelete}>Delete</button>
              </div>
            </div>
          </div>
        )}

        {valueStreams.length === 0 ? (
          <div className="vs-empty" data-testid="empty-state">
            <svg className="vs-empty-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M22 12H18L15 21L9 3L6 12H2" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            <h3>No value streams yet</h3>
            <p>Value streams model how your organization delivers value end-to-end. Create your first value stream to get started.</p>
            {canCreate && (
              <button
                type="button"
                className="btn btn-primary"
                onClick={() => {
                  setFormData({ name: '', description: '' });
                  setShowCreateForm(true);
                }}
              >
                Create your first value stream
              </button>
            )}
          </div>
        ) : (
          <div className="vs-list" data-testid="value-streams-list">
            {valueStreams.map((stream) => (
              <div key={stream.id} className="vs-card" data-testid={`value-stream-${stream.id}`}>
                <div>
                  <div className="vs-card-name">{stream.name}</div>
                  {stream.description && (
                    <div className="vs-card-description">{stream.description}</div>
                  )}
                  <div className="vs-card-meta">
                    <span>{stream.stageCount} stages</span>
                    <span>Created {new Date(stream.createdAt).toLocaleDateString()}</span>
                  </div>
                </div>
                <div className="vs-card-actions">
                  {canWrite && hasLink(stream, 'edit') && (
                    <button
                      type="button"
                      className="btn-small"
                      onClick={() => startEdit(stream)}
                      data-testid={`edit-${stream.id}`}
                    >
                      Edit
                    </button>
                  )}
                  {canDelete && hasLink(stream, 'delete') && (
                    <button
                      type="button"
                      className="btn-small btn-danger"
                      onClick={() => setDeletingStream(stream)}
                      data-testid={`delete-${stream.id}`}
                    >
                      Delete
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
