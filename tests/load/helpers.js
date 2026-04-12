import http from 'k6/http';
import { check } from 'k6';
import { Rate } from 'k6/metrics';

// ============================================
// CONFIGURATION
// ============================================

export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
export const SERVICE_NAME = __ENV.SERVICE_NAME || '';
export const SERVICE_KEY = __ENV.SERVICE_KEY || '';

// Custom metric: tracks assertion failures across all smoke groups.
// Smoke threshold: errors rate == 0 (any failure = test failure).
export const errorRate = new Rate('errors');

// ============================================
// HEADERS
// ============================================

export function baseHeaders() {
  const h = { 'Content-Type': 'application/json' };
  if (SERVICE_NAME && SERVICE_KEY) {
    h['X-Service-Name'] = SERVICE_NAME;
    h['X-Service-Key'] = SERVICE_KEY;
  }
  return h;
}

export function headersWithIdempotency(idempotencyKey) {
  const h = baseHeaders();
  h['X-Idempotency-Key'] = idempotencyKey;
  return h;
}

// ============================================
// HTTP HELPERS
// ============================================

export function post(url, body, headers) {
  return http.post(url, JSON.stringify(body), { headers: headers || baseHeaders() });
}

export function get(url, headers) {
  return http.get(url, { headers: headers || baseHeaders() });
}

export function put(url, body, headers) {
  return http.put(url, JSON.stringify(body), { headers: headers || baseHeaders() });
}

export function del(url, headers) {
  return http.del(url, null, { headers: headers || baseHeaders() });
}

// ============================================
// RESPONSE PARSING
// ============================================

export function parseData(res) {
  try {
    const body = JSON.parse(res.body);
    return body.data || null;
  } catch (_) {
    return null;
  }
}

export function parseErrorMessage(res) {
  try {
    const body = JSON.parse(res.body);
    return (body.errors && body.errors.message) || null;
  } catch (_) {
    return null;
  }
}

// ============================================
// UUID
// ============================================

export function uuid() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

// ============================================
// ASSERTIONS (with automatic error tracking)
// ============================================

export function assertStatus(res, expected, label) {
  const passed = check(res, {
    [label]: (r) => r.status === expected,
  });
  if (!passed) {
    errorRate.add(1);
    console.error(`FAIL: ${label} — got ${res.status}, expected ${expected}`);
  } else {
    errorRate.add(0);
  }
  return passed;
}

export function assertField(res, field, expected, label) {
  const data = parseData(res);
  const passed = check(res, {
    [label]: () => data && data[field] === expected,
  });
  if (!passed) {
    errorRate.add(1);
    console.error(`FAIL: ${label} — field "${field}" got ${data ? data[field] : 'null'}, expected ${expected}`);
  } else {
    errorRate.add(0);
  }
  return passed;
}

export function assertErrorContains(res, substring, label) {
  const msg = parseErrorMessage(res);
  const passed = check(res, {
    [label]: () => msg && msg.includes(substring),
  });
  if (!passed) {
    errorRate.add(1);
    console.error(`FAIL: ${label} — error message "${msg}" does not contain "${substring}"`);
  } else {
    errorRate.add(0);
  }
  return passed;
}

export function assertFieldExists(res, field, label) {
  const data = parseData(res);
  const passed = check(res, {
    [label]: () => data && data[field] !== undefined && data[field] !== null && data[field] !== '',
  });
  if (!passed) {
    errorRate.add(1);
    console.error(`FAIL: ${label} — field "${field}" is missing or empty`);
  } else {
    errorRate.add(0);
  }
  return passed;
}
