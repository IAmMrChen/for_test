<template>
  <view class="sun-page tasks-page">
    <view class="page-head">
      <view>
        <text class="eyebrow">任务工作台</text>
        <text class="title">任务</text>
        <text class="subtitle">{{ isParentRole ? '发布、管理和持续领取任务' : '领取任务，完成后提交审核' }}</text>
      </view>
    </view>

    <view v-if="isParentRole && !showTaskForm" class="head-actions">
      <button class="btn primary" size="mini" @click="openTaskForm">发布任务</button>
    </view>

    <view v-if="loading" class="card state-card">
      <text>正在加载任务...</text>
    </view>

    <view v-else-if="loadError" class="card state-card">
      <text>{{ loadError }}</text>
      <button class="btn light retry-btn" size="mini" @click="loadTaskPage">重新加载</button>
    </view>

    <view v-else class="tasks-content">
      <view v-if="isParentRole && showTaskForm && !editingTaskId" class="card hero-card create-panel">
        <view class="card-title">
          <text>发布任务</text>
        </view>
        <text class="card-copy">给孩子增加一个可领取的任务</text>

        <view class="create-form">
          <view class="form-row">
            <text class="form-label">任务标题</text>
            <view class="input-shell">
              <input v-model.trim="taskForm.title" class="form-input hero-input" placeholder="例如：阅读 30 分钟" placeholder-class="hero-placeholder" />
            </view>
          </view>
          <view class="form-row">
            <text class="form-label">奖励积分</text>
            <view class="input-shell">
              <input v-model="taskForm.points" class="form-input hero-input" type="number" placeholder="10" placeholder-class="hero-placeholder" />
            </view>
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
            <text class="form-hint">同名任务会提醒，可确认继续</text>
            <view class="form-buttons">
              <button class="btn light" size="mini" :disabled="submittingTask" @click="cancelTaskForm">取消</button>
              <button class="btn light" size="mini" :disabled="submittingTask" @click="publishTask">
                {{ submittingTask ? '保存中' : '确认发布' }}
              </button>
            </view>
          </view>
        </view>
      </view>

      <view v-if="isParentRole" class="card">
        <view class="card-title">
          <text>任务管理</text>
          <text>{{ tasks.length }} 项</text>
        </view>
        <view v-if="tasks.length === 0" class="empty-row">
          <text>暂无可管理任务</text>
        </view>
        <view v-else class="list">
          <view class="task-item" v-for="task in tasks" :key="task.id">
            <view class="row task-row">
              <view class="row-main">
                <text class="row-title">{{ task.title }}</text>
                <text class="row-meta">+{{ task.points }} 积分 · {{ cycleLabel(task.cycleType) }}</text>
              </view>
              <view class="row-actions">
                <button class="btn light" size="mini" :disabled="editingTaskId === task.id" @click="editTask(task)">
                  {{ editingTaskId === task.id ? '编辑中' : '编辑' }}
                </button>
                <button class="btn danger" size="mini" @click="archiveExistingTask(task)">归档</button>
              </view>
            </view>

            <view v-if="editingTaskId === task.id" class="inline-edit-panel">
              <view class="card-title inline-edit-title">
                <text>编辑任务</text>
              </view>
              <view class="create-form inline-edit-form">
                <view class="form-row">
                  <text class="form-label">任务标题</text>
                  <view class="input-shell">
                    <input v-model.trim="taskForm.title" class="form-input hero-input" placeholder="例如：阅读 30 分钟" placeholder-class="hero-placeholder" />
                  </view>
                </view>
                <view class="form-row">
                  <text class="form-label">奖励积分</text>
                  <view class="input-shell">
                    <input v-model="taskForm.points" class="form-input hero-input" type="number" placeholder="10" placeholder-class="hero-placeholder" />
                  </view>
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
                  <text class="form-hint">修改后会影响之后孩子看到的任务信息</text>
                  <view class="form-buttons">
                    <button class="btn light" size="mini" :disabled="submittingTask" @click="cancelTaskForm">取消</button>
                    <button class="btn light" size="mini" :disabled="submittingTask" @click="publishTask">
                      {{ submittingTask ? '保存中' : '确认保存' }}
                    </button>
                  </view>
                </view>
              </view>
            </view>
          </view>
        </view>
      </view>

      <view v-if="isChild" class="card hero-card points-card">
        <text class="section-name">当前积分</text>
        <text class="metric-main">{{ currentPoints }}</text>
        <text class="metric-sub">{{ familyName }} · 可执行任务 {{ taskRows.length }} 项</text>
      </view>

      <view v-if="isChild" class="card green-card">
        <view class="card-title">
          <text>孩子领取状态</text>
          <text>{{ taskRows.length }} 项</text>
        </view>
        <view v-if="taskRows.length === 0" class="empty-row">
          <text>暂无可执行任务</text>
        </view>
        <view v-else class="list">
          <view class="row task-row" v-for="task in taskRows" :key="task.id">
            <view class="row-main">
              <text class="row-title">{{ task.title }}</text>
              <text class="row-meta">+{{ task.points }} 积分 · {{ cycleLabel(task.cycleType) }} · {{ taskButtonText(task) }}</text>
            </view>
            <view class="row-actions child-task-actions">
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
                class="btn light"
                size="mini"
                @click="stopRecurringTaskClaim(task)"
              >
                停止领取
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
  archiveTask,
  claimTask,
  createTask,
  listTaskClaims,
  listTaskRecords,
  listTasks,
  startTaskClaim,
  stopTaskClaim,
  submitTask,
  updateTask
} from '../../api/task.js'
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
const showTaskForm = ref(false)
const submittingTask = ref(false)
const taskForm = ref(defaultTaskForm())
const editingTaskId = ref(0)

