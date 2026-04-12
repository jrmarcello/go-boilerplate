import { group, sleep } from 'k6';
import { Trend } from 'k6/metrics';

import {
  BASE_URL,
  SERVICE_NAME,
  errorRate,
  post,
  get,
  put,
  del,
  parseData,
  parseErrorMessage,
  assertStatus,
  assertField,
  assertFieldExists,
  uuid,
} from './helpers.js';

// ============================================
// CUSTOM METRICS
// ============================================

const createUserDuration = new Trend('create_user_duration');
const getUserDuration = new Trend('get_user_duration');
const listUsersDuration = new Trend('list_users_duration');

// ============================================
// LOCAL HELPERS
// ============================================

function randomEmail() {
  return `loadtest_${uuid().substring(0, 8)}@example.com`;
}

function randomName() {
  return `Load Test User ${uuid().substring(0, 8)}`;
}

// ============================================
// CRUD OPERATIONS
// ============================================

export function createUser() {
  const res = post(`${BASE_URL}/users`, {
    name: randomName(),
    email: randomEmail(),
  });

  createUserDuration.add(res.timings.duration);
  return res;
}

export function getUser(id) {
  const res = get(`${BASE_URL}/users/${id}`);

  getUserDuration.add(res.timings.duration);
  return res;
}

export function listUsers(page, limit) {
  const res = get(`${BASE_URL}/users?page=${page}&limit=${limit}`);

  listUsersDuration.add(res.timings.duration);
  return res;
}

export function updateUser(id) {
  const res = put(`${BASE_URL}/users/${id}`, {
    name: randomName(),
    email: randomEmail(),
  });

  return res;
}

export function deleteUser(id) {
  const res = del(`${BASE_URL}/users/${id}`);

  return res;
}

export function healthCheck() {
  const res = get(`${BASE_URL}/health`);

  return res;
}

// ============================================
// SMOKE GROUPS
// ============================================

// TC-S-01: GET /health returns 200 + status "ok"
// TC-S-02: GET /ready returns 200
export function smokeHealthCheck() {
  group('01 - Health Check', function () {
    // TC-S-01: health endpoint
    const healthRes = get(`${BASE_URL}/health`);
    assertStatus(healthRes, 200, 'TC-S-01: health returns 200');
    assertField(healthRes, 'status', 'ok', 'TC-S-01: health status is ok');

    // TC-S-02: ready endpoint
    const readyRes = get(`${BASE_URL}/ready`);
    assertStatus(readyRes, 200, 'TC-S-02: ready returns 200');
  });
  sleep(0.1);
}

// TC-S-03: request without service key returns 401 (if auth configured)
export function smokeAuthErrors() {
  group('02 - Auth Errors', function () {
    if (!SERVICE_NAME) {
      console.log('TC-S-03: SKIP — SERVICE_NAME not set, auth not configured');
      return;
    }

    // TC-S-03: request without service key returns 401
    const headers = { 'Content-Type': 'application/json' };
    const res = get(`${BASE_URL}/users`, headers);
    assertStatus(res, 401, 'TC-S-03: missing service key returns 401');
  });
  sleep(0.1);
}

// TC-S-04 to TC-S-08: full CRUD lifecycle
export function smokeUserCRUD() {
  group('03 - User CRUD Lifecycle', function () {
    // TC-S-04: create user returns 201 with id and created_at
    const createRes = createUser();
    assertStatus(createRes, 201, 'TC-S-04: create returns 201');
    assertFieldExists(createRes, 'id', 'TC-S-04: create returns id');
    assertFieldExists(createRes, 'created_at', 'TC-S-04: create returns created_at');

    const data = parseData(createRes);
    if (!data || !data.id) {
      console.error('TC-S-04: SKIP remaining — create did not return id');
      return;
    }
    const userId = data.id;

    // TC-S-05: get user returns 200
    const getRes = getUser(userId);
    assertStatus(getRes, 200, 'TC-S-05: get returns 200');

    // TC-S-07: update user returns 200
    const updateRes = updateUser(userId);
    assertStatus(updateRes, 200, 'TC-S-07: update returns 200');

    // TC-S-08: delete user returns 200
    const deleteRes = deleteUser(userId);
    assertStatus(deleteRes, 200, 'TC-S-08: delete returns 200');
  });
  sleep(0.1);
}

