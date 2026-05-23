# 积分流水独立页面 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将积分流水从“我的”页长列表拆成独立普通页面，并在“我的”页保留轻量入口。

**Architecture:** 不改后端接口。小程序新增 `pages/points/index.vue`，复用现有 `listPointLogs`、`listMembers` 和 `getDashboardSummary`。`profile/index.vue` 不再加载完整流水，只提供入口卡片。

**Tech Stack:** uni-app Vue3、现有 point/member/dashboard API、Node 源码行为测试、`.vue` script 语法检查、Go 全量测试。

---

### Task 1: 增加积分流水页面行为测试

**Files:**
- Create: `parent-child-miniprogram/pages/points/index.behavior.test.mjs`
- Test: `node parent-child-miniprogram/pages/points/index.behavior.test.mjs`

- [ ] **Step 1: 写失败测试**

创建 `parent-child-miniprogram/pages/points/index.behavior.test.mjs`：

```js
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
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```powershell
node parent-child-miniprogram/pages/points/index.behavior.test.mjs
```

Expected: FAIL，因为积分流水页尚不存在。

- [ ] **Step 3: 提交失败测试**

```powershell
git add parent-child-miniprogram/pages/points/index.behavior.test.mjs
git commit -m "test: 增加积分流水页面行为约束"
```

### Task 2: 新增积分流水页

**Files:**
- Create: `parent-child-miniprogram/pages/points/index.vue`
- Modify: `parent-child-miniprogram/pages.json`
- Test: `parent-child-miniprogram/pages/points/index.behavior.test.mjs`

- [ ] **Step 1: 注册普通页面**

在 `pages.json` 的 `pages` 中新增：

```json
{
  "path": "pages/points/index",
  "style": {
    "navigationBarTitleText": "积分流水",
    "navigationStyle": "custom"
  }
}
```

不加入 `tabBar.list`。

- [ ] **Step 2: 实现积分流水页**

页面需要：

1. 加载当前家庭。
2. 调用 `getDashboardSummary` 判断角色。
3. 家长调用 `listMembers` 并筛选孩子。
4. 调用 `listPointLogs({ familyId, memberId })`。
5. 切换筛选时只重新拉流水。
6. 显示加载态、错误态、空状态。

- [ ] **Step 3: 运行行为测试**

Run:

```powershell
node parent-child-miniprogram/pages/points/index.behavior.test.mjs
```

Expected: PASS。

- [ ] **Step 4: 提交积分流水页**

```powershell
git add parent-child-miniprogram/pages.json parent-child-miniprogram/pages/points/index.vue
git commit -m "feat: 增加积分流水页面"
```

### Task 3: 调整我的页为入口卡片

**Files:**
- Create: `parent-child-miniprogram/pages/profile/index.behavior.test.mjs`
- Modify: `parent-child-miniprogram/pages/profile/index.vue`
- Test: `node parent-child-miniprogram/pages/profile/index.behavior.test.mjs`

- [ ] **Step 1: 写失败测试**

创建 `parent-child-miniprogram/pages/profile/index.behavior.test.mjs`：

```js
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const source = fs.readFileSync(path.join(__dirname, 'index.vue'), 'utf8')

assert.ok(source.includes('goPointLogs'), '我的页应提供积分流水跳转函数')
assert.ok(source.includes('/pages/points/index'), '我的页应跳转到积分流水页')
assert.ok(source.includes('查看积分收入、兑换和调整记录'), '我的页应有积分流水入口说明')
assert.ok(!source.includes('listPointLogs'), '我的页不应直接拉取积分流水')
assert.ok(!source.includes('pointLogs'), '我的页不应维护完整积分流水列表')
assert.ok(!source.includes('log-card'), '我的页不应渲染完整流水卡片')
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```powershell
node parent-child-miniprogram/pages/profile/index.behavior.test.mjs
```

Expected: FAIL，因为当前“我的”页仍直接渲染流水列表。

- [ ] **Step 3: 修改我的页**

移除：

1. `listPointLogs` 导入。
2. `pointLogs` 状态。
3. `loadProfile` 中的流水请求。
4. 完整流水列表模板。
5. `sourceName`、`pointText` 等只服务于流水列表的函数。

新增入口卡片：

```vue
<view class="section">
  <view class="entry-card" @click="goPointLogs">
    <view>
      <text class="entry-title">积分流水</text>
      <text class="entry-subtitle">查看积分收入、兑换和调整记录</text>
    </view>
    <text class="entry-arrow">进入</text>
  </view>
</view>
```

新增：

```js
function goPointLogs() {
  uni.navigateTo({ url: '/pages/points/index' })
}
```

- [ ] **Step 4: 运行我的页行为测试**

Run:

```powershell
node parent-child-miniprogram/pages/profile/index.behavior.test.mjs
```

Expected: PASS。

- [ ] **Step 5: 提交我的页调整**

```powershell
git add parent-child-miniprogram/pages/profile/index.behavior.test.mjs parent-child-miniprogram/pages/profile/index.vue
git commit -m "feat: 调整我的页积分流水入口"
```

### Task 4: 最终验证

**Files:**
- Verify: `parent-child-miniprogram/pages/points/index.vue`
- Verify: `parent-child-miniprogram/pages/profile/index.vue`
- Verify: `parent-child-miniprogram/pages.json`

- [ ] **Step 1: 行为测试**

```powershell
node parent-child-miniprogram/pages/points/index.behavior.test.mjs
node parent-child-miniprogram/pages/profile/index.behavior.test.mjs
```

Expected: PASS。

- [ ] **Step 2: 后端全量测试**

```powershell
go test ./... -count=1
```

Working directory: `parent-child-api`

Expected: PASS。

- [ ] **Step 3: 前端 JS 语法检查**

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
```

Expected: exit 0。

- [ ] **Step 4: Vue script 语法检查**

用 Node 抽取 `pages/points/index.vue` 和 `pages/profile/index.vue` 的 `<script setup>` 后执行 `node --check`。

Expected: PASS。

- [ ] **Step 5: 乱码扫描**

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\points\index.vue parent-child-miniprogram\pages\profile\index.vue parent-child-miniprogram\pages.json
```

Expected: exit 1。

## 自检

1. 设计要求均有任务覆盖：独立页面、我的页入口、家长筛选、孩子权限、错误态、不进 tab。
2. 不改后端接口。
3. 行为测试覆盖“我的页不再渲染完整流水列表”。
