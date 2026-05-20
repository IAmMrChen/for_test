# 小程序奖励页真实数据接入 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让小程序奖励页接入后端真实奖励数据，支持孩子兑换、确认收到，以及家长发放、驳回奖励申请。

**Architecture:** 新增 `api/reward.js` 作为奖励接口 client，复用现有 `request.js`、`ensureDemoLogin`、`getCurrentFamily` 和 dashboard summary。奖励页单页完成角色分支、列表展示、操作按钮和刷新逻辑，不新增路由，不修改后端。

**Tech Stack:** uni-app Vue3、现有 `uni.request`/`uni.storage` API、后端 JWT demoLogin、Go 后端奖励接口、Node 语法检查、Go 测试。

---

## 文件结构

- Create: `parent-child-miniprogram/api/reward.js`
- Modify: `parent-child-miniprogram/pages/rewards/index.vue`
- Read only: `parent-child-miniprogram/api/auth.js`
- Read only: `parent-child-miniprogram/api/dashboard.js`
- Read only: `parent-child-miniprogram/utils/storage.js`
- Read only: `parent-child-miniprogram/utils/request.js`

---

### Task 1: 新增奖励 API client

**Files:**
- Create: `parent-child-miniprogram/api/reward.js`

- [ ] **Step 1: 创建 `api/reward.js`**

写入完整文件：

```js
import { request } from '../utils/request.js'

export function listRewards(familyId) {
  return request({
    url: '/api/reward/list',
    data: { familyId }
  })
}

export function listRewardRecords(params) {
  return request({
    url: '/api/reward/records',
    data: params
  })
}

export function applyReward(data) {
  return request({
    url: '/api/reward/apply',
    method: 'POST',
    data
  })
}

export function deliverReward(data) {
  return request({
    url: '/api/reward/deliver',
    method: 'POST',
    data
  })
}

export function rejectReward(data) {
  return request({
    url: '/api/reward/reject',
    method: 'POST',
    data
  })
}

export function receiveReward(data) {
  return request({
    url: '/api/reward/receive',
    method: 'POST',
    data
  })
}
```

- [ ] **Step 2: 检查 JS 语法**

Run:

```powershell
node --check parent-child-miniprogram\api\reward.js
```

Expected: exit 0。

- [ ] **Step 3: 提交**

```bash
git add parent-child-miniprogram/api/reward.js
git commit -m "feat: 增加小程序奖励接口"
```

---

### Task 2: 重写奖励页为真实数据页

**Files:**
- Modify: `parent-child-miniprogram/pages/rewards/index.vue`

- [ ] **Step 1: 替换奖励页模板、脚本和样式**

将 `parent-child-miniprogram/pages/rewards/index.vue` 替换为以下完整内容：

