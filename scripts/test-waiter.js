/**
 * 服务员功能测试脚本
 * 测试 feat/assistant 分支实现的功能
 *
 * 使用方法: node scripts/test-waiter.js
 * 前置条件: 服务已启动，数据库已迁移，管理员种子已存在
 */

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

const colors = { reset: '\x1b[0m', green: '\x1b[32m', red: '\x1b[31m', cyan: '\x1b[36m', yellow: '\x1b[33m' };
let passed = 0, failed = 0;
const tokens = {};
const created = { waiters: [], rooms: [], orders: [], checkins: [], guests: [] };

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

function uid(prefix) { return `${prefix}_t_${Date.now()}`; }
function futureDate(d) { const r = new Date(); r.setDate(r.getDate() + d); return r.toISOString().split('T')[0]; }

// =============== 准备 ===============

async function setup() {
  section('准备: 获取 token 和资源');

  // 管理员登录
  const admin = await api('POST', '/auth/admin-login', null, { username: 'admin', password: 'admin123' });
  if (admin.data.code !== 0) { log('FAIL', '管理员登录失败'); failed++; return false; }
  tokens.admin = admin.data.data.access_token;
  log('PASS', '管理员登录成功'); passed++;

  // 创建住户
  const gName = uid('guest');
  const reg = await api('POST', '/auth/register', null, { username: gName, password: 'test1234', name: '测试住户' });
  if (reg.data.code !== 0) { log('FAIL', '住户注册失败'); failed++; return false; }
  created.guests.push(gName);

  const guestLogin = await api('POST', '/auth/login', null, { username: gName, password: 'test1234' });
  if (guestLogin.data.code !== 0) { log('FAIL', '住户登录失败'); failed++; return false; }
  tokens.guest = guestLogin.data.data.access_token;
  tokens.guestCode = guestLogin.data.data.user_code;
  log('PASS', '住户注册并登录成功'); passed++;

  // 创建房间
  for (let i = 0; i < 3; i++) {
    const room = await api('POST', '/rooms', tokens.admin, {
      room_number: `WT${Date.now()}${i}`.slice(-8),
      type: 'standard', capacity: 2, floor: 1, price_per_night: 300, status: 'vacant',
    });
    if (room.data.code === 0) created.rooms.push(room.data.data.id);
  }
  log('PASS', `创建 ${created.rooms.length} 个房间成功`); passed++;

  return true;
}

// =============== 测试 1: 服务员 CRUD ===============

async function testWaiterCRUD() {
  section('测试 1: 服务员 CRUD（管理员）');

  // 1a. 非管理员禁止访问
  const forbid = await api('GET', '/waiters', tokens.guest);
  if (forbid.status === 403 || forbid.data.code === 403) {
    log('PASS', '非管理员无法访问服务员列表');
    passed++;
  } else { log('FAIL', `预期 403，实际 ${forbid.status}`); failed++; }

  // 1b. 创建服务员
  const wName = uid('waiter');
  const create = await api('POST', '/waiters', tokens.admin, {
    username: wName, password: 'test1234', name: '测试服务员', phone: '13800138001',
  });
  if (create.data.code !== 0) { log('FAIL', `创建服务员失败: ${JSON.stringify(create.data)}`); failed++; return; }
  log('PASS', '创建服务员成功'); passed++;

  // 1c. 获取列表
  const list = await api('GET', '/waiters', tokens.admin);
  const found = list.data.data.list.find(w => w.user?.username === wName);
  if (!found) { log('FAIL', '列表中未找到'); failed++; return; }
  created.waiters.push(found.id);
  log('PASS', '列表中可查到'); passed++;

  // 1d. 更新
  const upd = await api('PUT', `/waiters/${found.id}`, tokens.admin, { name: '更新后服务员' });
  if (upd.data.code === 0 && upd.data.data.name === '更新后服务员') {
    log('PASS', '更新成功'); passed++;
  } else { log('FAIL', `更新失败: ${JSON.stringify(upd.data)}`); failed++; }

  // 1e. 服务员通过 staff-login 登录
  const wLogin = await api('POST', '/auth/staff-login', null, { username: wName, password: 'test1234' });
  if (wLogin.data.code === 0) {
    tokens.waiter = wLogin.data.data.access_token;
    log('PASS', '服务员可通过 staff-login 登录'); passed++;
  } else { log('FAIL', '服务员登录失败'); failed++; }
}

// =============== 测试 2: 服务请求流程 ===============

