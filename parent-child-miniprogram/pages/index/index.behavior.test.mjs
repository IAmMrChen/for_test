import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const source = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')

assert.ok(source.includes('代孩子完成任务'), 'home page should keep proxy task shortcut panel')
assert.ok(source.includes('待发放奖励'), 'home page should keep pending reward panel')
assert.ok(source.indexOf('待发放奖励') < source.indexOf('代孩子完成任务'), 'proxy task panel should sit after pending rewards')
assert.ok(source.includes("member.roleType === 'CHILD'"), 'proxy operation should support all children')
assert.ok(!source.includes('member.isVirtual'), 'proxy operation should not only filter virtual children')
assert.ok(!source.includes('查看全部任务'), 'proxy shortcut panel should not provide view-all entry')
assert.ok(source.includes('/pages/tasks/index'), 'home page should navigate to task page')
assert.ok(source.includes('justify-content: center'), 'home stat card content should be vertically centered')
assert.ok(source.includes('font-size: 44rpx'), 'home stat numbers should have enough visual weight')
assert.ok(!source.includes('发布任务'), 'home page should not render task publish form')
assert.ok(!source.includes('任务管理'), 'home page should not render full task management list')
assert.ok(source.includes('支持真实孩子和虚拟孩子'), 'proxy task panel should explain support for real and virtual children')
assert.ok(source.includes('loadError'), 'home page should have loading error state to avoid blank screen')
assert.ok(!source.includes('throw error'), 'home page should not throw load errors into page runtime')
assert.ok(source.includes('applyClaimedTaskRecord'), 'claiming a task should update local state instead of refreshing whole page')
assert.ok(source.includes('applySubmittedTaskRecord'), 'submitting a task should update local state instead of refreshing whole page')
assert.ok(source.includes('taskClaims'), 'child view should keep recurring task claim state')
assert.ok(source.includes('startTaskClaim'), 'child view should call start claim API for recurring tasks')
assert.ok(source.includes('stopTaskClaim'), 'child view should support stopping recurring task claim')
assert.ok(source.includes('停止领取'), 'child view should show stop claim button')
assert.ok(source.includes('directApproveProxyTask'), 'proxy task panel should support direct approval')
assert.ok(source.includes("task.viewStatus === 'recurringActive'"), 'proxy recurring task should submit current period before approval')
assert.ok(source.includes('applyApprovedTaskRecord'), 'approval should update local state instead of refreshing whole page')
assert.ok(source.includes('removeRewardRecord'), 'reward operation should remove local record instead of refreshing whole page')
assert.ok(!source.includes('await loadHome()'), 'successful operations should not refresh the whole page')

const proxyRowsBlock = source.match(/const proxyTaskRows = computed\(\(\) => \{[\s\S]*?\n\}\)/)?.[0] || ''
assert.ok(proxyRowsBlock.includes('.filter(Boolean)'), 'proxy task rows should remove hidden null task rows before rendering')
