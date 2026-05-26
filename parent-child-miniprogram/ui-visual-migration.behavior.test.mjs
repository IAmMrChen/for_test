import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const pages = [
  { path: './pages/family-select/index.vue', scope: '.family-select-page button' },
  { path: './pages/index/index.vue', scope: '.home-page button' },
  { path: './pages/tasks/index.vue', scope: '.tasks-page button' },
  { path: './pages/rewards/index.vue', scope: '.rewards-page button' },
  { path: './pages/reward-proxy/index.vue', scope: '.proxy-page button' },
  { path: './pages/points/index.vue', scope: '.points-page button' },
  { path: './pages/profile/index.vue', scope: '.profile-page button' },
  { path: './pages/members/index.vue', scope: '.members-page button' }
]

for (const { path: pagePath, scope } of pages) {
  const source = readFileSync(new URL(pagePath, import.meta.url), 'utf8')
  assert.ok(source.includes('sun-page'), `${pagePath} 应使用晴日页面底色`)
  assert.ok(!source.includes('<view class="container">'), `${pagePath} 不应继续使用旧 container 根节点`)
  assert.ok(source.includes(scope), `${pagePath} 应定义页面级按钮覆盖，避免旧按钮样式覆盖全局基线`)
  assert.ok(!source.includes('var(--sun'), `${pagePath} 不应依赖微信小程序中不稳定的跨文件 CSS 变量`)
}
