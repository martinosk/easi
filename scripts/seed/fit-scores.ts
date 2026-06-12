import { apiCall, apiCallWithEtag, parallelBatch, API_URL, buildHeaders } from "./config.ts";
import type { StrategyPillar } from "./types.ts";

interface StrategyPillarsResponse {
  data: StrategyPillar[];
}

async function fetchPillarsWithEtag(): Promise<{ pillars: StrategyPillar[]; etag: string }> {
  const url = `${API_URL}/meta-model/strategy-pillars?includeInactive=false`;
  const response = await fetch(url, { method: "GET", headers: buildHeaders() });
  if (!response.ok) throw new Error(`Failed to fetch strategy pillars: ${response.status}`);
  const etag = response.headers.get("etag") || '"0"';
  const data: StrategyPillarsResponse = await response.json();
  return { pillars: data.data || [], etag };
}

interface PillarChange {
  operation: string;
  id?: string;
  name?: string;
  description?: string;
  fitScoringEnabled?: boolean;
  fitCriteria?: string;
}

async function batchUpdatePillars(changes: PillarChange[], etag: string): Promise<string> {
  const url = `${API_URL}/meta-model/strategy-pillars`;
  return apiCallWithEtag("PATCH", url, { changes }, etag, "Failed to batch update pillars");
}

async function batchUpdatePillarsWithRetry(changes: PillarChange[], maxRetries = 5): Promise<void> {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    const { etag } = await fetchPillarsWithEtag();
    try {
      await batchUpdatePillars(changes, etag);
      return;
    } catch (e) {
      if (attempt === maxRetries) throw e;
      await new Promise((resolve) => setTimeout(resolve, 100 * attempt));
    }
  }
}

interface PillarDef {
  name: string;
  description: string;
  fitCriteria: string;
}

const PILLAR_DEFS: PillarDef[] = [
  {
    name: "Cloud Native",
    description: "Embrace cloud-native technologies and patterns",
    fitCriteria: "Containerization, Kubernetes orchestration, auto-scaling, CI/CD pipelines",
  },
  {
    name: "API First",
    description: "Design and build APIs as first-class products",
    fitCriteria: "OpenAPI documentation, versioned APIs, RESTful design, developer portal",
  },
  {
    name: "Security",
    description: "Security-first approach to all systems",
    fitCriteria: "Authentication, authorization, encryption, audit logging, vulnerability scanning",
  },
];

async function ensurePillarsExist(existingNames: Set<string>): Promise<void> {
  for (const def of PILLAR_DEFS) {
    if (existingNames.has(def.name)) continue;
    try {
      const url = `${API_URL}/meta-model/strategy-pillars`;
      const response = await fetch(url, {
        method: "POST",
        headers: buildHeaders(),
        body: JSON.stringify({ name: def.name, description: def.description }),
      });
      if (!response.ok) {
        const text = await response.text();
        throw new Error(`Failed to create pillar: ${response.status}: ${text}`);
      }
    } catch {
      console.log(`    (Pillar ${def.name} may already exist)`);
    }
  }
}

async function enableFitScoring(pillars: StrategyPillar[]): Promise<Map<string, StrategyPillar>> {
  const pillarMap = new Map(pillars.map((p) => [p.name, p]));
  const changes: PillarChange[] = PILLAR_DEFS.filter((def) => {
    const pillar = pillarMap.get(def.name);
    return pillar && !pillar.fitScoringEnabled;
  }).map((def) => {
    const pillar = pillarMap.get(def.name)!;
    return {
      operation: "update",
      id: pillar.id,
      name: pillar.name,
      description: pillar.description,
      fitScoringEnabled: true,
      fitCriteria: def.fitCriteria,
    };
  });

  if (changes.length > 0) {
    try {
      await batchUpdatePillarsWithRetry(changes);
    } catch {
      console.log("    (Failed to enable fit scoring on some pillars)");
    }
  }

  const { pillars: finalPillars } = await fetchPillarsWithEtag();
  return new Map(finalPillars.filter((p) => p.fitScoringEnabled).map((p) => [p.name, p]));
}

interface FitScore {
  pillarName: string;
  score: number;
  rationale: string;
}

interface ComponentFitData {
  componentName: string;
  scores: FitScore[];
}

