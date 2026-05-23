<template>
  <view class="container">
    <view class="navbar">
      <view class="current-family">
        <text class="greeting">{{ greetingText }}</text>
        <text class="family-name">{{ familyName }}</text>
      </view>
      <view v-if="isChild" class="points-card">
        <text class="points-label">当前积分</text>
        <view class="points-value">{{ summary.currentPoints || 0 }} <text class="unit">分</text></view>
      </view>
    </view>

    <view v-if="loading" class="state-block">
      <text>正在加载首页...</text>
    </view>

    <view v-else-if="isChild" class="content-area">
      <view class="summary-row">
        <view class="summary-item">
          <text class="summary-num">{{ summary.myClaimedTaskCount || 0 }}</text>
          <text class="summary-label">已领取</text>
        </view>
        <view class="summary-item">
          <text class="summary-num">{{ summary.myPendingTaskCount || 0 }}</text>
          <text class="summary-label">审核中</text>
        </view>
      </view>

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

      <view class="section-title">可执行任务</view>
      <view v-if="taskRows.length === 0" class="state-block">
        <text>暂无可执行任务</text>
      </view>
      <view v-else class="task-list">
        <view class="task-item" v-for="task in taskRows" :key="task.id">
          <view class="task-info">
            <view class="task-title">{{ task.title }}</view>
            <view class="task-reward">+{{ task.points }} 积分</view>
          </view>
          <button
            class="action-btn"
            :class="task.viewStatus"
            :disabled="task.viewStatus === 'pending' || task.viewStatus === 'completed'"
            @click="handleTaskAction(task)"
          >
            {{ taskButtonText(task) }}
          </button>
        </view>
      </view>
    </view>

    <view v-else class="content-area">
      <view class="dashboard-stats">
        <view class="stat-item" @click="scrollToSection('#audit-section')">
          <view class="stat-num text-orange">{{ summary.familyPendingTaskCount || 0 }}</view>
          <view class="stat-desc">待审核</view>
        </view>
        <view class="stat-item" @click="scrollToSection('#reward-section')">
          <view class="stat-num text-green">{{ summary.familyAppliedRewardCount || 0 }}</view>
          <view class="stat-desc">待奖励</view>
        </view>
        <view class="stat-item" @click="goTaskPage">
          <view class="stat-num text-blue">{{ summary.activeTaskCount || 0 }}</view>
          <view class="stat-desc">可用任务</view>
        </view>
      </view>

      <view id="audit-section" class="section-title">待审核任务</view>
      <view v-if="parentPendingRecords.length === 0" class="state-block">
        <text>当前没有待审核任务</text>
      </view>
      <view v-else class="audit-list">
        <view class="audit-item" v-for="record in parentPendingRecords" :key="record.id">
          <view class="audit-header">
            <text class="child-name">{{ record.nickname || '孩子' }}</text>
            <text class="time">{{ formatTime(record.submitTime) }}</text>
          </view>
          <view class="audit-content">
            <text>提交了任务：</text>
            <text class="strong">{{ record.taskTitle }}</text>
          </view>
          <view v-if="record.submitRemark" class="audit-remark">{{ record.submitRemark }}</view>
          <view class="audit-actions">
            <button class="btn-reject" size="mini" plain @click="audit(record, false)">驳回</button>
            <button class="btn-approve" size="mini" type="primary" @click="audit(record, true)">通过</button>
          </view>
        </view>
      </view>

      <view id="reward-section" class="section-title reward-section-title">待发放奖励</view>
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

      <view v-if="isParentRole" class="proxy-panel">
        <view class="proxy-header">
          <view>
            <text class="proxy-title">代孩子完成任务</text>
            <text class="proxy-subtitle">选择孩子后处理常用任务</text>
          </view>
          <button class="plain-action-btn" size="mini" @click="goTaskPage">查看全部任务</button>
        </view>

        <view v-if="children.length === 0" class="state-block compact">
          <text>暂无孩子，请先在我的页面创建或邀请孩子</text>
        </view>
        <view v-else class="proxy-content">
          <view class="child-options">
            <button
              v-for="child in children"
              :key="child.id"
              class="child-option-btn"
              :class="{ active: selectedChildId === child.id }"
              size="mini"
              @click="selectChild(child.id)"
            >
              {{ child.nickname }}
            </button>
          </view>

          <view v-if="proxyVisibleTaskRows.length === 0" class="state-block compact">
            <text>当前孩子暂无可处理任务</text>
          </view>
          <view v-else class="proxy-task-list">
            <view class="task-item" v-for="task in proxyVisibleTaskRows" :key="task.id">
              <view class="task-info">
                <view class="task-title">{{ task.title }}</view>
                <view class="task-reward">+{{ task.points }} 积分 · {{ taskStatusText(task) }}</view>
              </view>
              <button
                class="action-btn"
                :class="task.viewStatus"
                :disabled="task.viewStatus === 'pending' || task.viewStatus === 'completed' || submittingProxyTaskId === task.id"
                @click="handleProxyTaskAction(task)"
              >
                {{ proxyTaskButtonText(task) }}
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
import { listMembers } from '../../api/member.js'
import { deliverReward, listRewardRecords, receiveReward, rejectReward } from '../../api/reward.js'
import { auditTask, claimTask, listTaskRecords, listTasks, submitTask } from '../../api/task.js'
import { getCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const currentFamily = ref(null)
const summary = ref({})
const tasks = ref([])
const claimedRecords = ref([])
const pendingRecords = ref([])
const approvedRecords = ref([])
const parentPendingRecords = ref([])
const parentAppliedRewardRecords = ref([])
const childDeliveredRewardRecords = ref([])
const members = ref([])
const proxyClaimedRecords = ref([])
const proxyApprovedRecords = ref([])
const selectedChildId = ref(0)
const submittingProxyTaskId = ref(0)

const roleType = computed(() => summary.value.roleType || currentFamily.value?.roleType || '')
const isChild = computed(() => roleType.value === 'CHILD')
const isParentRole = computed(() => ['OWNER', 'ADMIN', 'PARENT'].includes(roleType.value))
const familyName = computed(() => summary.value.familyName || currentFamily.value?.familyName || '当前家庭')
const greetingText = computed(() => {
  const nickname = summary.value.nickname || currentFamily.value?.nickname || roleName(roleType.value)
  return `Hi，${nickname}`
})

const taskRows = computed(() => {
  return tasks.value.map((task) => taskRowForRecords(task, claimedRecords.value, pendingRecords.value, approvedRecords.value))
})

const children = computed(() => {
  return members.value.filter((member) => member.roleType === 'CHILD')
})

const selectedChild = computed(() => {
  return children.value.find((member) => member.id === selectedChildId.value) || null
})

const proxyTaskRows = computed(() => {
  const child = selectedChild.value
  if (!child) {
    return []
  }

  const claimed = proxyClaimedRecords.value.filter((record) => record.memberId === child.id)
  const pending = parentPendingRecords.value.filter((record) => record.memberId === child.id)
  const approved = proxyApprovedRecords.value.filter((record) => record.memberId === child.id)
  return tasks.value.map((task) => taskRowForRecords(task, claimed, pending, approved))
})

const proxyVisibleTaskRows = computed(() => {
  return proxyTaskRows.value.slice(0, 5)
})

onShow(() => {
  loadHome()
})

async function loadHome(retried = false) {
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
    const [summaryData, taskList] = await Promise.all([
      getDashboardSummary(familyId),
      listTasks(familyId)
    ])
    summary.value = summaryData || {}
    tasks.value = taskList || []

    if (summary.value.roleType === 'CHILD') {
      const [claimed, pending, approved, deliveredRewards] = await Promise.all([
        listTaskRecords({ familyId, status: 'CLAIMED' }),
        listTaskRecords({ familyId, status: 'PENDING' }),
        listTaskRecords({ familyId, status: 'APPROVED' }),
        listRewardRecords({ familyId, status: 'DELIVERED' })
      ])
      claimedRecords.value = claimed || []
      pendingRecords.value = pending || []
      approvedRecords.value = approved || []
      childDeliveredRewardRecords.value = deliveredRewards || []
      parentPendingRecords.value = []
      parentAppliedRewardRecords.value = []
      members.value = []
      proxyClaimedRecords.value = []
      proxyApprovedRecords.value = []
      selectedChildId.value = 0
    } else if (isParentRole.value) {
      const [memberList, claimed, pending, approved, appliedRewards] = await Promise.all([
        listMembers(familyId),
        listTaskRecords({ familyId, status: 'CLAIMED' }),
        listTaskRecords({ familyId, status: 'PENDING' }),
        listTaskRecords({ familyId, status: 'APPROVED' }),
        listRewardRecords({ familyId, status: 'APPLIED' })
      ])
      members.value = memberList || []
      proxyClaimedRecords.value = claimed || []
      parentPendingRecords.value = pending || []
      proxyApprovedRecords.value = approved || []
      parentAppliedRewardRecords.value = appliedRewards || []
      childDeliveredRewardRecords.value = []
      syncSelectedChild()
      claimedRecords.value = []
      pendingRecords.value = []
      approvedRecords.value = []
    }
  } catch (error) {
    if (error.statusCode !== 401 || retried) {
      throw error
    }
    await ensureDemoLogin(true)
    await loadHome(true)
  } finally {
    loading.value = false
  }
}

