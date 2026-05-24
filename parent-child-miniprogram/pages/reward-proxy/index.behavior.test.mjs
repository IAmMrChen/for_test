import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

assert.ok(source.includes('代孩子兑换奖励'), '代兑页应有明确标题')
assert.ok(source.includes('currentPoints'), '代兑页应独立展示孩子积分')
assert.ok(source.includes("member.roleType === 'CHILD'"), '代兑页应支持所有孩子成员')
assert.ok(!source.includes('member.isVirtual'), '代兑页不应只限制虚拟孩子')
assert.ok(source.includes('async function applyProxyReward'), '代兑页应保留代孩子兑换动作')
assert.ok(!source.includes('{{ child.nickname }} ·'), '孩子选择项不应把积分拼进孩子名')
