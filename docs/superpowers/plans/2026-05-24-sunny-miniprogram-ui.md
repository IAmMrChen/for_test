# Sunny Mini Program UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 按“晴日家庭感”重设小程序所有已实现页面，并完成奖励代兑、成员管理、积分流水展示范围等页面结构调整。

**Architecture:** 先建立全局视觉基线和页面公共样式，再逐页迁移，避免每个页面重复写一套按钮、卡片、筛选胶囊和空状态。页面拆分仅新增独立业务页面，不改变 tabBar 的主结构。所有高频操作继续局部更新，不整页刷新。

**Tech Stack:** uni-app/Vue 单文件组件、微信小程序页面配置、现有前端 API 封装、Node 行为测试。

---

## File Structure

- Modify: `parent-child-miniprogram/App.vue`，加入全局晴日家庭感基础样式。
- Modify: `parent-child-miniprogram/pages.json`，注册 `pages/reward-proxy/index` 和 `pages/members/index`，修复页面中文标题。
- Modify: `parent-child-miniprogram/pages/family-select/index.vue`，重设选择家庭页。
- Modify: `parent-child-miniprogram/pages/index/index.vue`，重设首页家长/孩子视图，移除“查看全部任务”。
- Modify: `parent-child-miniprogram/pages/tasks/index.vue`，重设任务页视觉，保留发布、编辑、归档。
- Modify: `parent-child-miniprogram/pages/rewards/index.vue`，重设奖励页，代兑改入口，新建放入奖励库。
- Create: `parent-child-miniprogram/pages/reward-proxy/index.vue`，独立代孩子兑换奖励页。
- Modify: `parent-child-miniprogram/pages/points/index.vue`，展示最近一个月流水和成员筛选。
- Modify: `parent-child-miniprogram/pages/profile/index.vue`，成员模块改为摘要和管理入口。
- Create: `parent-child-miniprogram/pages/members/index.vue`，独立家庭成员管理页。
- Modify: `parent-child-miniprogram/pages/*.behavior.test.mjs`，补齐结构约束。

## Task 1: Global Visual Baseline

**Files:**
- Modify: `parent-child-miniprogram/App.vue`
- Test: `parent-child-miniprogram/app-style.behavior.test.mjs`

- [ ] **Step 1: Create failing style test**

Create `parent-child-miniprogram/app-style.behavior.test.mjs`:

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./App.vue', import.meta.url), 'utf8')

assert.ok(source.includes('--sun-primary'), '全局样式应定义阳光主色')
assert.ok(source.includes('.sun-page'), '全局样式应定义页面基类')
assert.ok(source.includes('.sun-card'), '全局样式应定义卡片基类')
assert.ok(source.includes('.sun-btn'), '全局样式应定义按钮基类')
assert.ok(source.includes('.sun-chip'), '全局样式应定义筛选胶囊基类')
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
node parent-child-miniprogram/app-style.behavior.test.mjs
```

Expected: FAIL with missing style markers.

- [ ] **Step 3: Add global styles**

Add to `parent-child-miniprogram/App.vue` style block:

```css
:root {
  --sun-primary: #ffb84d;
  --sun-action: #ff7a45;
  --sun-sky: #4b9fff;
  --sun-mint: #63c784;
  --sun-rose: #ef6b7a;
  --sun-ink: #2b2a28;
  --sun-muted: #7b8190;
  --sun-line: #edf0f5;
  --sun-paper: #ffffff;
  --sun-warm: #fff8ea;
}

.sun-page {
  min-height: 100vh;
  padding: 28rpx 28rpx 160rpx;
  background: linear-gradient(180deg, #fff7e7 0%, #eef7ff 42%, #f8fbff 100%);
  color: var(--sun-ink);
}

.sun-title {
  display: block;
  font-size: 46rpx;
  line-height: 1.18;
  font-weight: 800;
  color: var(--sun-ink);
}

.sun-subtitle {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  line-height: 1.55;
  color: var(--sun-muted);
}

.sun-card {
  margin-bottom: 22rpx;
  padding: 26rpx;
  border: 1rpx solid rgba(255, 184, 77, 0.28);
  border-radius: 28rpx;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 14rpx 34rpx rgba(43, 42, 40, 0.08);
}

.sun-card-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  font-size: 28rpx;
  font-weight: 800;
  color: var(--sun-ink);
}

