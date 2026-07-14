import { apiCall, parallelBatch } from "./config.ts";
import { buildRealizationPlan, type PlannedRealization } from "./realization-plan.ts";
import type { BusinessDomain, Capability } from "./types.ts";

type Grade = "Invest" | "Tolerate" | "Migrate" | "Eliminate";
type Role = "standard" | "legacy";
type RealizationLevel = "Full" | "Partial" | "Planned";
type JourneyKind = "migration" | "consolidation" | "carve-out" | "move";
type AdvanceTo = "planned" | "in-flight" | "done" | "abandoned";
type MilestoneStatus = "planned" | "in-flight" | "done";

interface TargetPeriod {
  year: number;
  quarter: number;
}

interface AppInvolvement {
  app: string;
  side: "from" | "to";
  level: RealizationLevel;
  role?: Role;
  grade?: Grade;
}

interface SubCapability {
  capability: string;
  realizations: { app: string; level: RealizationLevel }[];
}

interface MilestoneSpec {
  label: string;
  status: MilestoneStatus;
  targetPeriod: TargetPeriod;
}

interface JourneyScenario {
  capability: string;
  kind: JourneyKind;
  note: string;
  apps: AppInvolvement[];
  advanceTo: AdvanceTo;
  progress?: number;
  targetPeriod?: TargetPeriod;
  milestones?: MilestoneSpec[];
  move?: { targetDomain: string; resultingName: string };
  subCapabilities?: SubCapability[];
}

export interface ArchitectureDirectionContext {
  capabilities: Map<string, Capability>;
  components: Map<string, string>;
  domains: Map<string, BusinessDomain>;
}

const GRADES: Grade[] = ["Invest", "Tolerate", "Migrate", "Eliminate"];

const SCENARIOS: JourneyScenario[] = [
  {
    capability: "Order Management - Execution and Operations",
    kind: "migration",
    note: "Migrating order orchestration off the legacy stack onto the unified commerce platform, route by route.",
    advanceTo: "in-flight",
    progress: 60,
    targetPeriod: { year: 2027, quarter: 2 },
    apps: [
      { app: "Order Service", side: "from", level: "Full", role: "legacy", grade: "Migrate" },
      { app: "Order Orchestrator", side: "from", level: "Full", role: "legacy", grade: "Eliminate" },
      { app: "Digital Commerce Platform", side: "to", level: "Planned", role: "standard", grade: "Invest" },
    ],
    milestones: [
      { label: "Pilot routes migrated", status: "done", targetPeriod: { year: 2026, quarter: 2 } },
      { label: "Bulk cutover", status: "in-flight", targetPeriod: { year: 2027, quarter: 1 } },
      { label: "Decommission legacy order stack", status: "planned", targetPeriod: { year: 2027, quarter: 3 } },
    ],
    subCapabilities: [
      {
        capability: "Order Management - Execution and Operations - Process Management",
        realizations: [{ app: "Digital Commerce Platform", level: "Full" }],
      },
      {
        capability: "Order Management - Execution and Operations - Performance Optimization",
        realizations: [
          { app: "Order Service", level: "Full" },
          { app: "Digital Commerce Platform", level: "Planned" },
        ],
      },
    ],
  },
  {
    capability: "Payment Operations - Execution and Operations",
    kind: "consolidation",
    note: "Consolidated three payment tools onto a single orchestrator, completed 2025.",
    advanceTo: "done",
    progress: 100,
    targetPeriod: { year: 2025, quarter: 4 },
    apps: [
      { app: "Payment Gateway", side: "from", level: "Full", role: "legacy", grade: "Migrate" },
      { app: "Settlement Service", side: "from", level: "Full", role: "legacy", grade: "Tolerate" },
      { app: "Chargeback Handler", side: "from", level: "Full", role: "legacy", grade: "Eliminate" },
      { app: "Payment Orchestrator", side: "to", level: "Full", role: "standard", grade: "Invest" },
    ],
  },
  {
    capability: "Customer Support - Execution and Operations",
    kind: "migration",
    note: "Planned migration from the ticketing system onto the omnichannel contact centre.",
    advanceTo: "planned",
    targetPeriod: { year: 2027, quarter: 4 },
    apps: [
      { app: "Support Ticket System", side: "from", level: "Full", role: "legacy", grade: "Tolerate" },
      { app: "Contact Center Hub", side: "to", level: "Planned", role: "standard", grade: "Invest" },
    ],
  },
  {
    capability: "Revenue Management - Execution and Operations",
    kind: "move",
    note: "Relocating revenue operations into Sales as a dedicated billing capability.",
    advanceTo: "planned",
    targetPeriod: { year: 2027, quarter: 1 },
    move: { targetDomain: "Sales", resultingName: "Billing & Revenue Operations" },
    apps: [
      { app: "Revenue Recognition Engine", side: "from", level: "Full", role: "legacy", grade: "Tolerate" },
      { app: "Billing Platform", side: "to", level: "Planned", role: "standard", grade: "Invest" },
    ],
  },
  {
    capability: "Financial Reporting - Governance and Compliance",
    kind: "carve-out",
    note: "Carving regulatory reporting out of the general ledger onto a dedicated compliance engine.",
    advanceTo: "planned",
    targetPeriod: { year: 2028, quarter: 1 },
    apps: [
      { app: "General Ledger Service", side: "from", level: "Full", role: "legacy", grade: "Tolerate" },
      { app: "Compliance Reporting Engine", side: "to", level: "Planned", role: "standard", grade: "Invest" },
    ],
  },
  {
    capability: "Brand Management - Strategy and Planning",
    kind: "migration",
    note: "Abandoned migration — the asset manager rollout was shelved (kept as history).",
    advanceTo: "abandoned",
    apps: [
      { app: "Content Management", side: "from", level: "Full", grade: "Migrate" },
      { app: "Digital Asset Manager", side: "to", level: "Planned", grade: "Tolerate" },
    ],
  },
];

