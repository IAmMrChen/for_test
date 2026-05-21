# 小程序家长代虚拟孩子操作 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在首页和奖励页补齐家长代虚拟孩子领取/提交任务、申请兑换奖励的 MVP 闭环。

**Architecture:** 不改服务端，复用现有 `claimTask`、`submitTask`、`applyReward` 的 `memberId` 参数。首页和奖励页各自加载家庭成员，筛选虚拟孩子，并在页面本地维护选中孩子和提交中状态。

**Tech Stack:** uni-app、Vue 3 `<script setup>`、现有小程序 API 封装、现有 Go 服务端接口。

---

## 文件结构

- Modify: `parent-child-miniprogram/pages/index/index.vue`
  - 导入 `listMembers`。
  - 家长视图加载虚拟孩子和 `CLAIMED` 任务记录。
  - 增加“代孩子完成任务”区块。
  - 代孩子领取任务、提交任务时传入 `memberId`。
- Modify: `parent-child-miniprogram/pages/rewards/index.vue`
  - 导入 `listMembers`。
  - 家长视图加载虚拟孩子。
  - 增加“代孩子兑换奖励”区块。
  - 代孩子申请奖励时传入 `memberId`。

## 行为边界

- 只代理 `roleType === 'CHILD' && isVirtual === true` 的成员。
- 不代理真实孩子。
- 不新增后端接口。
- 不新增页面路由。
- 不支持批量操作。
- 不支持代孩子确认收到奖励。
- 不支持提交备注输入。

---

### Task 1: 首页增加代虚拟孩子任务区

**Files:**
- Modify: `parent-child-miniprogram/pages/index/index.vue`

- [ ] **Step 1: 增加成员 API 导入**

把导入区补成：

```js
import { listMembers } from '../../api/member.js'
```

- [ ] **Step 2: 增加状态和计算属性**

在现有状态后加入：

```js
const members = ref([])
const proxyClaimedRecords = ref([])
const selectedVirtualChildId = ref(0)
const submittingProxyTaskId = ref(0)
```

在 `taskRows` 之后加入：

```js
const virtualChildren = computed(() => {
  return members.value.filter((member) => member.roleType === 'CHILD' && member.isVirtual)
})

const selectedVirtualChild = computed(() => {
  return virtualChildren.value.find((member) => member.id === selectedVirtualChildId.value) || null
})

const proxyTaskRows = computed(() => {
  const child = selectedVirtualChild.value
  if (!child) {
    return []
  }

  return tasks.value.map((task) => {
    const claimed = proxyClaimedRecords.value.find((record) => record.taskId === task.id && record.memberId === child.id)
    const pending = parentPendingRecords.value.find((record) => record.taskId === task.id && record.memberId === child.id)
    if (pending) {
      return { ...task, viewStatus: 'pending', record: pending }
    }
    if (claimed) {
      return { ...task, viewStatus: 'claimed', record: claimed }
    }
    return { ...task, viewStatus: 'claimable', record: null }
  })
})
```

- [ ] **Step 3: 家长加载数据时补成员和已领取记录**

在 `loadHome` 的家长分支中，把当前逻辑改为同时加载：

```js
    } else if (isParentRole.value) {
      const [memberList, claimed, pending] = await Promise.all([
        listMembers(familyId),
        listTaskRecords({ familyId, status: 'CLAIMED' }),
        listTaskRecords({ familyId, status: 'PENDING' })
      ])
      members.value = memberList || []
      proxyClaimedRecords.value = claimed || []
      parentPendingRecords.value = pending || []
      syncSelectedVirtualChild()
      claimedRecords.value = []
      pendingRecords.value = []
    }
```

在孩子分支中清空新增家长数据：

```js
      members.value = []
      proxyClaimedRecords.value = []
      selectedVirtualChildId.value = 0
```

- [ ] **Step 4: 增加代操作模板**

在发布任务 `create-panel` 之后、待办事项标题之前插入：