```vue
<template>
  <view class="container">
    <view class="header">
      <view>
        <text class="title">奖励</text>
        <text class="subtitle">{{ familyName }}</text>
      </view>
      <view class="summary-pill">
        <text class="summary-label">{{ isChild ? '当前积分' : '待发放' }}</text>
        <text class="summary-value">{{ isChild ? currentPoints : appliedRecords.length }}</text>
      </view>
    </view>

    <view v-if="loading" class="state-block">
      <text>正在加载奖励...</text>
    </view>

    <view v-else class="content">
      <view v-if="isChild && deliveredRecords.length > 0" class="section">
        <view class="section-title">待确认收到</view>
        <view class="record-card" v-for="record in deliveredRecords" :key="record.id">
          <view class="record-main">
            <text class="record-title">{{ record.rewardName }}</text>
            <text class="record-meta">消耗 {{ record.pointsCost }} 积分</text>
          </view>
          <button class="primary-btn" size="mini" @click="receive(record)">确认收到</button>
        </view>
      </view>

      <view v-if="isParentRole" class="section">
        <view class="section-title">待发放申请</view>
        <view v-if="appliedRecords.length === 0" class="state-block compact">
          <text>当前没有待发放奖励</text>
        </view>
        <view v-else class="record-card" v-for="record in appliedRecords" :key="record.id">
          <view class="record-main">
            <text class="record-title">{{ record.nickname || '孩子' }} 申请 {{ record.rewardName }}</text>
            <text class="record-meta">{{ record.pointsCost }} 积分 · {{ formatTime(record.applyTime) }}</text>
          </view>
          <view class="record-actions">
            <button class="plain-btn" size="mini" @click="reject(record)">驳回</button>
            <button class="primary-btn" size="mini" @click="deliver(record)">发放</button>
          </view>
        </view>
      </view>

      <view class="section">
        <view class="section-title">可兑换奖励</view>
        <view v-if="rewards.length === 0" class="state-block compact">
          <text>暂无可兑换奖励</text>
        </view>
        <view v-else class="reward-list">
          <view class="reward-card" v-for="reward in rewardRows" :key="reward.id">
            <view class="reward-info">
              <text class="reward-name">{{ reward.name }}</text>
              <text class="reward-stock">{{ stockText(reward.stock) }}</text>
            </view>
            <view class="reward-side">
              <text class="points">{{ reward.pointsCost }} 积分</text>
              <button
                v-if="isChild"
                class="exchange-btn"
                size="mini"
                :class="reward.actionStatus"
                :disabled="reward.actionStatus !== 'available'"
                @click="apply(reward)"
              >
                {{ rewardButtonText(reward) }}
              </button>
            </view>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'

import { ensureDemoLogin } from '../../api/auth.js'
import { getDashboardSummary } from '../../api/dashboard.js'
import {
  applyReward,
  deliverReward,
  listRewardRecords,
  listRewards,
  receiveReward,
  rejectReward
} from '../../api/reward.js'
import { getCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const currentFamily = ref(null)
const summary = ref({})
const rewards = ref([])
const appliedRecords = ref([])
const deliveredRecords = ref([])

const roleType = computed(() => summary.value.roleType || currentFamily.value?.roleType || '')
const isChild = computed(() => roleType.value === 'CHILD')
const isParentRole = computed(() => ['OWNER', 'ADMIN', 'PARENT'].includes(roleType.value))
const familyName = computed(() => summary.value.familyName || currentFamily.value?.familyName || '当前家庭')
const currentPoints = computed(() => summary.value.currentPoints || 0)

const rewardRows = computed(() => {
  return rewards.value.map((reward) => {
    const pending = appliedRecords.value.find((record) => record.rewardId === reward.id)
    if (pending) {
      return { ...reward, actionStatus: 'applied' }
    }
    if (reward.stock === 0) {
      return { ...reward, actionStatus: 'soldout' }
    }
    if (currentPoints.value < reward.pointsCost) {
      return { ...reward, actionStatus: 'insufficient' }
    }
    return { ...reward, actionStatus: 'available' }
  })
})

onShow(() => {
  loadRewardPage()
})

async function loadRewardPage(retried = false) {
  const family = getCurrentFamily()
  if (!family) {
    uni.reLaunch({ url: '/pages/family-select/index' })
    return
  }

  currentFamily.value = family
  loading.value = true
  try {
    await ensureDemoLogin()
    const familyId = family.familyId
    const [summaryData, rewardList] = await Promise.all([
      getDashboardSummary(familyId),
      listRewards(familyId)
    ])

    summary.value = summaryData || {}
    rewards.value = rewardList || []

    if (summary.value.roleType === 'CHILD') {
      const [applied, delivered] = await Promise.all([
        listRewardRecords({ familyId, status: 'APPLIED' }),
        listRewardRecords({ familyId, status: 'DELIVERED' })
      ])
      appliedRecords.value = applied || []
      deliveredRecords.value = delivered || []
      return
    }

    if (isParentRole.value) {
      appliedRecords.value = (await listRewardRecords({ familyId, status: 'APPLIED' })) || []
      deliveredRecords.value = []
      return
    }

    appliedRecords.value = []
    deliveredRecords.value = []
  } catch (error) {
    if (error.statusCode !== 401 || retried) {
      throw error
    }
    await ensureDemoLogin(true)
    await loadRewardPage(true)
  } finally {
    loading.value = false
  }
}

async function apply(reward) {
  await applyReward({
    familyId: currentFamily.value.familyId,
    rewardId: reward.id
  })
  uni.showToast({ title: '已申请兑换', icon: 'success' })
  await loadRewardPage()
}

async function deliver(record) {
  await deliverReward({ recordId: record.id })
  uni.showToast({ title: '已发放', icon: 'success' })
  await loadRewardPage()
}

async function reject(record) {
  await rejectReward({ recordId: record.id })
  uni.showToast({ title: '已驳回', icon: 'success' })
  await loadRewardPage()
}

async function receive(record) {
  await receiveReward({ recordId: record.id })
  uni.showToast({ title: '已确认收到', icon: 'success' })
  await loadRewardPage()
}

function rewardButtonText(reward) {
  const map = {
    available: '兑换',
    applied: '待发放',
    insufficient: '积分不足',
    soldout: '已兑完'
  }
  return map[reward.actionStatus] || '兑换'
}

function stockText(stock) {
  if (stock < 0) {
    return '不限库存'
  }
  if (stock === 0) {
    return '已兑完'
  }
  return `剩余 ${stock} 份`
}

function formatTime(value) {
  if (!value) {
    return ''
  }
  const date = new Date(value)
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  const hour = `${date.getHours()}`.padStart(2, '0')
  const minute = `${date.getMinutes()}`.padStart(2, '0')
  return `${month}-${day} ${hour}:${minute}`
}
</script>

<style>
.container {
  background: #f5f7fa;
  min-height: 100vh;
  padding: 48px 20px 24px;
}

.header {
  align-items: flex-start;
  display: flex;
  justify-content: space-between;
  margin-bottom: 24px;
}

.title,
.subtitle {
  display: block;
}

.title {
  color: #1f2937;
  font-size: 26px;
  font-weight: 700;
  margin-bottom: 6px;
}

.subtitle,
.summary-label,
.record-meta,
.reward-stock {
  color: #64748b;
  font-size: 12px;
}

.summary-pill {
  align-items: flex-end;
  background: #fff;
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  min-width: 78px;
  padding: 10px 12px;
}

.summary-value {
  color: #2563eb;
  font-size: 20px;
  font-weight: 700;
}

.content,
.section,
.reward-list {
  display: flex;
  flex-direction: column;
}

.content,
.section {
  gap: 18px;
}

.section-title {
  color: #1f2937;
  font-size: 16px;
  font-weight: 700;
}

.reward-list,
.record-card + .record-card {
  gap: 12px;
}

.reward-card,
.record-card,
.state-block {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
}

.reward-card,
.record-card {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.reward-info,
.reward-side,
.record-main {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.reward-side {
  align-items: flex-end;
}

.reward-name,
.record-title {
  color: #111827;
  font-size: 15px;
  font-weight: 700;
}

.points {
  color: #f59e0b;
  font-size: 13px;
  font-weight: 700;
}

.record-actions {
  display: flex;
  gap: 8px;
}

.exchange-btn,
.primary-btn,
.plain-btn {
  border-radius: 16px;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  min-width: 72px;
  padding: 0 12px;
}

.exchange-btn,
.primary-btn {
  background: #3b82f6;
  border: none;
  color: #fff;
}

.plain-btn {
  background: #fff;
  border: 1px solid #cbd5e1;
  color: #475569;
}

.exchange-btn.applied,
.exchange-btn.insufficient,
.exchange-btn.soldout {
  background: #e5e7eb;
  color: #64748b;
}

.state-block {
  color: #64748b;
  font-size: 14px;
  text-align: center;
}

.state-block.compact {
  padding: 20px 16px;
}
</style>
```

