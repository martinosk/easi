import { apiCall, parallelBatch } from "./config.ts";
import { generateCapabilityTree } from "./capability-tree.ts";
import { buildRealizationPlan, type PlannedRealization } from "./realization-plan.ts";
import type { Capability, CapNode } from "./types.ts";

type CapabilityLevel = "L1" | "L2" | "L3" | "L4";

interface CapabilitySpec {
  name: string;
  description: string;
  level: CapabilityLevel;
  parentId?: string;
}

async function createCapability(spec: CapabilitySpec): Promise<Capability> {
  return apiCall<Capability>("POST", "/capabilities", spec);
}

async function updateCapabilityMetadata(
  capabilityId: string,
  metadata: { status?: string; ownershipModel?: string; maturityValue?: number }
): Promise<void> {
  await apiCall("PUT", `/capabilities/${capabilityId}/metadata`, metadata);
}

interface LevelEntry {
  node: CapNode;
  parentId?: string;
}

function childrenOf(nodes: CapNode[]): CapNode[] {
  return nodes.flatMap((node) => node.children ?? []);
}

function childEntries(parents: CapNode[], parentMap: Map<string, Capability>): LevelEntry[] {
  const entries: LevelEntry[] = [];
  for (const parent of parents) {
    const created = parentMap.get(parent.name);
    if (!created) continue;
    for (const child of parent.children ?? []) {
      entries.push({ node: child, parentId: created.id });
    }
  }
  return entries;
}

async function seedLevel(entries: LevelEntry[], level: CapabilityLevel): Promise<Map<string, Capability>> {
  const map = new Map<string, Capability>();
  const results = await parallelBatch(entries, 10, ({ node, parentId }) =>
    createCapability({ name: node.name, description: node.description, level, parentId })
  );
  results.forEach((cap, i) => map.set(entries[i].node.name, cap));
  return map;
}

interface MetadataUpdate {
  name: string;
  status: string;
  maturityValue: number;
  ownershipModel: string;
}

const METADATA_UPDATES: MetadataUpdate[] = [
  { name: "Customer Acquisition", status: "Active", maturityValue: 65, ownershipModel: "TribeOwned" },
  { name: "Order Management", status: "Active", maturityValue: 80, ownershipModel: "EnterpriseService" },
  { name: "Payment Operations", status: "Active", maturityValue: 90, ownershipModel: "EnterpriseService" },
  { name: "Inventory Management", status: "Active", maturityValue: 70, ownershipModel: "TribeOwned" },
  { name: "Fraud Detection and Prevention", status: "Active", maturityValue: 75, ownershipModel: "EnterpriseService" },
  { name: "Predictive Analytics", status: "Planned", maturityValue: 30, ownershipModel: "TeamOwned" },
  { name: "Observability and Monitoring", status: "Active", maturityValue: 85, ownershipModel: "EnterpriseService" },
  { name: "API and Integration Management", status: "Active", maturityValue: 80, ownershipModel: "EnterpriseService" },
  { name: "Data Governance and Quality", status: "Active", maturityValue: 60, ownershipModel: "Shared" },
  { name: "Cybersecurity Management", status: "Active", maturityValue: 75, ownershipModel: "EnterpriseService" },
];

export async function seedCapabilities(): Promise<Map<string, Capability>> {
  console.log("\n🎯 Seeding Business Capabilities (L1–L4)...");
  const tree = generateCapabilityTree();
  const l2Nodes = childrenOf(tree);
  const l3Nodes = childrenOf(l2Nodes);

  console.log("  Creating L1 capabilities...");
  const l1Map = await seedLevel(tree.map((node) => ({ node })), "L1");
  console.log(`  ✓ ${l1Map.size} L1 capabilities created`);

  console.log("  Creating L2 capabilities...");
  const l2Map = await seedLevel(childEntries(tree, l1Map), "L2");
  console.log(`  ✓ ${l2Map.size} L2 capabilities created`);

  console.log("  Creating L3 capabilities...");
  const l3Map = await seedLevel(childEntries(l2Nodes, l2Map), "L3");
  console.log(`  ✓ ${l3Map.size} L3 capabilities created`);

  console.log("  Creating L4 capabilities...");
  const l4Map = await seedLevel(childEntries(l3Nodes, l3Map), "L4");
  console.log(`  ✓ ${l4Map.size} L4 capabilities created`);

  const allCapabilities = new Map<string, Capability>([...l1Map, ...l2Map, ...l3Map, ...l4Map]);
  console.log(`  Total: ${allCapabilities.size} capabilities`);

  console.log("  Applying metadata updates...");
  await parallelBatch(METADATA_UPDATES, 5, async (update) => {
    const cap = allCapabilities.get(update.name);
    if (!cap) return;
    await updateCapabilityMetadata(cap.id, {
      status: update.status,
      maturityValue: update.maturityValue,
      ownershipModel: update.ownershipModel,
    });
  });

  return allCapabilities;
}

