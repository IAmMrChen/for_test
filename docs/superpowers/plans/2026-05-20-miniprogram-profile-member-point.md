# 小程序个人页成员与积分流水接入 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让小程序个人页接入家庭成员和积分流水真实数据，替换“家庭成员”“积分流水”的占位提示。

**Architecture:** 新增 `api/member.js` 与 `api/point.js` 两个小程序接口 client，复用现有 `request.js`、`ensureDemoLogin`、`getCurrentFamily` 和 dashboard summary。个人页单页展示当前身份、家庭成员列表和积分流水，不新增路由，不引入成员管理写操作。

**Tech Stack:** uni-app Vue3、现有 `uni.request`/`uni.storage` API、后端 JWT demoLogin、Go 后端成员与积分接口、Node 语法检查、Go 测试。

---

## 文件结构

- Create: `parent-child-miniprogram/api/member.js`
- Create: `parent-child-miniprogram/api/point.js`
- Modify: `parent-child-miniprogram/pages/profile/index.vue`
- Read only: `parent-child-miniprogram/api/auth.js`
- Read only: `parent-child-miniprogram/api/dashboard.js`
- Read only: `parent-child-miniprogram/utils/storage.js`
- Read only: `parent-child-miniprogram/utils/request.js`

---

### Task 1: 新增成员与积分 API client

**Files:**
- Create: `parent-child-miniprogram/api/member.js`
- Create: `parent-child-miniprogram/api/point.js`

- [ ] **Step 1: 创建 `api/member.js`**

写入完整文件：

```js
import { request } from '../utils/request.js'

export function listMembers(familyId) {
  return request({
    url: '/api/member/list',
    data: { familyId }
  })
}
```

- [ ] **Step 2: 创建 `api/point.js`**

写入完整文件：

```js
import { request } from '../utils/request.js'

export function listPointLogs(params) {
  return request({
    url: '/api/point/logs',
    data: params
  })
}
```

- [ ] **Step 3: 检查 JS 语法**

Run:

```powershell
node --check parent-child-miniprogram\api\member.js
node --check parent-child-miniprogram\api\point.js
```

Expected: 两条命令均 exit 0。

- [ ] **Step 4: 提交**

```bash
git add parent-child-miniprogram/api/member.js parent-child-miniprogram/api/point.js
git commit -m "feat: 增加小程序成员与积分接口"
```

---

### Task 2: 重写个人页为真实数据页

**Files:**
- Modify: `parent-child-miniprogram/pages/profile/index.vue`

- [ ] **Step 1: 替换个人页模板、脚本和样式**

将 `parent-child-miniprogram/pages/profile/index.vue` 替换为以下完整内容：