function taskRowForRecords(task, claimedRecordsValue, pendingRecordsValue, approvedRecordsValue) {
  const claimed = claimedRecordsValue.find((record) => record.taskId === task.id)
  const pending = pendingRecordsValue.find((record) => record.taskId === task.id)
  const approved = approvedRecordsValue.find((record) => record.taskId === task.id && recordMatchesCycle(record, task))
  if (pending) {
    return { ...task, viewStatus: 'pending', record: pending }
  }
  if (claimed) {
    return { ...task, viewStatus: 'claimed', record: claimed }
  }
  if (approved) {
    return { ...task, viewStatus: 'completed', record: approved }
  }
  return { ...task, viewStatus: 'claimable', record: null }
}

async function handleTaskAction(task) {
  const familyId = currentFamily.value.familyId
  if (task.viewStatus === 'claimable') {
    await claimTask({ familyId, taskId: task.id })
    uni.showToast({ title: '已领取任务', icon: 'success' })
    await loadHome()
    return
  }
  if (task.viewStatus === 'claimed') {
    await submitTask({ familyId, recordId: task.record.id })
    uni.showToast({ title: '已提交审核', icon: 'success' })
    await loadHome()
  }
}

async function audit(record, approved) {
  await auditTask({
    recordId: record.id,
    approved
  })
  uni.showToast({
    title: approved ? '已通过' : '已驳回',
    icon: 'success'
  })
  await loadHome()
}

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

