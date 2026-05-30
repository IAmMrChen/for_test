<template>
  <view class="sun-page home-page">
    <view v-if="loading" class="card state-card">
      <text>正在加载首页...</text>
    </view>

    <view v-else-if="loadError" class="card state-card">
      <text>{{ loadError }}</text>
      <button class="btn light retry-btn" size="mini" @click="loadHome">重新加载</button>
    </view>

    <view v-else-if="isChild" class="home-content">
      <view class="page-head">
        <view>
          <text class="eyebrow">{{ familyName }}</text>
          <text class="title">{{ childHomeTitle }}</text>
          <text class="subtitle">完成任务后可以兑换奖励</text>
        </view>
      </view>

      <view class="card hero-card current-points-card">
        <text class="section-name">当前积分</text>
        <text class="metric-main">{{ summary.currentPoints || 0 }}</text>
        <text class="metric-sub">已领取 {{ summary.myClaimedTaskCount || 0 }} 项，审核中 {{ summary.myPendingTaskCount || 0 }} 项。</text>
      </view>

      <view class="card task-card">
        <view class="card-title">
          <text>可执行任务</text>
          <button class="btn light" size="mini" @click="goTaskPage">任务页</button>
        </view>
        <view v-if="taskRows.length === 0" class="empty-row">
          <text>暂无可执行任务</text>
        </view>
        <view v-else class="list">
          <view class="row task-row" v-for="task in taskRows" :key="task.id">
            <view class="row-main">
              <text class="row-title">{{ task.title }}</text>
              <text class="row-meta">+{{ task.points }} 积分 · {{ cycleText(task.cycleType) }} · {{ taskStatusText(task) }}</text>
            </view>
            <view class="row-actions task-action-stack">
              <button
                class="btn blue action-btn"
                :class="task.viewStatus"
                :disabled="task.viewStatus === 'pending' || task.viewStatus === 'completed'"
                size="mini"
                @click="handleTaskAction(task)"
              >
                {{ taskButtonText(task) }}
              </button>
              <button
                v-if="task.activeClaim && task.cycleType !== 'ONCE'"
                class="btn light stop-claim-btn"
                size="mini"
                @click="stopRecurringTaskClaim(task)"
              >
                停止领取
              </button>
            </view>
          </view>
        </view>
      </view>

      <view class="card blue-card reward-confirm-card">
        <view class="card-title">
          <text>待确认奖励</text>
          <text>{{ childDeliveredRewardRecords.length }} 项</text>
        </view>
        <view v-if="childDeliveredRewardRecords.length === 0" class="empty-row">
          <text>当前没有待确认奖励</text>
        </view>
        <view v-else class="list">
          <view class="row reward-row" v-for="record in childDeliveredRewardRecords" :key="record.id">
            <view class="row-main">
              <text class="row-title">{{ record.rewardName }}</text>
              <text class="row-meta">已发放 · 消耗 {{ record.pointsCost }} 积分 · {{ formatTime(record.operateTime || record.applyTime) }}</text>
            </view>
            <button class="btn green" size="mini" @click="receiveDeliveredReward(record)">确认</button>
          </view>
        </view>
      </view>
    </view>

    <view v-else class="home-content">
      <view class="page-head">
        <view>
          <text class="eyebrow">{{ familyName }}</text>
          <text class="title">今日看板</text>
          <text class="subtitle">先处理孩子正在等你的事</text>
        </view>
      </view>

      <view class="head-actions">
        <button class="btn light" size="mini" @click="switchFamily">切换家庭</button>
      </view>

      <view class="stats">
        <view class="stat" @click="scrollToSection('#audit-section')">
          <text class="stat-num orange-text">{{ summary.familyPendingTaskCount || 0 }}</text>
          <text class="stat-label">待审核</text>
        </view>
        <view class="stat" @click="scrollToSection('#reward-section')">
          <text class="stat-num green-text">{{ summary.familyAppliedRewardCount || 0 }}</text>
          <text class="stat-label">待奖励</text>
        </view>
        <view class="stat" @click="goTaskPage">
          <text class="stat-num blue-text">{{ summary.activeTaskCount || 0 }}</text>
          <text class="stat-label">可用任务</text>
        </view>
      </view>

      <view id="audit-section" class="card audit-card">
        <view class="card-title">
          <text>待审核任务</text>
          <text>{{ parentPendingRecords.length }} 项</text>
        </view>
        <view v-if="parentPendingRecords.length === 0" class="empty-row">
          <text>当前没有待审核任务</text>
        </view>
        <view v-else class="list">
          <view class="row audit-row" v-for="record in parentPendingRecords" :key="record.id">
            <view class="row-main">
              <text class="row-title">{{ record.nickname || '孩子' }}提交{{ record.taskTitle }}</text>
              <text class="row-meta">+{{ record.points || 0 }} 积分 · 今天 {{ formatTime(record.submitTime) }}</text>
              <text v-if="record.submitRemark" class="row-remark">{{ record.submitRemark }}</text>
            </view>
            <view class="row-actions">
              <button class="btn green" size="mini" @click="audit(record, true)">通过</button>
              <button class="btn danger small" size="mini" @click="audit(record, false)">驳回</button>
            </view>
          </view>
        </view>
      </view>

      <view id="reward-section" class="card blue-card reward-card">
        <view class="card-title">
          <text>待发放奖励</text>
          <text>{{ parentAppliedRewardRecords.length }} 项</text>
        </view>
        <view v-if="parentAppliedRewardRecords.length === 0" class="empty-row">
          <text>当前没有待发放奖励</text>
        </view>
        <view v-else class="list">
          <view class="row reward-row" v-for="record in parentAppliedRewardRecords" :key="record.id">
            <view class="row-main">
              <text class="row-title">{{ record.nickname || '孩子' }}申请{{ record.rewardName }}</text>
              <text class="row-meta">{{ record.pointsCost }} 积分 · 等待发放 · {{ formatTime(record.applyTime) }}</text>
            </view>
            <view class="row-actions">
              <button class="btn primary" size="mini" @click="operateReward(record, true)">发放</button>
              <button class="btn danger small" size="mini" @click="operateReward(record, false)">拒绝</button>
            </view>
          </view>
        </view>
      </view>

      <view v-if="isParentRole" class="card green-card proxy-panel">
        <view class="card-title">
          <text>代孩子完成任务</text>
        </view>
        <text class="card-copy">支持真实孩子和虚拟孩子</text>

        <view v-if="children.length === 0" class="empty-row">
          <text>暂无孩子，请先在我的页面创建或邀请孩子</text>
        </view>
        <view v-else class="proxy-content">
          <view class="chips child-options">
            <button
              v-for="child in children"
              :key="child.id"
              class="chip child-option-btn"
              :class="{ active: selectedChildId === child.id }"
              size="mini"
              @click="selectChild(child.id)"
            >
              {{ child.nickname }}
            </button>
          </view>

          <view v-if="proxyVisibleTaskRows.length === 0" class="empty-row">
            <text>当前孩子暂无可处理任务</text>
          </view>
          <view v-else class="list proxy-task-list">
            <view class="row task-row" v-for="task in proxyVisibleTaskRows" :key="task.id">
              <view class="row-main">
                <text class="row-title">{{ task.title }}</text>
                <text class="row-meta">+{{ task.points }} 积分 · {{ taskStatusText(task) }}</text>
              </view>
              <button
                class="btn green action-btn"
                :class="task.viewStatus"
                :disabled="task.viewStatus === 'completed' || submittingProxyTaskId === task.id"
                size="mini"
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
import { auditTask, claimTask, listTaskClaims, listTaskRecords, listTasks, startTaskClaim, stopTaskClaim, submitTask } from '../../api/task.js'
import { getCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const loadError = ref('')
const currentFamily = ref(null)
const summary = ref({})
const tasks = ref([])
const claimedRecords = ref([])
const pendingRecords = ref([])
const approvedRecords = ref([])
const taskClaims = ref([])
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
const childHomeTitle = computed(() => {
  const nickname = summary.value.nickname || currentFamily.value?.nickname || '小朋友'
  return `${nickname}，今天也很棒`
})

const taskRows = computed(() => {
  return tasks.value
    .map((task) => taskRowForRecords(task, claimedRecords.value, pendingRecords.value, approvedRecords.value, taskClaims.value))
    .filter(Boolean)
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
  return proxyTaskRows.value.filter((task) => task.viewStatus !== 'completed').slice(0, 5)
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
  loadError.value = ''
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
      const [claimed, pending, approved, claims, deliveredRewards] = await Promise.all([
        listTaskRecords({ familyId, status: 'CLAIMED' }),
        listTaskRecords({ familyId, status: 'PENDING' }),
        listTaskRecords({ familyId, status: 'APPROVED' }),
        listTaskClaims(familyId),
        listRewardRecords({ familyId, status: 'DELIVERED' })
      ])
      claimedRecords.value = claimed || []
      pendingRecords.value = pending || []
      approvedRecords.value = approved || []
      taskClaims.value = claims || []
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
      taskClaims.value = []
    }
  } catch (error) {
    if (error.statusCode !== 401 || retried) {
      loadError.value = error.message || '首页加载失败，请稍后重试'
      return
    }
    await ensureDemoLogin(true)
    await loadHome(true)
  } finally {
    loading.value = false
  }
}

