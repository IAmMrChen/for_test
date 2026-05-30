import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const source = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')

assert.ok(source.includes('goPointLogs'), '我的页应提供积分流水跳转函数')
assert.ok(source.includes('/pages/points/index'), '我的页应跳转到积分流水页')
assert.ok(source.includes('查看积分收入、兑换和调整记录'), '我的页应有积分流水入口说明')
assert.ok(source.includes('goMembers'), '我的页应提供成员管理跳转函数')
assert.ok(source.includes('/pages/members/index'), '我的页应跳转到成员管理页')
assert.ok(source.includes('listMembers'), '我的页家庭成员模块应加载当前成员列表')
assert.ok(source.includes('member-card'), '我的页家庭成员模块应直接展示成员卡片')
assert.ok(source.includes('member.currentPoints'), '我的页成员卡片应展示成员当前积分')
assert.ok(!source.includes('成员管理互动件'), '我的页不应再展示成员管理互动件模块')
assert.ok(!source.includes('listPointLogs'), '我的页不应直接拉取积分流水')
assert.ok(!source.includes('pointLogs'), '我的页不应维护完整积分流水列表')
assert.ok(!source.includes('log-card'), '我的页不应渲染完整流水卡片')
assert.ok(!source.includes('createVirtualChild'), '我的页不应直接创建虚拟孩子')
assert.ok(!source.includes('createInvite'), '我的页不应直接创建成员邀请')
