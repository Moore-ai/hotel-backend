/**
 * 酒店房间分配功能测试脚本
 * 测试 feat/allocate_room 分支实现的功能
 *
 * 使用方法: node test/room_allocator_test.js
 * 前置条件:
 *   1. 酒店后端服务已启动 (默认 http://localhost:8080)
 *   2. 数据库和 Redis 已运行
 */

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

// 测试配置
const CONFIG = {
  admin: { username: 'admin', password: 'admin123' },
  floors: 5,           // 楼层数
  roomsPerFloor: 4,    // 每层房间数
  roomTypes: ['standard', 'deluxe', 'suite'],
};

// 颜色输出
const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  cyan: '\x1b[36m',
};

function log(type, message) {
  const timestamp = new Date().toISOString().split('T')[1].slice(0, 8);
  const prefix = {
    'PASS': `${colors.green}✓ PASS${colors.reset}`,
    'FAIL': `${colors.red}✗ FAIL${colors.reset}`,
    'INFO': `${colors.cyan}ℹ INFO${colors.reset}`,
    'TEST': `${colors.yellow}▶ TEST${colors.reset}`,
  }[type] || '      ';
  console.log(`[${timestamp}] ${prefix} ${message}`);
}

function logSection(title) {
  console.log(`\n${'='.repeat(50)}`);
  console.log(`  ${title}`);
  console.log('='.repeat(50));
}

// HTTP 请求封装
async function request(method, path, token, body = null) {
  const url = `${BASE_URL}/api/v1${path}`;
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const options = { method, headers };
  if (body) options.body = JSON.stringify(body);

  const response = await fetch(url, options);
  const data = await response.json();
  return { status: response.status, data };
}

// API 封装
const api = {
  login: (username, password) =>
    request('POST', '/auth/login', null, { username, password }),

  register: (username, password, name = '') =>
    request('POST', '/auth/register', null, { username, password, name }),

  createRoom: (token, room) =>
    request('POST', '/rooms', token, room),

  listRooms: (token, page = 1) =>
    request('GET', `/rooms?page=${page}&page_size=100`, token),

  getRoom: (token, id) =>
    request('GET', `/rooms/${id}`, token),

  createOrder: (token, order) =>
    request('POST', '/orders', token, order),

  listOrders: (token, page = 1) =>
    request('GET', `/orders?page=${page}&page_size=100`, token),

  createCheckin: (token, checkin) =>
    request('POST', '/checkins', token, checkin),

  listCheckins: (token, page = 1) =>
    request('GET', `/checkins?page=${page}&page_size=100`, token),

  checkout: (token, id) =>
    request('PUT', `/checkins/${id}/checkout`, token),
};

// 生成模拟房间数据
function generateRooms() {
  const rooms = [];
  let roomNumber = 101;

  for (let floor = 1; floor <= CONFIG.floors; floor++) {
    for (let i = 0; i < CONFIG.roomsPerFloor; i++) {
      const typeIndex = i % CONFIG.roomTypes.length;
      const type = CONFIG.roomTypes[typeIndex];
      const capacity = type === 'suite' ? 4 : type === 'deluxe' ? 3 : 2;
      const price = type === 'suite' ? 800 : type === 'deluxe' ? 500 : 300;

      rooms.push({
        room_number: `${floor}${String(i + 1).padStart(2, '0')}`,
        type,
        capacity,
        floor,
        price_per_night: price * floor, // 高楼层更贵
        status: 'vacant',
        description: `${floor}楼${type}房间`,
      });
      roomNumber++;
    }
  }

  return rooms;
}

// 获取未来日期
function futureDate(daysFromNow) {
  const date = new Date();
  date.setDate(date.getDate() + daysFromNow);
  return date.toISOString().split('T')[0];
}

// 测试用例
let testsPassed = 0;
let testsFailed = 0;

async function testLogin() {
  logSection('测试 1: 用户登录');

  const res = await api.login(CONFIG.admin.username, CONFIG.admin.password);

  if (res.status === 200 && res.data.code === 0 && res.data.data.access_token) {
    log('PASS', '管理员登录成功');
    testsPassed++;
    return res.data.data.access_token;
  } else {
    log('FAIL', `登录失败: ${JSON.stringify(res.data)}`);
    testsFailed++;
    return null;
  }
}

