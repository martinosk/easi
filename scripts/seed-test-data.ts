#!/usr/bin/env npx tsx

/**
 * Test Data Seeding Script
 *
 * Populates the test database with realistic enterprise architecture data
 * by calling the backend API: a full L1–L4 capability tree, 300 applications
 * each realizing a mix of capability levels, business domains, enterprise
 * capabilities, views, fit scores, and origin entities.
 *
 * AUTHENTICATION MODES:
 *
 * 1. Session Cookie (local dev with DEX) — RECOMMENDED:
 *    Log in at http://localhost:3000, open DevTools > Application > Cookies,
 *    copy "easi_session" and run:
 *      npm run seed -- --cookie "your_session_cookie_value"
 *
 * 2. Bypass Mode (CI/testing):
 *    Start backend with AUTH_MODE=bypass, then run:
 *      npm run seed -- --bypass --tenant-id acme
 */

import { BASE_URL, SESSION_COOKIE, BYPASS_MODE, TENANT_ID } from "./seed/config.ts";
import { seedComponents, seedRelations } from "./seed/components.ts";
import { seedCapabilities, seedCapabilityDependencies, seedSystemRealizations } from "./seed/capabilities.ts";
import { seedBusinessDomains, seedEnterpriseCapabilities } from "./seed/domains.ts";
import { seedViews } from "./seed/views.ts";
import { seedApplicationFitScores } from "./seed/fit-scores.ts";
import { seedOriginEntities } from "./seed/origins.ts";
import { seedArchitectureDirection } from "./seed/architecture-direction.ts";

if (!SESSION_COOKIE && !BYPASS_MODE) {
  console.log(`
Usage: npm run seed -- [options]

Authentication (choose one):
  --cookie <value>    Session cookie from browser (local dev with DEX)
  --bypass            X-Tenant-ID header (requires AUTH_MODE=bypass on backend)

Options:
  --base-url <url>    Backend URL (default: http://localhost:8080)
  --tenant-id <id>    Tenant ID for bypass mode (default: acme)

Examples:
  npm run seed -- --cookie "MTcz..."
  npm run seed -- --bypass --tenant-id acme

To get session cookie:
  1. Open http://localhost:3000 and log in with testuser@acme.com / password
  2. Open DevTools (F12) > Application > Cookies > localhost
  3. Copy the value of "easi_session"
`);
  process.exit(0);
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
  console.log(`Base URL:  ${BASE_URL}`);
  if (SESSION_COOKIE) {
    console.log(`Auth Mode: Session Cookie (DEX)`);
    console.log(`Cookie:    ${SESSION_COOKIE.substring(0, 20)}...`);
  } else {
    console.log(`Auth Mode: Bypass (X-Tenant-ID)`);
    console.log(`Tenant ID: ${TENANT_ID}`);
  }
  console.log("");

  console.log("Checking API health...");
  if (!(await checkApiHealth())) {
    console.error(`❌ API is not reachable at ${BASE_URL}/health`);
    process.exit(1);
  }
  console.log("✅ API is healthy\n");

  try {
    const components = await seedComponents();
    await seedRelations(components);

    const capabilities = await seedCapabilities();

    const domains = await seedBusinessDomains(capabilities);
    await seedEnterpriseCapabilities();
    await seedCapabilityDependencies(capabilities);
    await seedSystemRealizations(capabilities, components);
    await seedArchitectureDirection({ capabilities, components, domains });

    await seedViews(components);
    await seedApplicationFitScores(components);
    await seedOriginEntities(components);

    console.log("\n✅ Seeding complete!");
    console.log("\nSummary:");
    console.log(`  ${components.size} applications created`);
    console.log(`  ${capabilities.size} capabilities created (L1–L4)`);
    console.log(`  Business domains, enterprise capabilities, and views created`);
    console.log(`  System realizations and capability dependencies linked`);
    console.log(`  TIME assessments, realization roles, and capability journeys created (specs 180–183)`);
    console.log(`  Fit scores set for strategic pillars`);
    console.log(`  Origin entities (acquired entities, vendors, internal teams) created`);
  } catch (error) {
    console.error("\n❌ Seeding failed:", error);
    process.exit(1);
  }
}

main();
