import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const source = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')

for (const name of ['apply', 'deliver', 'reject', 'receive']) {
  const start = source.indexOf(`async function ${name}`)
  assert.notEqual(start, -1, `奖励页应存在 ${name} 函数`)
  const end = source.indexOf('\n}\n', start)
  const body = source.slice(start, end)
  assert.ok(!body.includes('await loadRewardPage()'), `${name} 成功后不应整页刷新`)
}

assert.ok(!source.includes('async function applyProxyReward'), '代孩子兑换逻辑应迁移到独立页面')
assert.ok(source.includes('/pages/reward-proxy/index'), '奖励页应提供代孩子兑换入口')
assert.ok(source.includes('proxy-entry-head'), '奖励页代孩子兑换入口应使用专门的横向轻卡标题布局')
assert.ok(source.includes('rgba(75, 159, 255, 0.34)'), '奖励页代孩子兑换入口应贴近 UI 图的浅蓝卡片样式')
assert.ok(source.includes('addAppliedRewardRecord'), '兑换后应本地增加待发放记录')
assert.ok(source.includes('removeAppliedRewardRecord'), '发放或驳回后应本地移除待发放记录')
assert.ok(source.includes('removeDeliveredRewardRecord'), '确认收到后应本地移除待确认记录')
