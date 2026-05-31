import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const pageSource = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')
const pagesJson = fs.readFileSync(path.resolve(__dirname, '../../pages.json'), 'utf8')

assert.ok(pagesJson.includes('"path": "pages/points/index"'), 'pages.json should register point logs page')
assert.ok(!pagesJson.includes('"pagePath": "pages/points/index"'), 'point logs page should not be added to tab bar')
assert.ok(pageSource.includes('积分流水'), 'point logs page should show title')
assert.ok(pageSource.includes('>返回</button>'), 'point logs page header action should use short back label')
assert.ok(!pageSource.includes('返回我的'), 'point logs page should not show “返回我的”')
assert.ok(pageSource.includes('listPointLogs'), 'point logs page should load point logs')
assert.ok(pageSource.includes('listMembers'), 'point logs page should support parent filtering by child')
assert.ok(pageSource.includes('selectedMemberId'), 'point logs page should keep selected member filter')
assert.ok(pageSource.includes('groupedPointLogs'), 'point logs page should group by real record date')
assert.ok(pageSource.includes('dateGroupTitle'), 'point logs page should compute group title from record date')
assert.ok(pageSource.includes('empty-log-card'), 'empty point logs state should use dedicated centered card')
assert.ok(pageSource.includes('align-items: center'), 'empty state text should be vertically centered')
assert.ok(pageSource.includes('justify-content: center'), 'empty state text should be horizontally centered')
assert.ok(pageSource.includes('loadError'), 'point logs page should have error state')
assert.ok(!pageSource.includes('throw error'), 'point logs page should not throw load errors into runtime')