.sun-btn {
  min-height: 60rpx;
  border-radius: 999rpx;
  padding: 0 24rpx;
  background: var(--sun-action);
  color: #fff;
  font-size: 24rpx;
  font-weight: 700;
  line-height: 60rpx;
}

.sun-btn.secondary {
  background: #eef7ff;
  color: var(--sun-sky);
  border: 1rpx solid rgba(75, 159, 255, 0.22);
}

.sun-btn.danger {
  background: #fff1f3;
  color: #d95061;
  border: 1rpx solid rgba(239, 107, 122, 0.2);
}

.sun-chip {
  flex: 0 0 auto;
  padding: 14rpx 22rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.9);
  border: 1rpx solid var(--sun-line);
  color: #697180;
  font-size: 22rpx;
  font-weight: 700;
}

.sun-chip.active {
  color: #fff;
  background: var(--sun-sky);
  border-color: var(--sun-sky);
}
```

- [ ] **Step 4: Run style test**

Run:

```powershell
node parent-child-miniprogram/app-style.behavior.test.mjs
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add parent-child-miniprogram/App.vue parent-child-miniprogram/app-style.behavior.test.mjs
git commit -m "style: 增加晴日家庭感全局样式"
```

## Task 2: Page Registration And Titles

**Files:**
- Modify: `parent-child-miniprogram/pages.json`
- Test: `parent-child-miniprogram/pages.config.behavior.test.mjs`

- [ ] **Step 1: Create failing config test**

Create `parent-child-miniprogram/pages.config.behavior.test.mjs`:

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const config = JSON.parse(readFileSync(new URL('./pages.json', import.meta.url), 'utf8'))
const paths = config.pages.map((page) => page.path)

assert.ok(paths.includes('pages/reward-proxy/index'), '应注册代孩子兑换奖励页')
assert.ok(paths.includes('pages/members/index'), '应注册成员管理页')
assert.equal(config.pages.find((page) => page.path === 'pages/index/index').style.navigationBarTitleText, '首页')
assert.equal(config.pages.find((page) => page.path === 'pages/rewards/index').style.navigationBarTitleText, '奖励')
assert.equal(config.tabBar.list.find((item) => item.pagePath === 'pages/tasks/index').text, '任务')
```

- [ ] **Step 2: Run config test to verify it fails**

Run:

```powershell
node parent-child-miniprogram/pages.config.behavior.test.mjs
```

Expected: FAIL because new pages are missing or titles are garbled.

- [ ] **Step 3: Update `pages.json`**

Set Chinese titles and add non-tab pages:

```json
{
  "path": "pages/reward-proxy/index",
  "style": {
    "navigationBarTitleText": "代孩子兑换",
    "navigationStyle": "custom"
  }
},
{
  "path": "pages/members/index",
  "style": {
    "navigationBarTitleText": "成员管理",
    "navigationStyle": "custom"
  }
}
```

Keep tabBar paths unchanged: 首页、任务、奖励、我的.

- [ ] **Step 4: Run config test**

Run:

```powershell
node parent-child-miniprogram/pages.config.behavior.test.mjs
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add parent-child-miniprogram/pages.json parent-child-miniprogram/pages.config.behavior.test.mjs
git commit -m "chore: 注册奖励代兑和成员管理页面"
```

## Task 3: Reward Proxy Page

**Files:**
- Create: `parent-child-miniprogram/pages/reward-proxy/index.vue`
- Modify: `parent-child-miniprogram/pages/rewards/index.vue`
- Modify: `parent-child-miniprogram/pages/rewards/index.behavior.test.mjs`
- Test: `parent-child-miniprogram/pages/reward-proxy/index.behavior.test.mjs`

- [ ] **Step 1: Write behavior tests**

Create `parent-child-miniprogram/pages/reward-proxy/index.behavior.test.mjs`:

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

assert.ok(source.includes('代孩子兑换'), '页面应展示代孩子兑换标题')
assert.ok(source.includes('listMembers'), '页面应加载家庭成员')
assert.ok(source.includes('applyReward'), '页面应调用奖励兑换接口')
assert.ok(source.includes("member.roleType === 'CHILD'"), '页面应支持真实孩子和虚拟孩子')
assert.ok(!source.includes('currentPoints || 0 }} 积分</text></view><text'), '孩子选择项不应重复展示积分')
```

Update rewards test:

```js
assert.ok(source.includes('/pages/reward-proxy/index'), '奖励页应跳转到代孩子兑换页')
assert.ok(source.includes('奖励库'), '奖励页应保留奖励库模块')
assert.ok(source.includes('新建'), '奖励库模块应提供新建入口')
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```powershell
node parent-child-miniprogram/pages/rewards/index.behavior.test.mjs
node parent-child-miniprogram/pages/reward-proxy/index.behavior.test.mjs
```

