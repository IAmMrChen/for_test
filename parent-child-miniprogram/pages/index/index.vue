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
            :disabled="task.viewStatus === 'pending'"
            @click="handleTaskAction(task)"
          >
            {{ taskButtonText(task) }}
          </button>
        </view>
      </view>
    </view>

    <view v-else class="content-area">
      <view class="dashboard-stats">
        <view class="stat-item">
          <view class="stat-num text-orange">{{ summary.familyPendingTaskCount || 0 }}</view>
          <view class="stat-desc">待审核</view>
        </view>
        <view class="stat-item">
          <view class="stat-num text-green">{{ summary.familyAppliedRewardCount || 0 }}</view>
          <view class="stat-desc">待发奖</view>
        </view>
        <view class="stat-item">
          <view class="stat-num text-blue">{{ summary.activeTaskCount || 0 }}</view>
          <view class="stat-desc">可用任务</view>
        </view>
      </view>

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

      <view class="section-title">待办事项</view>
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
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'

import { ensureDemoLogin } from '../../api/auth.js'
import { getDashboardSummary } from '../../api/dashboard.js'
import { auditTask, claimTask, createTask, listTaskRecords, listTasks, submitTask } from '../../api/task.js'
import { getCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const currentFamily = ref(null)
const summary = ref({})
const tasks = ref([])
const claimedRecords = ref([])
const pendingRecords = ref([])
const parentPendingRecords = ref([])
const showTaskForm = ref(false)
const submittingTask = ref(false)
const taskForm = ref(defaultTaskForm())

const taskCycles = [
  { label: '一次性', value: 'ONCE' },
  { label: '每日', value: 'DAILY' },
  { label: '每周', value: 'WEEKLY' }
]

const roleType = computed(() => summary.value.roleType || currentFamily.value?.roleType || '')
const isChild = computed(() => roleType.value === 'CHILD')
const isParentRole = computed(() => ['OWNER', 'ADMIN', 'PARENT'].includes(roleType.value))
const familyName = computed(() => summary.value.familyName || currentFamily.value?.familyName || '当前家庭')
const greetingText = computed(() => {
  const nickname = summary.value.nickname || currentFamily.value?.nickname || roleName(roleType.value)
  return `Hi，${nickname}`
})

const taskRows = computed(() => {
  return tasks.value.map((task) => {
    const claimed = claimedRecords.value.find((record) => record.taskId === task.id)
    const pending = pendingRecords.value.find((record) => record.taskId === task.id)
    if (pending) {
      return { ...task, viewStatus: 'pending', record: pending }
    }
    if (claimed) {
      return { ...task, viewStatus: 'claimed', record: claimed }
    }
    return { ...task, viewStatus: 'claimable', record: null }
  })
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
    const familyId = family.familyId
    const [summaryData, taskList] = await Promise.all([
      getDashboardSummary(familyId),
      listTasks(familyId)
    ])
    summary.value = summaryData || {}
    tasks.value = taskList || []

    if (summary.value.roleType === 'CHILD') {
      const [claimed, pending] = await Promise.all([
        listTaskRecords({ familyId, status: 'CLAIMED' }),
        listTaskRecords({ familyId, status: 'PENDING' })
      ])
      claimedRecords.value = claimed || []
      pendingRecords.value = pending || []
      parentPendingRecords.value = []
    } else if (isParentRole.value) {
      parentPendingRecords.value = (await listTaskRecords({ familyId, status: 'PENDING' })) || []
      claimedRecords.value = []
      pendingRecords.value = []
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

function taskButtonText(task) {
  const map = {
    claimable: '领取',
    claimed: '提交',
    pending: '审核中'
  }
  return map[task.viewStatus] || '查看'
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
.audit-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-item,
.audit-item,
.state-block {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
}

.task-item {
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

.action-btn.pending {
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
</style>