function syncSelectedChild() {
  if (children.value.some((child) => child.id === selectedChildId.value)) {
    return
  }
  selectedChildId.value = children.value[0]?.id || 0
}

function selectChild(memberId) {
  selectedChildId.value = memberId
}

async function handleProxyTaskAction(task) {
  const child = selectedChild.value
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

function taskButtonText(task) {
  const map = {
    claimable: '领取',
    claimed: '提交',
    pending: '审核中',
    completed: '已完成'
  }
  return map[task.viewStatus] || '查看'
}

function taskStatusText(task) {
  const map = {
    claimable: '可领取',
    claimed: '已领取',
    pending: '审核中',
    completed: '已完成'
  }
  return map[task.viewStatus] || '可处理'
}

function recordMatchesCycle(record, task) {
  if (task.cycleType === 'ONCE') {
    return true
  }
  const value = record.auditTime || record.submitTime
  if (!value) {
    return false
  }
  const date = new Date(value)
  const now = new Date()
  if (task.cycleType === 'DAILY') {
    return date.toDateString() === now.toDateString()
  }
  if (task.cycleType === 'WEEKLY') {
    return weekKey(date) === weekKey(now)
  }
  return false
}

function weekKey(date) {
  const copy = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  const day = copy.getDay() || 7
  copy.setDate(copy.getDate() + 4 - day)
  const yearStart = new Date(copy.getFullYear(), 0, 1)
  const week = Math.ceil((((copy - yearStart) / 86400000) + 1) / 7)
  return `${copy.getFullYear()}-${week}`
}

function scrollToSection(selector) {
  uni.pageScrollTo({
    selector,
    duration: 240
  })
}

function goTaskPage() {
  uni.switchTab({ url: '/pages/tasks/index' })
}

function roleName(role) {
  const map = {
    OWNER: '家主',
    ADMIN: '管理员',
    PARENT: '家长',
    CHILD: '孩子'
  }
  return map[role] || '用户'
}

function formatTime(value) {
  if (!value) {
    return ''
  }
  const date = new Date(value)
  const hour = `${date.getHours()}`.padStart(2, '0')
  const minute = `${date.getMinutes()}`.padStart(2, '0')
  return `${hour}:${minute}`
}
</script>

<style>
.container {
  background-color: #f5f7fa;
  min-height: 100vh;
}

.navbar {
  align-items: flex-end;
  background: #fff;
  display: flex;
  justify-content: space-between;
  padding: 44px 20px 20px;
}

.current-family {
  display: flex;
  flex-direction: column;
}

.greeting {
  color: #1f2937;
  font-size: 20px;
  font-weight: 700;
}

.family-name {
  color: #94a3b8;
  font-size: 12px;
  margin-top: 4px;
}

.points-card {
  display: flex;
  flex-direction: column;
  text-align: right;
}

.points-label {
  color: #999;
  font-size: 10px;
}

.points-value {
  color: #f59e0b;
  font-size: 20px;
  font-weight: 700;
}

.unit {
  font-size: 12px;
}

.content-area {
  padding: 20px;
}

.section-title {
  color: #1f2937;
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 12px;
}

.section-block {
  margin-bottom: 24px;
}

.reward-section-title {
  margin-top: 24px;
}

.summary-row,
.dashboard-stats {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
}

.summary-item,
.stat-item {
  background: #fff;
  border-radius: 10px;
  flex: 1;
  padding: 16px;
  text-align: center;
}

.summary-num,
.stat-num {
  font-size: 24px;
  font-weight: 700;
}

.summary-num {
  color: #2563eb;
}

.summary-label,
.stat-desc {
  color: #64748b;
  font-size: 12px;
  margin-top: 4px;
}

.text-orange {
  color: #f97316;
}

.text-green {
  color: #10b981;
}

.text-blue {
  color: #2563eb;
}

.task-list,
.audit-list,
.reward-todo-list,
.proxy-task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-item,
.audit-item,
.reward-todo-item,
.state-block,
.proxy-panel {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
}

.task-item,
.reward-todo-item {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.task-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.task-title {
  color: #111827;
  font-size: 16px;
  font-weight: 600;
}

.task-reward {
  color: #f59e0b;
  font-size: 12px;
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

.action-btn {
  background: #3b82f6;
  border: none;
  border-radius: 18px;
  color: #fff;
  font-size: 12px;
  height: 32px;
  line-height: 32px;
  margin: 0;
  padding: 0 16px;
}

.action-btn.claimed {
  background: #10b981;
}

.action-btn.pending,
.action-btn.completed {
  background: #e5e7eb;
  color: #64748b;
}

.audit-header {
  color: #64748b;
  display: flex;
  font-size: 12px;
  justify-content: space-between;
  margin-bottom: 8px;
}

.child-name,
.strong {
  color: #1f2937;
  font-weight: 700;
}

.audit-content {
  color: #374151;
  font-size: 14px;
  margin-bottom: 10px;
}

.audit-remark {
  background: #f8fafc;
  border-radius: 6px;
  color: #64748b;
  font-size: 12px;
  margin-bottom: 12px;
  padding: 8px;
}

.audit-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.btn-reject,
.btn-approve {
  margin: 0;
}

.state-block {
  color: #64748b;
  font-size: 14px;
  text-align: center;
}

.state-block.compact {
  padding: 20px 16px;
}

.proxy-panel {
  margin-top: 24px;
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

.primary-action-btn {
  background: #2563eb;
  border: none;
  color: #fff;
}
</style>
