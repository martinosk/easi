import { Group, SimpleGrid, Text } from '@mantine/core';
import type { ApplicationsTile, CapabilitiesTile, DomainsTile, HomePortfolio } from '../types';
import { TimeTile } from './TimeTile';
import { Tile, TileTotal } from './Tile';

function CapabilitiesTileView({ capabilities }: { capabilities: CapabilitiesTile }) {
  const { active, planned, deprecated } = capabilities.byStatus;
  return (
    <Tile label="Capabilities" testId="home-tile-capabilities">
      <TileTotal value={capabilities.total} />
      <Group gap="sm">
        <Text size="xs" c="dimmed">{`Active ${active}`}</Text>
        <Text size="xs" c="dimmed">{`Planned ${planned}`}</Text>
        <Text size="xs" c="dimmed">{`Deprecated ${deprecated}`}</Text>
      </Group>
    </Tile>
  );
}

function ApplicationsTileView({ applications }: { applications: ApplicationsTile }) {
  return (
    <Tile label="Applications" testId="home-tile-applications">
      <TileTotal value={applications.total} />
    </Tile>
  );
}

function domainNamesText({ total, names }: { total: number; names: string[] }): string {
  const rest = total - names.length;
  return rest > 0 ? `${names.join(', ')} and ${rest} more` : names.join(', ');
}

function DomainsTileView({ domains }: { domains: DomainsTile }) {
  return (
    <Tile label="Domains" testId="home-tile-domains">
      <TileTotal value={domains.total} />
      {domains.names && domains.names.length > 0 && (
        <Text size="xs" c="dimmed">
          {domainNamesText({ total: domains.total, names: domains.names })}
        </Text>
      )}
    </Tile>
  );
}

export function PortfolioTiles({ portfolio }: { portfolio: HomePortfolio }) {
  const { capabilities, applications, domains, time } = portfolio;
  return (
    <SimpleGrid cols={{ base: 1, sm: 2, lg: 4 }} spacing="md" data-testid="home-tiles">
      {capabilities && <CapabilitiesTileView capabilities={capabilities} />}
      {applications && <ApplicationsTileView applications={applications} />}
      {domains && <DomainsTileView domains={domains} />}
      {time && <TimeTile time={time} />}
    </SimpleGrid>
  );
}