function taskRowForRecords(task, claimedRecordsValue, pendingRecordsValue, approvedRecordsValue, taskClaimsValue = []) {
  const claimed = claimedRecordsValue.find((record) => record.taskId === task.id)
  const pending = pendingRecordsValue.find((record) => record.taskId === task.id)
  const approved = approvedRecordsValue.find((record) => record.taskId === task.id && recordMatchesCycle(record, task))
  const activeClaim = taskClaimsValue.find((claim) => claim.taskId === task.id)
  if (approved && task.cycleType === 'ONCE') {
    return null
  }
  if (pending) {
    return { ...task, viewStatus: 'pending', record: pending, activeClaim }
  }
  if (claimed) {
    return { ...task, viewStatus: 'claimed', record: claimed, activeClaim }
  }
  if (approved) {
    return { ...task, viewStatus: 'completed', record: approved, activeClaim }
  }
  if (activeClaim && task.cycleType !== 'ONCE') {
    return { ...task, viewStatus: 'recurringActive', record: null, activeClaim }
  }
  return { ...task, viewStatus: 'claimable', record: null, activeClaim: null }
}

async function handleTaskAction(task) {
  const familyId = currentFamily.value.familyId
  if (task.viewStatus === 'claimable') {
    if (task.cycleType !== 'ONCE') {
      const claim = await startTaskClaim({ familyId, taskId: task.id })
      applyTaskClaim(claim)
      uni.showToast({ title: '已领取任务', icon: 'success' })
      return
    }
    const record = await claimTask({ familyId, taskId: task.id })
    applyClaimedTaskRecord(record)
    uni.showToast({ title: '已领取任务', icon: 'success' })
    return
  }
  if (task.viewStatus === 'recurringActive') {
    const record = await submitTask({ familyId, taskId: task.id })
    applySubmittedTaskRecord(record)
    uni.showToast({ title: '已提交审核', icon: 'success' })
    return
  }
  if (task.viewStatus === 'claimed') {
    const record = await submitTask({ familyId, recordId: task.record.id })
    applySubmittedTaskRecord(record)
    uni.showToast({ title: '已提交审核', icon: 'success' })
  }
}

