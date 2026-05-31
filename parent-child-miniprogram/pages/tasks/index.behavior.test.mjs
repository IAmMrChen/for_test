import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const pageSource = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')
const pagesJson = fs.readFileSync(path.resolve(__dirname, '../../pages.json'), 'utf8')

assert.ok(pagesJson.includes('"path": "pages/tasks/index"'), 'pages.json should register task page')
assert.ok(pagesJson.includes('"pagePath": "pages/tasks/index"'), 'tab bar should include task entry')
assert.ok(pageSource.includes('发布任务'), 'task page should provide publish task entry')
assert.ok(pageSource.includes('v-if="isParentRole && showTaskForm && !editingTaskId"'), 'publish form should only be the top create form')
assert.ok(pageSource.includes('inline-edit-panel'), 'task edit form should expand inline under the edited task')
assert.ok(pageSource.includes('editingTaskId === task.id'), 'inline edit form should be tied to the selected task row')
assert.ok(pageSource.includes('showTaskForm.value = false'), 'editing a task should close the top publish form')
assert.ok(pageSource.includes('applyUpdatedTask'), 'editing a task should update the local row after saving')
assert.ok(pageSource.includes('applyCreatedTask'), 'creating a task should update the local list after saving')
assert.ok(pageSource.includes('任务管理'), 'task page should provide task management area')
assert.ok(pageSource.includes('已存在同名任务，仍要继续吗？'), 'task page should warn about duplicate task names')
assert.ok(pageSource.includes('confirmDuplicateTaskTitle'), 'task page should allow duplicate names only after confirmation')
assert.ok(pageSource.includes('archiveExistingTask'), 'task page should support archiving tasks')
assert.ok(pageSource.includes('editTask'), 'task page should support editing tasks')
assert.ok(pageSource.includes('loadError'), 'task page should have loading error state to avoid blank screen')
assert.ok(!pageSource.includes('throw error'), 'task page should not throw load errors into page runtime')
assert.ok(pageSource.includes('applyClaimedTaskRecord'), 'claiming a task should update local state instead of refreshing whole page')
assert.ok(pageSource.includes('applySubmittedTaskRecord'), 'submitting a task should update local state instead of refreshing whole page')
assert.ok(pageSource.includes('taskClaims'), 'task page should keep recurring task claim state')
assert.ok(pageSource.includes('startTaskClaim'), 'recurring tasks should call start claim API')
assert.ok(pageSource.includes('stopTaskClaim'), 'recurring tasks should call stop claim API')
assert.ok(pageSource.includes('停止领取'), 'task page should provide stop claim button')
assert.ok(pageSource.includes('.child-task-actions'), 'child task operation area should have dedicated layout')
assert.ok(pageSource.includes('flex-direction: row'), 'submit and stop claim buttons should be arranged horizontally')
assert.ok(pageSource.includes('child-task-actions"'), 'child task action view should keep the dedicated action class')
assert.ok(pageSource.includes('compact: task.activeClaim'), 'recurring claimed tasks should get compact horizontal action layout')
assert.ok(pageSource.includes('stop-claim-btn'), 'stop claim button should have a dedicated compact style hook')
assert.ok(pageSource.indexOf('stop-claim-btn') < pageSource.indexOf('task-submit-btn'), 'stop claim button should appear before submit button')

const publishTaskSource = pageSource.slice(pageSource.indexOf('async function publishTask'), pageSource.indexOf('function editTask'))
assert.ok(!publishTaskSource.includes('await loadTaskPage()'), 'publishing or editing should not refresh the whole task page')

const handleTaskActionSource = pageSource.slice(pageSource.indexOf('async function handleTaskAction'), pageSource.indexOf('function applyTaskClaim'))
assert.ok(!handleTaskActionSource.includes('await loadTaskPage()'), 'claiming or submitting a task should not refresh the whole task page')
