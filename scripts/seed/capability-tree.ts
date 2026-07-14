import type { CapNode } from "./types.ts";

export const CAPABILITY_DOMAIN_NAMES = [
  "Customer",
  "Digital Commerce",
  "Supply Chain",
  "Finance",
  "Human Resources",
  "Marketing",
  "Sales",
  "Technology Platform",
  "Risk and Compliance",
  "Analytics and Intelligence",
];

const L1_DEFS: [string, string][] = [
  // Customer (10)
  ["Customer Acquisition", "Attract and convert prospects into paying customers"],
  ["Customer Onboarding", "Register, activate, and onboard new customers"],
  ["Customer Retention", "Maintain and grow existing customer relationships"],
  ["Customer Support", "Handle customer inquiries, issues, and service requests"],
  ["Customer Analytics", "Analyze customer behavior, preferences, and lifetime value"],
  ["Customer Segmentation", "Classify customers into meaningful groups for targeted engagement"],
  ["Customer Loyalty Management", "Design and operate loyalty programs and rewards"],
  ["Customer Communications", "Manage multi-channel outbound customer communications"],
  ["Customer Identity Management", "Manage customer identities, credentials, and access"],
  ["Customer Experience Design", "Design and optimize end-to-end customer journeys"],
  // Digital Commerce (10)
  ["Order Management", "Process and fulfill customer orders end-to-end"],
  ["Product Catalog Management", "Maintain product information, categories, and attributes"],
  ["Pricing and Promotions", "Set, manage, and optimize product pricing and offers"],
  ["Checkout and Payment", "Handle the checkout process and payment collection"],
  ["Returns and Refunds", "Process product returns, exchanges, and refunds"],
  ["Subscription Management", "Manage recurring subscriptions and billing cycles"],
  ["Marketplace Operations", "Operate multi-vendor marketplace capabilities"],
  ["Digital Merchandising", "Curate and present products for optimal conversion"],
  ["Commerce Analytics", "Measure and optimize commerce performance"],
  ["Cross-Channel Commerce", "Enable consistent commerce across online and offline channels"],
  // Supply Chain (10)
  ["Inventory Management", "Track and manage inventory levels across locations"],
  ["Procurement Management", "Source and purchase goods and services"],
  ["Supplier Relationship Management", "Manage supplier performance and collaboration"],
  ["Warehouse Management", "Operate inbound, storage, and outbound warehouse flows"],
  ["Transportation Management", "Plan and execute freight and transportation"],
  ["Demand Planning", "Forecast demand and optimize inventory positioning"],
  ["Supply Chain Analytics", "Monitor and optimize supply chain performance"],
  ["Quality Management", "Ensure product and process quality standards"],
  ["Trade Compliance", "Manage import, export, and regulatory compliance"],
  ["Reverse Logistics", "Handle returns, repairs, and end-of-life product flows"],
  // Finance (10)
  ["Accounts Payable", "Manage supplier invoices and outgoing payments"],
  ["Accounts Receivable", "Manage customer invoices and incoming payments"],
  ["Financial Reporting", "Produce accurate financial statements and disclosures"],
  ["Budgeting and Forecasting", "Plan, allocate, and track financial budgets"],
  ["Treasury Management", "Manage cash, liquidity, and financial risk"],
  ["Tax Compliance", "Calculate, file, and manage tax obligations"],
  ["Payment Operations", "Operate payment processing and settlement"],
  ["Revenue Management", "Recognize, track, and optimize revenue streams"],
  ["Cost Management", "Track, allocate, and reduce operational costs"],
  ["Financial Risk Management", "Identify and mitigate financial exposures"],
  // Human Resources (10)
  ["Talent Acquisition", "Source, attract, and hire qualified candidates"],
  ["Employee Onboarding", "Integrate new employees into the organization"],
  ["Performance Management", "Set goals, evaluate performance, and develop talent"],
  ["Learning and Development", "Provide training, upskilling, and career growth programs"],
  ["Compensation and Benefits", "Design and administer employee pay and benefits"],
  ["Workforce Analytics", "Analyze workforce data to inform HR decisions"],
  ["Workforce Planning", "Plan headcount, skills, and organizational structure"],
  ["Employee Engagement", "Foster a motivated and committed workforce"],
  ["Time and Attendance", "Track employee work hours and absence"],
  ["HR Compliance", "Ensure adherence to labor laws and HR regulations"],
  // Marketing (10)
  ["Campaign Management", "Plan, execute, and measure marketing campaigns"],
  ["Content Marketing", "Create and distribute valuable content to attract audiences"],
  ["Search and Discovery Marketing", "Optimize search engine presence and paid discovery"],
  ["Social Media Management", "Manage brand presence across social platforms"],
  ["Email and Messaging Marketing", "Run targeted email and messaging programs"],
  ["Brand Management", "Define, protect, and evolve brand identity"],
  ["Market Research and Intelligence", "Gather and analyze market and competitive data"],
  ["Event Marketing", "Plan and execute physical and virtual events"],
  ["Partner and Affiliate Marketing", "Manage co-marketing and affiliate programs"],
  ["Marketing Analytics", "Measure marketing effectiveness and ROI"],
  // Sales (10)
  ["Lead Management", "Capture, qualify, and route sales leads"],
  ["Opportunity Management", "Manage deals through the sales pipeline"],
  ["Account Management", "Grow and retain strategic customer accounts"],
  ["Sales Analytics", "Analyze sales performance and pipeline health"],
  ["Territory and Quota Management", "Assign territories and sales targets"],
  ["Sales Enablement", "Equip sales teams with tools, content, and training"],
  ["Contract Lifecycle Management", "Manage contracts from creation to renewal"],
  ["Configure-Price-Quote", "Configure complex products and generate accurate quotes"],
  ["Channel Sales Management", "Manage indirect sales through partners and resellers"],
  ["Sales Operations", "Support and optimize sales process and systems"],
  // Technology Platform (10)
  ["API and Integration Management", "Manage APIs and system integrations"],
  ["Data Platform Management", "Operate and govern data storage and processing"],
  ["Cloud Infrastructure Management", "Provision and manage cloud infrastructure"],
  ["Security and Identity Management", "Manage security controls and digital identities"],
  ["DevOps and Continuous Delivery", "Automate software build, test, and deployment"],
  ["Observability and Monitoring", "Monitor system health, performance, and reliability"],
  ["Developer Experience", "Improve developer productivity and tooling"],
  ["Architecture Governance", "Define and enforce architectural standards"],
  ["Technology Asset Management", "Track and optimize technology assets and licenses"],
  ["Platform Engineering", "Build and operate internal developer platforms"],
  // Risk and Compliance (10)
  ["Regulatory Compliance", "Meet and evidence regulatory requirements"],
  ["Data Privacy and Protection", "Protect personal data and comply with privacy laws"],
  ["Fraud Detection and Prevention", "Detect and prevent fraudulent activity"],
  ["Cybersecurity Management", "Protect systems and data from cyber threats"],
  ["Business Continuity Management", "Ensure operations continue through disruptions"],
  ["Audit and Internal Controls", "Maintain effective internal controls and audit processes"],
  ["Third-Party Risk Management", "Assess and mitigate vendor and partner risk"],
  ["Operational Risk Management", "Identify and manage operational risks"],
  ["Environmental Compliance", "Manage environmental impact and sustainability reporting"],
  ["Legal and Contract Management", "Manage legal matters and contractual obligations"],
  // Analytics and Intelligence (10)
  ["Business Intelligence", "Deliver data-driven insights for business decisions"],
  ["Predictive Analytics", "Forecast future outcomes using statistical models"],
  ["Real-Time Analytics", "Analyze streaming data for immediate insights"],
  ["Data Engineering and Pipelines", "Build and operate data ingestion and transformation"],
  ["AI and Machine Learning Platform", "Develop, train, and deploy machine learning models"],
  ["Experimentation and Testing", "Design and run controlled experiments"],
  ["Customer Intelligence", "Generate deep understanding of customer needs and behavior"],
  ["Operational Intelligence", "Monitor and optimize operational processes in real time"],
  ["Strategic Insights and Reporting", "Synthesize data into executive-level strategic insights"],
  ["Data Governance and Quality", "Ensure data is accurate, consistent, and trustworthy"],
];

