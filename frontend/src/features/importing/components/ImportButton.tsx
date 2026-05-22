import { Button } from '@mantine/core';

interface ImportButtonProps {
  onClick: () => void;
}

export function ImportButton({ onClick }: ImportButtonProps) {
  return (
    <Button
      variant="default"
      onClick={onClick}
      data-testid="import-button"
      title="Import from ArchiMate Open Exchange"
    >
      Import
    </Button>
  );
}
