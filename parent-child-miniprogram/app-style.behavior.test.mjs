import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./App.vue', import.meta.url), 'utf8')

assert.ok(source.includes('--sun-primary'), '全局样式应定义阳光主色')
assert.ok(source.includes('.sun-page'), '全局样式应定义页面基类')
assert.ok(source.includes('.sun-card'), '全局样式应定义卡片基类')
assert.ok(source.includes('.sun-btn'), '全局样式应定义按钮基类')
assert.ok(source.includes('.sun-chip'), '全局样式应定义筛选胶囊基类')
