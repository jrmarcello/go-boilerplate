import { sleep } from 'k6';

import { BASE_URL, get } from './helpers.js';
import {
  smokeHealthCheck,
  smokeAuthErrors,
  smokeUserCRUD,
  smokeUserList,
  smokeValidationErrors,
  smokeBusinessErrors,
  smokeResponseFormat,
  loadUserOperations,
  createUser,
  getUser,
  listUsers,
} from './users.js';
import { smokeRoleCRUD, smokeRoleErrors } from './roles.js';

// ============================================
// SCENARIO CONFIGURATION
// ============================================
// Usage: k6 run --env SCENARIO=smoke tests/load/main.js

const SCENARIO = __ENV.SCENARIO || 'smoke';

const allScenarios = {
  // Smoke: functional validation (1 VU, 1 iteration)
  smoke: {
    executor: 'per-vu-iterations',
    vus: 1,
    iterations: 1,
    exec: 'smokeTest',
    tags: { scenario: 'smoke' },
  },

  // Load: progressive ramp (up to 50 VUs)
  load: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '30s', target: 10 },
      { duration: '1m', target: 30 },
      { duration: '1m', target: 50 },
      { duration: '30s', target: 10 },
      { duration: '30s', target: 0 },
    ],
    exec: 'loadTest',
    tags: { scenario: 'load' },
  },

  // Stress: find system limits (up to 200 VUs)
  stress: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '30s', target: 50 },
      { duration: '30s', target: 100 },
      { duration: '30s', target: 150 },
      { duration: '30s', target: 200 },
      { duration: '30s', target: 0 },
    ],
    exec: 'stressTest',
    tags: { scenario: 'stress' },
  },

  // Spike: sudden burst (0 -> 100 instantly)
  spike: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '10s', target: 5 },
      { duration: '5s', target: 100 },
      { duration: '30s', target: 100 },
      { duration: '10s', target: 5 },
      { duration: '10s', target: 0 },
    ],
    exec: 'spikeTest',
    tags: { scenario: 'spike' },
  },
};

// Select only the specified scenario
const selectedScenario = {};
selectedScenario[SCENARIO] = allScenarios[SCENARIO];

export const options = {
  scenarios: selectedScenario,

  thresholds: {
    // Smoke: assertion-based — any failure = test failure
    errors: [SCENARIO === 'smoke' ? 'rate==0' : 'rate<0.01'],
    http_req_failed: [SCENARIO === 'smoke' ? 'rate<0.50' : 'rate<0.01'],
    http_req_duration: ['p(95)<500'],
    create_user_duration: ['p(95)<800'],
    get_user_duration: ['p(95)<200'],
    list_users_duration: ['p(95)<300'],
  },
};

// ============================================
// SETUP / TEARDOWN
// ============================================

const startTime = Date.now();

export function setup() {
  console.log(`[setup] Scenario: ${SCENARIO}`);
  console.log(`[setup] Base URL: ${BASE_URL}`);

  const healthRes = get(`${BASE_URL}/health`);
  if (healthRes.status !== 200) {
    console.error(`[setup] Health check failed: ${healthRes.status}`);
  } else {
    console.log('[setup] Health check: OK');
  }
}

export function teardown() {
  const durationSec = ((Date.now() - startTime) / 1000).toFixed(1);
  console.log(`[teardown] Completed in ${durationSec}s`);
}

// ============================================
// SCENARIO EXECUTORS
// ============================================

// Smoke: runs ALL validation groups sequentially (1 VU, 1 iteration)
export function smokeTest() {
  // User domain
  smokeHealthCheck();
  smokeAuthErrors();
  smokeUserCRUD();
  smokeUserList();
  smokeValidationErrors();
  smokeBusinessErrors();
  smokeResponseFormat();

  // Role domain
  smokeRoleCRUD();
  smokeRoleErrors();
}

// Load: progressive traffic with read-heavy distribution
export function loadTest() {
  loadUserOperations();
  sleep(0.5);
}

// Stress: heavy write load
export function stressTest() {
  const rand = Math.random();
  if (rand < 0.5) {
    createUser();
  } else if (rand < 0.8) {
    listUsers(1, 10);
  } else {
    const created = createUser();
    try {
      const data = JSON.parse(created.body);
      if (data.data && data.data.id) {
        getUser(data.data.id);
      }
    } catch (_) { /* ignore parse errors during stress */ }
  }
  sleep(0.3);
}

// Spike: burst traffic (create + get + list)
export function spikeTest() {
  const created = createUser();
  try {
    const data = JSON.parse(created.body);
    if (data.data && data.data.id) {
      getUser(data.data.id);
    }
  } catch (_) { /* ignore parse errors during spike */ }
  listUsers(1, 5);
  sleep(0.1);
}
