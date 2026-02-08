const editGrantErrorMessages: Record<string, string> = {
  'Cannot grant edit access to yourself': 'You cannot grant edit access to yourself.',
  'An active edit grant already exists for this user and artifact': 'This user already has active edit access to this artifact.',
  'Edit grant has already been revoked': 'This edit grant has already been revoked.',
  'Edit grant has already expired': 'This edit grant has already expired.',
  'Edit grant not found': 'The edit grant was not found.',
  'Invalid artifact type': 'The artifact type is not valid.',
};

export function getEditGrantErrorMessage(errorMessage: string): string {
  return editGrantErrorMessages[errorMessage] ?? errorMessage;
}