const L2_TEMPLATES: [string, string][] = [
  ["Strategy and Planning", "Strategic direction, roadmaps, and resource planning"],
  ["Execution and Operations", "Day-to-day operational management and process execution"],
  ["Analytics and Reporting", "Performance measurement, dashboards, and insights"],
  ["Governance and Compliance", "Policies, controls, standards, and compliance management"],
];

const L3_BY_L2: Record<string, [string, string][]> = {
  "Strategy and Planning": [
    ["Roadmap Management", "Long-term planning and initiative prioritization"],
    ["Resource Allocation", "Budget and capacity assignment across initiatives"],
  ],
  "Execution and Operations": [
    ["Process Management", "Operational process definition, execution, and improvement"],
    ["Performance Optimization", "Continuous improvement of operational outcomes"],
  ],
  "Analytics and Reporting": [
    ["Metrics and Dashboards", "KPI definition, tracking, and visualization"],
    ["Insights and Analysis", "Data-driven analysis and actionable recommendations"],
  ],
  "Governance and Compliance": [
    ["Policy Management", "Policy definition, communication, and enforcement"],
    ["Audit and Controls", "Internal controls, audit processes, and evidence collection"],
  ],
};

const L4_TEMPLATES: [string, string][] = [
  ["Automation", "Automated execution, orchestration, and self-service tooling"],
];

function buildL4(l3Name: string, l3Suffix: string): CapNode[] {
  return L4_TEMPLATES.map(([l4Suffix, l4Desc]) => ({
    name: `${l3Name} - ${l4Suffix}`,
    description: `${l4Desc} for ${l3Suffix.toLowerCase()}`,
    level: "L4",
  }));
}

export function generateCapabilityTree(): CapNode[] {
  return L1_DEFS.map(([l1Name, l1Desc]) => ({
    name: l1Name,
    description: l1Desc,
    level: "L1",
    children: L2_TEMPLATES.map(([l2Suffix, l2Desc]) => {
      const l2Name = `${l1Name} - ${l2Suffix}`;
      return {
        name: l2Name,
        description: `${l2Desc} for ${l1Name.toLowerCase()}`,
        level: "L2",
        children: (L3_BY_L2[l2Suffix] ?? []).map(([l3Suffix, l3Desc]) => {
          const l3Name = `${l2Name} - ${l3Suffix}`;
          return {
            name: l3Name,
            description: `${l3Desc} within the ${l1Name.toLowerCase()} domain`,
            level: "L3",
            children: buildL4(l3Name, l3Suffix),
          };
        }),
      };
    }),
  }));
}
