# 家长端创建入口 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在小程序首页和奖励页为家长角色补齐“发布任务”和“新建奖励”的最小可用创建入口。

**Architecture:** 复用现有 uni-app 页面、API 封装和登录重试模式，不新增路由，不改服务端。首页只处理任务创建，奖励页只处理奖励创建，两个页面提交成功后刷新本页数据并恢复表单初始状态。

**Tech Stack:** uni-app、Vue 3 `<script setup>`、现有 `utils/request.js`、现有 Go 服务端接口。

---

## 文件结构

- Modify: `parent-child-miniprogram/api/task.js`
  - 增加 `createTask(data)`，封装 `POST /api/task/create`。
- Modify: `parent-child-miniprogram/api/reward.js`
  - 增加 `createReward(data)`，封装 `POST /api/reward/create`。
- Modify: `parent-child-miniprogram/pages/index/index.vue`
  - 家长角色统计卡下方增加发布任务面板。
  - 表单字段：任务标题、积分、周期。
  - 提交成功后关闭表单、重置表单、刷新首页。
- Modify: `parent-child-miniprogram/pages/rewards/index.vue`
  - 家长角色待发放区下方增加新建奖励面板。
  - 表单字段：奖励名称、所需积分、库存。
  - 库存填 `0` 表示不限库存，直接传给后端由现有接口归一化。
  - 提交成功后关闭表单、重置表单、刷新奖励页。

## 行为边界

- 只有 `OWNER`、`ADMIN`、`PARENT` 可以看到创建入口。
- `CHILD` 不显示发布任务入口，不显示新建奖励入口。
- 本次只做创建，不做编辑、下架、删除、归档。
- 本次不新增小程序页面路由。
- 表单校验失败使用 `uni.showToast({ icon: 'none' })`。
- 提交中按钮禁用，避免重复提交。
- 接口错误和 `401` 重试沿用页面已有加载函数和 `request.js` 行为。

---

### Task 1: 增加创建接口封装

**Files:**
- Modify: `parent-child-miniprogram/api/task.js`
- Modify: `parent-child-miniprogram/api/reward.js`

- [ ] **Step 1: 修改任务 API**

在 `parent-child-miniprogram/api/task.js` 末尾加入：

```js
export function createTask(data) {
  return request({
    url: '/api/task/create',
    method: 'POST',
    data
  })
}
```

- [ ] **Step 2: 修改奖励 API**

在 `parent-child-miniprogram/api/reward.js` 末尾加入：

```js
export function createReward(data) {
  return request({
    url: '/api/reward/create',
    method: 'POST',
    data
  })
}
```

- [ ] **Step 3: 运行 JS 语法检查**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
```

Expected: 命令退出码为 `0`，无 `SyntaxError`。

- [ ] **Step 4: 提交接口封装**

Run:

```powershell
git add parent-child-miniprogram\api\task.js parent-child-miniprogram\api\reward.js
git commit -m "feat: 增加小程序创建接口封装"
```

---

### Task 2: 首页增加发布任务入口

**Files:**
- Modify: `parent-child-miniprogram/pages/index/index.vue`

- [ ] **Step 1: 修改任务 API 导入**

把当前导入：

```js
import { auditTask, claimTask, listTaskRecords, listTasks, submitTask } from '../../api/task.js'
```

改为：

```js
import { auditTask, claimTask, createTask, listTaskRecords, listTasks, submitTask } from '../../api/task.js'
```

- [ ] **Step 2: 增加任务表单状态**

在已有 `parentPendingRecords` 状态后加入：

```js
const showTaskForm = ref(false)
const submittingTask = ref(false)
const taskForm = ref(defaultTaskForm())