interface RealizationTarget {
  capabilityId: string;
  componentId: string;
}

async function createRealization(target: RealizationTarget, level: RealizationLevel): Promise<void> {
  try {
    await apiCall("POST", `/capabilities/${target.capabilityId}/systems`, {
      componentId: target.componentId,
      realizationLevel: level,
      notes: "Seeded for capability journeys",
    });
  } catch {
    // A direct realization for this pair may already exist from the base seed.
  }
}

async function assessRealization(target: RealizationTarget, grade: Grade): Promise<void> {
  await apiCall("PUT", `/capabilities/${target.capabilityId}/components/${target.componentId}/time-assessment`, {
    grade,
    rationale: `Seeded ${grade} assessment`,
  });
}

async function assignRole(target: RealizationTarget, role: Role): Promise<void> {
  await apiCall("PUT", `/capabilities/${target.capabilityId}/components/${target.componentId}/realization-role`, { role });
}

async function applyInvolvement(target: RealizationTarget, app: AppInvolvement): Promise<void> {
  await createRealization(target, app.level);
  if (app.grade) await assessRealization(target, app.grade);
  if (app.role) await assignRole(target, app.role);
}

async function seedScenarioApps(capabilityId: string, ctx: ArchitectureDirectionContext, apps: AppInvolvement[]): Promise<void> {
  for (const app of apps) {
    const componentId = ctx.components.get(app.app);
    if (componentId) await applyInvolvement({ capabilityId, componentId }, app);
  }
}

async function seedSubCapabilityRealizations(
  capabilityId: string,
  ctx: ArchitectureDirectionContext,
  realizations: SubCapability["realizations"]
): Promise<void> {
  for (const realization of realizations) {
    const componentId = ctx.components.get(realization.app);
    if (componentId) await createRealization({ capabilityId, componentId }, realization.level);
  }
}

async function seedSubCapabilities(ctx: ArchitectureDirectionContext, subs: SubCapability[] = []): Promise<void> {
  for (const sub of subs) {
    const capability = ctx.capabilities.get(sub.capability);
    if (capability) await seedSubCapabilityRealizations(capability.id, ctx, sub.realizations);
  }
}

interface CapturePayload {
  kind: JourneyKind;
  fromComponentIds: string[];
  toComponentId: string;
  note: string;
  targetPeriod?: TargetPeriod;
  targetDomainId?: string;
  resultingName?: string;
}

function buildCapturePayload(scenario: JourneyScenario, ctx: ArchitectureDirectionContext): CapturePayload | null {
  const fromComponentIds = scenario.apps
    .filter((app) => app.side === "from")
    .map((app) => ctx.components.get(app.app))
    .filter((id): id is string => Boolean(id));
  const toComponentId = ctx.components.get(scenario.apps.find((app) => app.side === "to")?.app ?? "");
  if (!toComponentId) return null;

  const payload: CapturePayload = { kind: scenario.kind, fromComponentIds, toComponentId, note: scenario.note };
  if (scenario.targetPeriod) payload.targetPeriod = scenario.targetPeriod;
  if (scenario.move) {
    const domain = ctx.domains.get(scenario.move.targetDomain);
    if (!domain) return null;
    payload.targetDomainId = domain.id;
    payload.resultingName = scenario.move.resultingName;
  }
  return payload;
}

async function captureJourney(capabilityId: string, payload: CapturePayload): Promise<string> {
  const journey = await apiCall<{ id: string }>("POST", `/capabilities/${capabilityId}/journey`, payload);
  return journey.id;
}

async function transition(journeyId: string, action: "start" | "complete" | "abandon"): Promise<void> {
  await apiCall("POST", `/capability-journeys/${journeyId}/${action}`);
}

async function updateProgress(journeyId: string, progress: number): Promise<void> {
  await apiCall("PUT", `/capability-journeys/${journeyId}/progress`, { progress });
}

async function addMilestone(journeyId: string, milestone: MilestoneSpec): Promise<void> {
  await apiCall("POST", `/capability-journeys/${journeyId}/milestones`, {
    label: milestone.label,
    status: milestone.status,
    targetPeriod: milestone.targetPeriod,
  });
}

