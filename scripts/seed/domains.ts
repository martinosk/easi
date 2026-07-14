import { apiCall } from "./config.ts";
import type { BusinessDomain, Capability, EnterpriseCapability } from "./types.ts";

async function createBusinessDomain(name: string, description: string): Promise<BusinessDomain> {
  return apiCall<BusinessDomain>("POST", "/business-domains", { name, description });
}

async function assignCapabilityToDomain(domainId: string, capabilityId: string): Promise<void> {
  await apiCall("POST", `/business-domains/${domainId}/capabilities`, { capabilityId });
}

interface DomainSeed {
  name: string;
  description: string;
  capabilities: string[];
}

const DOMAIN_DATA: DomainSeed[] = [
  {
    name: "E-Commerce",
    description: "Online retail and shopping experience",
    capabilities: ["Order Management", "Product Catalog Management", "Pricing and Promotions", "Checkout and Payment"],
  },
  {
    name: "Customer Experience",
    description: "Customer-facing services and support",
    capabilities: ["Customer Acquisition", "Customer Retention", "Customer Support", "Customer Experience Design"],
  },
  {
    name: "Payments & Finance",
    description: "Financial transactions and accounting",
    capabilities: ["Payment Operations", "Financial Reporting", "Revenue Management", "Tax Compliance"],
  },
  {
    name: "Logistics",
    description: "Inventory, warehousing, and shipping operations",
    capabilities: ["Inventory Management", "Warehouse Management", "Transportation Management", "Demand Planning"],
  },
  {
    name: "Marketing",
    description: "Marketing and promotional activities",
    capabilities: ["Campaign Management", "Content Marketing", "Marketing Analytics", "Brand Management"],
  },
  {
    name: "Data & Analytics",
    description: "Business intelligence and data science",
    capabilities: ["Business Intelligence", "Predictive Analytics", "Data Governance and Quality", "Real-Time Analytics"],
  },
  {
    name: "Technology Platform",
    description: "Core technology platform and infrastructure",
    capabilities: ["API and Integration Management", "Platform Engineering", "DevOps and Continuous Delivery", "Observability and Monitoring"],
  },
  {
    name: "Risk & Compliance",
    description: "Risk management and regulatory compliance",
    capabilities: ["Fraud Detection and Prevention", "Cybersecurity Management", "Regulatory Compliance", "Data Privacy and Protection"],
  },
  {
    name: "Human Resources",
    description: "People management and organizational development",
    capabilities: ["Talent Acquisition", "Performance Management", "Learning and Development", "Workforce Analytics"],
  },
  {
    name: "Sales",
    description: "Sales operations and revenue growth",
    capabilities: ["Lead Management", "Opportunity Management", "Sales Analytics", "Account Management"],
  },
];

export async function seedBusinessDomains(
  capabilities: Map<string, Capability>
): Promise<Map<string, BusinessDomain>> {
  console.log("\n🏢 Seeding Business Domains...");
  const domainsByName = new Map<string, BusinessDomain>();

  for (const d of DOMAIN_DATA) {
    try {
      const domain = await createBusinessDomain(d.name, d.description);
      domainsByName.set(d.name, domain);
      for (const capName of d.capabilities) {
        const capability = capabilities.get(capName);
        if (capability) {
          try {
            await assignCapabilityToDomain(domain.id, capability.id);
          } catch {
            console.log(`    (Skipping capability assignment — may already exist)`);
          }
        }
      }
    } catch {
      console.log(`    (Skipping domain — may already exist)`);
    }
  }

  return domainsByName;
}

export async function seedEnterpriseCapabilities(): Promise<void> {
  console.log("\n🏛️  Seeding Enterprise Capabilities...");

  const enterpriseCapabilities = [
    { name: "Customer Identity", description: "Enterprise-wide customer identity and access management", category: "Customer" },
    { name: "Order Management", description: "Enterprise order processing and fulfillment platform", category: "Operations" },
    { name: "Payment Platform", description: "Enterprise payment processing and settlement infrastructure", category: "Finance" },
    { name: "Data Platform", description: "Enterprise data management, governance, and analytics", category: "Technology" },
    { name: "Integration Platform", description: "Enterprise API and integration services backbone", category: "Technology" },
    { name: "Security Platform", description: "Enterprise security, identity, and compliance services", category: "Security" },
    { name: "Analytics Platform", description: "Enterprise business intelligence and reporting platform", category: "Technology" },
  ];

  for (const ec of enterpriseCapabilities) {
    try {
      await apiCall<EnterpriseCapability>("POST", "/enterprise-capabilities", {
        name: ec.name,
        description: ec.description,
        category: ec.category,
      });
    } catch {
      console.log(`    (Skipping enterprise capability — may already exist)`);
    }
  }
}
