import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const pageSource = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')
const pagesJson = fs.readFileSync(path.resolve(__dirname, '../../pages.json'), 'utf8')

assert.ok(pagesJson.includes('"path": "pages/points/index"'), 'pages.json 应注册积分流水页')
assert.ok(!pagesJson.includes('"pagePath": "pages/points/index"'), '积分流水页不应加入底部 tab')
assert.ok(pageSource.includes('积分流水'), '积分流水页应展示页面标题')
assert.ok(pageSource.includes('listPointLogs'), '积分流水页应加载积分流水')
assert.ok(pageSource.includes('listMembers'), '积分流水页应支持家长按孩子筛选')
assert.ok(pageSource.includes('selectedMemberId'), '积分流水页应维护筛选成员')
assert.ok(pageSource.includes('loadError'), '积分流水页应有错误态')
assert.ok(!pageSource.includes('throw error'), '积分流水页不应把加载错误直接抛到运行时')
