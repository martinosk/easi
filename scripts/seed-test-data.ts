#!/usr/bin/env npx ts-node

/**
 * Test Data Seeding Script
 *
 * This script populates the test database with realistic enterprise architecture data
 * by calling the backend API. It requires AUTH_MODE=bypass to be set on the backend.
 *
 * Usage:
 *   npx ts-node scripts/seed-test-data.ts
 *   npx ts-node scripts/seed-test-data.ts --base-url http://localhost:8080
 *   npx ts-node scripts/seed-test-data.ts --tenant-id acme
 *
 * Prerequisites:
 *   - Backend running with AUTH_MODE=bypass
 *   - Test tenant "acme" exists (created by migration 041)
 */

const BASE_URL = process.argv.includes("--base-url")
  ? process.argv[process.argv.indexOf("--base-url") + 1]
  : "http://localhost:8080";

const TENANT_ID = process.argv.includes("--tenant-id")
  ? process.argv[process.argv.indexOf("--tenant-id") + 1]
  : "acme";

const API_URL = `${BASE_URL}/api/v1`;

interface ApiResponse<T> {
  data?: T;
  id?: string;
  _links?: Record<string, string>;
}

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

async function apiCall<T>(
  method: string,
  path: string,
  body?: unknown
): Promise<T> {
  const url = `${API_URL}${path}`;
  const options: RequestInit = {
    method,
    headers: {
      "Content-Type": "application/json",
      "X-Tenant-ID": TENANT_ID,
    },
  };

  if (body) {
    options.body = JSON.stringify(body);
  }

  const response = await fetch(url, options);

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`API call failed: ${method} ${path} - ${response.status}: ${text}`);
  }

  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
}

async function createComponent(name: string, description: string): Promise<Component> {
  console.log(`  Creating component: ${name}`);
  return apiCall<Component>("POST", "/components", { name, description });
}

