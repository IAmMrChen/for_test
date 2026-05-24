<template>
  <view class="sun-page points-page">
    <view class="header">
      <view>
        <text class="title">积分流水</text>
        <text class="subtitle">{{ familyName }}</text>
      </view>
      <button class="back-btn" size="mini" @click="goBack">返回</button>
    </view>

    <view v-if="isParentRole && children.length > 0" class="filter-row">
      <button
        class="filter-btn"
        :class="{ active: selectedMemberId === 0 }"
        size="mini"
        @click="selectMember(0)"
      >
        全部
      </button>
      <button
        v-for="child in children"
        :key="child.id"
        class="filter-btn"
        :class="{ active: selectedMemberId === child.id }"
        size="mini"
        @click="selectMember(child.id)"
      >
        {{ child.nickname }}
      </button>
    </view>

    <view v-if="loading" class="state-block">
      <text>正在加载积分流水...</text>
    </view>

    <view v-else-if="loadError" class="state-block">
      <text>{{ loadError }}</text>
      <button class="plain-action-btn retry-btn" size="mini" @click="loadPointLogs">重新加载</button>
    </view>

    <view v-else-if="pointLogs.length === 0" class="state-block">
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
const loadError = ref('')
const currentFamily = ref(null)
const summary = ref({})
const members = ref([])
const pointLogs = ref([])
const selectedMemberId = ref(0)

const roleType = computed(() => summary.value.roleType || currentFamily.value?.roleType || '')
const isParentRole = computed(() => ['OWNER', 'ADMIN', 'PARENT'].includes(roleType.value))
const familyName = computed(() => summary.value.familyName || currentFamily.value?.familyName || '当前家庭')
const children = computed(() => members.value.filter((member) => member.roleType === 'CHILD'))

onShow(() => {
  loadPage()
})

async function loadPage(retried = false) {
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
    const summaryData = await getDashboardSummary(familyId)
    summary.value = summaryData || {}

    if (isParentRole.value) {
      members.value = (await listMembers(familyId)) || []
    } else {
      members.value = []
      selectedMemberId.value = 0
    }

    await loadPointLogs()
  } catch (error) {
    if (error.statusCode === 401 && !retried) {
      await ensureDemoLogin(true)
      await loadPage(true)
      return
    }
    loadError.value = error.message || '积分流水加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

async function loadPointLogs() {
  if (!currentFamily.value) {
    return
  }
  loadError.value = ''
  const params = { familyId: currentFamily.value.familyId }
  if (selectedMemberId.value) {
    params.memberId = selectedMemberId.value
  }
  pointLogs.value = (await listPointLogs(params)) || []
}

async function selectMember(memberId) {
  if (selectedMemberId.value === memberId) {
    return
  }
  selectedMemberId.value = memberId
  loading.value = true
  try {
    await loadPointLogs()
  } catch (error) {
    loadError.value = error.message || '积分流水加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

function goBack() {
  uni.navigateBack()
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
  padding: 48px 20px 24px;
}

.header {
  align-items: flex-start;
  display: flex;
  justify-content: space-between;
  margin-bottom: 18px;
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
.log-meta {
  color: #64748b;
  font-size: 12px;
}

.back-btn,
.plain-action-btn {
  background: #fff;
  border: 1px solid #cbd5e1;
  border-radius: 16px;
  color: #475569;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  padding: 0 14px;
}

.filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 16px;
}

.filter-btn {
  background: #fff;
  border: 1px solid #dbe3ef;
  border-radius: 16px;
  color: #475569;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  padding: 0 14px;
}

.filter-btn.active {
  background: #eff6ff;
  border-color: #2563eb;
  color: #2563eb;
}

.log-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.log-card,
.state-block {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
}

.log-card {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.log-main {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.log-title-row {
  align-items: center;
  display: flex;
  gap: 6px;
}

.log-title {
  color: #111827;
  font-size: 15px;
  font-weight: 700;
}

.source-tag {
  background: #eff6ff;
  border-radius: 4px;
  color: #2563eb;
  font-size: 11px;
  padding: 2px 6px;
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

.retry-btn {
  margin-top: 12px;
}

.points-page .header {
  margin-bottom: 28rpx;
}

.points-page .title {
  color: var(--sun-ink);
  font-size: 46rpx;
  font-weight: 800;
}

.points-page .subtitle,
.points-page .log-meta {
  color: var(--sun-muted);
}

.points-page .log-card,
.points-page .state-block {
  background: rgba(255, 255, 255, 0.9);
  border: 1rpx solid rgba(255, 184, 77, 0.24);
  border-radius: 28rpx;
  box-shadow: 0 14rpx 34rpx rgba(43, 42, 40, 0.08);
}

.points-page .back-btn,
.points-page .plain-action-btn,
.points-page .filter-btn {
  background: #eef7ff;
  border-color: rgba(75, 159, 255, 0.22);
  border-radius: 999rpx;
  color: var(--sun-sky);
}

.points-page .filter-btn.active {
  background: var(--sun-sky);
  border-color: var(--sun-sky);
  color: #fff;
}

.points-page .source-tag {
  background: #eef7ff;
  color: var(--sun-sky);
}

.points-page .log-points {
  color: var(--sun-mint);
}
</style>
