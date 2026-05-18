/**
 * 客户取消订单功能测试脚本
 * 测试 feat/cancel-order 分支实现的功能
 *
 * 使用方法: node scripts/test-cancel-order.js
 * 前置条件: 服务已启动，数据库已迁移，管理员种子已存在
 */

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

const colors = { reset: '\x1b[0m', green: '\x1b[32m', red: '\x1b[31m', cyan: '\x1b[36m', yellow: '\x1b[33m' };
let passed = 0, failed = 0;
const tokens = {};

function log(type, msg) {
  const t = new Date().toISOString().split('T')[1].slice(0, 8);
  const p = { PASS: `${colors.green}✓ PASS${colors.reset}`, FAIL: `${colors.red}✗ FAIL${colors.reset}`, INFO: `${colors.cyan}ℹ INFO${colors.reset}`, TEST: `${colors.yellow}▶ TEST${colors.reset}` };
  console.log(`[${t}] ${p[type]||'      '} ${msg}`);
}
function section(title) { console.log(`\n${'═'.repeat(50)}\n  ${title}\n${'═'.repeat(50)}`); }

async function api(method, path, token, body) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  const res = await fetch(`${BASE_URL}/api/v1${path}`, { method, headers, body: body ? JSON.stringify(body) : undefined });
  return { status: res.status, data: await res.json() };
}

function uid(prefix) { return `${prefix}_test_${Date.now()}`; }

function futureDate(daysFromNow) {
  const d = new Date(); d.setDate(d.getDate() + daysFromNow); return d.toISOString().split('T')[0];
}

// =============== 测试用例 ===============

async function login() {
  section('准备: 获取测试 token');

  const admin = await api('POST', '/auth/admin-login', null, { username: 'admin', password: 'admin123' });
  if (admin.data.code === 0) {
    tokens.admin = admin.data.data.access_token;
    log('PASS', '管理员登录成功');
    passed++;
  } else { log('FAIL', `管理员登录失败: ${JSON.stringify(admin.data)}`); failed++; return false; }

  const guestName = uid('guest');
  const reg = await api('POST', '/auth/register', null, { username: guestName, password: 'test1234', name: '测试住户' });
  if (reg.data.code === 0) {
    tokens.guestUname = guestName;
    tokens.guestPwd = 'test1234';
    const login = await api('POST', '/auth/login', null, { username: guestName, password: 'test1234' });
    if (login.data.code === 0) {
      tokens.guest = login.data.data.access_token;
      log('PASS', '住户注册并登录成功');
      passed++;
    } else { log('FAIL', '住户登录失败'); failed++; return false; }
  } else { log('FAIL', '住户注册失败'); failed++; return false; }

  return true;
}

let createdRooms = [];
let createdOrders = [];

async function createRooms(count = 5) {
  section('准备: 创建测试房间');
  let ok = 0;
  for (let i = 0; i < count; i++) {
    const room = await api('POST', '/rooms', tokens.admin, {
      room_number: `CR${Date.now()}${i}`.slice(-8),
      type: 'standard', capacity: 2, floor: 1, price_per_night: 300, status: 'vacant',
    });
    if (room.data.code === 0) {
      createdRooms.push(room.data.data.id);
      ok++;
    }
  }
  if (ok > 0) { log('PASS', `创建 ${ok} 个房间成功`); passed++; return true; }
  log('FAIL', '创建房间失败'); failed++; return false;
}

async function testAutoCancel() {
  section('测试 1: 自动取消（距入住 > 24h）');

  // 创建一个很久以后的订单
  const order = await api('POST', '/orders', tokens.guest, {
    check_in_date: futureDate(30), check_out_date: futureDate(32), total_price: 600, guest_count: 1,
  });
  if (order.data.code !== 0) { log('FAIL', `创建订单失败: ${JSON.stringify(order.data)}`); failed++; return; }
  createdOrders.push(order.data.data.id);
  log('PASS', `创建订单 #${order.data.data.id}（入住日期: ${futureDate(30)}）`);
  passed++;

  // 自动取消
  const cancel = await api('POST', `/orders/${order.data.data.id}/cancel`, tokens.guest, { reason: '行程有变' });
  if (cancel.data.code === 0 && cancel.data.data.auto_cancelled === true && cancel.data.data.status === 'cancelled') {
    log('PASS', `自动取消成功，status=${cancel.data.data.status}, auto_cancelled=${cancel.data.data.auto_cancelled}`);
    passed++;
  } else {
    log('FAIL', `自动取消失败: ${JSON.stringify(cancel.data)}`);
    failed++;
  }
}

