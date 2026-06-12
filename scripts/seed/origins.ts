import { apiCall } from "./config.ts";
import type { AcquiredEntity, Vendor, InternalTeam } from "./types.ts";

async function tryLinkComponents(
  componentNames: string[],
  components: Map<string, string>,
  linkFn: (componentId: string) => Promise<void>
): Promise<void> {
  for (const name of componentNames) {
    const componentId = components.get(name);
    if (!componentId) continue;
    try {
      await linkFn(componentId);
    } catch {
      console.log(`    (Skipping link — may already exist)`);
    }
  }
}

async function seedAcquiredEntities(components: Map<string, string>): Promise<void> {
  const acquiredEntities = [
    {
      name: "DataTech Solutions",
      acquisitionDate: "2023-06-15",
      integrationStatus: "COMPLETED",
      notes: "Acquired for analytics capabilities. Integration completed Q4 2023.",
      components: ["Analytics Platform", "Recommendation Engine", "Data Warehouse"],
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
      components: ["API Gateway", "Cache Layer", "Service Mesh"],
    },
    {
      name: "RetailTech Corp",
      acquisitionDate: "2021-09-01",
      integrationStatus: "COMPLETED",
      notes: "Legacy retail systems acquisition. Fully integrated into e-commerce platform.",
      components: ["Product Catalog", "Inventory Service", "Warehouse Management System"],
    },
    {
      name: "SearchFirst AI",
      acquisitionDate: "2023-11-01",
      integrationStatus: "IN_PROGRESS",
      notes: "Acquired for AI-powered search capabilities. Integrating ML models.",
      components: ["Search Engine", "Recommendation Engine", "Personalization Engine"],
    },
  ];

  for (const ae of acquiredEntities) {
    try {
      const entity = await apiCall<AcquiredEntity>("POST", "/acquired-entities", {
        name: ae.name,
        acquisitionDate: ae.acquisitionDate,
        integrationStatus: ae.integrationStatus,
        notes: ae.notes,
      });
      await tryLinkComponents(ae.components, components, (compId) =>
        apiCall("PUT", `/components/${compId}/origin/acquired-via`, {
          acquiredEntityId: entity.id,
          notes: `Acquired from ${ae.name}`,
        })
      );
    } catch {
      console.log(`    (Skipping acquired entity — may already exist)`);
    }
  }
}

async function seedVendors(components: Map<string, string>): Promise<void> {
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
      components: ["Message Queue", "Event Bridge", "CDN Platform"],
    },
    {
      name: "Twilio",
      implementationPartner: "",
      notes: "SMS and communication APIs for customer notifications.",
      components: ["Notification Service", "SMS Gateway"],
    },
    {
      name: "Stripe",
      implementationPartner: "FinTech Partners",
      notes: "Payment processing integration. PCI DSS compliant.",
      components: ["Payment Gateway"],
    },
    {
      name: "Salesforce",
      implementationPartner: "CRM Consultants Ltd",
      notes: "CRM integration for customer data synchronization.",
      components: ["Customer Portal", "CRM Core Platform"],
    },
    {
      name: "Snowflake",
      implementationPartner: "Data Solutions Group",
      notes: "Cloud data warehouse for analytics workloads.",
      components: ["Data Warehouse"],
    },
    {
      name: "HashiCorp",
      implementationPartner: "",
      notes: "Infrastructure tooling for secrets and IaC management.",
      components: ["Secrets Manager", "Infrastructure as Code Platform"],
    },
  ];

  for (const v of vendors) {
    try {
      const vendor = await apiCall<Vendor>("POST", "/vendors", {
        name: v.name,
        implementationPartner: v.implementationPartner,
        notes: v.notes,
      });
      await tryLinkComponents(v.components, components, (compId) =>
        apiCall("PUT", `/components/${compId}/origin/purchased-from`, {
          vendorId: vendor.id,
          notes: `Purchased from ${v.name}`,
        })
      );
    } catch {
      console.log(`    (Skipping vendor — may already exist)`);
    }
  }
}

async function seedInternalTeams(components: Map<string, string>): Promise<void> {
  const internalTeams = [
    {
      name: "Core Platform Team",
      department: "Engineering",
      contactPerson: "Jane Smith",
      notes: "Owns core microservices and platform infrastructure.",
      components: ["User Service", "Order Service", "Inventory Service", "API Gateway"],
    },
    {
      name: "Customer Experience Team",
      department: "Product",
      contactPerson: "John Doe",
      notes: "Owns customer-facing applications and user journey.",
      components: ["Customer Portal", "Shopping Cart", "Customer Onboarding Hub"],
    },
    {
      name: "Data Engineering Team",
      department: "Engineering",
      contactPerson: "Alice Johnson",
      notes: "Builds data pipelines, analytics, and ML infrastructure.",
      components: ["Reporting Service", "Analytics Platform", "Recommendation Engine", "ETL Pipeline Service"],
    },
    {
      name: "Operations Team",
      department: "IT Operations",
      contactPerson: "Bob Williams",
      notes: "Manages internal tools and operational systems.",
      components: ["Admin Dashboard", "Pricing Service", "Shipping Service", "IT Service Management System"],
    },
    {
      name: "Security Team",
      department: "Engineering",
      contactPerson: "Carol Chen",
      notes: "Responsible for security, fraud prevention, and compliance.",
      components: ["Fraud Detection", "Identity Provider", "Single Sign-On Service"],
    },
    {
      name: "Content Team",
      department: "Marketing",
      contactPerson: "David Lee",
      notes: "Manages product content and digital assets.",
      components: ["Content Management", "Product Catalog", "Digital Asset Manager"],
    },
    {
      name: "Payments Team",
      department: "Engineering",
      contactPerson: "Eve Martinez",
      notes: "Owns payment processing, billing, and financial integrations.",
      components: ["Payment Gateway", "Billing Platform", "Payment Orchestrator"],
    },
    {
      name: "Infrastructure Team",
      department: "Engineering",
      contactPerson: "Frank Wilson",
      notes: "Owns cloud infrastructure, platform tooling, and developer experience.",
      components: ["Service Mesh", "Kubernetes Orchestrator", "Load Balancer", "Container Registry"],
    },
  ];

  for (const team of internalTeams) {
    try {
      const internalTeam = await apiCall<InternalTeam>("POST", "/internal-teams", {
        name: team.name,
        department: team.department,
        contactPerson: team.contactPerson,
        notes: team.notes,
      });
      await tryLinkComponents(team.components, components, (compId) =>
        apiCall("PUT", `/components/${compId}/origin/built-by`, {
          internalTeamId: internalTeam.id,
          notes: `Built by ${team.name}`,
        })
      );
    } catch {
      console.log(`    (Skipping internal team — may already exist)`);
    }
  }
}

export async function seedOriginEntities(components: Map<string, string>): Promise<void> {
  console.log("\n🏭 Seeding Origin Entities...");
  await seedAcquiredEntities(components);
  await seedVendors(components);
  await seedInternalTeams(components);
}