```vue
      <view v-if="isParentRole" class="proxy-panel">
        <view class="proxy-header">
          <view>
            <text class="proxy-title">代孩子完成任务</text>
            <text class="proxy-subtitle">仅支持虚拟孩子</text>
          </view>
          <text v-if="selectedVirtualChild" class="proxy-points">{{ selectedVirtualChild.currentPoints || 0 }} 积分</text>
        </view>

        <view v-if="virtualChildren.length === 0" class="state-block compact">
          <text>暂无虚拟孩子，请先在个人页创建</text>
        </view>
        <view v-else class="proxy-content">
          <view class="child-options">
            <button
              v-for="child in virtualChildren"
              :key="child.id"
              class="child-option-btn"
              :class="{ active: selectedVirtualChildId === child.id }"
              size="mini"
              @click="selectVirtualChild(child.id)"
            >
              {{ child.nickname }}
            </button>
          </view>

          <view class="proxy-task-list">
            <view class="task-item" v-for="task in proxyTaskRows" :key="task.id">
              <view class="task-info">
                <view class="task-title">{{ task.title }}</view>
                <view class="task-reward">+{{ task.points }} 积分</view>
              </view>
              <button
                class="action-btn"
                :class="task.viewStatus"
                :disabled="task.viewStatus === 'pending' || submittingProxyTaskId === task.id"
                @click="handleProxyTaskAction(task)"
              >
                {{ proxyTaskButtonText(task) }}
              </button>
            </view>
          </view>
        </view>
      </view>
```

- [ ] **Step 5: 增加代操作函数**

在 `publishTask()` 后加入：

```js
function syncSelectedVirtualChild() {
  if (virtualChildren.value.some((child) => child.id === selectedVirtualChildId.value)) {
    return
  }
  selectedVirtualChildId.value = virtualChildren.value[0]?.id || 0
}

function selectVirtualChild(memberId) {
  selectedVirtualChildId.value = memberId
}

async function handleProxyTaskAction(task) {
  const child = selectedVirtualChild.value
  if (!child) {
    uni.showToast({ title: '请先选择孩子', icon: 'none' })
    return
  }

  submittingProxyTaskId.value = task.id
  try {
    const familyId = currentFamily.value.familyId
    if (task.viewStatus === 'claimable') {
      await claimTask({ familyId, taskId: task.id, memberId: child.id })
      uni.showToast({ title: '已为孩子领取任务', icon: 'success' })
      await loadHome()
      return
    }
    if (task.viewStatus === 'claimed') {
      await submitTask({ familyId, recordId: task.record.id, memberId: child.id })
      uni.showToast({ title: '已提交审核', icon: 'success' })
      await loadHome()
    }
  } finally {
    submittingProxyTaskId.value = 0
  }
}

function proxyTaskButtonText(task) {
  if (submittingProxyTaskId.value === task.id) {
    return '处理中'
  }
  return taskButtonText(task)
}
```

- [ ] **Step 6: 增加样式**

在 `<style>` 中追加：

```css
.proxy-panel {
  background: #fff;
  border-radius: 10px;
  margin-bottom: 24px;
  padding: 16px;
}

.proxy-header {
  align-items: center;
  display: flex;
  justify-content: space-between;
  margin-bottom: 14px;
}

.proxy-title,
.proxy-subtitle {
  display: block;
}

.proxy-title {
  color: #111827;
  font-size: 16px;
  font-weight: 700;
}

.proxy-subtitle {
  color: #64748b;
  font-size: 12px;
  margin-top: 4px;
}

.proxy-points {
  color: #f59e0b;
  font-size: 13px;
  font-weight: 700;
}

.proxy-content,
.proxy-task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.child-options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.child-option-btn {
  background: #f8fafc;
  border: 1px solid #dbe3ef;
  border-radius: 16px;
  color: #475569;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  padding: 0 14px;
}

.child-option-btn.active {
  background: #eff6ff;
  border-color: #2563eb;
  color: #2563eb;
}

.state-block.compact {
  padding: 20px 16px;
}
```