async function testManualReview() {
  section('测试 2: 提交审核（距入住 <= 24h）');

  const order = await api('POST', '/orders', tokens.guest, {
    check_in_date: futureDate(0),  // 今天入住，必定 <= 24h
    check_out_date: futureDate(2), total_price: 600, guest_count: 1,
  });
  if (order.data.code !== 0) { log('FAIL', `创建订单失败: ${JSON.stringify(order.data)}`); failed++; return; }
  createdOrders.push(order.data.data.id);
  log('PASS', `创建订单 #${order.data.data.id}（入住日期: ${futureDate(0)}）`);
  passed++;

  const cancel = await api('POST', `/orders/${order.data.data.id}/cancel`, tokens.guest, { reason: '临时有事' });
  if (cancel.data.code === 0 && cancel.data.data.auto_cancelled === false && cancel.data.data.status === 'cancel_requested') {
    log('PASS', `进入待审核状态，status=${cancel.data.data.status}`);
    passed++;
  } else {
    log('FAIL', `预期 cancel_requested，实际: ${JSON.stringify(cancel.data)}`);
    failed++;
    return;
  }

  // 验证待审列表中能看到
  const cancelRequests = await api('GET', '/orders/cancel-requests', tokens.admin);
  if (cancelRequests.data.code === 0) {
    const found = cancelRequests.data.data.list.find(o => o.id === order.data.data.id);
    if (found) {
      log('PASS', '待审列表中存在该订单');
      passed++;
    } else {
      log('FAIL', '待审列表中未找到该订单');
      failed++;
    }
  } else {
    log('FAIL', `获取待审列表失败: ${JSON.stringify(cancelRequests.data)}`);
    failed++;
  }

  return order.data.data.id;  // 返回订单ID供后续审批测试
}

async function testRejectPastDate() {
  section('测试 3: 入住当天及之后禁止取消');

  const order = await api('POST', '/orders', tokens.guest, {
    check_in_date: futureDate(-1),  // 昨天
    check_out_date: futureDate(1), total_price: 600, guest_count: 1,
  });
  if (order.data.code !== 0) { log('FAIL', `创建订单失败: ${JSON.stringify(order.data)}`); failed++; return; }
  createdOrders.push(order.data.data.id);
  passed++;

  const cancel = await api('POST', `/orders/${order.data.data.id}/cancel`, tokens.guest, { reason: '想取消' });
  if (cancel.data.code !== 0) {
    log('PASS', `入住日期已过的订单被拒绝取消（code=${cancel.data.code}）`);
    passed++;
  } else {
    log('FAIL', `预期拒绝，实际成功: ${JSON.stringify(cancel.data)}`);
    failed++;
  }
}

async function testForbidden() {
  section('测试 4: 非本人订单禁止取消');

  // 管理员创建一个订单
  const order = await api('POST', '/orders', tokens.admin, {
    check_in_date: futureDate(10), check_out_date: futureDate(12), total_price: 600, guest_count: 1,
  });
  if (order.data.code !== 0) { log('FAIL', `创建订单失败`); failed++; return; }
  createdOrders.push(order.data.data.id);
  passed++;

  // 住户试图取消管理员的订单
  const cancel = await api('POST', `/orders/${order.data.data.id}/cancel`, tokens.guest, { reason: '别人的订单' });
  if (cancel.data.code === 403) {
    log('PASS', `非本人取消返回 403`);
    passed++;
  } else {
    log('FAIL', `预期 403，实际: ${JSON.stringify(cancel.data)}`);
    failed++;
  }
}

let cancelReqOrderId;

async function testStaffApprove() {
  section('测试 5: 员工审批通过');

  if (!cancelReqOrderId) {
    // 创建一个新的待审订单
    const order = await api('POST', '/orders', tokens.guest, {
      check_in_date: futureDate(0), check_out_date: futureDate(2), total_price: 600, guest_count: 1,
    });
    if (order.data.code !== 0) { log('FAIL', '创建订单失败'); failed++; return; }
    createdOrders.push(order.data.data.id);

    const cancel = await api('POST', `/orders/${order.data.data.id}/cancel`, tokens.guest, { reason: '急需审批' });
    cancelReqOrderId = order.data.data.id;
    passed++;
  }

  // 管理员审批通过 → cancelled
  const approve = await api('PUT', `/orders/${cancelReqOrderId}`, tokens.admin, { status: 'cancelled' });
  if (approve.data.code === 0 && approve.data.data.status === 'cancelled') {
    log('PASS', `审批通过，订单状态变为 cancelled`);
    passed++;
  } else {
    log('FAIL', `审批通过失败: ${JSON.stringify(approve.data)}`);
    failed++;
  }

  // 验证待审列表已移除
  const cancelRequests = await api('GET', '/orders/cancel-requests', tokens.admin);
  const found = cancelRequests.data.data.list.find(o => o.id === cancelReqOrderId);
  if (!found) {
    log('PASS', '审批通过后待审列表中已移除');
    passed++;
  } else {
    log('FAIL', '待审列表中仍存在');
    failed++;
  }
  cancelReqOrderId = null;
}