// TC-S-06: create 2 users, list with page=1&limit=10, verify 200 + data array
export function smokeUserList() {
  group('04 - User List', function () {
    // Create 2 users to ensure data exists
    createUser();
    createUser();

    // TC-S-06: list users
    const listRes = listUsers(1, 10);
    assertStatus(listRes, 200, 'TC-S-06: list returns 200');

    const data = parseData(listRes);
    const hasArray = data !== null && Array.isArray(data);
    if (!hasArray) {
      errorRate.add(1);
      console.error('TC-S-06: FAIL — data is not an array');
    } else {
      errorRate.add(0);
    }
  });
  sleep(0.1);
}

// TC-S-09: create with invalid email returns 400
// TC-S-10: get with "invalid-uuid" returns 400
export function smokeValidationErrors() {
  group('05 - Validation Errors', function () {
    // TC-S-09: invalid email
    const invalidEmailRes = post(`${BASE_URL}/users`, {
      name: randomName(),
      email: 'not-an-email',
    });
    assertStatus(invalidEmailRes, 400, 'TC-S-09: invalid email returns 400');

    // TC-S-10: invalid UUID format
    const invalidUUIDRes = getUser('invalid-uuid');
    assertStatus(invalidUUIDRes, 400, 'TC-S-10: invalid UUID returns 400');
  });
  sleep(0.1);
}

// TC-S-11: get nonexistent UUID returns 404
// TC-S-12: create duplicate email returns 409
export function smokeBusinessErrors() {
  group('06 - Business Errors', function () {
    // TC-S-11: get nonexistent user returns 404
    const notFoundRes = getUser('018e4a2c-6b4d-7000-9410-abcdef123456');
    assertStatus(notFoundRes, 404, 'TC-S-11: nonexistent user returns 404');

    // TC-S-12: duplicate email returns 409
    const email = randomEmail();
    const firstRes = post(`${BASE_URL}/users`, {
      name: randomName(),
      email: email,
    });
    assertStatus(firstRes, 201, 'TC-S-12: first create succeeds');

    const duplicateRes = post(`${BASE_URL}/users`, {
      name: randomName(),
      email: email,
    });
    assertStatus(duplicateRes, 409, 'TC-S-12: duplicate email returns 409');
  });
  sleep(0.1);
}

// TC-S-13: verify error responses use {"errors":{"message":...}} format
export function smokeResponseFormat() {
  group('07 - Response Format', function () {
    // TC-S-13: trigger an error and verify the response structure
    const res = post(`${BASE_URL}/users`, {
      name: '',
      email: 'bad-email',
    });

    const errorMsg = parseErrorMessage(res);
    const hasErrorFormat = errorMsg !== null;
    if (!hasErrorFormat) {
      errorRate.add(1);
      console.error('TC-S-13: FAIL — response does not use {"errors":{"message":...}} format');
    } else {
      errorRate.add(0);
    }
    assertStatus(res, 400, 'TC-S-13: validation error returns 400');
  });
  sleep(0.1);
}

// ============================================
// LOAD OPERATIONS
// ============================================

export function loadUserOperations() {
  const rand = Math.random();

  if (rand < 0.4) {
    // 40% - reads (list + get)
    const listRes = listUsers(1, 10);
    const listData = parseData(listRes);

    if (listData && Array.isArray(listData) && listData.length > 0) {
      const randomIndex = Math.floor(Math.random() * listData.length);
      const userId = listData[randomIndex].id;
      if (userId) {
        getUser(userId);
      }
    }
  } else if (rand < 0.7) {
    // 30% - creates
    createUser();
  } else if (rand < 0.9) {
    // 20% - updates (create + update)
    const createRes = createUser();
    const data = parseData(createRes);

    if (data && data.id) {
      updateUser(data.id);
    }
  } else {
    // 10% - deletes (create + delete)
    const createRes = createUser();
    const data = parseData(createRes);

    if (data && data.id) {
      deleteUser(data.id);
    }
  }
}