export async function seedCapabilityDependencies(capabilities: Map<string, Capability>): Promise<void> {
  console.log("\n🔗 Seeding Capability Dependencies...");

  const dependencies = [
    { source: "Customer Acquisition", target: "Customer Identity Management", type: "requires", description: "Acquisition requires identity for registration" },
    { source: "Order Management", target: "Customer Identity Management", type: "requires", description: "Orders require authenticated customers" },
    { source: "Order Management", target: "Inventory Management", type: "requires", description: "Orders need inventory verification" },
    { source: "Checkout and Payment", target: "Payment Operations", type: "requires", description: "Checkout triggers payment processing" },
    { source: "Checkout and Payment", target: "Shipping and Logistics", type: "requires", description: "Fulfilled orders need shipping coordination" },
    { source: "Payment Operations", target: "Fraud Detection and Prevention", type: "supports", description: "Fraud detection supports payment authorization" },
    { source: "Customer Analytics", target: "Business Intelligence", type: "informs", description: "Customer analytics feeds enterprise BI" },
    { source: "Predictive Analytics", target: "Customer Analytics", type: "requires", description: "Predictions consume customer behavioral data" },
    { source: "Campaign Management", target: "Customer Analytics", type: "requires", description: "Campaigns need customer segment insights" },
    { source: "Demand Planning", target: "Inventory Management", type: "informs", description: "Demand forecasts drive inventory targets" },
  ];

  for (const dep of dependencies) {
    const source = capabilities.get(dep.source);
    const target = capabilities.get(dep.target);
    if (!source || !target) continue;
    try {
      await apiCall("POST", "/capability-dependencies", {
        sourceCapabilityId: source.id,
        targetCapabilityId: target.id,
        dependencyType: dep.type,
        description: dep.description,
      });
    } catch {
      console.log(`    (Skipping dependency — may already exist)`);
    }
  }
}

async function linkRealization(capabilityId: string, componentId: string, item: PlannedRealization): Promise<boolean> {
  for (let attempt = 0; attempt < 3; attempt++) {
    try {
      await apiCall("POST", `/capabilities/${capabilityId}/systems`, {
        componentId,
        realizationLevel: item.realizationLevel,
        notes: item.notes,
      });
      return true;
    } catch {
      await new Promise((resolve) => setTimeout(resolve, 40 * (attempt + 1)));
    }
  }
  return false;
}

export async function seedSystemRealizations(
  capabilities: Map<string, Capability>,
  components: Map<string, string>
): Promise<void> {
  console.log("\n⚙️  Seeding System Realizations (all applications, L1–L4)...");

  const plan = buildRealizationPlan([...components.keys()]);
  let created = 0;
  let skipped = 0;

  await parallelBatch(plan, 6, async (item) => {
    const capability = capabilities.get(item.capabilityName);
    const componentId = components.get(item.componentName);
    if (!capability || !componentId) {
      skipped++;
      return;
    }
    if (await linkRealization(capability.id, componentId, item)) created++;
    else skipped++;
  });

  const appsLinked = new Set(plan.map((p) => p.componentName)).size;
  console.log(`  ✓ ${created} direct realizations across L1–L4 for ${appsLinked} applications (${skipped} skipped)`);
}