async function testStaffReject() {
  section('测试 6: 员工审批驳回');

  const order = await api('POST', '/orders', tokens.guest, {
    check_in_date: futureDate(0), check_out_date: futureDate(2), total_price: 600, guest_count: 1,
  });
  if (order.data.code !== 0) { log('FAIL', '创建订单失败'); failed++; return; }
  createdOrders.push(order.data.data.id);

  const cancel = await api('POST', `/orders/${order.data.data.id}/cancel`, tokens.guest, { reason: '试试驳回' });
  if (cancel.data.code !== 0) { log('FAIL', '提交取消申请失败'); failed++; return; }

  // 管理员驳回 → pending
  const reject = await api('PUT', `/orders/${order.data.data.id}`, tokens.admin, { status: 'pending' });
  if (reject.data.code === 0 && reject.data.data.status === 'pending') {
    log('PASS', `审批驳回，订单状态恢复为 pending`);
    passed++;
  } else {
    log('FAIL', `审批驳回失败: ${JSON.stringify(reject.data)}`);
    failed++;
  }

  // 已恢复的订单不应出现在待审列表
  const cancelRequests = await api('GET', '/orders/cancel-requests', tokens.admin);
  const found = cancelRequests.data.data.list.find(o => o.id === order.data.data.id);
  if (!found) {
    log('PASS', '驳回后待审列表中已移除');
    passed++;
  } else {
    log('FAIL', '待审列表中仍存在');
    failed++;
  }
}

async function testCancelNonPending() {
  section('测试 7: 非 pending 状态禁止取消');

  // 管理员创建一个订单并手动确认
  const order = await api('POST', '/orders', tokens.admin, {
    check_in_date: futureDate(10), check_out_date: futureDate(12), total_price: 600, guest_count: 1,
  });
  if (order.data.code !== 0) { log('FAIL', '创建订单失败'); failed++; return; }
  createdOrders.push(order.data.data.id);

  const upd = await api('PUT', `/orders/${order.data.data.id}`, tokens.admin, { status: 'confirmed' });
  if (upd.data.code !== 0) { log('FAIL', '确认订单失败'); failed++; return; }

  // 尝试取消已确认的订单
  const cancel = await api('POST', `/orders/${order.data.data.id}/cancel`, tokens.admin, { reason: '想取消已确认订单' });
  if (cancel.data.code === 4002) {
    log('PASS', `非 pending 订单无法取消（code=${cancel.data.code}）`);
    passed++;
  } else {
    log('FAIL', `预期 4002，实际: ${JSON.stringify(cancel.data)}`);
    failed++;
  }
}

async function testGuestAccessForbidden() {
  section('测试 8: 住户无法访问待审列表');

  const res = await api('GET', '/orders/cancel-requests', tokens.guest);
  if (res.status === 403 || res.data.code === 403) {
    log('PASS', '住户无法访问待审列表');
    passed++;
  } else {
    log('FAIL', `预期 403，实际: ${res.status}`);
    failed++;
  }
}

// =============== 清理 ===============

async function cleanup() {
  section('清理: 删除测试数据');

  // 删除创建的订单
  for (const id of createdOrders.reverse()) {
    await api('DELETE', `/orders/${id}`, tokens.admin);
  }
  // 删除创建的测试房间
  for (const id of createdRooms) {
    await api('DELETE', `/rooms/${id}`, tokens.admin);
  }
  // 删除注册的测试住户
  const guests = await api('GET', '/guests', tokens.admin);
  const guest = guests.data?.data?.list?.find(g => g.user?.username === tokens.guestUname);
  if (guest) {
    await api('DELETE', `/guests/${guest.id}`, tokens.admin);
  }
  log('PASS', `清理完成（订单 ${createdOrders.length} 条，房间 ${createdRooms.length} 间，住户 1 个）`);
  passed++;
}

// =============== 主流程 ===============

async function main() {
  console.log('\n╔══════════════════════════════════════════════╗');
  console.log('║     取消订单功能测试 - feat/cancel-order     ║');
  console.log('╚══════════════════════════════════════════════╝\n');

  try {
    if (!await login()) { printSummary(); process.exit(1); }
    await createRooms(10);
    await testAutoCancel();
    cancelReqOrderId = await testManualReview();
    await testRejectPastDate();
    await testForbidden();
    await testStaffApprove();
    await testStaffReject();
    await testCancelNonPending();
    await testGuestAccessForbidden();
  } catch (err) {
    log('FAIL', `测试异常: ${err.message}`);
  }

  await cleanup();
  printSummary();
}

function printSummary() {
  section('测试结果汇总');
  console.log(`  通过: ${colors.green}${passed}${colors.reset}`);
  console.log(`  失败: ${colors.red}${failed}${colors.reset}`);
  console.log(`  总计: ${passed + failed}\n`);
  process.exit(failed > 0 ? 1 : 0);
}

main();
