import { apiCall, parallelBatch } from "./config.ts";
import { COMPONENTS } from "./component-list.ts";
import type { Component } from "./types.ts";

async function createComponent(name: string, description: string): Promise<Component> {
  return apiCall<Component>("POST", "/components", { name, description });
}

export async function seedComponents(): Promise<Map<string, string>> {
  console.log(`\n📦 Seeding Application Components (${COMPONENTS.length} total)...`);
  const idByName = new Map<string, string>();

  const results = await parallelBatch(COMPONENTS, 10, (c) => createComponent(c.name, c.description));

  results.forEach((comp, i) => {
    idByName.set(COMPONENTS[i].name, comp.id);
    if ((i + 1) % 50 === 0 || i + 1 === COMPONENTS.length) {
      console.log(`  Progress: ${i + 1}/${COMPONENTS.length} components`);
    }
  });

  return idByName;
}

export async function seedRelations(components: Map<string, string>): Promise<void> {
  console.log("\n🔗 Seeding Component Relations...");

  const relations = [
    { name: "User Authentication", source: "API Gateway", target: "User Service", type: "Triggers", description: "Authenticates API requests via User Service" },
    { name: "Order Processing", source: "Order Service", target: "Payment Gateway", type: "Triggers", description: "Processes order payments" },
    { name: "Inventory Check", source: "Order Service", target: "Inventory Service", type: "Triggers", description: "Validates inventory availability before confirming" },
    { name: "Order Notifications", source: "Order Service", target: "Notification Service", type: "Triggers", description: "Publishes order events for customer notifications" },
    { name: "Cart Checkout", source: "Shopping Cart", target: "Order Service", type: "Triggers", description: "Creates orders from cart on checkout" },
    { name: "Product Search Index", source: "Search Engine", target: "Product Catalog", type: "Serves", description: "Indexes and searches product data" },
    { name: "Analytics Events", source: "API Gateway", target: "Analytics Platform", type: "Triggers", description: "Forwards analytics events from inbound requests" },
    { name: "Product Cache", source: "Cache Layer", target: "Product Catalog", type: "Serves", description: "Caches frequently accessed product data" },
    { name: "Personalized Recommendations", source: "Product Catalog", target: "Recommendation Engine", type: "Triggers", description: "Fetches product recommendations for catalog pages" },
    { name: "Fraud Check", source: "Payment Gateway", target: "Fraud Detection", type: "Triggers", description: "Validates payment transactions for fraud signals" },
    { name: "Shipping Rates", source: "Shopping Cart", target: "Shipping Service", type: "Triggers", description: "Retrieves shipping rate quotes during checkout" },
    { name: "Price Calculation", source: "Shopping Cart", target: "Pricing Service", type: "Triggers", description: "Calculates final prices with applicable discounts" },
    { name: "Domain Events", source: "Order Service", target: "Message Queue", type: "Triggers", description: "Publishes order domain events asynchronously" },
    { name: "Admin Reports", source: "Admin Dashboard", target: "Reporting Service", type: "Triggers", description: "Generates operational reports on demand" },
    { name: "Content Retrieval", source: "Customer Portal", target: "Content Management", type: "Triggers", description: "Retrieves dynamic marketing content" },
    { name: "Session Cache", source: "Cache Layer", target: "User Service", type: "Serves", description: "Caches user session data for fast lookup" },
    { name: "Notification Queue", source: "Notification Service", target: "Message Queue", type: "Serves", description: "Consumes notification events from the queue" },
    { name: "Analytics Reporting", source: "Analytics Platform", target: "Reporting Service", type: "Serves", description: "Provides aggregated data for reports" },
    { name: "Search Recommendations", source: "Search Engine", target: "Recommendation Engine", type: "Serves", description: "Surfaces recommended products in search results" },
    { name: "Gateway Rate Limit Cache", source: "Cache Layer", target: "API Gateway", type: "Serves", description: "Backs API Gateway rate limit counters" },
  ];

  for (const r of relations) {
    const sourceId = components.get(r.source);
    const targetId = components.get(r.target);
    if (!sourceId || !targetId) continue;
    try {
      await apiCall("POST", "/relations", {
        name: r.name,
        description: r.description,
        sourceComponentId: sourceId,
        targetComponentId: targetId,
        relationType: r.type,
      });
    } catch {
      console.log(`    (Skipping relation '${r.name}' — may already exist)`);
    }
  }
}
