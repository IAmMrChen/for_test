<template>
  <view class="sun-page points-page">
    <view class="page-head">
      <view>
        <text class="eyebrow">积分记录</text>
        <text class="title">积分流水</text>
        <text class="subtitle">收入、兑换和调整都在这里</text>
      </view>
      <button class="btn light" size="mini" @click="goBack">返回</button>
    </view>

    <view v-if="loading" class="card state-card">
      <text>正在加载积分流水...</text>
    </view>

    <view v-else-if="loadError" class="card state-card">
      <text>{{ loadError }}</text>
      <button class="btn light retry-btn" size="mini" @click="loadPointLogs">重新加载</button>
    </view>

    <view v-else class="points-content">
      <view v-if="isParentRole && children.length > 0" class="chips filter-row">
        <button
          class="chip"
          :class="{ active: selectedMemberId === 0 }"
          size="mini"
          @click="selectMember(0)"
        >
          全部
        </button>
        <button
          v-for="child in children"
          :key="child.id"
          class="chip"
          :class="{ active: selectedMemberId === child.id }"
          size="mini"
          @click="selectMember(child.id)"
        >
          {{ child.nickname }}
        </button>
      </view>

      <view class="card blue-card intro-card">
        <view class="card-title">
          <text>最近一个月流水</text>
          <button class="btn light" size="mini">筛选</button>
        </view>
        <text class="card-copy">先按成员筛选并按时间倒序展示；近一年分页和日期筛选放到后续规划。</text>
      </view>

      <view v-if="pointLogs.length === 0" class="card state-card">
        <text>暂无积分流水</text>
      </view>

      <view v-else class="log-groups">
        <view class="card log-section" v-for="group in groupedPointLogs" :key="group.key">
          <view class="card-title">
            <text>{{ group.title }}</text>
          </view>
          <view class="list">
            <view class="row log-card" v-for="log in group.items" :key="log.id">
              <view class="row-main">
                <view class="log-title-row">
                  <text class="row-title">{{ log.sourceTitle || '未命名来源' }}</text>
                  <text class="source-tag">{{ sourceName(log.sourceType) }}</text>
                </view>
                <text class="row-meta">{{ log.nickname || '成员' }} · {{ formatTime(log.createdAt) }}</text>
              </view>
              <text class="log-points" :class="{ negative: log.points < 0 }">{{ pointText(log.points) }}</text>
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
const groupedPointLogs = computed(() => {
  const groups = []
  const groupMap = new Map()
  for (const log of pointLogs.value) {
    const key = dateKey(log.createdAt)
    if (!groupMap.has(key)) {
      const group = {
        key,
        title: dateGroupTitle(log.createdAt),
        items: []
      }
      groups.push(group)
      groupMap.set(key, group)
    }
    groupMap.get(key).items.push(log)
  }
  return groups
})

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

function dateKey(value) {
  if (!value) {
    return 'unknown'
  }
  const date = new Date(value)
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}

function dateGroupTitle(value) {
  if (!value) {
    return '未知日期'
  }
  const date = new Date(value)
  const today = new Date()
  if (date.toDateString() === today.toDateString()) {
    return '今天'
  }
  const yesterday = new Date(today.getFullYear(), today.getMonth(), today.getDate() - 1)
  if (date.toDateString() === yesterday.toDateString()) {
    return '昨天'
  }
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${month}-${day}`
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
  color: #172033;
  font-size: 46rpx;
  font-weight: 800;
}

.points-page .subtitle,
.points-page .log-meta {
  color: #707887;
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
  color: #4b9fff;
}

.points-page .filter-btn.active {
  background: #4b9fff;
  border-color: #4b9fff;
  color: #fff;
}

.points-page .source-tag {
  background: #eef7ff;
  color: #4b9fff;
}

.points-page .log-points {
  color: #63c784;
}

.points-page button {
  min-height: 64rpx;
  border-radius: 999rpx;
  font-size: 24rpx;
  font-weight: 800;
  line-height: 64rpx;
  padding: 0 26rpx;
}

.points-page {
  padding-left: 40rpx;
  padding-right: 40rpx;
}

.points-content {
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

.eyebrow,
.title,
.subtitle,
.card-copy,
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

.points-page .title {
  color: #172033;
  font-size: 50rpx;
  font-weight: 950;
  line-height: 1.12;
  margin-bottom: 0;
}

.points-page .subtitle,
.row-meta,
.card-copy {
  color: #707887;
  font-size: 24rpx;
  line-height: 1.45;
}

.points-page .subtitle {
  margin-top: 14rpx;
}

.card {
  background: rgba(255, 255, 255, 0.9);
  border: 1rpx solid rgba(255, 184, 77, 0.3);
  border-radius: 36rpx;
  box-shadow: 0 24rpx 56rpx rgba(43, 42, 40, 0.08);
  padding: 26rpx;
}

.blue-card {
  background: linear-gradient(135deg, #eef7ff 0%, #fff 100%);
  border-color: rgba(75, 159, 255, 0.24);
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

.list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.log-groups {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
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
}

.row-meta {
  margin-top: 8rpx;
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 14rpx;
}

.state-card {
  color: #7b8190;
  font-size: 26rpx;
  min-height: 104rpx;
  text-align: center;
}

.points-page button,
.points-page .btn,
.points-page .chip {
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
}

.points-page .btn.light,
.points-page .chip {
  background: #eef7ff;
  border: 1rpx solid rgba(75, 159, 255, 0.18);
  box-shadow: none;
  color: #2f80ed;
}

.points-page .chip {
  background: rgba(255, 255, 255, 0.9);
  border-color: #edf0f5;
  color: #697180;
  min-width: 88rpx;
}

.points-page .chip.active {
  background: #4b9fff;
  border-color: #4b9fff;
  box-shadow: 0 16rpx 36rpx rgba(75, 159, 255, 0.2);
  color: #fff;
}
</style>
