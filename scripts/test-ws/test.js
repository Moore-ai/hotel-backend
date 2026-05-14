/**
 * WebSocket 实时通知测试脚本
 *
 * 用法:
 *   1. 确保后端已启动 (go run main.go)
 *   2. 安装依赖: cd scripts/test-ws && npm install
 *   3. 运行: node test.js [options]
 *
 * 选项:
 *   --host     API 地址，默认 http://localhost:8080
 *   --user     登录用户名，默认 admin
 *   --pass     登录密码，默认 admin123
 *   --timeout  监听超时(秒)，默认 60
 *
 * 测试流程:
 *   1. 登录获取 token
 *   2. 连接 WebSocket
 *   3. 创建测试场景（住户 + 房间 + 入住）
 *   4. 监听 WebSocket 消息
 */

const fetch = globalThis.fetch;
import { WebSocket } from 'ws';

// 解析参数
const args = {};
process.argv.slice(2).forEach((arg) => {
  const m = arg.match(/^--(\w+)=(.+)/);
  if (m) args[m[1]] = m[2];
});

const HOST = args.host || 'http://localhost:8080';
const WS_HOST = HOST.replace(/^http/, 'ws');
const USER = args.user || 'admin';
const PASS = args.pass || 'admin123';
const TIMEOUT = parseInt(args.timeout || '60', 10);

let token = '';
let ws = null;
let stepCount = 0;

async function step(msg, fn) {
  stepCount++;
  process.stdout.write(`\n[${stepCount}] ${msg} ... `);
  try {
    await fn();
    console.log('✓');
  } catch (err) {
    console.log(`✗ 失败: ${err.message}`);
    cleanup();
    process.exit(1);
  }
}

function api(path, options = {}) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return fetch(`${HOST}/api/v1${path}`, { ...options, headers }).then(async (r) => {
    const body = await r.json();
    if (body.code !== 0) throw new Error(`${r.status} ${JSON.stringify(body)}`);
    return body.data;
  });
}

function cleanup() {
  if (ws) {
    try { ws.close(); } catch (_) {}
    ws = null;
  }
}

async function waitForScheduler() {
  console.log(`\n   等待调度器检测超时入住（${TIMEOUT}s 内）...`);
  const start = Date.now();
  return new Promise((resolve) => {
    const timeout = setTimeout(() => {
      console.log(`   超时：${TIMEOUT}s 内未收到通知`);
      resolve(false);
    }, TIMEOUT * 1000);

    ws.on('message', (raw) => {
      try {
        const msg = JSON.parse(raw.toString());
        if (msg.type !== 'notification') {
          console.log(`\n   收到未知消息:`, JSON.stringify(msg));
          return;
        }
        const n = msg.data;
        console.log(`\n   ┌─ 收到 WebSocket 推送`);
        console.log(`   │  类型: ${n.type}`);
        console.log(`   │  标题: ${n.title}`);
        console.log(`   │  内容: ${n.content}`);
        console.log(`   │  时间: ${n.CreatedAt || n.created_at}`);
        console.log(`   └─ 耗时: ${((Date.now() - start) / 1000).toFixed(1)}s`);
        clearTimeout(timeout);
        resolve(true);
      } catch (_) {}
    });
  });
}

(async () => {
  console.log(`
╔══════════════════════════════════════╗
║   WebSocket 实时通知功能测试          ║
║   服务器: ${HOST.padEnd(31)}║
║   用户: ${(USER + '@' + HOST).padEnd(33)}║
╚══════════════════════════════════════╝`);

  // 1. 登录
  await step('登录获取 token', async () => {
    const data = await api('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username: USER, password: PASS }),
    });
    token = data.access_token;
    if (!token) throw new Error('未获取到 access_token');
  });

  // 2. 连接 WebSocket
  await step('连接 WebSocket', async () => {
    ws = new WebSocket(`${WS_HOST}/api/v1/ws?token=${token}`);
    await new Promise((resolve, reject) => {
      ws.on('open', resolve);
      ws.on('error', reject);
      ws.on('unexpected-response', (_, res) =>
        reject(new Error(`HTTP ${res.statusCode}`))
      );
      // 一秒内未连接成功也报错
      setTimeout(() => reject(new Error('连接超时')), 3000);
    });
    console.log('WebSocket 已连接');
  });

  // 3. 创建测试数据
  let guestUser, room;

  await step('创建测试住户', async () => {
    guestUser = await api('/users', {
      method: 'POST',
      body: JSON.stringify({
        username: `test_guest_${Date.now()}`,
        password: 'test1234',
        role: 'guest',
        name: '测试住户',
      }),
    });
    console.log(`ID=${guestUser.id} Name=${guestUser.name}`);
  });

  await step('创建测试房间', async () => {
    room = await api('/rooms', {
      method: 'POST',
      body: JSON.stringify({
        room_number: `T${Date.now() % 10000}`,
        room_type: '标准间',
        price: 100,
        status: 'available',
      }),
    });
    console.log(`ID=${room.id} 房号=${room.room_number}`);
  });

  await step('创建超时入住记录', async () => {
    // expected_checkout_time 设为 10 分钟前，调度器下次扫描时就会触发
    const past = new Date(Date.now() - 10 * 60 * 1000).toISOString();
    const checkin = await api('/checkins', {
      method: 'POST',
      body: JSON.stringify({
        user_id: guestUser.id,
        room_id: room.id,
        expected_checkout_time: past,
      }),
    });
    console.log(`入住 ID=${checkin.id}`);
  });

  // 4. 等待通知
  await step('等待 WebSocket 推送', async () => {
    const received = await waitForScheduler();
    if (!received) {
      console.log('\n   提示: 调度器默认每 300s 扫描一次.');
      console.log('   如需快速测试，可在 config.yaml 中调整:');
      console.log('     checkout.scheduler_interval: 30');
      console.log('   然后重启服务并重新运行此脚本.');
    }
  });

  // 5. 清理
  cleanup();

  console.log(`\n✅ 测试完成`);
})().catch((err) => {
  console.error(`\n❌ 脚本异常: ${err.message}`);
  cleanup();
  process.exit(1);
});
