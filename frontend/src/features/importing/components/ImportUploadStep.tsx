import { Alert, Button, FileInput, Group, NativeSelect, Stack, Text } from '@mantine/core';
import { useState } from 'react';
import type { BusinessDomain } from '../../../api/types';
import type { User } from '../../users/types';

interface ImportUploadStepProps {
  businessDomains: BusinessDomain[];
  eaOwnerCandidates: User[];
  isLoading: boolean;
  error: string | null;
  onUpload: (file: File, businessDomainId?: string, capabilityEAOwner?: string) => void;
  onCancel: () => void;
}

function buildDomainOptions(domains: BusinessDomain[]) {
  return [
    { value: '', label: 'None - Do not assign to domain' },
    ...domains.map((d) => ({ value: d.id, label: d.name })),
  ];
}

function buildOwnerOptions(candidates: User[]) {
  return [
    { value: '', label: 'Select EA Owner (optional)' },
    ...candidates.map((u) => ({ value: u.id, label: u.name || u.email })),
  ];
}

export function ImportUploadStep({
  businessDomains,
  eaOwnerCandidates,
  isLoading,
  error,
  onUpload,
  onCancel,
}: ImportUploadStepProps) {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [selectedDomain, setSelectedDomain] = useState<string>('');
  const [selectedEAOwner, setSelectedEAOwner] = useState<string>('');

  const handleSubmit = () => {
    if (selectedFile) {
      onUpload(selectedFile, selectedDomain || undefined, selectedEAOwner || undefined);
    }
  };

  return (
    <Stack gap="md">
      <Text c="dimmed" size="sm">
        Select an ArchiMate Open Exchange XML file to import capabilities and components.
      </Text>

      <FileInput
        label="File"
        placeholder="Choose .xml file"
        accept=".xml,application/xml,text/xml"
        value={selectedFile}
        onChange={setSelectedFile}
        disabled={isLoading}
        withAsterisk
        data-testid="file-input"
      />

      <NativeSelect
        label="Business Domain (Optional)"
        description="If selected, L1 capabilities will be assigned to this business domain."
        data={buildDomainOptions(businessDomains)}
        value={selectedDomain}
        onChange={(event) => setSelectedDomain(event.currentTarget.value)}
        disabled={isLoading}
        data-testid="domain-select"
      />

      <NativeSelect
        label="EA Owner for Capabilities (Optional)"
        description="If selected, this user will be assigned as EA Owner to all imported capabilities."
        data={buildOwnerOptions(eaOwnerCandidates)}
        value={selectedEAOwner}
        onChange={(event) => setSelectedEAOwner(event.currentTarget.value)}
        disabled={isLoading}
        data-testid="ea-owner-select"
      />

      {error && (
        <Alert color="red" data-testid="upload-error">
          {error}
        </Alert>
      )}

      <Group justify="flex-end" gap="sm">
        <Button variant="default" onClick={onCancel} disabled={isLoading} data-testid="cancel-button">
          Cancel
        </Button>
        <Button
          onClick={handleSubmit}
          loading={isLoading}
          disabled={!selectedFile}
          data-testid="upload-button"
        >
          Upload
        </Button>
      </Group>
    </Stack>
  );
}
