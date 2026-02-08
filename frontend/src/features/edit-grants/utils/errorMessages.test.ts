import { describe, it, expect } from 'vitest';
import { getEditGrantErrorMessage } from './errorMessages';

describe('getEditGrantErrorMessage', () => {
  it('should map self-grant error to user-facing message', () => {
    expect(getEditGrantErrorMessage('Cannot grant edit access to yourself')).toBe(
      'You cannot grant edit access to yourself.'
    );
  });

  it('should map duplicate grant error to user-facing message', () => {
    expect(
      getEditGrantErrorMessage('An active edit grant already exists for this user and artifact')
    ).toBe('This user already has active edit access to this artifact.');
  });

  it('should map already revoked error to user-facing message', () => {
    expect(getEditGrantErrorMessage('Edit grant has already been revoked')).toBe(
      'This edit grant has already been revoked.'
    );
  });

  it('should map already expired error to user-facing message', () => {
    expect(getEditGrantErrorMessage('Edit grant has already expired')).toBe(
      'This edit grant has already expired.'
    );
  });

  it('should map not found error to user-facing message', () => {
    expect(getEditGrantErrorMessage('Edit grant not found')).toBe(
      'The edit grant was not found.'
    );
  });

  it('should map invalid artifact type error to user-facing message', () => {
    expect(getEditGrantErrorMessage('Invalid artifact type')).toBe(
      'The artifact type is not valid.'
    );
  });

  it('should return the original message for unmapped errors', () => {
    const unknownError = 'Some unexpected server error';
    expect(getEditGrantErrorMessage(unknownError)).toBe(unknownError);
  });

  it('should return empty string for empty input', () => {
    expect(getEditGrantErrorMessage('')).toBe('');
  });
});
