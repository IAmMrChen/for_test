import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'

const config = JSON.parse(readFileSync(new URL('./pages.json', import.meta.url), 'utf8'))
const paths = config.pages.map((page) => page.path)

assert.ok(paths.includes('pages/reward-proxy/index'), '应注册代孩子兑换奖励页')
assert.ok(paths.includes('pages/members/index'), '应注册成员管理页')
assert.equal(config.pages.find((page) => page.path === 'pages/index/index')?.style.navigationBarTitleText, '首页')
assert.equal(config.pages.find((page) => page.path === 'pages/rewards/index')?.style.navigationBarTitleText, '奖励')
assert.equal(config.tabBar.list.find((item) => item.pagePath === 'pages/tasks/index')?.text, '任务')
assert.equal(config.tabBar.selectedColor, '#ff7a45', '底部菜单选中态应使用晴日行动色')
assert.equal(config.tabBar.backgroundColor, '#fffaf0', '底部菜单背景应使用暖白色')

for (const item of config.tabBar.list) {
  assert.notEqual(item.iconPath, 'static/logo.png', `${item.text} 不应继续使用绿色 logo 作为 tab 图标`)
  assert.notEqual(item.selectedIconPath, 'static/logo.png', `${item.text} 不应继续使用绿色 logo 作为选中 tab 图标`)
  assert.ok(existsSync(new URL(`./${item.iconPath}`, import.meta.url)), `${item.text} 普通图标文件应存在`)
  assert.ok(existsSync(new URL(`./${item.selectedIconPath}`, import.meta.url)), `${item.text} 选中图标文件应存在`)
}
