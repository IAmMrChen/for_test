import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const source = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')

assert.ok(source.includes('已有邀请码？'), '选择家庭页应保留低权重的邀请码入口')
assert.ok(source.includes("nickname: '家主'"), '创建家庭应固定传入家主昵称')
assert.ok(!source.includes('acceptVirtualChildBindInvite'), '选择家庭页不应处理虚拟孩子绑定')
assert.ok(!source.includes('showBindForm'), '选择家庭页不应展示绑定表单')
assert.ok(!source.includes('toggleBindForm'), '选择家庭页不应有绑定入口切换逻辑')
assert.ok(!source.includes('bindForm'), '选择家庭页不应维护绑定码表单状态')
assert.ok(!source.includes('关联孩子'), '选择家庭页不应展示关联孩子入口')
assert.ok(!source.includes('我的昵称'), '创建家庭表单不应要求输入创建人昵称')
assert.ok(source.includes('create-family-card'), '选择家庭页应固定展示创建家庭大卡片入口')
assert.ok(source.includes('创建一个新家庭'), '创建家庭入口应是页面首屏重点模块')
