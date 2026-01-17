#!/usr/bin/env npx ts-node

/**
 * Test Data Seeding Script
 *
 * This script populates the test database with realistic enterprise architecture data
 * by calling the backend API.
 *
 * AUTHENTICATION MODES:
 *
 * 1. Session Cookie (local dev with DEX) - RECOMMENDED for local dev:
 *    - Log in to the app at http://localhost:3000 with testuser@acme.com
 *    - Open DevTools > Application > Cookies > copy "easi_session" value
 *    - Run: npm run seed -- --cookie "your_session_cookie_value"
 *
 * 2. Bypass Mode (CI/testing):
 *    - Start backend with AUTH_MODE=bypass
 *    - Run: npm run seed -- --bypass --tenant-id acme
 *
 * Usage:
 *   npm run seed -- --cookie "session_cookie_value"
 *   npm run seed -- --bypass --tenant-id acme
 *   npm run seed -- --bypass --base-url http://localhost:8080
 *
 * Prerequisites:
 *   - Backend running (with DEX for cookie mode, or AUTH_MODE=bypass)
 *   - Test tenant "acme" exists (created by migration 041)
 */

function getArg(flag: string): string | undefined {
  const idx = process.argv.indexOf(flag);
  if (idx !== -1 && process.argv[idx + 1]) {
    return process.argv[idx + 1];
  }
  return undefined;
}

const BASE_URL = getArg("--base-url") ?? "http://localhost:8080";
const TENANT_ID = getArg("--tenant-id") ?? "acme";
const SESSION_COOKIE = getArg("--cookie");
const BYPASS_MODE = process.argv.includes("--bypass");

if (!SESSION_COOKIE && !BYPASS_MODE) {
  console.log(`
Usage: npm run seed -- [options]

Authentication (choose one):
  --cookie <value>    Use session cookie from browser (local dev with DEX)
  --bypass            Use X-Tenant-ID header (requires AUTH_MODE=bypass on backend)

Options:
  --base-url <url>    Backend URL (default: http://localhost:8080)
  --tenant-id <id>    Tenant ID for bypass mode (default: acme)

Examples:
  # Local dev with DEX (get cookie from browser DevTools):
  npm run seed -- --cookie "MTcz..."

  # CI/testing with bypass mode:
  npm run seed -- --bypass --tenant-id acme

To get session cookie:
  1. Open http://localhost:3000 and log in with testuser@acme.com / password
  2. Open DevTools (F12) > Application > Cookies > localhost
  3. Copy the value of "easi_session"
`);
  process.exit(0);
}

const API_URL = `${BASE_URL}/api/v1`;

interface Component {
  id: string;
  name: string;
  description: string;
}

interface Capability {
  id: string;
  name: string;
  description: string;
  level: string;
  parentId?: string;
}

interface BusinessDomain {
  id: string;
  name: string;
  description: string;
}

interface EnterpriseCapability {
  id: string;
  name: string;
  description: string;
  category: string;
}

interface View {
  id: string;
  name: string;
  description: string;
}

interface StrategyPillar {
  id: string;
  name: string;
  description: string;
  active: boolean;
  fitScoringEnabled: boolean;
}

interface AcquiredEntity {
  id: string;
  name: string;
  acquisitionDate?: string;
  integrationStatus: string;
  notes?: string;
}

interface Vendor {
  id: string;
  name: string;
  implementationPartner?: string;
  notes?: string;
}

interface InternalTeam {
  id: string;
  name: string;
  department?: string;
  contactPerson?: string;
  notes?: string;
}

function buildHeaders(): Record<string, string> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  if (SESSION_COOKIE) {
    headers["Cookie"] = `easi_session=${SESSION_COOKIE}`;
  } else if (BYPASS_MODE) {
    headers["X-Tenant-ID"] = TENANT_ID;
  }

  return headers;
}

async function apiCall<T>(
  method: string,
  path: string,
  body?: unknown
): Promise<T> {
  const url = `${API_URL}${path}`;
  const options: RequestInit = {
    method,
    headers: buildHeaders(),
  };

  if (body) {
    options.body = JSON.stringify(body);
  }

  const response = await fetch(url, options);

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`API call failed: ${method} ${path} - ${response.status}: ${text}`);
  }

  if (response.status === 204 || response.status === 201) {
    const text = await response.text();
    if (!text) {
      return {} as T;
    }
    try {
      return JSON.parse(text);
    } catch {
      return {} as T;
    }
  }

  return response.json();
}

async function createComponent(name: string, description: string): Promise<Component> {
  console.log(`  Creating component: ${name}`);
  return apiCall<Component>("POST", "/components", { name, description });
}

interface RelationParams {
  name: string;
  description: string;
  sourceId: string;
  targetId: string;
  relationType: string;
}

async function createRelation(params: RelationParams): Promise<void> {
  console.log(`  Creating relation: ${params.name}`);
  await apiCall("POST", "/relations", {
    name: params.name,
    description: params.description,
    sourceComponentId: params.sourceId,
    targetComponentId: params.targetId,
    relationType: params.relationType,
  });
}

async function createCapability(
  name: string,
  description: string,
  level: string,
  parentId?: string
): Promise<Capability> {
  console.log(`  Creating capability: ${name}`);
  return apiCall<Capability>("POST", "/capabilities", {
    name,
    description,
    level,
    parentId,
  });
}

async function updateCapabilityMetadata(
  capabilityId: string,
  metadata: {
    status?: string;
    ownershipModel?: string;
    maturityValue?: number;
    primaryOwner?: string;
  }
): Promise<void> {
  console.log(`  Updating capability metadata: ${capabilityId}`);
  await apiCall("PUT", `/capabilities/${capabilityId}/metadata`, metadata);
}

