import { apiCall, parallelBatch } from "./config.ts";
import { generateCapabilityTree } from "./capability-tree.ts";
import type { Capability, CapNode } from "./types.ts";

async function createCapability(
  name: string,
  description: string,
  level: string,
  parentId?: string
): Promise<Capability> {
  return apiCall<Capability>("POST", "/capabilities", { name, description, level, parentId });
}

async function updateCapabilityMetadata(
  capabilityId: string,
  metadata: { status?: string; ownershipModel?: string; maturityValue?: number }
): Promise<void> {
  await apiCall("PUT", `/capabilities/${capabilityId}/metadata`, metadata);
}

async function seedL1(tree: CapNode[]): Promise<Map<string, Capability>> {
  const map = new Map<string, Capability>();
  const results = await parallelBatch(tree, 10, (node) =>
    createCapability(node.name, node.description, "L1")
  );
  results.forEach((cap, i) => map.set(tree[i].name, cap));
  return map;
}

async function seedL2(
  tree: CapNode[],
  l1Map: Map<string, Capability>
): Promise<Map<string, Capability>> {
  const map = new Map<string, Capability>();
  const allL2: { node: CapNode; parentId: string }[] = [];

  for (const l1Node of tree) {
    const parent = l1Map.get(l1Node.name);
    if (!parent) continue;
    for (const l2Node of l1Node.children ?? []) {
      allL2.push({ node: l2Node, parentId: parent.id });
    }
  }

  const results = await parallelBatch(allL2, 10, ({ node, parentId }) =>
    createCapability(node.name, node.description, "L2", parentId)
  );
  results.forEach((cap, i) => map.set(allL2[i].node.name, cap));
  return map;
}

async function seedL3(
  tree: CapNode[],
  l2Map: Map<string, Capability>
): Promise<Map<string, Capability>> {
  const map = new Map<string, Capability>();
  const allL3: { node: CapNode; parentId: string }[] = [];

  for (const l1Node of tree) {
    for (const l2Node of l1Node.children ?? []) {
      const parent = l2Map.get(l2Node.name);
      if (!parent) continue;
      for (const l3Node of l2Node.children ?? []) {
        allL3.push({ node: l3Node, parentId: parent.id });
      }
    }
  }

  const results = await parallelBatch(allL3, 10, ({ node, parentId }) =>
    createCapability(node.name, node.description, "L3", parentId)
  );
  results.forEach((cap, i) => map.set(allL3[i].node.name, cap));
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
  console.log("\n🎯 Seeding Business Capabilities (1300 total)...");
  const tree = generateCapabilityTree();

  console.log("  Creating L1 capabilities (100)...");
  const l1Map = await seedL1(tree);
  console.log(`  ✓ ${l1Map.size} L1 capabilities created`);

  console.log("  Creating L2 capabilities (400)...");
  const l2Map = await seedL2(tree, l1Map);
  console.log(`  ✓ ${l2Map.size} L2 capabilities created`);

  console.log("  Creating L3 capabilities (800)...");
  const l3Map = await seedL3(tree, l2Map);
  console.log(`  ✓ ${l3Map.size} L3 capabilities created`);

  const allCapabilities = new Map<string, Capability>([...l1Map, ...l2Map, ...l3Map]);
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

export async function seedSystemRealizations(
  capabilities: Map<string, Capability>,
  components: Map<string, string>
): Promise<void> {
  console.log("\n⚙️  Seeding System Realizations...");

  const realizations = [
    { capability: "Customer Identity Management", component: "User Service", level: "Full", notes: "Primary customer identity system" },
    { capability: "Customer Acquisition", component: "Customer Portal", level: "Partial", notes: "Digital acquisition entry point" },
    { capability: "Order Management", component: "Order Service", level: "Full", notes: "Core order processing" },
    { capability: "Checkout and Payment", component: "Shopping Cart", level: "Partial", notes: "Cart to order conversion" },
    { capability: "Checkout and Payment", component: "Order Service", level: "Full", notes: "Checkout orchestration" },
    { capability: "Payment Operations", component: "Payment Gateway", level: "Full", notes: "Payment processing hub" },
    { capability: "Fraud Detection and Prevention", component: "Fraud Detection", level: "Full", notes: "Real-time fraud decisioning" },
    { capability: "Product Catalog Management", component: "Product Catalog", level: "Full", notes: "Central product repository" },
    { capability: "Product Catalog Management", component: "Search Engine", level: "Partial", notes: "Searchable product index" },
    { capability: "Inventory Management", component: "Inventory Service", level: "Full", notes: "Inventory tracking and reservation" },
    { capability: "Transportation Management", component: "Shipping Service", level: "Full", notes: "Shipping calculation and carrier integration" },
    { capability: "Business Intelligence", component: "Reporting Service", level: "Full", notes: "Report generation and delivery" },
    { capability: "Business Intelligence", component: "Analytics Platform", level: "Partial", notes: "Analytics data source" },
    { capability: "Customer Analytics", component: "Analytics Platform", level: "Full", notes: "Customer behavior analytics" },
    { capability: "Predictive Analytics", component: "Recommendation Engine", level: "Partial", notes: "ML-based prediction engine" },
    { capability: "API and Integration Management", component: "API Gateway", level: "Full", notes: "API routing, rate limiting, and security" },
    { capability: "Observability and Monitoring", component: "Analytics Platform", level: "Partial", notes: "System metrics aggregation" },
    { capability: "Content Marketing", component: "Content Management", level: "Full", notes: "Content authoring and publishing" },
    { capability: "Pricing and Promotions", component: "Pricing Service", level: "Full", notes: "Pricing rules and discount engine" },
  ];

  for (const r of realizations) {
    const capability = capabilities.get(r.capability);
    const componentId = components.get(r.component);
    if (!capability || !componentId) continue;
    try {
      await apiCall("POST", `/capabilities/${capability.id}/systems`, {
        componentId,
        realizationLevel: r.level,
        notes: r.notes,
      });
    } catch {
      console.log(`    (Skipping realization — may already exist)`);
    }
  }
}