const taskCycles = [
  { label: '一次性', value: 'ONCE' },
  { label: '每日', value: 'DAILY' },
  { label: '每周', value: 'WEEKLY' }
]
```

- [ ] **Step 3: 增加任务表单模板**

在家长视图的 `dashboard-stats` 区块之后、`<view class="section-title">待办事项</view>` 之前插入：

```vue
      <view v-if="isParentRole" class="create-panel">
        <view class="create-header">
          <view>
            <text class="create-title">发布任务</text>
            <text class="create-subtitle">给孩子增加一个可以领取的任务</text>
          </view>
          <button v-if="!showTaskForm" class="small-primary-btn" size="mini" @click="openTaskForm">发布</button>
        </view>

        <view v-if="showTaskForm" class="create-form">
          <view class="form-row">
            <text class="form-label">任务标题</text>
            <input v-model.trim="taskForm.title" class="form-input" placeholder="例如：阅读 30 分钟" />
          </view>
          <view class="form-row">
            <text class="form-label">奖励积分</text>
            <input v-model="taskForm.points" class="form-input" type="number" placeholder="10" />
          </view>
          <view class="form-row">
            <text class="form-label">任务周期</text>
            <view class="cycle-options">
              <button
                v-for="cycle in taskCycles"
                :key="cycle.value"
                class="cycle-btn"
                :class="{ active: taskForm.cycleType === cycle.value }"
                size="mini"
                @click="taskForm.cycleType = cycle.value"
              >
                {{ cycle.label }}
              </button>
            </view>
          </view>
          <view class="form-actions">
            <button class="plain-action-btn" size="mini" :disabled="submittingTask" @click="cancelTaskForm">取消</button>
            <button class="primary-action-btn" size="mini" :disabled="submittingTask" @click="publishTask">
              {{ submittingTask ? '发布中' : '确认发布' }}
            </button>
          </view>
        </view>
      </view>
```

- [ ] **Step 4: 增加任务表单函数**

在 `audit(record, approved)` 函数之后、`taskButtonText(task)` 之前加入：

```js
function defaultTaskForm() {
  return {
    title: '',
    points: 10,
    cycleType: 'DAILY'
  }
}

function resetTaskForm() {
  taskForm.value = defaultTaskForm()
}

function openTaskForm() {
  showTaskForm.value = true
}

function cancelTaskForm() {
  showTaskForm.value = false
  resetTaskForm()
}

function normalizePositiveInteger(value) {
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) ? parsed : 0
}

function validateTaskForm() {
  const title = taskForm.value.title.trim()
  const points = normalizePositiveInteger(taskForm.value.points)
  const cycleType = taskForm.value.cycleType

  if (!title) {
    uni.showToast({ title: '请填写任务标题', icon: 'none' })
    return null
  }
  if (points <= 0) {
    uni.showToast({ title: '奖励积分必须大于 0', icon: 'none' })
    return null
  }
  if (!taskCycles.some((cycle) => cycle.value === cycleType)) {
    uni.showToast({ title: '请选择任务周期', icon: 'none' })
    return null
  }

  return {
    title,
    points,
    cycleType
  }
}

async function publishTask() {
  if (submittingTask.value) {
    return
  }

  const payload = validateTaskForm()
  if (!payload) {
    return
  }

  submittingTask.value = true
  try {
    await createTask({
      ...payload,
      familyId: currentFamily.value.familyId
    })
    uni.showToast({ title: '任务已发布', icon: 'success' })
    showTaskForm.value = false
    resetTaskForm()
    await loadHome()
  } finally {
    submittingTask.value = false
  }
}
```

- [ ] **Step 5: 增加任务表单样式**

在 `parent-child-miniprogram/pages/index/index.vue` 的 `<style>` 中追加：

```css
.create-panel {
  background: #fff;
  border-radius: 10px;
  margin-bottom: 24px;
  padding: 16px;
}

.create-header {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.create-title,
.create-subtitle {
  display: block;
}

.create-title {
  color: #111827;
  font-size: 16px;
  font-weight: 700;
}

.create-subtitle {
  color: #64748b;
  font-size: 12px;
  margin-top: 4px;
}

.small-primary-btn,
.primary-action-btn {
  background: #2563eb;
  border: none;
  color: #fff;
}

.small-primary-btn {
  border-radius: 16px;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  padding: 0 14px;
}

.create-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-top: 16px;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  color: #374151;
  font-size: 13px;
  font-weight: 600;
}

.form-input {
  background: #f8fafc;
  border-radius: 8px;
  color: #111827;
  font-size: 14px;
  height: 40px;
  padding: 0 12px;
}

.cycle-options {
  display: flex;
  gap: 8px;
}

.cycle-btn {
  background: #f8fafc;
  border: 1px solid #dbe3ef;
  border-radius: 16px;
  color: #475569;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  padding: 0 14px;
}

.cycle-btn.active {
  background: #eff6ff;
  border-color: #2563eb;
  color: #2563eb;
}

.form-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

.plain-action-btn,
.primary-action-btn {
  border-radius: 16px;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  padding: 0 14px;
}

