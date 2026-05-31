import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const source = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')

assert.ok(source.includes('已有邀请码？'), 'family select page should keep low-priority invite code entry')
assert.ok(source.includes("nickname: '家主'"), 'creating a family should pass default owner nickname')
assert.ok(source.includes(':class="{ expanded: showInviteForm }"'), 'invite entry should have compact and expanded states')
assert.ok(source.includes('.invite-card .muted'), 'collapsed invite entry should compress the description vertically')
assert.ok(source.includes('padding-top: 18rpx'), 'collapsed invite entry should reduce vertical padding')
assert.ok(source.includes('white-space: nowrap'), 'collapsed invite description should stay compact on one line')
assert.ok(!source.includes('max-width: 520rpx'), 'collapsed invite entry should keep full card width')
assert.ok(source.includes('.invite-form .form-actions'), 'invite form actions should have dedicated alignment')
assert.ok(source.includes('justify-content: flex-end'), 'invite actions should align to the right')
assert.ok(source.includes('background: #fff'), 'expanded invite input should use a clear white background')
assert.ok(!source.includes('acceptVirtualChildBindInvite'), 'family select page should not handle virtual child binding')
assert.ok(!source.includes('showBindForm'), 'family select page should not show binding form')
assert.ok(!source.includes('toggleBindForm'), 'family select page should not have binding toggle logic')
assert.ok(!source.includes('bindForm'), 'family select page should not keep binding form state')
assert.ok(!source.includes('关联孩子'), 'family select page should not show child binding entry')
assert.ok(!source.includes('我的昵称'), 'family creation form should not ask for creator nickname')
assert.ok(source.includes('create-family-card'), 'family select page should always show create family card')
assert.ok(source.includes('创建一个新家庭'), 'create family entry should be a first-screen focus module')