Expected: FAIL because new page is missing.

- [ ] **Step 3: Create reward proxy page**

Move the current proxy reward state from `pages/rewards/index.vue` into `pages/reward-proxy/index.vue`: `members`、`selectedRewardVirtualChildId` should become `selectedChildId`; child filter should use `member.roleType === 'CHILD'`; reward rows use `applyReward({ familyId, rewardId, memberId: selectedChild.id })`.

Use `.sun-page`、`.sun-card`、`.sun-chip`、`.sun-btn` classes.

- [ ] **Step 4: Simplify rewards page**

In `pages/rewards/index.vue`:

- Replace proxy reward list with an entry card:

```vue
<view v-if="isParentRole" class="sun-card">
  <view class="sun-card-title">
    <text>代孩子兑换奖励</text>
    <button class="sun-btn secondary" size="mini" @click="goRewardProxy">进入</button>
  </view>
  <text class="sun-subtitle">先选孩子，再处理可兑换奖励，适合奖励较多的家庭。</text>
</view>
```

- Add:

```js
function goRewardProxy() {
  uni.navigateTo({ url: '/pages/reward-proxy/index' })
}
```

- Put “新建” inside the reward library header, not page header.

- [ ] **Step 5: Run reward tests**

Run:

```powershell
node parent-child-miniprogram/pages/rewards/index.behavior.test.mjs
node parent-child-miniprogram/pages/reward-proxy/index.behavior.test.mjs
```

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add parent-child-miniprogram/pages/rewards/index.vue parent-child-miniprogram/pages/rewards/index.behavior.test.mjs parent-child-miniprogram/pages/reward-proxy/index.vue parent-child-miniprogram/pages/reward-proxy/index.behavior.test.mjs
git commit -m "feat: 拆分代孩子兑换奖励页面"
```

## Task 4: Members Management Page

**Files:**
- Create: `parent-child-miniprogram/pages/members/index.vue`
- Modify: `parent-child-miniprogram/pages/profile/index.vue`
- Modify: `parent-child-miniprogram/pages/profile/index.behavior.test.mjs`
- Test: `parent-child-miniprogram/pages/members/index.behavior.test.mjs`

- [ ] **Step 1: Write behavior tests**

Create `parent-child-miniprogram/pages/members/index.behavior.test.mjs`:

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

assert.ok(source.includes('成员管理'), '页面应展示成员管理标题')
assert.ok(source.includes('createVirtualChild'), '成员管理页应能创建虚拟孩子')
assert.ok(source.includes('createInvite'), '成员管理页应能邀请成员')
assert.ok(source.includes('createVirtualChildBindInvite'), '成员管理页应能为虚拟孩子生成关联邀请')
```

Update profile test:

```js
assert.ok(source.includes('/pages/members/index'), '我的页应跳转到成员管理页')
assert.ok(!source.includes('先为没有微信账号的孩子建立积分身份'), '我的页不应直接展示创建虚拟孩子管理块')
assert.ok(!source.includes('创建一个带角色的邀请码'), '我的页不应直接展示邀请成员管理块')
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```powershell
node parent-child-miniprogram/pages/profile/index.behavior.test.mjs
node parent-child-miniprogram/pages/members/index.behavior.test.mjs
```

Expected: FAIL because members page is missing.

- [ ] **Step 3: Create members page**

Move virtual child creation, invite creation, invite token copy, member list, and virtual child bind invite logic from `profile/index.vue` to `members/index.vue`. Keep profile with summary only.

- [ ] **Step 4: Update profile page**

Add entry:

```vue
<view class="sun-card">
  <view class="sun-card-title">
    <text>家庭成员</text>
    <button class="sun-btn secondary" size="mini" @click="goMembers">管理</button>
  </view>
  <view class="member-list">
    <view class="member-card" v-for="member in members" :key="member.id">
      ...
    </view>
  </view>