.plain-action-btn {
  background: #fff;
  border: 1px solid #cbd5e1;
  color: #475569;
}
```

- [ ] **Step 6: 运行页面文本和语法检查**

Run:

```powershell
node --check parent-child-miniprogram\api\task.js
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\index\index.vue
```

Expected:
- `node --check` 退出码为 `0`。
- `rg` 退出码为 `1`，表示没有命中乱码特征。

- [ ] **Step 7: 提交首页创建入口**

Run:

```powershell
git add parent-child-miniprogram\pages\index\index.vue
git commit -m "feat: 增加首页发布任务入口"
```

---

### Task 3: 奖励页增加新建奖励入口

**Files:**
- Modify: `parent-child-miniprogram/pages/rewards/index.vue`

- [ ] **Step 1: 修改奖励 API 导入**

把当前导入：

```js
import {
  applyReward,
  deliverReward,
  listRewardRecords,
  listRewards,
  receiveReward,
  rejectReward
} from '../../api/reward.js'
```

改为：

```js
import {
  applyReward,
  createReward,
  deliverReward,
  listRewardRecords,
  listRewards,
  receiveReward,
  rejectReward
} from '../../api/reward.js'
```

- [ ] **Step 2: 增加奖励表单状态**

在已有 `deliveredRecords` 状态后加入：

```js
const showRewardForm = ref(false)
const submittingReward = ref(false)
const rewardForm = ref(defaultRewardForm())
```

- [ ] **Step 3: 增加奖励表单模板**

在家长角色的“待发放申请”区块之后、奖励列表区块之前插入：

```vue
      <view v-if="isParentRole" class="create-panel">
        <view class="create-header">
          <view>
            <text class="create-title">新建奖励</text>
            <text class="create-subtitle">创建孩子可以用积分兑换的奖励</text>
          </view>
          <button v-if="!showRewardForm" class="small-primary-btn" size="mini" @click="openRewardForm">新建</button>
        </view>

        <view v-if="showRewardForm" class="create-form">
          <view class="form-row">
            <text class="form-label">奖励名称</text>
            <input v-model.trim="rewardForm.name" class="form-input" placeholder="例如：周末电影票" />
          </view>
          <view class="form-row">
            <text class="form-label">所需积分</text>
            <input v-model="rewardForm.pointsCost" class="form-input" type="number" placeholder="30" />
          </view>
          <view class="form-row">
            <text class="form-label">库存</text>
            <input v-model="rewardForm.stock" class="form-input" type="number" placeholder="0" />
            <text class="form-hint">库存填 0 表示不限库存</text>
          </view>
          <view class="form-actions">
            <button class="plain-action-btn" size="mini" :disabled="submittingReward" @click="cancelRewardForm">取消</button>
            <button class="primary-action-btn" size="mini" :disabled="submittingReward" @click="createNewReward">
              {{ submittingReward ? '创建中' : '确认创建' }}
            </button>
          </view>
        </view>
      </view>
```

- [ ] **Step 4: 增加奖励表单函数**

在 `receive(record)` 函数之后、`rewardButtonText(reward)` 之前加入：

```js
function defaultRewardForm() {
  return {
    name: '',
    pointsCost: 30,
    stock: 0
  }
}

function resetRewardForm() {
  rewardForm.value = defaultRewardForm()
}

function openRewardForm() {
  showRewardForm.value = true
}

function cancelRewardForm() {
  showRewardForm.value = false
  resetRewardForm()
}

function normalizePositiveInteger(value) {
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) ? parsed : 0
}

function normalizeNonNegativeInteger(value) {
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) ? parsed : -1
}

function validateRewardForm() {
  const name = rewardForm.value.name.trim()
  const pointsCost = normalizePositiveInteger(rewardForm.value.pointsCost)
  const stock = normalizeNonNegativeInteger(rewardForm.value.stock)

  if (!name) {
    uni.showToast({ title: '请填写奖励名称', icon: 'none' })
    return null
  }
  if (pointsCost <= 0) {
    uni.showToast({ title: '所需积分必须大于 0', icon: 'none' })
    return null
  }
  if (stock < 0) {
    uni.showToast({ title: '库存不能小于 0', icon: 'none' })
    return null
  }

  return {
    name,
    pointsCost,
    stock
  }
}

async function createNewReward() {
  if (submittingReward.value) {
    return
  }

  const payload = validateRewardForm()
  if (!payload) {
    return
  }

  submittingReward.value = true
  try {
    await createReward({
      ...payload,
      familyId: currentFamily.value.familyId
    })
    uni.showToast({ title: '奖励已创建', icon: 'success' })
    showRewardForm.value = false
    resetRewardForm()
    await loadRewardPage()
  } finally {
    submittingReward.value = false
  }
}
```

- [ ] **Step 5: 增加奖励表单样式**

在 `parent-child-miniprogram/pages/rewards/index.vue` 的 `<style>` 中追加：

```css
.create-panel {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
}

