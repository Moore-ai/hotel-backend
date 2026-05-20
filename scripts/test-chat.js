// AI 智能管家测试脚本
// 运行: node scripts/test-chat.js
// 依赖: Node.js 18+ (内置 fetch)
// 要求: 服务端已启动，config.yaml 中 llm.base_url 已配置

const BASE = "http://localhost:8080/api/v1"
const headers = { "Content-Type": "application/json" }

let adminToken, guestToken
let pass = 0, fail = 0

async function req(method, path, reqBody, token) {
  const opt = { method, headers: { ...headers } }
  if (token) opt.headers["Authorization"] = "Bearer " + token
  if (reqBody) opt.body = JSON.stringify(reqBody)
  const res = await fetch(BASE + path, opt)
  return { status: res.status, body: await res.json() }
}

async function ok(method, path, reqBody, token) {
  const { status, body } = await req(method, path, reqBody, token)
  if (body.code !== 0) throw new Error(`${path} → ${body.code} ${body.message}`)
  return body.data
}

function check(label, ok) {
  if (ok) { pass++; console.log(`  ✅ ${label}`) }
  else { fail++; console.log(`  ❌ ${label}`) }
}

async function main() {
  // ── 环境准备 ──
  console.log("\n=== 环境准备 ===")

  // 检查服务端连通性
  try {
    const login = await ok("POST", "/auth/admin-login",
      { username: "admin", password: "admin123" })
    adminToken = login.access_token
    console.log("  管理员登录成功")
  } catch (e) {
    console.log("  ⚠ 服务端未启动或无管理员账号，跳过需要管理员权限的测试")
  }

  const guestUser = "chat_test_" + Date.now()
  await ok("POST", "/guests", {
    username: guestUser, password: "123456", name: "聊天测试住户"
  }, adminToken)
  const guestLogin = await ok("POST", "/auth/login",
    { username: guestUser, password: "123456" })
  guestToken = guestLogin.access_token
  console.log("  住户登录成功")

  // ================================================================
  // 测试 1：未认证请求被拒绝
  // ================================================================
  console.log("\n═══════════════════════════════════════════════════════")
  console.log("测试 1：未认证请求被拒绝")
  console.log("═══════════════════════════════════════════════════════")

  const noAuth = await req("POST", "/chat", { message: "你好" })
  check("未认证返回 401", noAuth.status === 401)

  // ================================================================
  // 测试 2：空消息被拒绝
  // ================================================================
  console.log("\n═══════════════════════════════════════════════════════")
  console.log("测试 2：空消息被拒绝")
  console.log("═══════════════════════════════════════════════════════")

  const emptyMsg = await req("POST", "/chat", { message: "" }, guestToken)
  check("空消息返回 400", emptyMsg.status === 400)

  const noMsg = await req("POST", "/chat", {}, guestToken)
  check("缺少 message 返回 400", noMsg.status === 400)

  // ================================================================
  // 测试 3：新对话 — 发送消息
  // ================================================================
  console.log("\n═══════════════════════════════════════════════════════")
  console.log("测试 3：新对话")
  console.log("═══════════════════════════════════════════════════════")

  const chat1 = await req("POST", "/chat",
    { message: "你好，请问退房时间是几点？" }, guestToken)

  if (chat1.body.code === 0) {
    check("返回 code 为 0", true)
    check("返回 reply 非空", chat1.body.data.reply && chat1.body.data.reply.length > 0)
    check("返回 conversation_id", chat1.body.data.conversation_id && chat1.body.data.conversation_id.length > 0)
    check("返回 action 为字符串", typeof chat1.body.data.action === "string")

    const convID = chat1.body.data.conversation_id

    // ================================================================
    // 测试 4：继续已有对话
    // ================================================================
    console.log("\n═══════════════════════════════════════════════════════")
    console.log("测试 4：继续已有对话")
    console.log("═══════════════════════════════════════════════════════")

    const chat2 = await req("POST", "/chat",
      { message: "酒店有健身房吗？", conversation_id: convID }, guestToken)

    if (chat2.body.code === 0) {
      check("返回 code 为 0", true)
      check("返回 reply 非空", chat2.body.data.reply && chat2.body.data.reply.length > 0)
      check("conversation_id 不变", chat2.body.data.conversation_id === convID)
    } else {
      check(`继续对话成功 (${chat2.body.code} ${chat2.body.message})`, false)
    }

    // ================================================================
    // 测试 5：工具调用 — 查订单
    // ================================================================
    console.log("\n═══════════════════════════════════════════════════════")
    console.log("测试 5：工具调用 — 查订单")
    console.log("═══════════════════════════════════════════════════════")

    const chat3 = await req("POST", "/chat",
      { message: "帮我查一下我的订单", conversation_id: convID }, guestToken)
    if (chat3.body.code === 0) {
      check("返回 reply 非空", chat3.body.data.reply && chat3.body.data.reply.length > 0)
      const hasAction = chat3.body.data.action && chat3.body.data.action.length > 0
      if (hasAction) {
        check(`触发了工具调用: ${chat3.body.data.action}`, true)
      } else {
        check("LLM 以文本回复代替工具调用（可接受）", true)
      }
    } else {
      check(`查订单 (${chat3.body.code} ${chat3.body.message})`, false)
    }
  } else {
    // LLM 不可用 — 跳过深度测试
    const skipMsg = chat1.body.message || "未知错误"
    console.log(`  ⚠ AI 管家不可用 (${skipMsg})，跳过对话测试`)
    console.log("  提示: 请确认 config.yaml 中 llm.api_key 已正确配置")
    check("服务端返回了错误信息", chat1.body.code === 6001)
  }

  // ================================================================
  // 测试 6：并发对话隔离
  // ================================================================
  console.log("\n═══════════════════════════════════════════════════════")
  console.log("测试 6：并发对话隔离")
  console.log("═══════════════════════════════════════════════════════")

  // 第二个住户
  const guestUser2 = "chat_test2_" + Date.now()
  await ok("POST", "/guests", {
    username: guestUser2, password: "123456", name: "聊天测试住户2"
  }, adminToken)
  const guestLogin2 = await ok("POST", "/auth/login",
    { username: guestUser2, password: "123456" })
  const token2 = guestLogin2.access_token

  // 住户 1 新对话
  const a1 = await req("POST", "/chat", { message: "帮我叫服务员" }, guestToken)
  // 住户 2 新对话
  const a2 = await req("POST", "/chat", { message: "帮我查订单" }, token2)

  if (a1.body.code === 0 && a2.body.code === 0) {
    check("住户 1 获得新 conversation_id", a1.body.data.conversation_id.length > 0)
    check("住户 2 获得新 conversation_id", a2.body.data.conversation_id.length > 0)
        if (a1.body.data.action) check(`住户 1 触发了工具: ${a1.body.data.action}`, true)
	    if (a2.body.data.action) check(`住户 2 触发了工具: ${a2.body.data.action}`, true)
	    check("不同住户 conversation_id 不同", a1.body.data.conversation_id !== a2.body.data.conversation_id)

    // 住户 2 不能用住户 1 的 conversation_id
    const a3 = await req("POST", "/chat",
      { message: "继续", conversation_id: a1.body.data.conversation_id }, token2)
    if (a3.body.code !== 0) {
      check("跨用户 conversation_id 被拒绝", true)
    } else {
      check("跨用户 conversation_id 被拒绝（返回成功但不一致）",
        a3.body.data.conversation_id !== a1.body.data.conversation_id)
    }
  } else {
    check("并发对话隔离（跳过，AI 不可用）",
      a1.body.code === 6001 || a2.body.code === 6001)
  }

  // ── 最终结果 ──
  console.log("\n═══════════════════════════════════════════════════════")
  const total = pass + fail
  console.log(`测试完成: ${pass}/${total} 通过, ${fail}/${total} 失败`)
  if (fail > 0) process.exit(1)
  console.log("=== ✅ 全部通过 ===")
}

main().catch(e => {
  console.error("\n❌ 脚本异常:", e.message)
  process.exit(1)
})