const FIT_SCORE_DATA: ComponentFitData[] = [
  {
    componentName: "User Service",
    scores: [
      { pillarName: "Cloud Native", score: 4, rationale: "Fully containerized with Kubernetes orchestration" },
      { pillarName: "API First", score: 5, rationale: "Well-documented REST APIs with OpenAPI specs" },
      { pillarName: "Security", score: 4, rationale: "OAuth2/OIDC implementation with regular audits" },
    ],
  },
  {
    componentName: "Order Service",
    scores: [
      { pillarName: "Cloud Native", score: 3, rationale: "Containerized but with some legacy dependencies" },
      { pillarName: "API First", score: 4, rationale: "REST API with documentation, minor inconsistencies" },
      { pillarName: "Security", score: 3, rationale: "Basic authentication, needs improved audit logging" },
    ],
  },
  {
    componentName: "Payment Gateway",
    scores: [
      { pillarName: "Cloud Native", score: 4, rationale: "Cloud-hosted with auto-scaling capabilities" },
      { pillarName: "API First", score: 5, rationale: "Industry-standard payment APIs with full documentation" },
      { pillarName: "Security", score: 5, rationale: "PCI-DSS compliant, encryption at rest and in transit" },
    ],
  },
  {
    componentName: "Inventory Service",
    scores: [
      { pillarName: "Cloud Native", score: 2, rationale: "Still running on VMs with manual scaling" },
      { pillarName: "API First", score: 3, rationale: "API exists but lacks proper versioning" },
      { pillarName: "Security", score: 3, rationale: "Basic access controls, needs improvement" },
    ],
  },
  {
    componentName: "Analytics Platform",
    scores: [
      { pillarName: "Cloud Native", score: 5, rationale: "Fully serverless architecture" },
      { pillarName: "API First", score: 4, rationale: "GraphQL and REST APIs available" },
      { pillarName: "Security", score: 4, rationale: "Role-based access with data encryption" },
    ],
  },
  {
    componentName: "API Gateway",
    scores: [
      { pillarName: "Cloud Native", score: 5, rationale: "Cloud-native gateway with auto-scaling and geo-distribution" },
      { pillarName: "API First", score: 5, rationale: "Central API management with full OpenAPI support" },
      { pillarName: "Security", score: 4, rationale: "JWT validation, rate limiting, and WAF integration" },
    ],
  },
  {
    componentName: "Search Engine",
    scores: [
      { pillarName: "Cloud Native", score: 4, rationale: "Elasticsearch cluster running on Kubernetes" },
      { pillarName: "API First", score: 4, rationale: "Standard search REST APIs" },
      { pillarName: "Security", score: 3, rationale: "API key authentication only, needs RBAC improvement" },
    ],
  },
  {
    componentName: "Notification Service",
    scores: [
      { pillarName: "Cloud Native", score: 5, rationale: "Event-driven serverless notification functions" },
      { pillarName: "API First", score: 3, rationale: "Internal APIs with limited external documentation" },
      { pillarName: "Security", score: 4, rationale: "Encrypted queues and secure message handling" },
    ],
  },
  {
    componentName: "Admin Dashboard",
    scores: [
      { pillarName: "Cloud Native", score: 2, rationale: "Monolithic deployment with manual updates required" },
      { pillarName: "API First", score: 2, rationale: "Primarily server-rendered with limited API exposure" },
      { pillarName: "Security", score: 3, rationale: "Basic RBAC, needs MFA and audit log improvements" },
    ],
  },
  {
    componentName: "Fraud Detection",
    scores: [
      { pillarName: "Cloud Native", score: 4, rationale: "Containerized ML inference with horizontal scaling" },
      { pillarName: "API First", score: 4, rationale: "Well-defined decision API with clear contracts" },
      { pillarName: "Security", score: 5, rationale: "By design security-first, encrypted data and audit trails" },
    ],
  },
];

export async function seedApplicationFitScores(components: Map<string, string>): Promise<void> {
  console.log("\n📊 Seeding Application Fit Scores...");

  const { pillars: existingPillars } = await fetchPillarsWithEtag();
  const existingNames = new Set(existingPillars.map((p) => p.name));
  await ensurePillarsExist(existingNames);

  const { pillars: allPillars } = await fetchPillarsWithEtag();
  const enabledPillars = await enableFitScoring(allPillars);

  if (enabledPillars.size === 0) {
    console.log("  No pillars with fit scoring enabled, skipping");
    return;
  }
  console.log(`  Setting fit scores against ${enabledPillars.size} pillars`);

  await parallelBatch(FIT_SCORE_DATA, 3, async (data) => {
    const componentId = components.get(data.componentName);
    if (!componentId) return;
    for (const scoreData of data.scores) {
      const pillar = enabledPillars.get(scoreData.pillarName);
      if (!pillar) continue;
      try {
        await apiCall("PUT", `/components/${componentId}/fit-scores/${pillar.id}`, {
          score: scoreData.score,
          rationale: scoreData.rationale,
        });
      } catch {
        console.log(`    (Skipping fit score — may already exist)`);
      }
    }
  });
}