async function stopRecurringTaskClaim(task) {
  const confirmed = await new Promise((resolve) => {
    uni.showModal({
      title: '停止领取',
      content: `停止后将不再把“${task.title}”作为已领取任务展示`,
      confirmText: '停止',
      cancelText: '取消',
      success: (res) => resolve(res.confirm)
    })
  })
  if (!confirmed) {
    return
  }

  await stopTaskClaim({
    familyId: currentFamily.value.familyId,
    taskId: task.id
  })
  removeTaskClaim(task.id)
  uni.showToast({ title: '已停止领取', icon: 'success' })
}

async function audit(record, approved) {
  const audited = await auditTask({
    recordId: record.id,
    approved
  })
  if (approved) {
    applyApprovedTaskRecord(audited, record)
  } else {
    removePendingTaskRecord(record.id)
  }
  uni.showToast({
    title: approved ? '已通过' : '已驳回',
    icon: 'success'
  })
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
  removeRewardRecord(record.id)
}

async function receiveDeliveredReward(record) {
  await receiveReward({ recordId: record.id })
  childDeliveredRewardRecords.value = childDeliveredRewardRecords.value.filter((item) => item.id !== record.id)
  uni.showToast({ title: '已确认收到', icon: 'success' })
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
    await directApproveProxyTask(task, child)
    uni.showToast({ title: '已通过并发放积分', icon: 'success' })
  } finally {
    submittingProxyTaskId.value = 0
  }
}