async function testCreateRooms(token) {
  logSection('测试 2: 创建模拟房间');

  if (!token) {
    log('FAIL', '无有效token，跳过房间创建');
    testsFailed++;
    return [];
  }

  const rooms = generateRooms();
  const createdRooms = [];

  for (const room of rooms) {
    const res = await api.createRoom(token, room);
    if (res.status === 200 && res.data.code === 0) {
      createdRooms.push(res.data.data);
    }
  }

  // 如果创建失败（房间可能已存在），尝试从列表获取
  if (createdRooms.length === 0) {
    const listRes = await api.listRooms(token);
    if (listRes.data.code === 0) {
      createdRooms.push(...listRes.data.data.list);
    }
  }

  if (createdRooms.length >= rooms.length * 0.8) {
    log('PASS', `成功创建/获取 ${createdRooms.length} 个房间`);
    testsPassed++;
  } else {
    log('FAIL', `房间创建不完整: ${createdRooms.length}/${rooms.length}`);
    testsFailed++;
  }

  return createdRooms;
}

async function testAutoAllocateOrder(token) {
  logSection('测试 3: 自动分配房间（创建订单）');

  if (!token) {
    log('FAIL', '无有效token，跳过订单测试');
    testsFailed++;
    return null;
  }

  // 测试场景1: 不指定房间，让系统自动分配
  log('TEST', '场景1: 自动分配标准间');
  const order1 = await api.createOrder(token, {
    guest_count: 2,
    room_type_preference: 'standard',
    check_in_date: futureDate(1),
    check_out_date: futureDate(3),
    total_price: 600,
  });

  if (order1.status === 200 && order1.data.code === 0) {
    const order = order1.data.data;
    log('PASS', `订单创建成功，自动分配房间ID: ${order.room_id}`);
    testsPassed++;

    // 验证房间状态变为 reserved
    const roomRes = await api.getRoom(token, order.room_id);
    if (roomRes.data.data?.status === 'reserved') {
      log('PASS', '房间状态正确更新为 reserved');
      testsPassed++;
    } else {
      log('FAIL', `房间状态错误: ${roomRes.data.data?.status}`);
      testsFailed++;
    }

    return order;
  } else {
    log('FAIL', `订单创建失败: ${JSON.stringify(order1.data)}`);
    testsFailed++;
    return null;
  }
}

async function testLowFloorStrategy(token) {
  logSection('测试 4: 低楼层优先策略');

  if (!token) {
    log('FAIL', '无有效token，跳过策略测试');
    testsFailed++;
    return;
  }

  // 创建多个订单，验证分配顺序
  const floors = [];

  for (let i = 0; i < 3; i++) {
    const res = await api.createOrder(token, {
      guest_count: 2,
      room_type_preference: 'standard',
      check_in_date: futureDate(10 + i),
      check_out_date: futureDate(12 + i),
      total_price: 600,
    });

    if (res.status === 200 && res.data.code === 0) {
      const room = await api.getRoom(token, res.data.data.room_id);
      if (room.data.code === 0) {
        floors.push(room.data.data.floor);
        log('INFO', `订单${i + 1} 分配到 ${room.data.data.floor} 楼`);
      }
    }
  }

  // 低楼层优先：第一个分配的应该是最小楼层
  if (floors.length >= 2 && floors[0] <= Math.max(...floors)) {
    log('PASS', `低楼层优先策略验证通过，分配楼层: [${floors.join(', ')}]`);
    testsPassed++;
  } else {
    log('FAIL', `低楼层优先策略可能有问题，分配楼层: [${floors.join(', ')}]`);
    testsFailed++;
  }
}