- [ ] **Step 7: 验证并提交首页**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\index\index.vue
git diff --check
```

Expected:
- `rg` 退出码为 `1`。
- `git diff --check` 退出码为 `0`。

Commit:

```powershell
git add parent-child-miniprogram\pages\index\index.vue
git commit -m "feat: 增加代虚拟孩子任务入口"
```

---

### Task 2: 奖励页增加代虚拟孩子兑换区

**Files:**
- Modify: `parent-child-miniprogram/pages/rewards/index.vue`

- [ ] **Step 1: 增加成员 API 导入**

在导入区加入：

```js
import { listMembers } from '../../api/member.js'
```

- [ ] **Step 2: 增加状态和计算属性**

在现有状态后加入：

```js
const members = ref([])
const selectedRewardVirtualChildId = ref(0)
const submittingProxyRewardId = ref(0)
```

在 `rewardRows` 之后加入：

```js
const virtualChildren = computed(() => {
  return members.value.filter((member) => member.roleType === 'CHILD' && member.isVirtual)
})

const selectedRewardVirtualChild = computed(() => {
  return virtualChildren.value.find((member) => member.id === selectedRewardVirtualChildId.value) || null
})

const proxyRewardRows = computed(() => {
  const child = selectedRewardVirtualChild.value
  if (!child) {
    return []
  }

  return rewards.value.map((reward) => {
    const pending = appliedRecords.value.find((record) => record.rewardId === reward.id && record.memberId === child.id)
    if (pending) {
      return { ...reward, actionStatus: 'applied' }
    }
    if (reward.stock === 0) {
      return { ...reward, actionStatus: 'soldout' }
    }
    if ((child.currentPoints || 0) < reward.pointsCost) {
      return { ...reward, actionStatus: 'insufficient' }
    }
    return { ...reward, actionStatus: 'available' }
  })
})
```

- [ ] **Step 3: 家长加载数据时补成员**

在 `loadRewardPage` 的家长分支中，把当前逻辑改为：

```js
    if (isParentRole.value) {
      const [applied, memberList] = await Promise.all([
        listRewardRecords({ familyId, status: 'APPLIED' }),
        listMembers(familyId)
      ])
      appliedRecords.value = applied || []
      members.value = memberList || []
      deliveredRecords.value = []
      syncSelectedRewardVirtualChild()
      return
    }
```

在孩子分支中清空：

```js
      members.value = []
      selectedRewardVirtualChildId.value = 0
```

- [ ] **Step 4: 增加代兑换模板**

在新建奖励 `create-panel` 之后、奖励列表之前插入：

```vue
      <view v-if="isParentRole" class="proxy-panel">
        <view class="proxy-header">
          <view>
            <text class="proxy-title">代孩子兑换奖励</text>
            <text class="proxy-subtitle">仅支持虚拟孩子</text>
          </view>
          <text v-if="selectedRewardVirtualChild" class="proxy-points">{{ selectedRewardVirtualChild.currentPoints || 0 }} 积分</text>
        </view>

        <view v-if="virtualChildren.length === 0" class="state-block compact">
          <text>暂无虚拟孩子，请先在个人页创建</text>
        </view>
        <view v-else class="proxy-content">
          <view class="child-options">
            <button
              v-for="child in virtualChildren"
              :key="child.id"
              class="child-option-btn"
              :class="{ active: selectedRewardVirtualChildId === child.id }"
              size="mini"
              @click="selectRewardVirtualChild(child.id)"
            >
              {{ child.nickname }}
            </button>
          </view>

          <view class="reward-list">
            <view class="reward-card" v-for="reward in proxyRewardRows" :key="reward.id">
              <view class="reward-info">
                <text class="reward-name">{{ reward.name }}</text>
                <text class="reward-stock">{{ stockText(reward.stock) }}</text>
              </view>
              <view class="reward-side">
                <text class="points">{{ reward.pointsCost }} 积分</text>
                <button
                  class="exchange-btn"
                  size="mini"
                  :class="reward.actionStatus"
                  :disabled="reward.actionStatus !== 'available' || submittingProxyRewardId === reward.id"
                  @click="applyProxyReward(reward)"
                >
                  {{ proxyRewardButtonText(reward) }}
                </button>
              </view>
            </view>
          </view>
        </view>
      </view>
```

- [ ] **Step 5: 增加代兑换函数**

在 `createNewReward()` 后加入：

```js
function syncSelectedRewardVirtualChild() {
  if (virtualChildren.value.some((child) => child.id === selectedRewardVirtualChildId.value)) {
    return
  }
  selectedRewardVirtualChildId.value = virtualChildren.value[0]?.id || 0
}