async function directApproveProxyTask(task, child) {
  const familyId = currentFamily.value.familyId
  let pendingRecord = task.record
  if (task.viewStatus === 'claimable') {
    pendingRecord = await submitTask({ familyId, taskId: task.id, memberId: child.id })
  } else if (task.viewStatus === 'recurringActive') {
    pendingRecord = await submitTask({ familyId, taskId: task.id, memberId: child.id })
  } else if (task.viewStatus === 'claimed') {
    pendingRecord = await submitTask({ familyId, recordId: task.record.id, memberId: child.id })
  }
  const approvedRecord = await auditTask({
    recordId: pendingRecord.id,
    approved: true
  })
  applyApprovedTaskRecord(approvedRecord, {
    ...pendingRecord,
    taskId: task.id,
    taskTitle: task.title,
    points: task.points,
    memberId: child.id,
    nickname: child.nickname
  })
}

function applyClaimedTaskRecord(record) {
  claimedRecords.value = [
    record,
    ...claimedRecords.value.filter((item) => item.id !== record.id)
  ]
  summary.value = {
    ...summary.value,
    myClaimedTaskCount: (summary.value.myClaimedTaskCount || 0) + 1
  }
}

function applyTaskClaim(claim) {
  taskClaims.value = [
    claim,
    ...taskClaims.value.filter((item) => item.id !== claim.id && item.taskId !== claim.taskId)
  ]
}

function removeTaskClaim(taskId) {
  taskClaims.value = taskClaims.value.filter((item) => item.taskId !== taskId)
  summary.value = {
    ...summary.value,
    myClaimedTaskCount: Math.max((summary.value.myClaimedTaskCount || 0) - 1, 0)
  }
}

function applySubmittedTaskRecord(record) {
  claimedRecords.value = claimedRecords.value.filter((item) => item.id !== record.id)
  pendingRecords.value = [
    record,
    ...pendingRecords.value.filter((item) => item.id !== record.id)
  ]
  summary.value = {
    ...summary.value,
    myClaimedTaskCount: Math.max((summary.value.myClaimedTaskCount || 0) - 1, 0),
    myPendingTaskCount: (summary.value.myPendingTaskCount || 0) + 1
  }
}

function applyProxyClaimedTaskRecord(record) {
  proxyClaimedRecords.value = [
    record,
    ...proxyClaimedRecords.value.filter((item) => item.id !== record.id)
  ]
}

function applyProxySubmittedTaskRecord(record, task, child) {
  proxyClaimedRecords.value = proxyClaimedRecords.value.filter((item) => item.id !== record.id)
  const pendingRecord = {
    ...record,
    taskTitle: task.title,
    points: task.points,
    nickname: child.nickname
  }
  parentPendingRecords.value = [
    pendingRecord,
    ...parentPendingRecords.value.filter((item) => item.id !== record.id)
  ]
  summary.value = {
    ...summary.value,
    familyPendingTaskCount: (summary.value.familyPendingTaskCount || 0) + 1
  }
}

function applyApprovedTaskRecord(record, sourceRecord) {
  removePendingTaskRecord(sourceRecord.id)
  proxyClaimedRecords.value = proxyClaimedRecords.value.filter((item) => item.id !== sourceRecord.id)
  proxyApprovedRecords.value = [
    {
      ...sourceRecord,
      ...record,
      status: 'APPROVED'
    },
    ...proxyApprovedRecords.value.filter((item) => item.id !== sourceRecord.id)
  ]
  members.value = members.value.map((member) => {
    if (member.id !== sourceRecord.memberId) {
      return member
    }
    return {
      ...member,
      currentPoints: (member.currentPoints || 0) + (sourceRecord.points || 0),
      totalEarnedPoints: (member.totalEarnedPoints || 0) + (sourceRecord.points || 0)
    }
  })
}

function removePendingTaskRecord(recordId) {
  const before = parentPendingRecords.value.length
  parentPendingRecords.value = parentPendingRecords.value.filter((item) => item.id !== recordId)
  const removed = before - parentPendingRecords.value.length
  if (removed > 0) {
    summary.value = {
      ...summary.value,
      familyPendingTaskCount: Math.max((summary.value.familyPendingTaskCount || 0) - removed, 0)
    }
  }
}