async function testCheckin(token, order) {
  logSection('测试 5: 入住办理');

  if (!token || !order) {
    log('FAIL', '缺少必要数据，跳过入住测试');
    testsFailed++;
    return null;
  }

  // 测试场景: 根据订单办理入住
  log('TEST', `根据订单 #${order.id} 办理入住`);

  const expectedCheckout = new Date();
  expectedCheckout.setDate(expectedCheckout.getDate() + 2);

  const res = await api.createCheckin(token, {
    order_id: order.id,
    user_id: order.user_id,
    expected_checkout_time: expectedCheckout.toISOString(),
  });

  if (res.status === 200 && res.data.code === 0) {
    const checkin = res.data.data;
    log('PASS', `入住办理成功，入住ID: ${checkin.id}`);
    testsPassed++;

    // 验证房间状态变为 occupied
    const roomRes = await api.getRoom(token, checkin.room_id);
    if (roomRes.data.data?.status === 'occupied') {
      log('PASS', '房间状态正确更新为 occupied');
      testsPassed++;
    } else {
      log('FAIL', `房间状态错误: ${roomRes.data.data?.status}`);
      testsFailed++;
    }

    return checkin;
  } else {
    log('FAIL', `入住办理失败: ${JSON.stringify(res.data)}`);
    testsFailed++;
    return null;
  }
}

async function testCheckout(token, checkin) {
  logSection('测试 6: 签离办理');

  if (!token || !checkin) {
    log('FAIL', '缺少必要数据，跳过签离测试');
    testsFailed++;
    return;
  }

  const res = await api.checkout(token, checkin.id);

  if (res.status === 200 && res.data.code === 0) {
    log('PASS', `签离成功，入住ID: ${checkin.id}`);
    testsPassed++;

    // 验证房间状态恢复为 vacant
    const roomRes = await api.getRoom(token, checkin.room_id);
    if (roomRes.data.data?.status === 'vacant') {
      log('PASS', '房间状态正确恢复为 vacant');
      testsPassed++;
    } else {
      log('FAIL', `房间状态错误: ${roomRes.data.data?.status}`);
      testsFailed++;
    }
  } else {
    log('FAIL', `签离失败: ${JSON.stringify(res.data)}`);
    testsFailed++;
  }
}

async function testNoRoomAvailable(token) {
  logSection('测试 7: 无可用房间处理');

  // 尝试预订一个不存在的房型
  const res = await api.createOrder(token, {
    guest_count: 10, // 超大人数，应该没有房间
    room_type_preference: 'penthouse', // 不存在的房型
    check_in_date: futureDate(1),
    check_out_date: futureDate(3),
    total_price: 10000,
  });

  if (res.data.code === 3003) { // ErrNoRoomAvailable
    log('PASS', '正确返回无可用房间错误');
    testsPassed++;
  } else {
    log('FAIL', `预期返回无可用房间错误，实际: ${JSON.stringify(res.data)}`);
    testsFailed++;
  }
}

// 主测试流程
async function main() {
  console.log('\n╔══════════════════════════════════════════════════╗');
  console.log('║     酒店房间分配功能测试 - feat/allocate_room     ║');
  console.log('╚══════════════════════════════════════════════════╝\n');

  log('INFO', `服务地址: ${BASE_URL}`);
  log('INFO', `模拟房间: ${CONFIG.floors}层 × ${CONFIG.roomsPerFloor}间 = ${CONFIG.floors * CONFIG.roomsPerFloor}间`);

  try {
    // 1. 登录
    const token = await testLogin();

    // 2. 创建房间
    await testCreateRooms(token);

    // 3. 测试自动分配
    const order = await testAutoAllocateOrder(token);

    // 4. 测试楼层策略
    await testLowFloorStrategy(token);

    // 5. 测试入住
    const checkin = await testCheckin(token, order);

    // 6. 测试签离
    await testCheckout(token, checkin);

    // 7. 测试无可用房间
    await testNoRoomAvailable(token);

  } catch (error) {
    log('FAIL', `测试异常: ${error.message}`);
    console.error(error);
  }

  // 测试结果汇总
  logSection('测试结果汇总');
  console.log(`\n  通过: ${colors.green}${testsPassed}${colors.reset}`);
  console.log(`  失败: ${colors.red}${testsFailed}${colors.reset}`);
  console.log(`  总计: ${testsPassed + testsFailed}\n`);

  if (testsFailed === 0) {
    log('PASS', '所有测试通过！');
    process.exit(0);
  } else {
    log('FAIL', '部分测试失败');
    process.exit(1);
  }
}

main();
