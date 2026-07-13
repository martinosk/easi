import { CAPABILITY_DOMAIN_NAMES, generateCapabilityTree } from "./capability-tree.ts";

export interface PlannedRealization {
  componentName: string;
  capabilityName: string;
  realizationLevel: string;
  notes: string;
}

type Level = "L1" | "L2" | "L3" | "L4";

interface Subtree {
  l1: string;
  l2: string[];
  l3: string[];
  l4: string[];
}

const APP_DOMAIN_SCHEDULE: { count: number; capDomains: string[] }[] = [
  { count: 20, capDomains: ["Digital Commerce", "Customer", "Finance", "Technology Platform"] },
  { count: 28, capDomains: ["Customer"] },
  { count: 28, capDomains: ["Digital Commerce"] },
  { count: 28, capDomains: ["Finance"] },
  { count: 28, capDomains: ["Supply Chain"] },
  { count: 28, capDomains: ["Analytics and Intelligence"] },
  { count: 28, capDomains: ["Technology Platform"] },
  { count: 27, capDomains: ["Human Resources"] },
  { count: 28, capDomains: ["Marketing", "Sales"] },
  { count: 28, capDomains: ["Risk and Compliance"] },
  { count: 29, capDomains: ["Technology Platform", "Analytics and Intelligence"] },
];

const LEVEL_RECIPES: Level[][] = [
  ["L1", "L2", "L3"],
  ["L2", "L3", "L4"],
  ["L1", "L3", "L4"],
  ["L1", "L2", "L4"],
  ["L2", "L4"],
  ["L1", "L3"],
  ["L3", "L4"],
  ["L1", "L2", "L3", "L4"],
];

const REALIZATION_LEVELS = ["Full", "Partial", "Full", "Planned"];

function buildDomainPools(): Map<string, Subtree[]> {
  const pools = new Map<string, Subtree[]>();
  for (const domain of CAPABILITY_DOMAIN_NAMES) pools.set(domain, []);

  const tree = generateCapabilityTree();
  const perDomain = Math.ceil(tree.length / CAPABILITY_DOMAIN_NAMES.length);
  tree.forEach((l1, i) => {
    const domainIdx = Math.min(Math.floor(i / perDomain), CAPABILITY_DOMAIN_NAMES.length - 1);
    const subtree: Subtree = { l1: l1.name, l2: [], l3: [], l4: [] };
    for (const l2 of l1.children ?? []) {
      subtree.l2.push(l2.name);
      for (const l3 of l2.children ?? []) {
        subtree.l3.push(l3.name);
        for (const l4 of l3.children ?? []) subtree.l4.push(l4.name);
      }
    }
    pools.get(CAPABILITY_DOMAIN_NAMES[domainIdx])?.push(subtree);
  });

  return pools;
}

function appDomains(count: number): string[] {
  const domains: string[] = [];
  for (const entry of APP_DOMAIN_SCHEDULE) {
    for (let k = 0; k < entry.count && domains.length < count; k++) {
      domains.push(entry.capDomains[domains.length % entry.capDomains.length]);
    }
  }
  while (domains.length < count) {
    domains.push(CAPABILITY_DOMAIN_NAMES[domains.length % CAPABILITY_DOMAIN_NAMES.length]);
  }
  return domains;
}

function levelArray(subtree: Subtree, level: Level): string[] {
  switch (level) {
    case "L1":
      return [subtree.l1];
    case "L2":
      return subtree.l2;
    case "L3":
      return subtree.l3;
    case "L4":
      return subtree.l4;
  }
}

function pickCapability(subtrees: Subtree[], subtreeIdx: number, level: Level, salt: number): string | undefined {
  if (subtrees.length === 0) return undefined;
  const arr = levelArray(subtrees[subtreeIdx % subtrees.length], level);
  return arr.length === 0 ? undefined : arr[salt % arr.length];
}

export function buildRealizationPlan(componentNames: string[]): PlannedRealization[] {
  const pools = buildDomainPools();
  const domainsByApp = appDomains(componentNames.length);
  const plan: PlannedRealization[] = [];

  componentNames.forEach((componentName, i) => {
    const subtrees = pools.get(domainsByApp[i]) ?? [];
    const recipe = LEVEL_RECIPES[i % LEVEL_RECIPES.length];

    recipe.forEach((level, k) => {
      const capabilityName = pickCapability(subtrees, i + k, level, i);
      if (!capabilityName) return;
      plan.push({
        componentName,
        capabilityName,
        realizationLevel: REALIZATION_LEVELS[(i + k) % REALIZATION_LEVELS.length],
        notes: `${componentName} realizes ${capabilityName}`,
      });
    });
  });

  return plan;
}
