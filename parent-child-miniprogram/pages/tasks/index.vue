<template>
  <view class="container">
    <view class="header">
      <view>
        <text class="title">任务</text>
        <text class="subtitle">{{ familyName }}</text>
      </view>
      <view class="summary-pill">
        <text class="summary-label">{{ isChild ? '当前积分' : '可用任务' }}</text>
        <text class="summary-value">{{ isChild ? currentPoints : tasks.length }}</text>
      </view>
    </view>

    <view v-if="loading" class="state-block">
      <text>正在加载任务...</text>
    </view>

    <view v-else class="content">
      <view v-if="isParentRole" class="create-panel">
        <view class="create-header">
          <view>
            <text class="create-title">发布任务</text>
            <text class="create-subtitle">给孩子增加一个可领取的任务</text>
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
              {{ submittingTask ? '保存中' : submitTaskText }}
            </button>
          </view>
        </view>
      </view>

      <view v-if="isParentRole" class="section">
        <view class="section-header">
          <text class="section-title">任务管理</text>
          <text class="section-count">{{ tasks.length }} 项</text>
        </view>
        <view v-if="tasks.length === 0" class="state-block compact">
          <text>暂无可管理任务</text>
        </view>
        <view v-else class="task-list">
          <view class="task-card" v-for="task in tasks" :key="task.id">
            <view class="task-main">
              <text class="task-title">{{ task.title }}</text>
              <text class="task-meta">+{{ task.points }} 积分 · {{ cycleLabel(task.cycleType) }}</text>
            </view>
            <view class="task-actions">
              <button class="plain-action-btn" size="mini" @click="editTask(task)">编辑</button>
              <button class="danger-action-btn" size="mini" @click="archiveExistingTask(task)">归档</button>
            </view>
          </view>
        </view>
      </view>

      <view v-if="isChild" class="section">
        <view class="section-header">
          <text class="section-title">我的任务</text>
          <text class="section-count">{{ taskRows.length }} 项</text>
        </view>
        <view v-if="taskRows.length === 0" class="state-block compact">
          <text>暂无可执行任务</text>
        </view>
        <view v-else class="task-list">
          <view class="task-card" v-for="task in taskRows" :key="task.id">
            <view class="task-main">
              <text class="task-title">{{ task.title }}</text>
              <text class="task-meta">+{{ task.points }} 积分 · {{ cycleLabel(task.cycleType) }}</text>
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
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'

import { ensureDemoLogin } from '../../api/auth.js'
import { getDashboardSummary } from '../../api/dashboard.js'
import { archiveTask, claimTask, createTask, listTaskRecords, listTasks, submitTask, updateTask } from '../../api/task.js'
import { getCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const currentFamily = ref(null)
const summary = ref({})
const tasks = ref([])
const claimedRecords = ref([])
const pendingRecords = ref([])
const approvedRecords = ref([])
const showTaskForm = ref(false)
const submittingTask = ref(false)
const taskForm = ref(defaultTaskForm())
const editingTaskId = ref(0)

const taskCycles = [
  { label: '一次性', value: 'ONCE' },
  { label: '每日', value: 'DAILY' },
  { label: '每周', value: 'WEEKLY' }
]

const roleType = computed(() => summary.value.roleType || currentFamily.value?.roleType || '')
const isChild = computed(() => roleType.value === 'CHILD')
const isParentRole = computed(() => ['OWNER', 'ADMIN', 'PARENT'].includes(roleType.value))
const familyName = computed(() => summary.value.familyName || currentFamily.value?.familyName || '当前家庭')
const currentPoints = computed(() => summary.value.currentPoints || 0)
const submitTaskText = computed(() => editingTaskId.value ? '确认保存' : '确认发布')

const taskRows = computed(() => {
  return tasks.value.map((task) => {
    const claimed = claimedRecords.value.find((record) => record.taskId === task.id)
    const pending = pendingRecords.value.find((record) => record.taskId === task.id)
    const approved = approvedRecords.value.find((record) => record.taskId === task.id && recordMatchesCycle(record, task))
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
  })
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
      const [claimed, pending, approved] = await Promise.all([
        listTaskRecords({ familyId, status: 'CLAIMED' }),
        listTaskRecords({ familyId, status: 'PENDING' }),
        listTaskRecords({ familyId, status: 'APPROVED' })
      ])
      claimedRecords.value = claimed || []
      pendingRecords.value = pending || []
      approvedRecords.value = approved || []
    } else {
      claimedRecords.value = []
      pendingRecords.value = []
      approvedRecords.value = []
    }
  } catch (error) {
    if (error.statusCode !== 401 || retried) {
      throw error
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
      await updateTask({
        ...payload,
        familyId,
        taskId: editingTaskId.value
      })
      uni.showToast({ title: '任务已更新', icon: 'success' })
    } else {
      await createTask({
        ...payload,
        familyId
      })
      uni.showToast({ title: '任务已发布', icon: 'success' })
    }
    showTaskForm.value = false
    resetTaskForm()
    await loadTaskPage()
  } finally {
    submittingTask.value = false
  }
}

function editTask(task) {
  editingTaskId.value = task.id
  taskForm.value = {
    title: task.title,
    points: task.points,
    cycleType: task.cycleType
  }
  showTaskForm.value = true
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
    await claimTask({ familyId, taskId: task.id })
    uni.showToast({ title: '已领取任务', icon: 'success' })
    await loadTaskPage()
    return
  }
  if (task.viewStatus === 'claimed') {
    await submitTask({ familyId, recordId: task.record.id })
    uni.showToast({ title: '已提交审核', icon: 'success' })
    await loadTaskPage()
  }
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
    pending: '审核中',
    completed: '已完成'
  }
  return map[task.viewStatus] || '查看'
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
.task-meta,
.section-count {
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
.task-list {
  display: flex;
  flex-direction: column;
}

.content,
.section {
  gap: 18px;
}

.section-header {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.section-title {
  color: #1f2937;
  font-size: 16px;
  font-weight: 700;
}

.task-list {
  gap: 12px;
}

.task-card,
.state-block,
.create-panel {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
}

.task-card {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.task-main {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 5px;
  min-width: 0;
}

.task-title {
  color: #111827;
  font-size: 15px;
  font-weight: 700;
}

.task-actions {
  display: flex;
  gap: 8px;
  margin-left: 12px;
}

.action-btn,
.plain-action-btn,
.primary-action-btn,
.danger-action-btn,
.small-primary-btn {
  border-radius: 16px;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  padding: 0 14px;
}

.action-btn,
.primary-action-btn,
.small-primary-btn {
  background: #2563eb;
  border: none;
  color: #fff;
}

.action-btn.claimed {
  background: #10b981;
}

.action-btn.pending,
.action-btn.completed {
  background: #e5e7eb;
  color: #64748b;
}

.plain-action-btn {
  background: #fff;
  border: 1px solid #cbd5e1;
  color: #475569;
}

.danger-action-btn {
  background: #fff;
  border: 1px solid #fecaca;
  color: #dc2626;
}

.state-block {
  color: #64748b;
  font-size: 14px;
  text-align: center;
}

.state-block.compact {
  padding: 20px 16px;
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
</style>
