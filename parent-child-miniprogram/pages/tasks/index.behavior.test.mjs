import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const pageSource = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')
const pagesJson = fs.readFileSync(path.resolve(__dirname, '../../pages.json'), 'utf8')

assert.ok(pagesJson.includes('"path": "pages/tasks/index"'), 'pages.json 应注册任务页')
assert.ok(pagesJson.includes('"pagePath": "pages/tasks/index"'), '底部菜单应包含任务入口')
assert.ok(pageSource.includes('发布任务'), '任务页应提供发布任务入口')
assert.ok(pageSource.includes('任务管理'), '任务页应提供任务管理区域')
assert.ok(pageSource.includes('已存在同名任务，仍要继续吗？'), '任务页应对同名任务做前端提醒')
assert.ok(pageSource.includes('confirmDuplicateTaskTitle'), '任务页应通过确认函数允许同名任务继续提交')
assert.ok(pageSource.includes('archiveExistingTask'), '任务页应支持归档任务')
assert.ok(pageSource.includes('editTask'), '任务页应支持编辑任务')
assert.ok(pageSource.includes('loadError'), '任务页应有加载错误态，避免接口异常白屏')
assert.ok(!pageSource.includes('throw error'), '任务页不应把加载错误直接抛到页面运行时')
assert.ok(pageSource.includes('applyClaimedTaskRecord'), '领取任务后应更新本地状态而不是整页刷新')
assert.ok(pageSource.includes('applySubmittedTaskRecord'), '提交任务后应更新本地状态而不是整页刷新')
