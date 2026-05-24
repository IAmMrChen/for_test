import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./task.js', import.meta.url), 'utf8')

assert.ok(source.includes('listTaskClaims'), '任务 API 应封装领取关系列表接口')
assert.ok(source.includes('/api/task/claims'), '任务 API 应请求领取关系列表路由')
assert.ok(source.includes('startTaskClaim'), '任务 API 应封装开始领取接口')
assert.ok(source.includes('/api/task/startClaim'), '任务 API 应请求开始领取路由')
assert.ok(source.includes('stopTaskClaim'), '任务 API 应封装停止领取接口')
assert.ok(source.includes('/api/task/stopClaim'), '任务 API 应请求停止领取路由')