async function createRelation(
  name: string,
  description: string,
  sourceId: string,
  targetId: string,
  relationType: string
): Promise<void> {
  console.log(`  Creating relation: ${name}`);
  await apiCall("POST", "/relations", {
    name,
    description,
    sourceComponentId: sourceId,
    targetComponentId: targetId,
    relationType,
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
    { name: "User Authentication", source: "API Gateway", target: "User Service", type: "calls", description: "Authenticates API requests" },
    { name: "Order Processing", source: "Order Service", target: "Payment Gateway", type: "calls", description: "Processes order payments" },
    { name: "Inventory Check", source: "Order Service", target: "Inventory Service", type: "calls", description: "Validates inventory availability" },
    { name: "Order Notifications", source: "Order Service", target: "Notification Service", type: "publishes", description: "Publishes order events for notifications" },
    { name: "Cart Checkout", source: "Shopping Cart", target: "Order Service", type: "calls", description: "Creates orders from cart" },
    { name: "Product Search", source: "Product Catalog", target: "Search Engine", type: "uses", description: "Indexes and searches products" },
    { name: "Analytics Events", source: "API Gateway", target: "Analytics Platform", type: "publishes", description: "Sends analytics events" },
    { name: "Cache Products", source: "Product Catalog", target: "Cache Layer", type: "uses", description: "Caches product data" },
    { name: "Recommendations", source: "Product Catalog", target: "Recommendation Engine", type: "calls", description: "Gets product recommendations" },
    { name: "Fraud Check", source: "Payment Gateway", target: "Fraud Detection", type: "calls", description: "Validates transactions for fraud" },
    { name: "Shipping Rates", source: "Shopping Cart", target: "Shipping Service", type: "calls", description: "Gets shipping rate quotes" },
    { name: "Price Calculation", source: "Shopping Cart", target: "Pricing Service", type: "calls", description: "Calculates final prices with discounts" },
    { name: "Event Publishing", source: "Order Service", target: "Message Queue", type: "publishes", description: "Publishes domain events" },
    { name: "Admin Reports", source: "Admin Dashboard", target: "Reporting Service", type: "calls", description: "Generates admin reports" },
    { name: "Content Serving", source: "Customer Portal", target: "Content Management", type: "calls", description: "Retrieves dynamic content" },
  ];

  for (const r of relations) {
    const source = components.get(r.source);
    const target = components.get(r.target);
    if (source && target) {
      await createRelation(r.name, r.description, source.id, target.id, r.type);
    }
  }
}

async function seedCapabilities(): Promise<Map<string, Capability>> {
  console.log("\n🎯 Seeding Business Capabilities...");
  const capabilities = new Map<string, Capability>();

  const l1Capabilities = [
    { name: "Customer Management", description: "Acquire, retain, and manage customer relationships", level: "L1" },
    { name: "Order Fulfillment", description: "Process and fulfill customer orders end-to-end", level: "L1" },
    { name: "Product Management", description: "Manage product lifecycle and catalog", level: "L1" },
    { name: "Financial Operations", description: "Manage payments, invoicing, and financial transactions", level: "L1" },
    { name: "Supply Chain", description: "Manage inventory, suppliers, and logistics", level: "L1" },
    { name: "Marketing & Sales", description: "Drive customer acquisition and revenue growth", level: "L1" },
    { name: "Analytics & Insights", description: "Generate business intelligence and insights", level: "L1" },
    { name: "Platform Operations", description: "Maintain and operate technology platform", level: "L1" },
  ];

  for (const cap of l1Capabilities) {
    const capability = await createCapability(cap.name, cap.description, cap.level);
    capabilities.set(cap.name, capability);
  }

  const l2Capabilities = [
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

  for (const cap of l2Capabilities) {
    const parent = capabilities.get(cap.parent);
    if (parent) {
      const capability = await createCapability(cap.name, cap.description, "L2", parent.id);
      capabilities.set(cap.name, capability);
    }
  }

  const metadataUpdates = [
    { name: "Customer Authentication", status: "Active", maturityValue: 80, ownershipModel: "Platform" },
    { name: "Order Creation", status: "Active", maturityValue: 75, ownershipModel: "TribeOwned" },
    { name: "Payment Processing", status: "Active", maturityValue: 90, ownershipModel: "Platform" },
    { name: "Product Search", status: "Active", maturityValue: 60, ownershipModel: "Shared" },
    { name: "Inventory Management", status: "Active", maturityValue: 70, ownershipModel: "TribeOwned" },
    { name: "Fraud Prevention", status: "Evolving", maturityValue: 55, ownershipModel: "Platform" },
    { name: "Predictive Analytics", status: "Emerging", maturityValue: 30, ownershipModel: "Experimental" },
    { name: "System Monitoring", status: "Active", maturityValue: 85, ownershipModel: "Platform" },
  ];

  for (const update of metadataUpdates) {
    const capability = capabilities.get(update.name);
    if (capability) {
      await updateCapabilityMetadata(capability.id, {
        status: update.status,
        maturityValue: update.maturityValue,
        ownershipModel: update.ownershipModel,
      });
    }
  }

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
      capabilities: ["Order Creation", "Order Processing", "Shopping Cart", "Product Search"],
    },
    {
      name: "Customer Experience",
      description: "Customer-facing services and support",
      capabilities: ["Customer Onboarding", "Customer Authentication", "Customer Support", "Customer Loyalty"],
    },
    {
      name: "Payments & Finance",
      description: "Financial transactions and accounting",
      capabilities: ["Payment Processing", "Invoice Management", "Fraud Prevention"],
    },
    {
      name: "Logistics",
      description: "Inventory and shipping operations",
      capabilities: ["Inventory Management", "Shipping & Logistics", "Order Tracking"],
    },
    {
      name: "Marketing",
      description: "Marketing and promotional activities",
      capabilities: ["Campaign Management", "Promotions & Discounts", "Content Publishing"],
    },
    {
      name: "Data & Analytics",
      description: "Business intelligence and data science",
      capabilities: ["Business Reporting", "Customer Analytics", "Predictive Analytics"],
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
  console.log(`Tenant ID: ${TENANT_ID}`);
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

    console.log("\n✅ Test data seeding complete!");
    console.log("\nSummary:");
    console.log(`  - ${components.size} components created`);
    console.log(`  - ${capabilities.size} capabilities created`);
    console.log(`  - Business domains, enterprise capabilities, and views created`);
    console.log(`  - System realizations and dependencies linked`);
  } catch (error) {
    console.error("\n❌ Seeding failed:", error);
    process.exit(1);
  }
}

main();