```vue
<template>
  <view class="container">
    <view class="user-card">
      <view class="avatar">{{ avatarText }}</view>
      <view class="info">
        <text class="nickname">{{ nickname }}</text>
        <text class="current-family">{{ familyName }} · {{ roleName(roleType) }}</text>
      </view>
      <button class="switch-btn" size="mini" @click="switchFamily">切换</button>
    </view>

    <view class="points-row">
      <view class="point-card">
        <text class="point-value">{{ currentPoints }}</text>
        <text class="point-label">当前积分</text>
      </view>
      <view class="point-card">
        <text class="point-value">{{ totalEarnedPoints }}</text>
        <text class="point-label">累计积分</text>
      </view>
    </view>

    <view v-if="loading" class="state-block">
      <text>正在加载个人页...</text>
    </view>

    <view v-else class="content">
      <view class="section">
        <view class="section-title">家庭成员</view>
        <view v-if="members.length === 0" class="state-block compact">
          <text>暂无家庭成员</text>
        </view>
        <view v-else class="member-list">
          <view class="member-card" v-for="member in members" :key="member.id">
            <view class="member-main">
              <text class="member-name">{{ member.nickname }}</text>
              <view class="tag-row">
                <text class="role-tag">{{ roleName(member.roleType) }}</text>
                <text v-if="member.isVirtual" class="virtual-tag">虚拟账号</text>
              </view>
            </view>
            <view class="member-points">
              <text class="member-score">{{ member.currentPoints || 0 }}</text>
              <text class="member-score-label">当前积分</text>
            </view>
          </view>
        </view>
      </view>

      <view class="section">
        <view class="section-title">积分流水</view>
        <view v-if="pointLogs.length === 0" class="state-block compact">
          <text>暂无积分流水</text>
        </view>
        <view v-else class="log-list">
          <view class="log-card" v-for="log in pointLogs" :key="log.id">
            <view class="log-main">
              <view class="log-title-row">
                <text class="log-title">{{ log.sourceTitle || '未命名来源' }}</text>
                <text class="source-tag">{{ sourceName(log.sourceType) }}</text>
              </view>
              <text class="log-meta">{{ log.nickname || '成员' }} · {{ formatTime(log.createdAt) }}</text>
            </view>
            <text class="log-points" :class="{ negative: log.points < 0 }">{{ pointText(log.points) }}</text>
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
import { listMembers } from '../../api/member.js'
import { listPointLogs } from '../../api/point.js'
import { getCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const currentFamily = ref(null)
const summary = ref({})
const members = ref([])
const pointLogs = ref([])

const familyName = computed(() => summary.value.familyName || currentFamily.value?.familyName || '未选择家庭')
const nickname = computed(() => summary.value.nickname || '我的')
const roleType = computed(() => summary.value.roleType || currentFamily.value?.roleType || '')
const currentPoints = computed(() => summary.value.currentPoints || 0)
const totalEarnedPoints = computed(() => summary.value.totalEarnedPoints || 0)
const avatarText = computed(() => (nickname.value || '我').slice(0, 1))

onShow(() => {
  loadProfile()
})

async function loadProfile(retried = false) {
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
    const [summaryData, memberList, logs] = await Promise.all([
      getDashboardSummary(familyId),
      listMembers(familyId),
      listPointLogs({ familyId })
    ])
    summary.value = summaryData || {}
    members.value = memberList || []
    pointLogs.value = logs || []
  } catch (error) {
    if (error.statusCode !== 401 || retried) {
      throw error
    }
    await ensureDemoLogin(true)
    await loadProfile(true)
  } finally {
    loading.value = false
  }
}

function switchFamily() {
  uni.reLaunch({
    url: '/pages/family-select/index'
  })
}

function roleName(role) {
  const map = {
    OWNER: '家主',
    ADMIN: '管理员',
    PARENT: '家长',
    CHILD: '孩子'
  }
  return map[role] || '成员'
}

function sourceName(sourceType) {
  const map = {
    TASK: '任务',
    REWARD: '奖励',
    ADJUST: '调整'
  }
  return map[sourceType] || '积分'
}

function pointText(points) {
  if (points > 0) {
    return `+${points}`
  }
  return `${points || 0}`
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
  background-color: #f5f7fa;
  min-height: 100vh;
  padding: 24px 20px;
}

.user-card {
  align-items: center;
  background: #fff;
  border-radius: 12px;
  display: flex;
  gap: 14px;
  margin-bottom: 16px;
  padding: 24px 18px;
}

.avatar {
  align-items: center;
  background: #dbeafe;
  border-radius: 50%;
  color: #2563eb;
  display: flex;
  flex-shrink: 0;
  font-size: 22px;
  font-weight: 700;
  height: 58px;
  justify-content: center;
  width: 58px;
}

.info {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.nickname {
  color: #1f2937;
  font-size: 20px;
  font-weight: 700;
}

.current-family {
  color: #64748b;
  font-size: 13px;
}

.switch-btn {
  background: #eff6ff;
  border: none;
  color: #2563eb;
  margin: 0;
}

.points-row {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.point-card {
  background: #fff;
  border-radius: 10px;
  flex: 1;
  padding: 16px;
  text-align: center;
}

.point-value {
  color: #2563eb;
  display: block;
  font-size: 24px;
  font-weight: 700;
}

.point-label,
.member-score-label,
.log-meta {
  color: #64748b;
  font-size: 12px;
}

.content,
.section,
.member-list,
.log-list {
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

.member-list,
.log-list {
  gap: 12px;
}

.member-card,
.log-card,
.state-block {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
}

.member-card,
.log-card {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.member-main,
.log-main {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.member-name,
.log-title {
  color: #111827;
  font-size: 15px;
  font-weight: 700;
}

.tag-row,
.log-title-row {
  align-items: center;
  display: flex;
  gap: 6px;
}

.role-tag,
.virtual-tag,
.source-tag {
  border-radius: 4px;
  font-size: 11px;
  padding: 2px 6px;
}

.role-tag,
.source-tag {
  background: #eff6ff;
  color: #2563eb;
}

.virtual-tag {
  background: #fef3c7;
  color: #b45309;
}

.member-points {
  align-items: flex-end;
  display: flex;
  flex-direction: column;
  margin-left: 12px;
}

.member-score {
  color: #f59e0b;
  font-size: 18px;
  font-weight: 700;
}

.log-points {
  color: #10b981;
  font-size: 18px;
  font-weight: 700;
  margin-left: 12px;
}

.log-points.negative {
  color: #ef4444;
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

- [ ] **Step 2: 检查个人页源码没有明显乱码片段**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\profile\index.vue
```

Expected: exit 1，无匹配。

- [ ] **Step 3: 提交**

```bash
git add parent-child-miniprogram/pages/profile/index.vue
git commit -m "feat: 接入个人页真实数据"
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

- [ ] **Step 2: 检查个人页源码没有明显乱码**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\profile\index.vue
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

- Spec coverage: 覆盖成员 API client、积分 API client、个人页当前身份、成员列表、积分流水、切换家庭、loading、空状态、401 重试和验证命令。
- Placeholder scan: 计划中没有未决占位词或含糊步骤。
- Type consistency: `listMembers`、`listPointLogs`、`loadProfile`、`roleName`、`sourceName`、`pointText` 和页面模板中的调用一致。
- Scope control: 不新增成员管理写操作、积分调整、筛选、分页、导出和新路由。
