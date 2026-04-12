import { group, sleep, check } from 'k6';
import {
  BASE_URL,
  post,
  get,
  del,
  parseData,
  assertStatus,
  assertFieldExists,
  assertErrorContains,
  uuid,
} from './helpers.js';

// ============================================
// LOCAL HELPERS
// ============================================

function createRole(name) {
  return post(`${BASE_URL}/roles`, { name });
}

function listRoles() {
  return get(`${BASE_URL}/roles`);
}

function deleteRole(id) {
  return del(`${BASE_URL}/roles/${id}`);
}

// ============================================
// SMOKE GROUPS
// ============================================

// TC-S-14, TC-S-15, TC-S-16: Role CRUD (create, list, delete)
export function smokeRoleCRUD() {
  let roleId;

  group('14 - Role CRUD', function () {
    // TC-S-14: Create role -> 201, has id and name
    const name = `Load Test Role ${uuid().substring(0, 8)}`;
    const createRes = createRole(name);
    assertStatus(createRes, 201, 'create role returns 201');
    assertFieldExists(createRes, 'id', 'create role has id');
    const createData = parseData(createRes);
    if (createData) roleId = createData.id;
  });

  sleep(0.1);

  group('15 - Role List', function () {
    // TC-S-15: List roles -> 200, data is array
    const listRes = listRoles();
    assertStatus(listRes, 200, 'list roles returns 200');
    // List response has {data: [...], meta: {...}} — verify data is present
    const body = JSON.parse(listRes.body);
    check(listRes, {
      'list roles has data array': () => Array.isArray(body.data),
    });
  });

  sleep(0.1);

  group('16 - Role Delete', function () {
    // TC-S-16: Delete role -> 200
    if (roleId) {
      const deleteRes = deleteRole(roleId);
      assertStatus(deleteRes, 200, 'delete role returns 200');
    }
  });

  sleep(0.1);
}

// TC-S-17: Role error scenarios
export function smokeRoleErrors() {
  group('17 - Role Errors', function () {
    // TC-S-17: Create role with duplicate name -> 409
    const name = `Dup Role ${uuid().substring(0, 8)}`;
    const firstRes = createRole(name);
    assertStatus(firstRes, 201, 'first role creation returns 201');

    const dupRes = createRole(name);
    assertStatus(dupRes, 409, 'duplicate role name returns 409');
    assertErrorContains(dupRes, 'already exists', 'duplicate role error message');

    // Cleanup: delete the created role
    const firstData = parseData(firstRes);
    if (firstData && firstData.id) {
      deleteRole(firstData.id);
    }
  });

  sleep(0.1);
}
