import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const source = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')

assert.ok(source.includes('代孩子完成任务'), '首页应保留代孩子完成任务快捷区')
assert.ok(source.includes('待发放奖励'), '首页应保留待发放奖励区')
assert.ok(source.indexOf('待发放奖励') < source.indexOf('代孩子完成任务'), '代孩子完成任务应位于待发放奖励之后')
assert.ok(source.includes("member.roleType === 'CHILD'"), '首页代操作应支持所有孩子')
assert.ok(!source.includes('member.isVirtual'), '首页代操作不应只筛选虚拟孩子')
assert.ok(source.includes('查看全部任务'), '首页代孩子快捷区应提供查看全部任务入口')
assert.ok(source.includes('/pages/tasks/index'), '首页应能跳转到任务页')
assert.ok(!source.includes('发布任务'), '首页不应展示发布任务表单')
assert.ok(!source.includes('任务管理'), '首页不应展示完整任务管理列表')