const taskCycles = [
  { label: '每日', value: 'DAILY' },
  { label: '每周', value: 'WEEKLY' },
  { label: '一次性', value: 'ONCE' }
]

const roleType = computed(() => summary.value.roleType || currentFamily.value?.roleType || '')
const isChild = computed(() => roleType.value === 'CHILD')
const isParentRole = computed(() => ['OWNER', 'ADMIN', 'PARENT'].includes(roleType.value))
const familyName = computed(() => summary.value.familyName || currentFamily.value?.familyName || '当前家庭')
const currentPoints = computed(() => summary.value.currentPoints || 0)

const taskRows = computed(() => {
  return tasks.value.map((task) => {
    const claimed = claimedRecords.value.find((record) => record.taskId === task.id)
    const pending = pendingRecords.value.find((record) => record.taskId === task.id)
    const approved = approvedRecords.value.find((record) => record.taskId === task.id && recordMatchesCycle(record, task))
    const activeClaim = taskClaims.value.find((claim) => claim.taskId === task.id)
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
  }).filter(Boolean)
})

onShow(() => {
  loadTaskPage()
})

async function loadTaskPage(retried = false) {
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
      const [claimed, pending, approved, claims] = await Promise.all([
        listTaskRecords({ familyId, status: 'CLAIMED' }),
        listTaskRecords({ familyId, status: 'PENDING' }),
        listTaskRecords({ familyId, status: 'APPROVED' }),
        listTaskClaims(familyId)
      ])
      claimedRecords.value = claimed || []
      pendingRecords.value = pending || []
      approvedRecords.value = approved || []
      taskClaims.value = claims || []
    } else {
      claimedRecords.value = []
      pendingRecords.value = []
      approvedRecords.value = []
      taskClaims.value = []
    }
  } catch (error) {
    if (error.statusCode !== 401 || retried) {
      loadError.value = error.message || '任务加载失败，请稍后重试'
      return
    }
    await ensureDemoLogin(true)
    await loadTaskPage(true)
  } finally {
    loading.value = false
  }
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
  editingTaskId.value = 0
}

