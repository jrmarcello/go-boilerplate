import { group } from 'k6';
import { check } from 'k6';
import grpc from 'k6/net/grpc';
import { Rate } from 'k6/metrics';

// ============================================
// CONFIGURATION
// ============================================

const GRPC_ADDR = __ENV.GRPC_ADDR || 'localhost:50051';
const errorRate = new Rate('grpc_errors');

// ============================================
// PROTO LOADING
// ============================================

const client = new grpc.Client();
client.load(
  ['../../proto'],
  'appmax/user/v1/user_service.proto',
  'appmax/role/v1/role_service.proto',
);

// ============================================
// HELPERS
// ============================================

function uuid() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

function randomEmail() {
  return `grpc_smoke_${uuid().substring(0, 8)}@example.com`;
}

function randomName() {
  return `gRPC Smoke User ${uuid().substring(0, 8)}`;
}

function grpcCheck(response, label, checks) {
  const passed = check(response, checks);
  if (!passed) {
    errorRate.add(1);
    console.error(`FAIL: ${label} — status: ${response.status}, error: ${response.error}`);
  } else {
    errorRate.add(0);
  }
  return passed;
}

// ============================================
// SMOKE TEST GROUPS
// ============================================

// TC-S-G-01: CreateUser + GetUser round-trip
export function smokeGRPCUserCRUD() {
  group('gRPC: User CRUD round-trip', function () {
    client.connect(GRPC_ADDR, { plaintext: true });

    // Create
    const name = randomName();
    const email = randomEmail();
    const createRes = client.invoke('appmax.user.v1.UserService/CreateUser', {
      name: name,
      email: email,
    });
    grpcCheck(createRes, 'gRPC CreateUser', {
      'CreateUser status OK': (r) => r && r.status === grpc.StatusOK,
      'CreateUser has id': (r) => r && r.message && r.message.id !== '',
      'CreateUser has created_at': (r) => r && r.message && r.message.createdAt !== '',
    });

    if (createRes.status !== grpc.StatusOK) {
      client.close();
      return;
    }

    const userId = createRes.message.id;

    // Get
    const getRes = client.invoke('appmax.user.v1.UserService/GetUser', {
      id: userId,
    });
    grpcCheck(getRes, 'gRPC GetUser', {
      'GetUser status OK': (r) => r && r.status === grpc.StatusOK,
      'GetUser has correct name': (r) => r && r.message && r.message.user && r.message.user.name === name,
      'GetUser has correct email': (r) => r && r.message && r.message.user && r.message.user.email === email,
      'GetUser is active': (r) => r && r.message && r.message.user && r.message.user.active === true,
    });

    // Delete (cleanup)
    const deleteRes = client.invoke('appmax.user.v1.UserService/DeleteUser', {
      id: userId,
    });
    grpcCheck(deleteRes, 'gRPC DeleteUser', {
      'DeleteUser status OK': (r) => r && r.status === grpc.StatusOK,
    });

    client.close();
  });
}

// TC-S-G-02: CreateRole + ListRoles
export function smokeGRPCRoleCRUD() {
  group('gRPC: Role CRUD round-trip', function () {
    client.connect(GRPC_ADDR, { plaintext: true });

    // Create
    const roleName = `gRPC Smoke Role ${uuid().substring(0, 8)}`;
    const createRes = client.invoke('appmax.role.v1.RoleService/CreateRole', {
      name: roleName,
    });
    grpcCheck(createRes, 'gRPC CreateRole', {
      'CreateRole status OK': (r) => r && r.status === grpc.StatusOK,
      'CreateRole has id': (r) => r && r.message && r.message.id !== '',
    });

    if (createRes.status !== grpc.StatusOK) {
      client.close();
      return;
    }

    const roleId = createRes.message.id;

    // List
    const listRes = client.invoke('appmax.role.v1.RoleService/ListRoles', {
      page: 1,
      limit: 50,
    });
    grpcCheck(listRes, 'gRPC ListRoles', {
      'ListRoles status OK': (r) => r && r.status === grpc.StatusOK,
      'ListRoles has roles': (r) => r && r.message && r.message.roles && r.message.roles.length > 0,
    });

    // Delete (cleanup)
    const deleteRes = client.invoke('appmax.role.v1.RoleService/DeleteRole', {
      id: roleId,
    });
    grpcCheck(deleteRes, 'gRPC DeleteRole', {
      'DeleteRole status OK': (r) => r && r.status === grpc.StatusOK,
    });

    client.close();
  });
}

// TC-S-G-03: Request without auth metadata (when auth enabled)
export function smokeGRPCAuthError() {
  group('gRPC: Auth error (no metadata)', function () {
    // This test only validates when SERVICE_KEYS_ENABLED=true on the server.
    // When auth is disabled, calls succeed — skip validation in that case.
    client.connect(GRPC_ADDR, { plaintext: true });

    const res = client.invoke('appmax.user.v1.UserService/ListUsers', {
      page: 1,
      limit: 10,
    });

    // If auth is enabled, we expect Unauthenticated. If disabled, OK is fine.
    const authEnabled = __ENV.SERVICE_KEYS_ENABLED === 'true';
    if (authEnabled) {
      grpcCheck(res, 'gRPC Auth Error', {
        'Unauthenticated without metadata': (r) =>
          r && r.status === grpc.StatusUnauthenticated,
      });
    } else {
      console.log('gRPC auth test skipped (SERVICE_KEYS_ENABLED != true)');
      errorRate.add(0);
    }

    client.close();
  });
}

// TC-S-G-04: CreateUser with invalid email
export function smokeGRPCValidationError() {
  group('gRPC: Validation error (invalid email)', function () {
    client.connect(GRPC_ADDR, { plaintext: true });

    const res = client.invoke('appmax.user.v1.UserService/CreateUser', {
      name: 'Valid Name',
      email: 'not-an-email',
    });

    grpcCheck(res, 'gRPC Validation Error', {
      'InvalidArgument for bad email': (r) =>
        r && r.status === grpc.StatusInvalidArgument,
    });

    client.close();
  });
}
