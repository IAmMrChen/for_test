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