.create-header {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.create-title,
.create-subtitle {
  display: block;
}

.create-title {
  color: #111827;
  font-size: 16px;
  font-weight: 700;
}

.create-subtitle,
.form-hint {
  color: #64748b;
  font-size: 12px;
  margin-top: 4px;
}

.small-primary-btn,
.primary-action-btn {
  background: #2563eb;
  border: none;
  color: #fff;
}

.small-primary-btn {
  border-radius: 16px;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  padding: 0 14px;
}

.create-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-top: 16px;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  color: #374151;
  font-size: 13px;
  font-weight: 600;
}

.form-input {
  background: #f8fafc;
  border-radius: 8px;
  color: #111827;
  font-size: 14px;
  height: 40px;
  padding: 0 12px;
}

.form-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

.plain-action-btn,
.primary-action-btn {
  border-radius: 16px;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  padding: 0 14px;
}

.plain-action-btn {
  background: #fff;
  border: 1px solid #cbd5e1;
  color: #475569;
}
```

- [ ] **Step 6: 运行页面文本和语法检查**

Run:

```powershell
node --check parent-child-miniprogram\api\reward.js
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\rewards\index.vue
```

Expected:
- `node --check` 退出码为 `0`。
- `rg` 退出码为 `1`，表示没有命中乱码特征。

- [ ] **Step 7: 提交奖励页创建入口**

Run:

```powershell
git add parent-child-miniprogram\pages\rewards\index.vue
git commit -m "feat: 增加奖励页新建奖励入口"
```

---

### Task 4: 集成验证

**Files:**
- Verify: `parent-child-miniprogram/api/task.js`
- Verify: `parent-child-miniprogram/api/reward.js`
- Verify: `parent-child-miniprogram/pages/index/index.vue`
- Verify: `parent-child-miniprogram/pages/rewards/index.vue`
- Verify: `parent-child-api`

- [ ] **Step 1: 运行前端 JS 语法检查**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
```

Expected: 命令退出码为 `0`，无 `SyntaxError`。

- [ ] **Step 2: 扫描两个页面的乱码特征**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\index\index.vue parent-child-miniprogram\pages\rewards\index.vue
```

Expected: 退出码为 `1`，无匹配结果。

- [ ] **Step 3: 运行服务端测试**

Run:

```powershell
go test ./... -count=1
```

Working directory: `parent-child-api`

Expected: 所有 package 显示 `ok` 或 `?`，命令退出码为 `0`。如果沙箱因为 Go build cache 无权限失败，使用同一命令申请提升权限后重跑。

- [ ] **Step 4: 人工路径验证**

在微信开发者工具或 H5 预览中验证：

1. 家长角色进入首页，可以看到“发布任务”面板。
2. 点击“发布”，空标题提交会提示“请填写任务标题”。
3. 积分填 `0` 提交会提示“奖励积分必须大于 0”。
4. 标题、积分、周期有效时提交成功，提示“任务已发布”，首页任务列表刷新。
5. 孩子角色进入首页，看不到“发布任务”面板。
6. 家长角色进入奖励页，可以看到“新建奖励”面板。
7. 点击“新建”，空名称提交会提示“请填写奖励名称”。
8. 所需积分填 `0` 提交会提示“所需积分必须大于 0”。
9. 库存填负数提交会提示“库存不能小于 0”。
10. 名称、所需积分、库存有效时提交成功，提示“奖励已创建”，奖励列表刷新。
11. 孩子角色进入奖励页，看不到“新建奖励”面板。

- [ ] **Step 5: 提交验证修正**

如果集成验证产生修正，按实际修改文件提交：

```powershell
git add parent-child-miniprogram\api\task.js parent-child-miniprogram\api\reward.js parent-child-miniprogram\pages\index\index.vue parent-child-miniprogram\pages\rewards\index.vue
git commit -m "fix: 修正家长端创建入口"
```

如果没有产生修正，不创建空提交。

---

## 自检结果

- 设计稿中的任务创建、奖励创建、角色可见性、字段校验、提交禁用、成功刷新、错误处理边界均在 Task 1 至 Task 4 覆盖。
- 本计划不引入新路由，不扩展编辑、删除、下架能力。
- 接口名称和字段保持为 `createTask(data)`、`createReward(data)`、`familyId`、`title`、`points`、`cycleType`、`name`、`pointsCost`、`stock`。
- 计划中的验证命令覆盖 API 语法、页面乱码扫描和服务端测试。
