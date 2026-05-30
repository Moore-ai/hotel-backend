// 申诉功能完整测试脚本
// 运行: node test/appeal-flow.js
// 依赖: Node.js 18+ (内置 fetch)

const BASE = "http://localhost:8080/api/v1"
const headers = { "Content-Type": "application/json" }

function today() { return new Date().toISOString().slice(0, 10) }
function future(days) {
  const d = new Date()
  d.setDate(d.getDate() + days)
  return d.toISOString().slice(0, 10)
}

let adminToken, guestToken, guestID
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
  const login = await ok("POST", "/auth/admin-login",
    { username: "admin", password: "admin123" })
  adminToken = login.access_token

  const guestUser = "test_" + Date.now()
  await ok("POST", "/guests", {
    username: guestUser, password: "123456",
    name: "测试住户"
  }, adminToken)

  const guestLogin = await ok("POST", "/auth/login",
    { username: guestUser, password: "123456" })
  guestToken = guestLogin.access_token
  guestID = guestLogin.user_code

  const todayStr = today()
  const day3 = future(3)
  const day4 = future(4)

  // ================================================================
  // 测试 1：正常流程 — 申诉通过 → 订单取消
  // ================================================================
  console.log("\n═══════════════════════════════════════════════════════")
  console.log("测试 1：申诉通过 → 订单取消")
  console.log("═══════════════════════════════════════════════════════")

  const order1 = await ok("POST", "/orders", {
    guest_count: 1, room_type_preference: "standard",
    check_in_date: todayStr, check_out_date: day3, total_price: 600
  }, guestToken)
  const code1 = order1.id
  console.log("  创建订单:", code1)

  // 取消 → cancel_requested
  const cancel1 = await ok("POST", `/orders/${code1}/cancel`,
    { reason: "行程变更" }, guestToken)
  check("取消进入审核", cancel1.status === "cancel_requested" && !cancel1.auto_cancelled)

  // 驳回
  await ok("POST", `/orders/${code1}/reject-cancel`,
    { reason: "已过免费取消期限" }, adminToken)
  const ordAfterReject = await ok("GET", `/orders/${code1}`, null, adminToken)
  check("驳回后状态为 pending", ordAfterReject.status === "pending")

  // 申诉
  const appeal1 = await ok("POST", `/orders/${code1}/appeal`,
    { reason: "请求人工审核" }, guestToken)
  check("申诉创建成功", appeal1.status === "pending" && appeal1.id > 0)

  // 查看待审列表
  const list1 = await ok("GET", "/appeals?status=pending", null, adminToken)
  check("待审列表包含该申诉", list1.total >= 1 && list1.list.some(a => a.id === appeal1.id))

  // 通过申诉
  const review1 = await ok("POST", `/appeals/${appeal1.id}/review`,
    { action: "approved", review_note: "同意取消" }, adminToken)
  check("申诉状态变为 approved", review1.status === "approved")
  check("审核人有记录", review1.reviewer_id !== null)
  check("审核备注已保存", review1.review_note === "同意取消")

  // 验证订单已取消
  const ord1Final = await ok("GET", `/orders/${code1}`, null, adminToken)
  check("订单已取消", ord1Final.status === "cancelled")

  // ================================================================
  // 测试 2：正常流程 — 申诉驳回 → 订单保持 pending
  // ================================================================
  console.log("\n═══════════════════════════════════════════════════════")
  console.log("测试 2：申诉驳回 → 订单保持 pending")
  console.log("═══════════════════════════════════════════════════════")

  const order2 = await ok("POST", "/orders", {
    guest_count: 1, room_type_preference: "standard",
    check_in_date: todayStr, check_out_date: day4, total_price: 500
  }, guestToken)
  const code2 = order2.id

  await ok("POST", `/orders/${code2}/cancel`, { reason: "计划有变" }, guestToken)
  await ok("POST", `/orders/${code2}/reject-cancel`, { reason: "不符合条件" }, adminToken)
  await ok("POST", `/orders/${code2}/appeal`, { reason: "请重新考虑" }, guestToken)

  const list2 = await ok("GET", "/appeals?status=pending", null, adminToken)
  const appeal2 = list2.list[list2.list.length - 1] // 取最后一个（最新创建）

  const review2 = await ok("POST", `/appeals/${appeal2.id}/review`,
    { action: "rejected", review_note: "政策不允许" }, adminToken)
  check("驳回后状态为 rejected", review2.status === "rejected")

  const ord2Final = await ok("GET", `/orders/${code2}`, null, adminToken)
  check("驳回后订单保持 pending", ord2Final.status === "pending")

  // ================================================================
  // 测试 3：重复申诉拦截
  // ================================================================
  console.log("\n═══════════════════════════════════════════════════════")
  console.log("测试 3：重复申诉拦截")
  console.log("═══════════════════════════════════════════════════════")

  const order3 = await ok("POST", "/orders", {
    guest_count: 1, room_type_preference: "standard",
    check_in_date: todayStr, check_out_date: day4, total_price: 400
  }, guestToken)
  const code3 = order3.id

  await ok("POST", `/orders/${code3}/cancel`, { reason: "行程冲突" }, guestToken)
  await ok("POST", `/orders/${code3}/reject-cancel`, { reason: "不可取消" }, adminToken)

  // 第一次申诉 — 应该成功
  const a3first = await ok("POST", `/orders/${code3}/appeal`, { reason: "第一次申诉" }, guestToken)

  // 第二次申诉（同订单，前一次 pending）— 应该被拒绝
  const dup = await req("POST", `/orders/${code3}/appeal`,
    { reason: "第二次申诉" }, guestToken)
  check("重复申诉被拒绝", dup.body.code !== 0)
  check("错误码为 ErrAppealExists", dup.body.code === 4005)

  // 驳回第一次申诉
  await ok("POST", `/appeals/${a3first.id}/review`,
    { action: "rejected", review_note: "不通过" }, adminToken)

  // 驳回后，再次申诉 — 应该成功（前一次已结束）
  const retry = await ok("POST", `/orders/${code3}/appeal`,
    { reason: "再次申诉（前一次已驳回）" }, guestToken)
  check("驳回后可重新申诉", retry.status === "pending")
  await ok("POST", `/appeals/${retry.id}/review`,
    { action: "rejected", review_note: "不通过" }, adminToken)

  // ================================================================
  // 测试 4：已审核的申诉不可再审
  // ================================================================
  console.log("\n═══════════════════════════════════════════════════════")
  console.log("测试 4：已审核的申诉不可再审")
  console.log("═══════════════════════════════════════════════════════")

  // 用 appeal1（已 approved）再审一次
  const reReview = await req("POST", `/appeals/${appeal1.id}/review`,
    { action: "rejected", review_note: "再驳回" }, adminToken)
  check("已审核申诉被拒绝", reReview.body.code !== 0)
  check("错误码为 ErrAppealNotPending", reReview.body.code === 4004)

  // ================================================================
  // 测试 5：无效操作参数
  // ================================================================
  console.log("\n═══════════════════════════════════════════════════════")
  console.log("测试 5：无效操作参数")
  console.log("═══════════════════════════════════════════════════════")

  // 先创建新订单 → 取消 → 驳回 → 申诉
  const order5 = await ok("POST", "/orders", {
    guest_count: 1, room_type_preference: "standard",
    check_in_date: todayStr, check_out_date: day4, total_price: 300
  }, guestToken)
  const code5 = order5.id
  await ok("POST", `/orders/${code5}/cancel`, { reason: "test" }, guestToken)
  await ok("POST", `/orders/${code5}/reject-cancel`, { reason: "test" }, adminToken)
  const a5 = await ok("POST", `/orders/${code5}/appeal`, { reason: "test" }, guestToken)

  const badAction = await req("POST", `/appeals/${a5.id}/review`,
    { action: "invalid_action", review_note: "" }, adminToken)
  check("无效 action 返回 400", badAction.status === 400)

  // ================================================================
  // 测试 6：不存在的申诉
  // ================================================================
  console.log("\n═══════════════════════════════════════════════════════")
  console.log("测试 6：不存在的申诉")
  console.log("═══════════════════════════════════════════════════════")

  const notFound = await req("POST", "/appeals/999999/review",
    { action: "approved", review_note: "" }, adminToken)
  check("不存在被拒绝", notFound.body.code !== 0)
  check("错误码为 ErrAppealNotFound", notFound.body.code === 4003)

  // ================================================================
  // 测试 7：非本人订单不可申诉
  // ================================================================
  console.log("\n═══════════════════════════════════════════════════════")
  console.log("测试 7：非本人订单不可申诉")
  console.log("═══════════════════════════════════════════════════════")

  const otherGuest = "other_" + Date.now()
  await ok("POST", "/guests", {
    username: otherGuest, password: "123456", name: "其他住户"
  }, adminToken)
  const otherLogin = await ok("POST", "/auth/login",
    { username: otherGuest, password: "123456" })
  const otherToken = otherLogin.access_token

  const order7 = await ok("POST", "/orders", {
    guest_count: 1, room_type_preference: "standard",
    check_in_date: todayStr, check_out_date: day4, total_price: 350
  }, guestToken)
  const code7 = order7.id
  await ok("POST", `/orders/${code7}/cancel`, { reason: "test" }, guestToken)
  await ok("POST", `/orders/${code7}/reject-cancel`, { reason: "test" }, adminToken)

  // 其他住户尝试申诉此订单
  const forbidden = await req("POST", `/orders/${code7}/appeal`,
    { reason: "恶意申诉" }, otherToken)
  check("非本人申诉返回 403", forbidden.status === 403)

  // ================================================================
  // 测试 8：非 pending 订单不可申诉
  // ================================================================
  console.log("\n═══════════════════════════════════════════════════════")
  console.log("测试 8：非 pending 订单不可申诉")
  console.log("═══════════════════════════════════════════════════════")

  const order8 = await ok("POST", "/orders", {
    guest_count: 1, room_type_preference: "standard",
    check_in_date: todayStr, check_out_date: day4, total_price: 250
  }, guestToken)
  const code8 = order8.id

  // 直接申诉（订单未取消过，是 pending 状态）— 不对，申诉逻辑要求先被驳回
  // 先取消、再驳回、再申诉通过，然后再次申诉应该不行
  await ok("POST", `/orders/${code8}/cancel`, { reason: "test" }, guestToken)
  await ok("POST", `/orders/${code8}/reject-cancel`, { reason: "test" }, adminToken)
  const a8 = await ok("POST", `/orders/${code8}/appeal`, { reason: "test" }, guestToken)
  await ok("POST", `/appeals/${a8.id}/review`,
    { action: "approved", review_note: "通过" }, adminToken)

  // 订单已取消，非 pending，不可再申诉
  const afterCancelled = await req("POST", `/orders/${code8}/appeal`,
    { reason: "再申诉" }, guestToken)
  check("已取消订单不可申诉", afterCancelled.body.code !== 0)
  check("错误码为 ErrOrderNotPending", afterCancelled.body.code === 4002)

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