</view>
```

Add:

```js
function goMembers() {
  uni.navigateTo({ url: '/pages/members/index' })
}
```

- [ ] **Step 5: Run member tests**

Run:

```powershell
node parent-child-miniprogram/pages/profile/index.behavior.test.mjs
node parent-child-miniprogram/pages/members/index.behavior.test.mjs
```

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add parent-child-miniprogram/pages/profile/index.vue parent-child-miniprogram/pages/profile/index.behavior.test.mjs parent-child-miniprogram/pages/members/index.vue parent-child-miniprogram/pages/members/index.behavior.test.mjs
git commit -m "feat: 拆分家庭成员管理页面"
```

## Task 5: Home, Family Select, Tasks, Points Visual Pass

**Files:**
- Modify: `parent-child-miniprogram/pages/family-select/index.vue`
- Modify: `parent-child-miniprogram/pages/index/index.vue`
- Modify: `parent-child-miniprogram/pages/tasks/index.vue`
- Modify: `parent-child-miniprogram/pages/points/index.vue`
- Modify: related behavior tests.

- [ ] **Step 1: Add behavior assertions**

Add assertions:

```js
assert.ok(source.includes('sun-page'), '页面应使用晴日家庭感页面基类')
assert.ok(source.includes('sun-card'), '页面应使用晴日家庭感卡片基类')
```

For home test, replace old assertion:

```js
assert.ok(!source.includes('查看全部任务'), '首页代孩子完成任务不应展示查看全部任务')
```

For points test:

```js
assert.ok(pageSource.includes('最近一个月'), '积分流水页应展示最近一个月范围')
assert.ok(pageSource.includes('selectedMemberId'), '积分流水页应保留成员筛选')
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```powershell
node parent-child-miniprogram/pages/family-select/index.behavior.test.mjs
node parent-child-miniprogram/pages/index/index.behavior.test.mjs
node parent-child-miniprogram/pages/tasks/index.behavior.test.mjs
node parent-child-miniprogram/pages/points/index.behavior.test.mjs
```

Expected: FAIL until pages are restyled.

- [ ] **Step 3: Restyle pages**

Apply the HTML wireframe structure from `docs/ui-wireframes/sunny-family-app.html`:

- `family-select`: create entry always visible; invite code folded and low weight.
- `index`: parent dashboard stats, child points card, no “查看全部任务”.
- `tasks`: form and list use sunny cards; parent actions stay “编辑/归档”.
- `points`: header says “最近一个月”， keeps member chips and today/yesterday grouping when data allows.

- [ ] **Step 4: Run page behavior tests**

Run:

```powershell
node parent-child-miniprogram/pages/family-select/index.behavior.test.mjs
node parent-child-miniprogram/pages/index/index.behavior.test.mjs
node parent-child-miniprogram/pages/tasks/index.behavior.test.mjs
node parent-child-miniprogram/pages/points/index.behavior.test.mjs
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add parent-child-miniprogram/pages/family-select/index.vue parent-child-miniprogram/pages/index/index.vue parent-child-miniprogram/pages/tasks/index.vue parent-child-miniprogram/pages/points/index.vue parent-child-miniprogram/pages/family-select/index.behavior.test.mjs parent-child-miniprogram/pages/index/index.behavior.test.mjs parent-child-miniprogram/pages/tasks/index.behavior.test.mjs parent-child-miniprogram/pages/points/index.behavior.test.mjs
git commit -m "style: 重设核心页面晴日家庭感界面"
```

## Task 6: Full Verification

**Files:**
- No source changes expected.

- [ ] **Step 1: Run all mini program behavior tests**

Run:

```powershell
Get-ChildItem parent-child-miniprogram/pages -Recurse -Filter *.behavior.test.mjs | ForEach-Object { node $_.FullName }
node parent-child-miniprogram/app-style.behavior.test.mjs
node parent-child-miniprogram/pages.config.behavior.test.mjs
```

Expected: PASS.

- [ ] **Step 2: Run Vue syntax checks**

Run:

```powershell
Get-ChildItem parent-child-miniprogram/pages -Recurse -Filter index.vue | ForEach-Object { node --check $_.FullName }
```

Expected: PASS for extracted JavaScript or no syntax errors where supported.

- [ ] **Step 3: Run Go tests**

Run:

```powershell
cd parent-child-api
go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 4: Commit only if verification updates files**

If no files changed, do not commit.

## Self-Review

- Spec coverage: Covers global style, family select, home, tasks, rewards, reward proxy, points, profile, and members management.
- Placeholder scan: No unresolved markers or vague styling steps.
- Type consistency: Uses existing API names and new page paths consistently: `pages/reward-proxy/index` and `pages/members/index`.
