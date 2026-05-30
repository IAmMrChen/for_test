import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const source = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')

assert.ok(source.includes('goPointLogs'), 'profile page should provide point log navigation')
assert.ok(source.includes('/pages/points/index'), 'profile page should navigate to point log page')
assert.ok(source.includes('查看积分收入、兑换和调整记录'), 'profile page should show point log entry copy')
assert.ok(source.includes('goMembers'), 'profile page should provide member management navigation')
assert.ok(source.includes('/pages/members/index'), 'profile page should navigate to member management page')
assert.ok(source.includes('listMembers'), 'profile page should load current family members')
assert.ok(source.includes('member-card'), 'profile page should render member cards directly')
assert.ok(source.includes('member.currentPoints'), 'profile page member card should show current points')
assert.ok(source.includes('createVirtualChildBindInvite'), 'profile page should expose virtual child bind invite action')
assert.ok(source.includes('latestBindInvite.memberId === member.id'), 'bind invite should render under the matching child')
assert.ok(source.includes('@click.stop="createBindInvite(member)"'), 'bind invite button should be an explicit member action')
assert.ok(source.includes('@click.stop="goMembers"'), 'only the manage button should enter member management')
assert.ok(!source.includes('class="card entry-card" @click="goMembers"'), 'family member card should not navigate when tapped')
assert.ok(!source.includes('成员管理互动件'), 'profile page should not show member management action module')
assert.ok(!source.includes('listPointLogs'), 'profile page should not fetch complete point logs directly')
assert.ok(!source.includes('pointLogs'), 'profile page should not keep complete point log list state')
assert.ok(!source.includes('log-card'), 'profile page should not render complete log cards')
assert.ok(!source.includes('createVirtualChild({'), 'profile page should not directly create virtual children')
assert.ok(!source.includes('createInvite('), 'profile page should not directly create member invites')