- [ ] **Step 2: 检查奖励页源码没有明显乱码片段**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\rewards\index.vue
```

Expected: exit 1，无匹配。

- [ ] **Step 3: 提交**

```bash
git add parent-child-miniprogram/pages/rewards/index.vue
git commit -m "feat: 接入奖励页真实数据"
```

---

### Task 3: 收尾验证

**Files:**
- No source changes expected.

- [ ] **Step 1: 检查前端 JS 语法**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
```

Expected: exit 0。

- [ ] **Step 2: 检查奖励页源码没有明显乱码**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\rewards\index.vue
```

Expected: exit 1，无匹配。

- [ ] **Step 3: 运行后端全量测试**

Run:

```powershell
go test ./... -count=1
```

Workdir: `parent-child-api`

Expected: PASS。若沙箱拦截 Go build cache，按权限规则在沙箱外重新运行同一命令。

- [ ] **Step 4: 检查工作区状态**

Run:

```bash
git status --short --branch
```

Expected: `## main`。

---

## 自检

- Spec coverage: 覆盖奖励 API client、奖励列表、孩子兑换、孩子确认收到、家长发放、家长驳回、loading、空状态、401 重试和验证命令。
- Placeholder scan: 计划中没有未决占位词或含糊步骤。
- Type consistency: `listRewards`、`listRewardRecords`、`applyReward`、`deliverReward`、`rejectReward`、`receiveReward` 与奖励页导入名一致。
- Scope control: 不新增奖励创建、编辑、上下架、图片、分类、搜索、分页和新路由。