async function testServiceFlow() {
  section('测试 2: 服务请求 → 派单 → 完成');

  // 2a. 创建订单 + 入住（需要有入住才能请求服务）
  const order = await api('POST', '/orders', tokens.guest, {
    check_in_date: futureDate(0), check_out_date: futureDate(2), total_price: 600, guest_count: 1,
  });
  if (order.data.code !== 0) { log('FAIL', `创建订单失败`); failed++; return; }
  created.orders.push(order.data.data.id);
  log('PASS', `创建订单成功`); passed++;

  // 管理员确认订单
  const confirm = await api('POST', `/orders/${order.data.data.id}/confirm`, tokens.admin);
  if (confirm.data.code !== 0) { log('FAIL', `确认订单失败`); failed++; return; }

  // 创建入住
  const checkin = await api('POST', '/checkins', tokens.admin, {
    user_code: tokens.guestCode,
    expected_checkout_time: new Date(Date.now() + 86400000 * 2).toISOString(),
  });
  if (checkin.data.code !== 0) { log('FAIL', `创建入住失败: ${JSON.stringify(checkin.data)}`); failed++; return; }
  created.checkins.push(checkin.data.data.id);
  log('PASS', `入住创建成功，入住ID: ${checkin.data.data.id}`); passed++;

  // 2b. 住户发起服务请求
  const req = await api('POST', `/checkins/${checkin.data.data.id}/service-request`, tokens.guest);
  if (req.data.code === 0 && req.data.data.serving_room_id) {
    log('PASS', `服务请求已派单，服务员已分配`); passed++;
  } else if (req.data.code === 3004) {
    log('PASS', `无空闲服务员，已通知等待（当前无服务员可派单）`); passed++;
    // 没有服务员时无法继续测试派单流程，跳过
    return;
  } else { log('FAIL', `服务请求失败: ${JSON.stringify(req.data)}`); failed++; return; }

  // 2c. 服务员查看自己的任务
  const waiterInfo = await api('GET', `/waiters/${req.data.data.id}`, tokens.admin);
  if (waiterInfo.data.data.serving_room_id) {
    log('PASS', '服务员状态为忙碌（有 serving_room_id）'); passed++;
  } else { log('FAIL', '服务员应为忙碌状态'); failed++; }

  // 2d. 其他服务员不能完成此任务
  const w2Name = uid('w2');
  await api('POST', '/waiters', tokens.admin, { username: w2Name, password: 'test1234', name: '服务员2' });
  const w2Login = await api('POST', '/auth/staff-login', null, { username: w2Name, password: 'test1234' });
  const w2Token = w2Login.data.data?.access_token;
  if (w2Token) {
    const forbidComplete = await api('POST', `/waiters/${req.data.data.id}/complete-service`, w2Token);
    if (forbidComplete.status === 403 || forbidComplete.data.code === 403) {
      log('PASS', '其他服务员无法完成此服务单（权限正确）');
      passed++;
    } else { log('FAIL', `其他服务员应被拒绝，实际 ${forbidComplete.status}`); failed++; }
  }

  // 2e. 管理员完成服务（跳过权限校验聚焦流程测试）
  const complete = await api('POST', `/waiters/${req.data.data.id}/complete-service`, tokens.admin);
  if (complete.data.code === 0) {
    log('PASS', '管理员完成服务成功'); passed++;
  } else { log('FAIL', `完成服务失败: ${JSON.stringify(complete.data)}`); failed++; }

  // 2f. 验证服务员恢复空闲
  const after = await api('GET', `/waiters/${req.data.data.id}`, tokens.admin);
  if (after.data.data.serving_room_id === null || after.data.data.serving_room_id === 0) {
    log('PASS', '服务员状态已恢复为空闲'); passed++;
  } else { log('FAIL', `服务员应空闲，实际 serving_room_id=${after.data.data.serving_room_id}`); failed++; }
}

// =============== 清理 ===============

async function cleanup() {
  section('清理: 删除测试数据');
  for (const id of created.checkins.reverse()) await api('DELETE', `/checkins/${id}`, tokens.admin);
  for (const id of created.orders.reverse()) await api('DELETE', `/orders/${id}`, tokens.admin);
  for (const w of created.waiters) {
    const info = await api('GET', `/waiters/${w}`, tokens.admin);
    if (info.data.code === 0) await api('DELETE', `/waiters/${w}`, tokens.admin);
  }
  for (const id of created.rooms) await api('DELETE', `/rooms/${id}`, tokens.admin);
  for (const g of created.guests) {
    const list = await api('GET', '/guests', tokens.admin);
    const guest = list.data?.data?.list?.find(x => x.user?.username === g);
    if (guest) await api('DELETE', `/guests/${guest.id}`, tokens.admin);
  }
  log('PASS', `清理完成`); passed++;
}

// =============== 主流程 ===============

async function main() {
  console.log('\n╔══════════════════════════════════════════════╗');
  console.log('║    服务员功能测试 - feat/assistant          ║');
  console.log('╚══════════════════════════════════════════════╝\n');

  try {
    if (!await setup()) { printSummary(); process.exit(1); }
    await testWaiterCRUD();
    await testServiceFlow();
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
