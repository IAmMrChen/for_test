import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

assert.ok(source.includes('成员管理'), '成员管理页应有明确标题')
assert.ok(source.includes('创建虚拟孩子'), '成员管理页应提供虚拟孩子创建能力')
assert.ok(source.includes('邀请成员'), '成员管理页应提供成员邀请能力')
assert.ok(source.includes('邀请关联'), '成员管理页应提供虚拟孩子关联邀请能力')
assert.ok(source.includes('createVirtualChild'), '成员管理页应调用创建虚拟孩子接口')
assert.ok(source.includes('createInvite'), '成员管理页应调用邀请接口')
assert.ok(source.includes('createVirtualChildBindInvite'), '成员管理页应调用虚拟孩子绑定邀请接口')
assert.ok(source.includes('latestBindInvite.memberId === member.id'), '虚拟孩子绑定码应显示在对应孩子卡片下')
