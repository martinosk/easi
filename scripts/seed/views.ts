import { apiCall } from "./config.ts";
import type { View } from "./types.ts";

async function createView(name: string, description: string): Promise<View> {
  return apiCall<View>("POST", "/views", { name, description });
}

async function addComponentToView(viewId: string, componentId: string, x: number, y: number): Promise<void> {
  await apiCall("POST", `/views/${viewId}/components`, { componentId, x, y });
}

export async function seedViews(components: Map<string, string>): Promise<void> {
  console.log("\n🖼️  Seeding Architecture Views...");

  const viewsData = [
    {
      name: "Order Flow",
      description: "End-to-end order processing architecture",
      components: [
        { name: "Shopping Cart", x: 100, y: 100 },
        { name: "Order Service", x: 400, y: 100 },
        { name: "Payment Gateway", x: 700, y: 100 },
        { name: "Inventory Service", x: 400, y: 300 },
        { name: "Shipping Service", x: 700, y: 300 },
        { name: "Notification Service", x: 400, y: 500 },
        { name: "Fraud Detection", x: 1000, y: 100 },
      ],
    },
    {
      name: "Customer Facing",
      description: "Customer-facing services and integrations",
      components: [
        { name: "API Gateway", x: 100, y: 200 },
        { name: "User Service", x: 400, y: 100 },
        { name: "Customer Portal", x: 400, y: 300 },
        { name: "Product Catalog", x: 700, y: 100 },
        { name: "Search Engine", x: 700, y: 300 },
        { name: "Recommendation Engine", x: 1000, y: 200 },
        { name: "Pricing Service", x: 1000, y: 400 },
      ],
    },
    {
      name: "Data Platform",
      description: "Analytics, reporting, and data services",
      components: [
        { name: "Analytics Platform", x: 400, y: 200 },
        { name: "Reporting Service", x: 100, y: 200 },
        { name: "Message Queue", x: 700, y: 100 },
        { name: "Cache Layer", x: 700, y: 300 },
        { name: "Data Warehouse", x: 400, y: 400 },
        { name: "ETL Pipeline Service", x: 100, y: 400 },
      ],
    },
    {
      name: "Security Architecture",
      description: "Security, identity, and fraud prevention services",
      components: [
        { name: "API Gateway", x: 100, y: 200 },
        { name: "User Service", x: 400, y: 100 },
        { name: "Fraud Detection", x: 400, y: 300 },
        { name: "Payment Gateway", x: 700, y: 200 },
        { name: "Identity Provider", x: 700, y: 400 },
        { name: "Single Sign-On Service", x: 1000, y: 200 },
      ],
    },
    {
      name: "Integration Layer",
      description: "API and integration platform components",
      components: [
        { name: "API Gateway", x: 100, y: 200 },
        { name: "Message Queue", x: 400, y: 100 },
        { name: "Event Bridge", x: 400, y: 300 },
        { name: "Enterprise Service Bus", x: 700, y: 200 },
        { name: "Integration Platform", x: 1000, y: 200 },
        { name: "Webhook Manager", x: 700, y: 400 },
      ],
    },
  ];

  for (const viewData of viewsData) {
    try {
      const view = await createView(viewData.name, viewData.description);
      for (const compData of viewData.components) {
        const componentId = components.get(compData.name);
        if (componentId) {
          await addComponentToView(view.id, componentId, compData.x, compData.y);
        }
      }
    } catch {
      console.log(`    (Skipping view — may already exist)`);
    }
  }
}
