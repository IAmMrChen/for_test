import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./sunny-family-true-pages.html', import.meta.url), 'utf8')

const pages = [
  'family-select',
  'home-parent',
  'home-child',
  'tasks',
  'rewards',
  'reward-proxy',
  'points',
  'profile'
]

for (const page of pages) {
  assert.ok(source.includes(`data-page="${page}"`), `静态真稿应包含 ${page} 页面`)
}

assert.ok(source.includes('grid-template-columns: repeat(4, 292px)'), '静态真稿应固定手机稿宽度')
assert.ok(source.includes('height: 50px'), '静态真稿应定义悬浮底部导航高度')
assert.ok(source.includes('创建一个新家庭'), '选择家庭页应包含创建家庭主卡片')
assert.ok(source.includes('领取一次后每天直接提交'), '任务页应表达循环任务领取语义')
assert.ok(source.includes('代孩子兑换'), '奖励代兑页应独立呈现')
assert.ok(source.includes('最近一个月流水'), '积分流水页应保留 MVP 最近一个月范围')