function openTaskForm() {
  editingTaskId.value = 0
  taskForm.value = defaultTaskForm()
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

async function confirmDuplicateTaskTitle(title) {
  const normalized = title.trim()
  const duplicated = tasks.value.some((task) => {
    return task.id !== editingTaskId.value && task.title.trim() === normalized
  })
  if (!duplicated) {
    return true
  }
  return new Promise((resolve) => {
    uni.showModal({
      title: '任务名称重复',
      content: '已存在同名任务，仍要继续吗？',
      confirmText: '继续',
      cancelText: '取消',
      success: (res) => resolve(res.confirm)
    })
  })
}

async function publishTask() {
  if (submittingTask.value) {
    return
  }

  const payload = validateTaskForm()
  if (!payload) {
    return
  }
  if (!(await confirmDuplicateTaskTitle(payload.title))) {
    return
  }

  submittingTask.value = true
  try {
    const familyId = currentFamily.value.familyId
    if (editingTaskId.value) {
      const taskId = editingTaskId.value
      const updatedTask = await updateTask({
        ...payload,
        familyId,
        taskId
      })
      applyUpdatedTask(updatedTask || { ...payload, id: taskId, familyId })
      uni.showToast({ title: '任务已更新', icon: 'success' })
    } else {
      const createdTask = await createTask({
        ...payload,
        familyId
      })
      applyCreatedTask(createdTask || { ...payload, familyId })
      uni.showToast({ title: '任务已发布', icon: 'success' })
    }
    showTaskForm.value = false
    resetTaskForm()
  } finally {
    submittingTask.value = false
  }
}

function editTask(task) {
  showTaskForm.value = false
  editingTaskId.value = task.id
  taskForm.value = {
    title: task.title,
    points: task.points,
    cycleType: task.cycleType
  }
  showTaskForm.value = true
}

function applyCreatedTask(task) {
  tasks.value = [
    task,
    ...tasks.value.filter((item) => item.id !== task.id)
  ]
}

function applyUpdatedTask(task) {
  tasks.value = tasks.value.map((item) => {
    if (item.id !== task.id) {
      return item
    }
    return {
      ...item,
      ...task
    }
  })
}

async function archiveExistingTask(task) {
  const confirmed = await new Promise((resolve) => {
    uni.showModal({
      title: '确认归档',
      content: `归档后孩子将不能再领取“${task.title}”`,
      success: (res) => resolve(res.confirm)
    })
  })
  if (!confirmed) {
    return
  }

  await archiveTask({
    familyId: currentFamily.value.familyId,
    taskId: task.id
  })
  uni.showToast({ title: '任务已归档', icon: 'success' })
  await loadTaskPage()
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

function applyTaskClaim(claim) {
  taskClaims.value = [
    claim,
    ...taskClaims.value.filter((item) => item.id !== claim.id && item.taskId !== claim.taskId)
  ]
}

function removeTaskClaim(taskId) {
  taskClaims.value = taskClaims.value.filter((item) => item.taskId !== taskId)
}

function applyClaimedTaskRecord(record) {
  claimedRecords.value = [
    record,
    ...claimedRecords.value.filter((item) => item.id !== record.id)
  ]
}

function applySubmittedTaskRecord(record) {
  claimedRecords.value = claimedRecords.value.filter((item) => item.id !== record.id)
  pendingRecords.value = [
    record,
    ...pendingRecords.value.filter((item) => item.id !== record.id)
  ]
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

function cycleLabel(value) {
  return taskCycles.find((cycle) => cycle.value === value)?.label || value
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
</script>

<style>
.tasks-page {
  padding-left: 40rpx;
  padding-right: 40rpx;
}

.tasks-content {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.page-head {
  align-items: flex-start;
  display: flex;
  gap: 20rpx;
  justify-content: space-between;
  margin-bottom: 28rpx;
}

.page-head > view {
  max-width: 460rpx;
}

.head-actions {
  display: flex;
  gap: 14rpx;
  justify-content: flex-end;
  margin: -98rpx 0 24rpx;
  min-height: 60rpx;
}

.eyebrow,
.title,
.subtitle,
.card-copy,
.section-name,
.metric-main,
.metric-sub,
.form-label,
.form-hint,
.row-title,
.row-meta {
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

.green-card {
  background: linear-gradient(135deg, #effbf3 0%, #fff 100%);
  border-color: rgba(99, 199, 132, 0.26);
}

.blue-card {
  background: linear-gradient(135deg, #eef7ff 0%, #fff 100%);
  border-color: rgba(75, 159, 255, 0.24);
}

.action-entry {
  align-items: center;
  display: flex;
  gap: 16rpx;
  justify-content: space-between;
}

.entry-main {
  flex: 1;
  min-width: 0;
}

.entry-title,
.entry-subtitle {
  display: block;
}

.entry-title {
  color: #172033;
  font-size: 28rpx;
  font-weight: 950;
}

.entry-subtitle {
  color: #707887;
  font-size: 24rpx;
  line-height: 1.45;
  margin-top: 8rpx;
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

.hero-card .card-title,
.hero-card .form-label,
.hero-card .form-hint,
.hero-card .section-name,
.hero-card .metric-main,
.hero-card .metric-sub {
  color: #fff;
}

.card-copy {
  color: rgba(255, 255, 255, 0.92);
  font-size: 24rpx;
  line-height: 1.45;
  margin: -4rpx 0 22rpx;
}

.section-name {
  font-size: 24rpx;
  font-weight: 900;
  margin-bottom: 12rpx;
}

.metric-main {
  font-size: 62rpx;
  font-weight: 950;
  line-height: 1;
}

.metric-sub {
  font-size: 24rpx;
  line-height: 1.35;
  margin-top: 14rpx;
}

.create-form,
.list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.task-item {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
}

.inline-edit-panel {
  background: linear-gradient(135deg, #ffb84d 0%, #ff7a45 66%);
  border: 1rpx solid rgba(255, 255, 255, 0.36);
  border-radius: 30rpx;
  box-shadow: 0 24rpx 48rpx rgba(255, 122, 69, 0.18);
  padding: 24rpx;
}

.inline-edit-title {
  color: #fff;
  margin-bottom: 18rpx;
}

.inline-edit-panel .form-label,
.inline-edit-panel .form-hint {
  color: #fff;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.form-label {
  font-size: 24rpx;
  font-weight: 900;
}

.input-shell {
  align-items: center;
  background: rgba(255, 255, 255, 0.2);
  border: 1rpx solid rgba(255, 255, 255, 0.36);
  border-radius: 24rpx;
  display: flex;
  min-height: 78rpx;
  padding: 0 24rpx;
}

.form-input {
  color: #fff;
  flex: 1;
  font-size: 24rpx;
  min-height: 76rpx;
}

.hero-placeholder {
  color: rgba(255, 255, 255, 0.72);
}

.cycle-options {
  display: grid;
  gap: 16rpx;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.form-actions {
  align-items: center;
  display: flex;
  gap: 16rpx;
  justify-content: space-between;
  margin-top: 2rpx;
}

.form-hint {
  color: rgba(255, 255, 255, 0.9);
  flex: 1;
  font-size: 22rpx;
  line-height: 1.4;
}

.form-buttons {
  display: flex;
  flex: 0 0 auto;
  gap: 12rpx;
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

.row-meta {
  color: #7b8190;
  font-size: 22rpx;
  line-height: 1.35;
  margin-top: 8rpx;
}

.row-actions,
.child-task-actions {
  align-items: flex-end;
  display: flex;
  flex: 0 0 auto;
  gap: 10rpx;
}

.child-task-actions {
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

.tasks-page button,
.tasks-page .btn {
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

.tasks-page .btn.primary {
  background: #ff7a45;
}

.tasks-page .btn.blue {
  background: #4b9fff;
  box-shadow: 0 16rpx 32rpx rgba(75, 159, 255, 0.18);
}

.tasks-page .btn.light {
  background: #eef7ff;
  border: 1rpx solid rgba(75, 159, 255, 0.18);
  box-shadow: none;
  color: #2f80ed;
}

.tasks-page .btn.danger {
  background: #fff1f3;
  border: 1rpx solid rgba(239, 107, 122, 0.2);
  box-shadow: none;
  color: #d95061;
}

.tasks-page .cycle-btn {
  background: rgba(255, 255, 255, 0.92);
  border: 1rpx solid rgba(255, 255, 255, 0.3);
  box-shadow: none;
  color: #172033;
  min-width: 0;
  padding: 0 12rpx;
}

.tasks-page .cycle-btn.active {
  background: #eef7ff;
  border-color: rgba(75, 159, 255, 0.24);
  color: #2f80ed;
}

.tasks-page .action-btn.claimed,
.tasks-page .action-btn.recurringActive {
  background: #4b9fff;
}

.tasks-page .action-btn.pending,
.tasks-page .action-btn.completed {
  background: #eef2f7;
  box-shadow: none;
  color: #8a95a5;
}

.retry-btn {
  margin-top: 4rpx;
}
</style>