async function createBusinessDomain(
  name: string,
  description: string
): Promise<BusinessDomain> {
  console.log(`  Creating business domain: ${name}`);
  return apiCall<BusinessDomain>("POST", "/business-domains", { name, description });
}

async function assignCapabilityToDomain(
  domainId: string,
  capabilityId: string
): Promise<void> {
  console.log(`  Assigning capability to domain`);
  await apiCall("POST", `/business-domains/${domainId}/capabilities`, { capabilityId });
}

async function createEnterpriseCapability(
  name: string,
  description: string,
  category: string
): Promise<EnterpriseCapability> {
  console.log(`  Creating enterprise capability: ${name}`);
  return apiCall<EnterpriseCapability>("POST", "/enterprise-capabilities", {
    name,
    description,
    category,
  });
}

async function linkEnterpriseCapability(
  enterpriseCapabilityId: string,
  capabilityId: string
): Promise<void> {
  console.log(`  Linking enterprise capability to domain capability`);
  await apiCall("POST", `/enterprise-capabilities/${enterpriseCapabilityId}/links`, {
    capabilityId,
  });
}

async function createView(name: string, description: string): Promise<View> {
  console.log(`  Creating view: ${name}`);
  return apiCall<View>("POST", "/views", { name, description });
}

async function addComponentToView(
  viewId: string,
  componentId: string,
  x: number,
  y: number
): Promise<void> {
  await apiCall("POST", `/views/${viewId}/components`, { componentId, x, y });
}

async function linkSystemToCapability(
  capabilityId: string,
  componentId: string,
  realizationLevel: string,
  notes: string
): Promise<void> {
  console.log(`  Linking system to capability`);
  await apiCall("POST", `/capabilities/${capabilityId}/systems`, {
    componentId,
    realizationLevel,
    notes,
  });
}

async function createDependency(
  sourceCapabilityId: string,
  targetCapabilityId: string,
  dependencyType: string,
  description: string
): Promise<void> {
  console.log(`  Creating capability dependency`);
  await apiCall("POST", "/capability-dependencies", {
    sourceCapabilityId,
    targetCapabilityId,
    dependencyType,
    description,
  });
}

interface StrategyPillarsResponse {
  data: StrategyPillar[];
  etag?: string;
}

