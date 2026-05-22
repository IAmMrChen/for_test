# 首页奖励待办 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在小程序首页补齐家长待发放奖励和孩子待确认奖励入口。

**Architecture:** 不新增后端能力，首页复用现有 `api/reward.js` 的奖励记录查询与操作接口。页面加载时按当前家庭和角色补充奖励记录数据，操作成功后重新加载首页。

**Tech Stack:** uni-app Vue 3 `<script setup>`、现有小程序 API 封装、Go 后端回归测试。

---

## 文件结构

- Modify: `parent-child-miniprogram/pages/index/index.vue`
  - 引入奖励 API。
  - 增加家长待发放奖励数据。
  - 增加孩子待确认奖励数据。
  - 增加发放、拒绝、确认收到操作。
  - 增加少量样式。

---

### Task 1: 首页奖励待办数据接入

**Files:**
- Modify: `parent-child-miniprogram/pages/index/index.vue`

- [ ] **Step 1: 引入奖励 API**

在首页脚本中增加：

```js
import { deliverReward, listRewardRecords, receiveReward, rejectReward } from '../../api/reward.js'
```

- [ ] **Step 2: 增加状态**

在现有 `parentPendingRecords` 附近增加：

```js
const parentAppliedRewardRecords = ref([])
const childDeliveredRewardRecords = ref([])
```

- [ ] **Step 3: 孩子视角加载待确认奖励**

在 `summary.value.roleType === 'CHILD'` 分支中，把 Promise 扩展为：

```js
const [claimed, pending, deliveredRewards] = await Promise.all([
  listTaskRecords({ familyId, status: 'CLAIMED' }),
  listTaskRecords({ familyId, status: 'PENDING' }),
  listRewardRecords({ familyId, status: 'DELIVERED' })
])
```

并设置：

```js
childDeliveredRewardRecords.value = deliveredRewards || []
parentAppliedRewardRecords.value = []
```

- [ ] **Step 4: 家长视角加载待发放奖励**

在家长分支中，把 Promise 扩展为：

```js
const [memberList, claimed, pending, appliedRewards] = await Promise.all([
  listMembers(familyId),
  listTaskRecords({ familyId, status: 'CLAIMED' }),
  listTaskRecords({ familyId, status: 'PENDING' }),
  listRewardRecords({ familyId, status: 'APPLIED' })
])
```

并设置：

```js
parentAppliedRewardRecords.value = appliedRewards || []
childDeliveredRewardRecords.value = []
```

---

### Task 2: 首页奖励待办 UI 和操作

**Files:**
- Modify: `parent-child-miniprogram/pages/index/index.vue`

- [ ] **Step 1: 孩子首页增加待确认奖励区块**

在孩子视角任务列表之前增加：

```vue
<view v-if="childDeliveredRewardRecords.length > 0" class="section-block">
  <view class="section-title">待确认奖励</view>
  <view class="reward-todo-list">
    <view class="reward-todo-item" v-for="record in childDeliveredRewardRecords" :key="record.id">
      <view class="reward-todo-info">
        <text class="reward-todo-title">{{ record.rewardName }}</text>
        <text class="reward-todo-meta">消耗 {{ record.pointsCost }} 积分 · {{ formatTime(record.operateTime || record.applyTime) }}</text>
      </view>
      <button class="primary-action-btn" size="mini" @click="receiveDeliveredReward(record)">确认收到</button>
    </view>
  </view>
</view>
```

- [ ] **Step 2: 家长首页增加待发放奖励区块**

在家长视角任务审核列表后增加：

```vue
<view class="section-title reward-section-title">待发放奖励</view>
<view v-if="parentAppliedRewardRecords.length === 0" class="state-block">
  <text>当前没有待发放奖励</text>
</view>
<view v-else class="reward-todo-list">
  <view class="reward-todo-item" v-for="record in parentAppliedRewardRecords" :key="record.id">
    <view class="reward-todo-info">
      <text class="reward-todo-title">{{ record.nickname || '孩子' }} 申请 {{ record.rewardName }}</text>
      <text class="reward-todo-meta">{{ record.pointsCost }} 积分 · {{ formatTime(record.applyTime) }}</text>
    </view>
    <view class="audit-actions">
      <button class="btn-reject" size="mini" plain @click="operateReward(record, false)">拒绝</button>
      <button class="btn-approve" size="mini" type="primary" @click="operateReward(record, true)">发放</button>
    </view>
  </view>
</view>
```

- [ ] **Step 3: 增加操作方法**

在 `audit` 方法后增加：

```js
async function operateReward(record, delivered) {
  if (delivered) {
    await deliverReward({ recordId: record.id })
  } else {
    await rejectReward({ recordId: record.id })
  }
  uni.showToast({
    title: delivered ? '已发放' : '已拒绝',
    icon: 'success'
  })
  await loadHome()
}

async function receiveDeliveredReward(record) {
  await receiveReward({ recordId: record.id })
  uni.showToast({ title: '已确认收到', icon: 'success' })
  await loadHome()
}
```

- [ ] **Step 4: 增加样式**

在首页样式中增加：

```css
.section-block {
  margin-bottom: 24px;
}

.reward-section-title {
  margin-top: 24px;
}

.reward-todo-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.reward-todo-item {
  align-items: center;
  background: #fff;
  border-radius: 10px;
  display: flex;
  justify-content: space-between;
  padding: 16px;
}

.reward-todo-info {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.reward-todo-title {
  color: #111827;
  font-size: 15px;
  font-weight: 700;
}

.reward-todo-meta {
  color: #64748b;
  font-size: 12px;
}
```

---

### Task 3: 验证和提交

**Files:**
- Verify: `parent-child-miniprogram/pages/index/index.vue`

- [ ] **Step 1: 前端 JS 语法检查**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
```

Expected: exit 0。

- [ ] **Step 2: 首页乱码扫描**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\index\index.vue
```

Expected: exit 1，无匹配。

- [ ] **Step 3: 后端回归测试**

Run:

```powershell
go test ./... -count=1
```

Working directory: `parent-child-api`

Expected: exit 0。

- [ ] **Step 4: 提交实现**

```powershell
git add parent-child-miniprogram\pages\index\index.vue
git commit -m "feat: 增加首页奖励待办入口"
```

## 自查

- 家长首页待发放奖励：Task 1、Task 2 覆盖。
- 孩子首页待确认奖励：Task 1、Task 2 覆盖。
- 不新增后端接口：计划只改首页。
- 验证包含前端语法、乱码扫描和后端回归。