function removeRewardRecord(recordId) {
  parentAppliedRewardRecords.value = parentAppliedRewardRecords.value.filter((item) => item.id !== recordId)
  summary.value = {
    ...summary.value,
    familyAppliedRewardCount: Math.max((summary.value.familyAppliedRewardCount || 0) - 1, 0)
  }
}

function proxyTaskButtonText(task) {
  if (submittingProxyTaskId.value === task.id) {
    return '处理中'
  }
  if (task.viewStatus === 'completed') {
    return '已完成'
  }
  return '通过'
}

function taskButtonText(task) {
  const map = {
    claimable: '领取',
    claimed: '提交',
    recurringActive: '提交',
    pending: '审核中',
    completed: '已完成'
  }
  return map[task.viewStatus] || '查看'
}

function taskStatusText(task) {
  const map = {
    claimable: '可领取',
    claimed: '已领取',
    recurringActive: '已领取',
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

function switchFamily() {
  uni.reLaunch({ url: '/pages/family-select/index' })
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

function cycleText(cycleType) {
  const map = {
    ONCE: '一次',
    DAILY: '每日',
    WEEKLY: '每周',
    MONTHLY: '每月'
  }
  return map[cycleType] || '任务'
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
.home-page {
  padding-left: 40rpx;
  padding-right: 40rpx;
}

.home-content {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.page-head {
  align-items: flex-start;
  display: flex;
  gap: 20rpx;
  justify-content: space-between;
  margin-bottom: 2rpx;
}

.page-head > view {
  max-width: 460rpx;
}

.head-actions {
  display: flex;
  gap: 14rpx;
  justify-content: flex-end;
  margin: -72rpx 0 24rpx;
  min-height: 60rpx;
}

.eyebrow,
.title,
.subtitle,
.card-copy,
.section-name,
.metric-main,
.metric-sub,
.row-title,
.row-meta,
.row-remark,
.stat-num,
.stat-label {
  display: block;
}

.eyebrow {
  color: #7a6c55;
  font-size: 22rpx;
  font-weight: 900;
  margin-bottom: 8rpx;
}

.title {
  color: #172033;
  font-size: 50rpx;
  font-weight: 950;
  letter-spacing: 0;
  line-height: 1.12;
}

.subtitle {
  color: #707887;
  font-size: 24rpx;
  line-height: 1.45;
  margin-top: 14rpx;
}

.card {
  background: rgba(255, 255, 255, 0.9);
  border: 1rpx solid rgba(255, 184, 77, 0.3);
  border-radius: 36rpx;
  box-shadow: 0 24rpx 56rpx rgba(43, 42, 40, 0.08);
  padding: 26rpx;
}

.hero-card {
  background: linear-gradient(135deg, #ffb84d 0%, #ff7a45 66%);
  border-color: rgba(255, 255, 255, 0.36);
  box-shadow: 0 32rpx 64rpx rgba(255, 122, 69, 0.22);
  color: #fff;
}

.blue-card {
  background: linear-gradient(135deg, #eef7ff 0%, #fff 100%);
  border-color: rgba(75, 159, 255, 0.24);
}

.green-card {
  background: linear-gradient(135deg, #effbf3 0%, #fff 100%);
  border-color: rgba(99, 199, 132, 0.26);
}

.card-title {
  align-items: center;
  color: #172033;
  display: flex;
  font-size: 28rpx;
  font-weight: 950;
  gap: 16rpx;
  justify-content: space-between;
  margin-bottom: 18rpx;
}

.card-copy {
  color: #707887;
  font-size: 24rpx;
  line-height: 1.45;
  margin: -4rpx 0 18rpx;
}

.section-name {
  color: #fff;
  font-size: 24rpx;
  font-weight: 900;
  margin-bottom: 12rpx;
}

.metric-main {
  color: #fff;
  font-size: 62rpx;
  font-weight: 950;
  line-height: 1;
}

.metric-sub {
  color: rgba(255, 255, 255, 0.92);
  font-size: 24rpx;
  line-height: 1.35;
  margin-top: 14rpx;
}

.stats {
  display: grid;
  gap: 16rpx;
  grid-template-columns: repeat(3, 1fr);
}

.stat {
  align-items: center;
  background: rgba(255, 255, 255, 0.9);
  border: 1rpx solid #edf0f5;
  border-radius: 30rpx;
  box-shadow: 0 20rpx 44rpx rgba(43, 42, 40, 0.06);
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-height: 136rpx;
  padding: 10rpx 12rpx 12rpx;
  text-align: center;
}

.stat-num {
  font-size: 44rpx;
  font-weight: 950;
  line-height: 1;
}

.stat-label {
  color: #7b8190;
  font-size: 22rpx;
  font-weight: 800;
  line-height: 1.2;
  margin-top: 12rpx;
}

.orange-text {
  color: #f97316;
}

.green-text {
  color: #36b86b;
}

.blue-text {
  color: #2f80ed;
}

.list,
.proxy-content {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.row {
  align-items: center;
  background: rgba(255, 255, 255, 0.84);
  border: 1rpx solid #edf0f5;
  border-radius: 28rpx;
  display: flex;
  gap: 16rpx;
  justify-content: space-between;
  min-height: 108rpx;
  padding: 20rpx 22rpx;
}

.row-main {
  flex: 1 1 auto;
  min-width: 0;
}

.row-title {
  color: #172033;
  font-size: 26rpx;
  font-weight: 900;
  line-height: 1.25;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.row-meta,
.row-remark {
  color: #7b8190;
  font-size: 22rpx;
  line-height: 1.35;
  margin-top: 8rpx;
}

.row-remark {
  background: #f6f8fb;
  border-radius: 14rpx;
  padding: 10rpx 12rpx;
}

.row-actions {
  align-items: flex-end;
  display: flex;
  flex: 0 0 auto;
  gap: 10rpx;
}

.task-action-stack {
  flex-direction: column;
}

.empty-row,
.state-card {
  align-items: center;
  background: rgba(255, 255, 255, 0.7);
  border: 1rpx solid #edf0f5;
  border-radius: 28rpx;
  color: #7b8190;
  display: flex;
  flex-direction: column;
  font-size: 26rpx;
  justify-content: center;
  min-height: 104rpx;
  text-align: center;
}

.state-card {
  gap: 18rpx;
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 14rpx;
  margin-bottom: 18rpx;
}

.home-page button,
.home-page .btn {
  align-items: center;
  border: 0;
  border-radius: 999rpx;
  box-shadow: 0 16rpx 32rpx rgba(255, 122, 69, 0.2);
  color: #fff;
  display: inline-flex;
  flex: 0 0 auto;
  font-size: 24rpx;
  font-weight: 900;
  height: 60rpx;
  justify-content: center;
  line-height: 60rpx;
  margin: 0;
  min-height: 60rpx;
  min-width: 112rpx;
  padding: 0 24rpx;
  white-space: nowrap;
}

.home-page .btn.primary {
  background: #ff7a45;
}

.home-page .btn.blue {
  background: #4b9fff;
  box-shadow: 0 16rpx 32rpx rgba(75, 159, 255, 0.18);
}

.home-page .btn.green {
  background: #63c784;
  box-shadow: 0 16rpx 32rpx rgba(99, 199, 132, 0.18);
}

.home-page .btn.light {
  background: #eef7ff;
  border: 1rpx solid rgba(75, 159, 255, 0.18);
  box-shadow: none;
  color: #2f80ed;
}

.home-page .btn.danger {
  background: #fff1f3;
  border: 1rpx solid rgba(239, 107, 122, 0.2);
  box-shadow: none;
  color: #d95061;
}

.home-page .btn.small {
  font-size: 22rpx;
  height: 56rpx;
  line-height: 56rpx;
  min-height: 56rpx;
  min-width: 88rpx;
  padding: 0 20rpx;
}

.home-page .action-btn.claimed,
.home-page .action-btn.recurringActive {
  background: #4b9fff;
}

.home-page .action-btn.pending,
.home-page .action-btn.completed {
  background: #eef2f7;
  box-shadow: none;
  color: #8a95a5;
}

.home-page .chip,
.home-page .child-option-btn {
  background: rgba(255, 255, 255, 0.9);
  border: 1rpx solid #edf0f5;
  box-shadow: none;
  color: #697180;
  font-size: 22rpx;
  height: 62rpx;
  line-height: 62rpx;
  min-height: 62rpx;
  min-width: 88rpx;
  padding: 0 22rpx;
}

.home-page .chip.active,
.home-page .child-option-btn.active {
  background: #4b9fff;
  border-color: #4b9fff;
  box-shadow: 0 16rpx 36rpx rgba(75, 159, 255, 0.2);
  color: #fff;
}

.retry-btn {
  margin-top: 4rpx;
}
</style>