async function advanceJourney(journeyId: string, scenario: JourneyScenario): Promise<void> {
  if (scenario.advanceTo === "planned") return;
  if (scenario.advanceTo === "abandoned") {
    await transition(journeyId, "abandon");
    return;
  }
  await transition(journeyId, "start");
  for (const milestone of scenario.milestones ?? []) await addMilestone(journeyId, milestone);
  if (scenario.progress !== undefined) await updateProgress(journeyId, scenario.progress);
  if (scenario.advanceTo === "done") await transition(journeyId, "complete");
}

async function seedScenario(ctx: ArchitectureDirectionContext, scenario: JourneyScenario): Promise<boolean> {
  const capability = ctx.capabilities.get(scenario.capability);
  if (!capability) {
    console.log(`    (Skipping journey — capability "${scenario.capability}" not found)`);
    return false;
  }
  await seedScenarioApps(capability.id, ctx, scenario.apps);
  await seedSubCapabilities(ctx, scenario.subCapabilities);

  const payload = buildCapturePayload(scenario, ctx);
  if (!payload) {
    console.log(`    (Skipping journey on "${scenario.capability}" — target app or domain unresolved)`);
    return false;
  }
  const journeyId = await captureJourney(capability.id, payload);
  await advanceJourney(journeyId, scenario);
  return true;
}

async function seedJourneys(ctx: ArchitectureDirectionContext): Promise<void> {
  console.log("  Creating capability journeys...");
  let created = 0;
  for (const scenario of SCENARIOS) {
    try {
      if (await seedScenario(ctx, scenario)) created++;
    } catch (error) {
      console.log(`    (Journey on "${scenario.capability}" failed: ${error})`);
    }
  }
  console.log(`  ✓ ${created}/${SCENARIOS.length} capability journeys created`);
}

interface GradeTask extends RealizationTarget {
  grade: Grade;
}

function buildGradeTasks(plan: PlannedRealization[], ctx: ArchitectureDirectionContext): GradeTask[] {
  const tasks: GradeTask[] = [];
  plan.forEach((item, i) => {
    if (i % 2 !== 0) return;
    const capability = ctx.capabilities.get(item.capabilityName);
    const componentId = ctx.components.get(item.componentName);
    if (capability && componentId) tasks.push({ capabilityId: capability.id, componentId, grade: GRADES[i % GRADES.length] });
  });
  return tasks;
}

function componentsByCapability(
  plan: PlannedRealization[],
  ctx: ArchitectureDirectionContext
): Map<string, { capabilityId: string; componentIds: string[] }> {
  const map = new Map<string, { capabilityId: string; componentIds: string[] }>();
  for (const item of plan) {
    const capability = ctx.capabilities.get(item.capabilityName);
    const componentId = ctx.components.get(item.componentName);
    if (!capability || !componentId) continue;
    const entry = map.get(capability.id) ?? { capabilityId: capability.id, componentIds: [] };
    if (!entry.componentIds.includes(componentId)) entry.componentIds.push(componentId);
    map.set(capability.id, entry);
  }
  return map;
}

interface CapabilityRolePlan {
  standard: RealizationTarget;
  legacy: RealizationTarget;
}

function buildRolePlans(plan: PlannedRealization[], ctx: ArchitectureDirectionContext): CapabilityRolePlan[] {
  const plans: CapabilityRolePlan[] = [];
  for (const { capabilityId, componentIds } of componentsByCapability(plan, ctx).values()) {
    if (componentIds.length < 2) continue;
    plans.push({
      standard: { capabilityId, componentId: componentIds[0] },
      legacy: { capabilityId, componentId: componentIds[1] },
    });
  }
  return plans;
}

// Roles live in one aggregate per capability, so the standard and legacy writes
// for a capability must run sequentially; only distinct capabilities go in parallel.
async function applyRolePlan(rolePlan: CapabilityRolePlan): Promise<void> {
  await assignRole(rolePlan.standard, "standard").catch(() => undefined);
  await assignRole(rolePlan.legacy, "legacy").catch(() => undefined);
}

async function seedBoardAssessmentsAndRoles(ctx: ArchitectureDirectionContext): Promise<void> {
  console.log("  Assessing TIME grades and realization roles across the board...");
  const plan = buildRealizationPlan([...ctx.components.keys()]);
  const gradeTasks = buildGradeTasks(plan, ctx);
  const rolePlans = buildRolePlans(plan, ctx);

  await parallelBatch(gradeTasks, 8, (task) => assessRealization(task, task.grade).catch(() => undefined));
  await parallelBatch(rolePlans, 8, (rolePlan) => applyRolePlan(rolePlan));
  console.log(`  ✓ ${gradeTasks.length} TIME assessments, ${rolePlans.length * 2} realization roles`);
}

export async function seedArchitectureDirection(ctx: ArchitectureDirectionContext): Promise<void> {
  console.log("\n🧭 Seeding Architecture Direction (TIME, roles, journeys — specs 180–183)...");
  await seedBoardAssessmentsAndRoles(ctx);
  await seedJourneys(ctx);
}