async function getStrategyPillarsWithETag(): Promise<{ pillars: StrategyPillar[]; etag: string }> {
  console.log(`  Fetching strategy pillars`);
  const url = `${API_URL}/meta-model/strategy-pillars?includeInactive=false`;
  const response = await fetch(url, {
    method: "GET",
    headers: buildHeaders(),
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch strategy pillars: ${response.status}`);
  }

  const etag = response.headers.get("etag") || '"0"';
  const data = await response.json();
  return { pillars: data.data || [], etag };
}

async function getStrategyPillars(): Promise<StrategyPillar[]> {
  const result = await getStrategyPillarsWithETag();
  return result.pillars;
}

async function apiCallWithEtag(
  method: "PUT" | "PATCH",
  url: string,
  body: unknown,
  etag: string,
  errorContext: string
): Promise<string> {
  const headers = buildHeaders();
  headers["If-Match"] = etag;

  const response = await fetch(url, { method, headers, body: JSON.stringify(body) });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(`${errorContext}: ${response.status}: ${text}`);
  }
  return response.headers.get("etag") || etag;
}

async function enableFitScoringOnPillar(pillarId: string, fitCriteria: string, etag: string): Promise<string> {
  console.log(`  Enabling fit scoring on pillar`);
  const url = `${API_URL}/meta-model/strategy-pillars/${pillarId}/fit-configuration`;
  return apiCallWithEtag("PUT", url, { fitScoringEnabled: true, fitCriteria }, etag, "Failed to enable fit scoring");
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
  console.log(`  Batch updating ${changes.length} pillars`);
  const url = `${API_URL}/meta-model/strategy-pillars`;
  return apiCallWithEtag("PATCH", url, { changes }, etag, "Failed to batch update pillars");
}

async function createStrategyPillar(
  name: string,
  description: string
): Promise<{ pillar: StrategyPillar; etag: string }> {
  console.log(`  Creating strategy pillar: ${name}`);
  const url = `${API_URL}/meta-model/strategy-pillars`;
  const response = await fetch(url, {
    method: "POST",
    headers: buildHeaders(),
    body: JSON.stringify({ name, description }),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`Failed to create pillar: ${response.status}: ${text}`);
  }

  const etag = response.headers.get("etag") || '"0"';
  const pillar = await response.json();
  return { pillar, etag };
}

async function setApplicationFitScore(
  componentId: string,
  pillarId: string,
  score: number,
  rationale: string
): Promise<void> {
  console.log(`  Setting fit score for component`);
  await apiCall("PUT", `/components/${componentId}/fit-scores/${pillarId}`, {
    score,
    rationale,
  });
}

async function createAcquiredEntity(
  name: string,
  acquisitionDate: string,
  integrationStatus: string,
  notes: string
): Promise<AcquiredEntity> {
  console.log(`  Creating acquired entity: ${name}`);
  return apiCall<AcquiredEntity>("POST", "/acquired-entities", {
    name,
    acquisitionDate,
    integrationStatus,
    notes,
  });
}

async function createVendor(
  name: string,
  implementationPartner: string,
  notes: string
): Promise<Vendor> {
  console.log(`  Creating vendor: ${name}`);
  return apiCall<Vendor>("POST", "/vendors", {
    name,
    implementationPartner,
    notes,
  });
}

async function createInternalTeam(
  name: string,
  department: string,
  contactPerson: string,
  notes: string
): Promise<InternalTeam> {
  console.log(`  Creating internal team: ${name}`);
  return apiCall<InternalTeam>("POST", "/internal-teams", {
    name,
    department,
    contactPerson,
    notes,
  });
}

async function linkComponentToAcquiredEntity(acquiredEntityId: string, componentId: string, notes: string): Promise<void> {
  console.log(`  Linking component to acquired entity`);
  await apiCall("POST", `/components/${componentId}/origin/acquired-via`, { acquiredEntityId, notes });
}

async function linkComponentToVendor(vendorId: string, componentId: string, notes: string): Promise<void> {
  console.log(`  Linking component to vendor`);
  await apiCall("POST", `/components/${componentId}/origin/purchased-from`, { vendorId, notes });
}

async function linkComponentToInternalTeam(internalTeamId: string, componentId: string, notes: string): Promise<void> {
  console.log(`  Linking component to internal team`);
  await apiCall("POST", `/components/${componentId}/origin/built-by`, { internalTeamId, notes });
}

async function seedComponents(): Promise<Map<string, Component>> {
  console.log("\n📦 Seeding Application Components...");
  const components = new Map<string, Component>();

  const componentData = [
    { name: "User Service", description: "Handles user authentication, authorization, and profile management" },
    { name: "Order Service", description: "Manages customer orders, order lifecycle, and order history" },
    { name: "Payment Gateway", description: "Processes payments, refunds, and integrates with payment providers" },
    { name: "Inventory Service", description: "Tracks product inventory levels and availability across warehouses" },
    { name: "Notification Service", description: "Sends emails, SMS, and push notifications to customers" },
    { name: "Product Catalog", description: "Manages product information, categories, and search functionality" },
    { name: "Shopping Cart", description: "Maintains shopping cart state and checkout process" },
    { name: "Analytics Platform", description: "Collects and processes business analytics and metrics" },
    { name: "Customer Portal", description: "Self-service portal for customer account management" },
    { name: "Admin Dashboard", description: "Internal tool for operations and support teams" },
    { name: "API Gateway", description: "Central entry point for all API requests with rate limiting and routing" },
    { name: "Message Queue", description: "Asynchronous message processing for event-driven architecture" },
    { name: "Cache Layer", description: "Redis-based caching for improved performance" },
    { name: "Search Engine", description: "Elasticsearch-powered full-text search capabilities" },
    { name: "Recommendation Engine", description: "ML-based product recommendations for customers" },
    { name: "Shipping Service", description: "Calculates shipping rates and tracks deliveries" },
    { name: "Pricing Service", description: "Dynamic pricing, discounts, and promotional rules engine" },
    { name: "Fraud Detection", description: "Real-time fraud detection for payment transactions" },
    { name: "Content Management", description: "Manages website content, banners, and marketing materials" },
    { name: "Reporting Service", description: "Generates business reports and data exports" },
  ];

  for (const c of componentData) {
    const component = await createComponent(c.name, c.description);
    components.set(c.name, component);
  }

  return components;
}

async function seedRelations(components: Map<string, Component>): Promise<void> {
  console.log("\n🔗 Seeding Component Relations...");

  const relations = [
    { name: "User Authentication", source: "API Gateway", target: "User Service", type: "Triggers", description: "Authenticates API requests" },
    { name: "Order Processing", source: "Order Service", target: "Payment Gateway", type: "Triggers", description: "Processes order payments" },
    { name: "Inventory Check", source: "Order Service", target: "Inventory Service", type: "Triggers", description: "Validates inventory availability" },
    { name: "Order Notifications", source: "Order Service", target: "Notification Service", type: "Triggers", description: "Publishes order events for notifications" },
    { name: "Cart Checkout", source: "Shopping Cart", target: "Order Service", type: "Triggers", description: "Creates orders from cart" },
    { name: "Product Search", source: "Search Engine", target: "Product Catalog", type: "Serves", description: "Indexes and searches products" },
    { name: "Analytics Events", source: "API Gateway", target: "Analytics Platform", type: "Triggers", description: "Sends analytics events" },
    { name: "Cache Products", source: "Cache Layer", target: "Product Catalog", type: "Serves", description: "Caches product data" },
    { name: "Recommendations", source: "Product Catalog", target: "Recommendation Engine", type: "Triggers", description: "Gets product recommendations" },
    { name: "Fraud Check", source: "Payment Gateway", target: "Fraud Detection", type: "Triggers", description: "Validates transactions for fraud" },
    { name: "Shipping Rates", source: "Shopping Cart", target: "Shipping Service", type: "Triggers", description: "Gets shipping rate quotes" },
    { name: "Price Calculation", source: "Shopping Cart", target: "Pricing Service", type: "Triggers", description: "Calculates final prices with discounts" },
    { name: "Event Publishing", source: "Order Service", target: "Message Queue", type: "Triggers", description: "Publishes domain events" },
    { name: "Admin Reports", source: "Admin Dashboard", target: "Reporting Service", type: "Triggers", description: "Generates admin reports" },
    { name: "Content Serving", source: "Customer Portal", target: "Content Management", type: "Triggers", description: "Retrieves dynamic content" },
  ];

  for (const r of relations) {
    const source = components.get(r.source);
    const target = components.get(r.target);
    if (source && target) {
      await createRelation({
        name: r.name,
        description: r.description,
        sourceId: source.id,
        targetId: target.id,
        relationType: r.type,
      });
    }
  }
}

interface L1CapabilityDef {
  name: string;
  description: string;
  level: string;
}

interface L2CapabilityDef {
  name: string;
  description: string;
  parent: string;
}

interface MetadataUpdateDef {
  name: string;
  status: string;
  maturityValue: number;
  ownershipModel: string;
}

const L1_CAPABILITIES: L1CapabilityDef[] = [
  { name: "Customer Management", description: "Acquire, retain, and manage customer relationships", level: "L1" },
  { name: "Order Fulfillment", description: "Process and fulfill customer orders end-to-end", level: "L1" },
  { name: "Product Management", description: "Manage product lifecycle and catalog", level: "L1" },
  { name: "Financial Operations", description: "Manage payments, invoicing, and financial transactions", level: "L1" },
  { name: "Supply Chain", description: "Manage inventory, suppliers, and logistics", level: "L1" },
  { name: "Marketing & Sales", description: "Drive customer acquisition and revenue growth", level: "L1" },
  { name: "Analytics & Insights", description: "Generate business intelligence and insights", level: "L1" },
  { name: "Platform Operations", description: "Maintain and operate technology platform", level: "L1" },
];

const L2_CAPABILITIES: L2CapabilityDef[] = [
  { name: "Customer Onboarding", description: "Register and onboard new customers", parent: "Customer Management" },
  { name: "Customer Authentication", description: "Verify customer identity and manage access", parent: "Customer Management" },
  { name: "Customer Support", description: "Handle customer inquiries and issues", parent: "Customer Management" },
  { name: "Customer Loyalty", description: "Manage loyalty programs and rewards", parent: "Customer Management" },
  { name: "Order Creation", description: "Create and validate new orders", parent: "Order Fulfillment" },
  { name: "Order Processing", description: "Process orders through fulfillment stages", parent: "Order Fulfillment" },
  { name: "Order Tracking", description: "Track order status and delivery", parent: "Order Fulfillment" },
  { name: "Returns & Refunds", description: "Handle product returns and refunds", parent: "Order Fulfillment" },
  { name: "Product Catalog Management", description: "Maintain product information and categories", parent: "Product Management" },
  { name: "Pricing Management", description: "Set and manage product prices", parent: "Product Management" },
  { name: "Product Search", description: "Enable product discovery and search", parent: "Product Management" },
  { name: "Payment Processing", description: "Process customer payments", parent: "Financial Operations" },
  { name: "Invoice Management", description: "Generate and manage invoices", parent: "Financial Operations" },
  { name: "Fraud Prevention", description: "Detect and prevent fraudulent transactions", parent: "Financial Operations" },
  { name: "Inventory Management", description: "Track and manage inventory levels", parent: "Supply Chain" },
  { name: "Supplier Management", description: "Manage supplier relationships", parent: "Supply Chain" },
  { name: "Shipping & Logistics", description: "Handle shipping and delivery", parent: "Supply Chain" },
  { name: "Campaign Management", description: "Create and manage marketing campaigns", parent: "Marketing & Sales" },
  { name: "Promotions & Discounts", description: "Manage promotional offers", parent: "Marketing & Sales" },
  { name: "Content Publishing", description: "Publish marketing content", parent: "Marketing & Sales" },
  { name: "Business Reporting", description: "Generate business reports", parent: "Analytics & Insights" },
  { name: "Customer Analytics", description: "Analyze customer behavior", parent: "Analytics & Insights" },
  { name: "Predictive Analytics", description: "ML-based predictions and forecasting", parent: "Analytics & Insights" },
  { name: "System Monitoring", description: "Monitor system health and performance", parent: "Platform Operations" },
  { name: "API Management", description: "Manage and secure APIs", parent: "Platform Operations" },
  { name: "Data Management", description: "Manage data storage and access", parent: "Platform Operations" },
];

const METADATA_UPDATES: MetadataUpdateDef[] = [
  { name: "Customer Authentication", status: "Active", maturityValue: 80, ownershipModel: "EnterpriseService" },
  { name: "Order Creation", status: "Active", maturityValue: 75, ownershipModel: "TribeOwned" },
  { name: "Payment Processing", status: "Active", maturityValue: 90, ownershipModel: "EnterpriseService" },
  { name: "Product Search", status: "Active", maturityValue: 60, ownershipModel: "Shared" },
  { name: "Inventory Management", status: "Active", maturityValue: 70, ownershipModel: "TribeOwned" },
  { name: "Fraud Prevention", status: "Active", maturityValue: 55, ownershipModel: "EnterpriseService" },
  { name: "Predictive Analytics", status: "Planned", maturityValue: 30, ownershipModel: "TeamOwned" },
  { name: "System Monitoring", status: "Active", maturityValue: 85, ownershipModel: "EnterpriseService" },
];

async function createL1Capabilities(capabilities: Map<string, Capability>): Promise<void> {
  for (const cap of L1_CAPABILITIES) {
    const capability = await createCapability(cap.name, cap.description, cap.level);
    capabilities.set(cap.name, capability);
  }
}

async function createL2Capabilities(capabilities: Map<string, Capability>): Promise<void> {
  for (const cap of L2_CAPABILITIES) {
    const parent = capabilities.get(cap.parent);
    if (!parent) continue;
    const capability = await createCapability(cap.name, cap.description, "L2", parent.id);
    capabilities.set(cap.name, capability);
  }
}

async function applyMetadataUpdates(capabilities: Map<string, Capability>): Promise<void> {
  for (const update of METADATA_UPDATES) {
    const capability = capabilities.get(update.name);
    if (!capability) continue;
    await updateCapabilityMetadata(capability.id, {
      status: update.status,
      maturityValue: update.maturityValue,
      ownershipModel: update.ownershipModel,
    });
  }
}

async function seedCapabilities(): Promise<Map<string, Capability>> {
  console.log("\n🎯 Seeding Business Capabilities...");
  const capabilities = new Map<string, Capability>();

  await createL1Capabilities(capabilities);
  await createL2Capabilities(capabilities);
  await applyMetadataUpdates(capabilities);

  return capabilities;
}

async function seedBusinessDomains(
  capabilities: Map<string, Capability>
): Promise<Map<string, BusinessDomain>> {
  console.log("\n🏢 Seeding Business Domains...");
  const domains = new Map<string, BusinessDomain>();

  const domainData = [
    {
      name: "E-Commerce",
      description: "Online retail and shopping experience",
      capabilities: ["Order Fulfillment", "Product Management"],
    },
    {
      name: "Customer Experience",
      description: "Customer-facing services and support",
      capabilities: ["Customer Management"],
    },
    {
      name: "Payments & Finance",
      description: "Financial transactions and accounting",
      capabilities: ["Financial Operations"],
    },
    {
      name: "Logistics",
      description: "Inventory and shipping operations",
      capabilities: ["Supply Chain"],
    },
    {
      name: "Marketing",
      description: "Marketing and promotional activities",
      capabilities: ["Marketing & Sales"],
    },
    {
      name: "Data & Analytics",
      description: "Business intelligence and data science",
      capabilities: ["Analytics & Insights", "Platform Operations"],
    },
  ];

  for (const d of domainData) {
    const domain = await createBusinessDomain(d.name, d.description);
    domains.set(d.name, domain);

    for (const capName of d.capabilities) {
      const capability = capabilities.get(capName);
      if (capability) {
        await assignCapabilityToDomain(domain.id, capability.id);
      }
    }
  }

  return domains;
}

async function seedEnterpriseCapabilities(
  capabilities: Map<string, Capability>
): Promise<void> {
  console.log("\n🏛️ Seeding Enterprise Capabilities...");

  const enterpriseCapabilities = [
    {
      name: "Customer Identity",
      description: "Enterprise-wide customer identity and access management",
      category: "Customer",
      linkedCapabilities: ["Customer Authentication", "Customer Onboarding"],
    },
    {
      name: "Order Management",
      description: "Enterprise order processing and fulfillment",
      category: "Operations",
      linkedCapabilities: ["Order Creation", "Order Processing", "Order Tracking"],
    },
    {
      name: "Payment Platform",
      description: "Enterprise payment processing infrastructure",
      category: "Finance",
      linkedCapabilities: ["Payment Processing", "Fraud Prevention"],
    },
    {
      name: "Data Platform",
      description: "Enterprise data management and analytics",
      category: "Technology",
      linkedCapabilities: ["Business Reporting", "Customer Analytics", "Predictive Analytics"],
    },
    {
      name: "Integration Platform",
      description: "Enterprise API and integration services",
      category: "Technology",
      linkedCapabilities: ["API Management", "System Monitoring"],
    },
  ];

  for (const ec of enterpriseCapabilities) {
    const enterprise = await createEnterpriseCapability(ec.name, ec.description, ec.category);

    for (const capName of ec.linkedCapabilities) {
      const capability = capabilities.get(capName);
      if (capability) {
        try {
          await linkEnterpriseCapability(enterprise.id, capability.id);
        } catch (e) {
          console.log(`    (Skipping link - may already exist or not eligible)`);
        }
      }
    }
  }
}

async function seedCapabilityDependencies(capabilities: Map<string, Capability>): Promise<void> {
  console.log("\n🔗 Seeding Capability Dependencies...");

  const dependencies = [
    { source: "Order Creation", target: "Customer Authentication", type: "requires", description: "Orders require authenticated customers" },
    { source: "Order Creation", target: "Inventory Management", type: "requires", description: "Must verify inventory availability" },
    { source: "Order Processing", target: "Payment Processing", type: "requires", description: "Orders need payment processing" },
    { source: "Order Processing", target: "Shipping & Logistics", type: "requires", description: "Fulfilled orders need shipping" },
    { source: "Payment Processing", target: "Fraud Prevention", type: "supports", description: "Fraud detection supports payments" },
    { source: "Customer Analytics", target: "Business Reporting", type: "informs", description: "Customer analytics feeds reporting" },
    { source: "Predictive Analytics", target: "Customer Analytics", type: "requires", description: "Predictions need customer data" },
    { source: "Campaign Management", target: "Customer Analytics", type: "requires", description: "Campaigns need customer insights" },
  ];

  for (const dep of dependencies) {
    const source = capabilities.get(dep.source);
    const target = capabilities.get(dep.target);
    if (source && target) {
      try {
        await createDependency(source.id, target.id, dep.type, dep.description);
      } catch (e) {
        console.log(`    (Skipping dependency - may already exist)`);
      }
    }
  }
}

async function seedSystemRealizations(
  capabilities: Map<string, Capability>,
  components: Map<string, Component>
): Promise<void> {
  console.log("\n⚙️ Seeding System Realizations...");

  const realizations = [
    { capability: "Customer Authentication", component: "User Service", level: "Full", notes: "Primary authentication system" },
    { capability: "Customer Onboarding", component: "User Service", level: "Full", notes: "Handles user registration" },
    { capability: "Order Creation", component: "Order Service", level: "Full", notes: "Core order creation" },
    { capability: "Order Creation", component: "Shopping Cart", level: "Partial", notes: "Cart to order conversion" },
    { capability: "Order Processing", component: "Order Service", level: "Full", notes: "Order lifecycle management" },
    { capability: "Payment Processing", component: "Payment Gateway", level: "Full", notes: "Payment processing hub" },
    { capability: "Fraud Prevention", component: "Fraud Detection", level: "Full", notes: "Real-time fraud detection" },
    { capability: "Product Search", component: "Search Engine", level: "Full", notes: "Full-text search" },
    { capability: "Product Search", component: "Product Catalog", level: "Partial", notes: "Product data source" },
    { capability: "Inventory Management", component: "Inventory Service", level: "Full", notes: "Inventory tracking" },
    { capability: "Shipping & Logistics", component: "Shipping Service", level: "Full", notes: "Shipping calculation and tracking" },
    { capability: "Business Reporting", component: "Reporting Service", level: "Full", notes: "Report generation" },
    { capability: "Business Reporting", component: "Analytics Platform", level: "Partial", notes: "Data source" },
    { capability: "Customer Analytics", component: "Analytics Platform", level: "Full", notes: "Customer behavior analytics" },
    { capability: "Predictive Analytics", component: "Recommendation Engine", level: "Partial", notes: "ML-based predictions" },
    { capability: "API Management", component: "API Gateway", level: "Full", notes: "API routing and security" },
    { capability: "System Monitoring", component: "Analytics Platform", level: "Partial", notes: "System metrics" },
    { capability: "Content Publishing", component: "Content Management", level: "Full", notes: "Content management" },
    { capability: "Promotions & Discounts", component: "Pricing Service", level: "Full", notes: "Promotional pricing" },
  ];

  for (const r of realizations) {
    const capability = capabilities.get(r.capability);
    const component = components.get(r.component);
    if (capability && component) {
      try {
        await linkSystemToCapability(capability.id, component.id, r.level, r.notes);
      } catch (e) {
        console.log(`    (Skipping realization - may already exist)`);
      }
    }
  }
}

async function seedViews(components: Map<string, Component>): Promise<void> {
  console.log("\n🖼️ Seeding Architecture Views...");

  const viewsData = [
    {
      name: "Order Flow",
      description: "Order processing architecture",
      components: [
        { name: "Shopping Cart", x: 100, y: 100 },
        { name: "Order Service", x: 400, y: 100 },
        { name: "Payment Gateway", x: 700, y: 100 },
        { name: "Inventory Service", x: 400, y: 300 },
        { name: "Shipping Service", x: 700, y: 300 },
        { name: "Notification Service", x: 400, y: 500 },
      ],
    },
    {
      name: "Customer Facing",
      description: "Customer-facing services",
      components: [
        { name: "API Gateway", x: 100, y: 200 },
        { name: "User Service", x: 400, y: 100 },
        { name: "Customer Portal", x: 400, y: 300 },
        { name: "Product Catalog", x: 700, y: 100 },
        { name: "Search Engine", x: 700, y: 300 },
        { name: "Recommendation Engine", x: 1000, y: 200 },
      ],
    },
    {
      name: "Data Platform",
      description: "Analytics and data services",
      components: [
        { name: "Analytics Platform", x: 400, y: 200 },
        { name: "Reporting Service", x: 100, y: 200 },
        { name: "Message Queue", x: 700, y: 100 },
        { name: "Cache Layer", x: 700, y: 300 },
      ],
    },
  ];

  for (const viewData of viewsData) {
    const view = await createView(viewData.name, viewData.description);

    for (const compData of viewData.components) {
      const component = components.get(compData.name);
      if (component) {
        await addComponentToView(view.id, component.id, compData.x, compData.y);
      }
    }
  }
}

async function batchUpdatePillarsWithRetry(
  changes: PillarChange[],
  maxRetries: number = 5
): Promise<void> {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    const { etag } = await getStrategyPillarsWithETag();
    try {
      await batchUpdatePillars(changes, etag);
      return;
    } catch (e) {
      if (attempt === maxRetries) {
        throw e;
      }
      console.log(`    (Retry ${attempt}/${maxRetries} for batch update...)`);
      await new Promise(resolve => setTimeout(resolve, 100 * attempt));
    }
  }
}

interface PillarDefinition {
  name: string;
  description: string;
  fitCriteria: string;
}

const PILLAR_DEFINITIONS: PillarDefinition[] = [
  { name: "Cloud Native", description: "Embrace cloud-native technologies and patterns", fitCriteria: "Containerization, Kubernetes orchestration, auto-scaling, CI/CD pipelines" },
  { name: "API First", description: "Design and build APIs as first-class products", fitCriteria: "OpenAPI documentation, versioned APIs, RESTful design, developer portal" },
  { name: "Security", description: "Security-first approach to all systems", fitCriteria: "Authentication, authorization, encryption, audit logging, vulnerability scanning" },
];

async function createMissingPillars(existingNames: Set<string>): Promise<void> {
  for (const def of PILLAR_DEFINITIONS) {
    if (existingNames.has(def.name)) continue;
    try {
      await createStrategyPillar(def.name, def.description);
    } catch (e) {
      console.log(`    (Pillar ${def.name} may already exist)`);
    }
  }
}

function buildFitScoringChanges(pillarMap: Map<string, StrategyPillar>): PillarChange[] {
  return PILLAR_DEFINITIONS
    .filter(def => {
      const pillar = pillarMap.get(def.name);
      return pillar && !pillar.fitScoringEnabled;
    })
    .map(def => {
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
}

function extractEnabledPillars(pillars: StrategyPillar[]): Map<string, StrategyPillar> {
  return new Map(
    pillars.filter(p => p.fitScoringEnabled).map(p => [p.name, p])
  );
}

async function ensureStrategyPillarsWithFitScoring(): Promise<Map<string, StrategyPillar>> {
  console.log("\n🎯 Ensuring Strategy Pillars with Fit Scoring...");

  const { pillars: existingPillars } = await getStrategyPillarsWithETag();
  const existingNames = new Set(existingPillars.map(p => p.name));

  await createMissingPillars(existingNames);

  const { pillars: allPillars } = await getStrategyPillarsWithETag();
  const pillarMap = new Map(allPillars.map(p => [p.name, p]));

  const fitScoringChanges = buildFitScoringChanges(pillarMap);
  if (fitScoringChanges.length > 0) {
    try {
      await batchUpdatePillarsWithRetry(fitScoringChanges);
    } catch (e) {
      console.log(`    (Failed to batch enable fit scoring: ${e})`);
    }
  }

  const { pillars: finalPillars } = await getStrategyPillarsWithETag();
  const enabledPillars = extractEnabledPillars(finalPillars);

  console.log(`  ${enabledPillars.size} pillars with fit scoring enabled`);
  return enabledPillars;
}

interface FitScoreEntry {
  pillarName: string;
  score: number;
  rationale: string;
}

interface ComponentFitScoreData {
  componentName: string;
  scores: FitScoreEntry[];
}

const FIT_SCORE_DATA: ComponentFitScoreData[] = [
  {
    componentName: "User Service",
    scores: [
      { pillarName: "Cloud Native", score: 4, rationale: "Fully containerized with Kubernetes orchestration" },
      { pillarName: "API First", score: 5, rationale: "Well-documented REST APIs with OpenAPI specs" },
      { pillarName: "Security", score: 4, rationale: "OAuth2/OIDC implementation, regular security audits" },
    ],
  },
  {
    componentName: "Order Service",
    scores: [
      { pillarName: "Cloud Native", score: 3, rationale: "Containerized but with some legacy dependencies" },
      { pillarName: "API First", score: 4, rationale: "REST API with documentation, some inconsistencies" },
      { pillarName: "Security", score: 3, rationale: "Basic authentication, needs improved audit logging" },
    ],
  },
  {
    componentName: "Payment Gateway",
    scores: [
      { pillarName: "Cloud Native", score: 4, rationale: "Cloud-hosted with auto-scaling capabilities" },
      { pillarName: "API First", score: 5, rationale: "Industry-standard payment APIs" },
      { pillarName: "Security", score: 5, rationale: "PCI-DSS compliant, encryption at rest and transit" },
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
      { pillarName: "Security", score: 4, rationale: "Role-based access, data encryption" },
    ],
  },
  {
    componentName: "Search Engine",
    scores: [
      { pillarName: "Cloud Native", score: 4, rationale: "Elasticsearch cluster on Kubernetes" },
      { pillarName: "API First", score: 4, rationale: "Standard search APIs" },
      { pillarName: "Security", score: 3, rationale: "API keys only, needs better auth" },
    ],
  },
  {
    componentName: "Notification Service",
    scores: [
      { pillarName: "Cloud Native", score: 5, rationale: "Event-driven serverless functions" },
      { pillarName: "API First", score: 3, rationale: "Internal APIs, limited documentation" },
      { pillarName: "Security", score: 4, rationale: "Secure message handling, encrypted queues" },
    ],
  },
  {
    componentName: "Admin Dashboard",
    scores: [
      { pillarName: "Cloud Native", score: 2, rationale: "Monolithic deployment, manual updates" },
      { pillarName: "API First", score: 2, rationale: "Server-rendered pages, limited API usage" },
      { pillarName: "Security", score: 3, rationale: "Basic RBAC, needs MFA implementation" },
    ],
  },
];

async function applyFitScoresForComponent(
  component: Component,
  scores: FitScoreEntry[],
  enabledPillars: Map<string, StrategyPillar>
): Promise<void> {
  for (const scoreData of scores) {
    const pillar = enabledPillars.get(scoreData.pillarName);
    if (!pillar) continue;

    try {
      await setApplicationFitScore(component.id, pillar.id, scoreData.score, scoreData.rationale);
    } catch (e) {
      console.log(`    (Skipping fit score - may already exist or pillar not enabled)`);
    }
  }
}

async function seedApplicationFitScores(components: Map<string, Component>): Promise<void> {
  console.log("\n📊 Seeding Application Fit Scores...");

  const enabledPillars = await ensureStrategyPillarsWithFitScoring();
  if (enabledPillars.size === 0) {
    console.log("  No pillars with fit scoring enabled, skipping fit scores");
    return;
  }

  console.log(`  Setting fit scores for ${enabledPillars.size} pillars`);

  for (const data of FIT_SCORE_DATA) {
    const component = components.get(data.componentName);
    if (!component) continue;
    await applyFitScoresForComponent(component, data.scores, enabledPillars);
  }
}

async function tryLinkComponents(
  componentNames: string[],
  components: Map<string, Component>,
  linkFn: (componentId: string) => Promise<void>
): Promise<void> {
  for (const compName of componentNames) {
    const component = components.get(compName);
    if (component) {
      try {
        await linkFn(component.id);
      } catch {
        console.log(`    (Skipping link - may already exist)`);
      }
    }
  }
}

async function seedAcquiredEntities(components: Map<string, Component>): Promise<void> {
  const acquiredEntities = [
    {
      name: "DataTech Solutions",
      acquisitionDate: "2023-06-15",
      integrationStatus: "COMPLETED",
      notes: "Acquired for their analytics capabilities. Integration completed Q4 2023.",
      components: ["Analytics Platform", "Recommendation Engine"],
    },
    {
      name: "SecurePay Inc",
      acquisitionDate: "2022-03-20",
      integrationStatus: "IN_PROGRESS",
      notes: "Acquired for payment processing expertise. Currently migrating to unified auth.",
      components: ["Payment Gateway", "Fraud Detection"],
    },
    {
      name: "CloudScale Systems",
      acquisitionDate: "2024-01-10",
      integrationStatus: "NOT_STARTED",
      notes: "Recent acquisition. Integration planning phase Q2 2024.",
      components: ["API Gateway", "Cache Layer"],
    },
    {
      name: "RetailTech Corp",
      acquisitionDate: "2021-09-01",
      integrationStatus: "COMPLETED",
      notes: "Legacy retail systems acquisition. Fully integrated into e-commerce platform.",
      components: ["Product Catalog", "Inventory Service"],
    },
  ];

  for (const ae of acquiredEntities) {
    try {
      const entity = await createAcquiredEntity(ae.name, ae.acquisitionDate, ae.integrationStatus, ae.notes);
      await tryLinkComponents(ae.components, components, (compId) =>
        linkComponentToAcquiredEntity(entity.id, compId, `Acquired from ${ae.name}`)
      );
    } catch {
      console.log(`    (Skipping acquired entity - may already exist)`);
    }
  }
}

async function seedVendors(components: Map<string, Component>): Promise<void> {
  const vendors = [
    {
      name: "Elastic NV",
      implementationPartner: "SearchTech Consulting",
      notes: "Enterprise search platform. Contract renewal due 2025.",
      components: ["Search Engine"],
    },
    {
      name: "Redis Labs",
      implementationPartner: "",
      notes: "In-memory data store provider. Redis Enterprise license.",
      components: ["Cache Layer"],
    },
    {
      name: "AWS",
      implementationPartner: "Cloud Solutions Inc",
      notes: "Primary cloud infrastructure provider. Enterprise agreement.",
      components: ["Message Queue"],
    },
    {
      name: "Twilio",
      implementationPartner: "",
      notes: "SMS and communication APIs for customer notifications.",
      components: ["Notification Service"],
    },
    {
      name: "Stripe",
      implementationPartner: "FinTech Partners",
      notes: "Payment processing integration. PCI compliant.",
      components: ["Payment Gateway"],
    },
    {
      name: "Salesforce",
      implementationPartner: "CRM Consultants Ltd",
      notes: "CRM integration for customer data sync.",
      components: ["Customer Portal"],
    },
  ];

  for (const v of vendors) {
    try {
      const vendor = await createVendor(v.name, v.implementationPartner, v.notes);
      await tryLinkComponents(v.components, components, (compId) =>
        linkComponentToVendor(vendor.id, compId, `Purchased from ${v.name}`)
      );
    } catch {
      console.log(`    (Skipping vendor - may already exist)`);
    }
  }
}

async function seedInternalTeams(components: Map<string, Component>): Promise<void> {
  const internalTeams = [
    {
      name: "Core Platform Team",
      department: "Engineering",
      contactPerson: "Jane Smith",
      notes: "Responsible for core microservices and platform infrastructure.",
      components: ["User Service", "Order Service", "Inventory Service", "API Gateway"],
    },
    {
      name: "Customer Experience Team",
      department: "Product",
      contactPerson: "John Doe",
      notes: "Owns customer-facing applications and user journey.",
      components: ["Customer Portal", "Shopping Cart"],
    },
    {
      name: "Data Engineering Team",
      department: "Engineering",
      contactPerson: "Alice Johnson",
      notes: "Builds data pipelines, analytics, and ML infrastructure.",
      components: ["Reporting Service", "Analytics Platform", "Recommendation Engine"],
    },
    {
      name: "Operations Team",
      department: "IT Operations",
      contactPerson: "Bob Williams",
      notes: "Manages internal tools and operational systems.",
      components: ["Admin Dashboard", "Pricing Service", "Shipping Service"],
    },
    {
      name: "Security Team",
      department: "Engineering",
      contactPerson: "Carol Chen",
      notes: "Responsible for security, fraud prevention, and compliance.",
      components: ["Fraud Detection"],
    },
    {
      name: "Content Team",
      department: "Marketing",
      contactPerson: "David Lee",
      notes: "Manages product content and digital assets.",
      components: ["Content Management", "Product Catalog"],
    },
  ];

  for (const team of internalTeams) {
    try {
      const internalTeam = await createInternalTeam(team.name, team.department, team.contactPerson, team.notes);
      await tryLinkComponents(team.components, components, (compId) =>
        linkComponentToInternalTeam(internalTeam.id, compId, `Built by ${team.name}`)
      );
    } catch {
      console.log(`    (Skipping internal team - may already exist)`);
    }
  }
}

async function seedOriginEntities(components: Map<string, Component>): Promise<void> {
  console.log("\n🏭 Seeding Origin Entities...");
  await seedAcquiredEntities(components);
  await seedVendors(components);
  await seedInternalTeams(components);
}

async function checkApiHealth(): Promise<boolean> {
  try {
    const response = await fetch(`${BASE_URL}/health`);
    return response.ok;
  } catch {
    return false;
  }
}

async function main(): Promise<void> {
  console.log("🌱 EASI Test Data Seeding Script");
  console.log("================================");
  console.log(`Base URL: ${BASE_URL}`);

  if (SESSION_COOKIE) {
    console.log(`Auth Mode: Session Cookie (DEX)`);
    console.log(`Cookie: ${SESSION_COOKIE.substring(0, 20)}...`);
  } else {
    console.log(`Auth Mode: Bypass (X-Tenant-ID header)`);
    console.log(`Tenant ID: ${TENANT_ID}`);
  }
  console.log("");

  console.log("Checking API health...");
  const healthy = await checkApiHealth();
  if (!healthy) {
    console.error("❌ API is not reachable. Make sure the backend is running.");
    console.error(`   Tried: ${BASE_URL}/health`);
    process.exit(1);
  }
  console.log("✅ API is healthy\n");

  try {
    const components = await seedComponents();
    await seedRelations(components);

    const capabilities = await seedCapabilities();
    const _domains = await seedBusinessDomains(capabilities);

    await seedEnterpriseCapabilities(capabilities);
    await seedCapabilityDependencies(capabilities);
    await seedSystemRealizations(capabilities, components);

    await seedViews(components);

    await seedApplicationFitScores(components);

    await seedOriginEntities(components);

    console.log("\n✅ Test data seeding complete!");
    console.log("\nSummary:");
    console.log(`  - ${components.size} components created`);
    console.log(`  - ${capabilities.size} capabilities created`);
    console.log(`  - Business domains, enterprise capabilities, and views created`);
    console.log(`  - System realizations and dependencies linked`);
    console.log(`  - Application fit scores set for strategic pillars`);
    console.log(`  - Origin entities (acquired entities, vendors, internal teams) created`);
  } catch (error) {
    console.error("\n❌ Seeding failed:", error);
    process.exit(1);
  }
}

main();
