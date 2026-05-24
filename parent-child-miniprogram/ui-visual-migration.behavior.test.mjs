import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const pagePaths = [
  './pages/family-select/index.vue',
  './pages/index/index.vue',
  './pages/tasks/index.vue',
  './pages/rewards/index.vue',
  './pages/points/index.vue',
  './pages/profile/index.vue'
]

for (const pagePath of pagePaths) {
  const source = readFileSync(new URL(pagePath, import.meta.url), 'utf8')
  assert.ok(source.includes('sun-page'), `${pagePath} 应使用晴日页面底色`)
  assert.ok(!source.includes('<view class="container">'), `${pagePath} 不应继续使用旧 container 根节点`)
}