function selectRewardVirtualChild(memberId) {
  selectedRewardVirtualChildId.value = memberId
}

async function applyProxyReward(reward) {
  const child = selectedRewardVirtualChild.value
  if (!child) {
    uni.showToast({ title: '请先选择孩子', icon: 'none' })
    return
  }

  submittingProxyRewardId.value = reward.id
  try {
    await applyReward({
      familyId: currentFamily.value.familyId,
      rewardId: reward.id,
      memberId: child.id
    })
    uni.showToast({ title: '已为孩子申请兑换', icon: 'success' })
    await loadRewardPage()
  } finally {
    submittingProxyRewardId.value = 0
  }
}

function proxyRewardButtonText(reward) {
  if (submittingProxyRewardId.value === reward.id) {
    return '处理中'
  }
  if (reward.actionStatus === 'available') {
    return '代兑换'
  }
  return rewardButtonText(reward)
}
```

- [ ] **Step 6: 增加样式**

在 `<style>` 中追加与首页一致的代理区样式：

```css
.proxy-panel {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
}

.proxy-header {
  align-items: center;
  display: flex;
  justify-content: space-between;
  margin-bottom: 14px;
}

.proxy-title,
.proxy-subtitle {
  display: block;
}

.proxy-title {
  color: #111827;
  font-size: 16px;
  font-weight: 700;
}

.proxy-subtitle {
  color: #64748b;
  font-size: 12px;
  margin-top: 4px;
}

.proxy-points {
  color: #f59e0b;
  font-size: 13px;
  font-weight: 700;
}

.proxy-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.child-options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.child-option-btn {
  background: #f8fafc;
  border: 1px solid #dbe3ef;
  border-radius: 16px;
  color: #475569;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  padding: 0 14px;
}

.child-option-btn.active {
  background: #eff6ff;
  border-color: #2563eb;
  color: #2563eb;
}
```

- [ ] **Step 7: 验证并提交奖励页**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\rewards\index.vue
git diff --check
```

Expected:
- `rg` 退出码为 `1`。
- `git diff --check` 退出码为 `0`。

Commit:

```powershell
git add parent-child-miniprogram\pages\rewards\index.vue
git commit -m "feat: 增加代虚拟孩子兑换入口"
```

---

### Task 3: 集成验证

**Files:**
- Verify: `parent-child-miniprogram/pages/index/index.vue`
- Verify: `parent-child-miniprogram/pages/rewards/index.vue`
- Verify: `parent-child-api`

- [ ] **Step 1: 运行前端 JS 语法检查**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
```

Expected: 命令退出码为 `0`，无 `SyntaxError`。

- [ ] **Step 2: 扫描页面乱码特征**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\index\index.vue parent-child-miniprogram\pages\rewards\index.vue
```

Expected: 退出码为 `1`，没有乱码特征命中。

- [ ] **Step 3: 运行服务端测试**

Run:

```powershell
go test ./... -count=1
```

Working directory: `parent-child-api`

Expected: 所有 package 显示 `ok` 或 `?`，命令退出码为 `0`。如果沙箱因为 Go build cache 无权限失败，使用同一命令申请提升权限后重跑。

- [ ] **Step 4: 人工路径验证**

在微信开发者工具或 H5 预览中验证：

1. 家长首页有虚拟孩子时显示“代孩子完成任务”。
2. 家长首页无虚拟孩子时显示空态。
3. 选择虚拟孩子后可以代其领取任务。
4. 已领取任务可代其提交，提交后进入待审核。
5. 奖励页有虚拟孩子时显示“代孩子兑换奖励”。
6. 奖励页无虚拟孩子时显示空态。
7. 积分不足、待发放、已兑完按钮禁用。
8. 可兑换奖励可代虚拟孩子发起兑换申请。
9. 真实孩子不出现在代理选择中。

---

## 自检结果

- 设计规格中的首页代领取、代提交、奖励页代兑换、虚拟孩子过滤、选中状态同步、错误处理和验证均已覆盖。
- 本计划不修改服务端，不新增路由，不代理真实孩子。
- 字段名与现有接口保持一致：`familyId`、`taskId`、`recordId`、`rewardId`、`memberId`。
