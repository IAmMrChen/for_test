<template>
  <view class="sun-page proxy-page">
    <view class="page-head">
      <view>
        <text class="eyebrow">代操作</text>
        <text class="title">代孩子兑换奖励</text>
        <text class="subtitle">支持真实孩子和虚拟孩子</text>
      </view>
    </view>

    <view class="head-actions">
      <button class="btn light" size="mini" @click="goBack">返回奖励页</button>
    </view>

    <view v-if="loading" class="card state-card">
      <text>正在加载奖励...</text>
    </view>

    <view v-else class="proxy-content">
      <view class="card child-card">
        <view class="card-title">
          <text>选择孩子</text>
        </view>
        <view v-if="children.length === 0" class="empty-row">
          <text>暂无孩子，请先到成员管理中创建或邀请孩子</text>
        </view>
        <view v-else class="chips child-options">
          <button
            v-for="child in children"
            :key="child.id"
            class="chip"
            :class="{ active: selectedChildId === child.id }"
            size="mini"
            @click="selectChild(child.id)"
          >
            {{ child.nickname }}
          </button>
        </view>
      </view>

      <view v-if="selectedChild" class="card hero-card points-card">
        <text class="section-name">{{ selectedChild.nickname }}可用积分</text>
        <text class="metric-main">{{ selectedChild.currentPoints || 0 }}</text>
        <text class="metric-sub">筛选后只处理这个孩子的奖励兑换。</text>
      </view>

      <view class="filters">
        <button class="chip active" size="mini">全部</button>
        <button class="chip" size="mini">可兑换</button>
        <button class="chip" size="mini">积分不足</button>
        <button class="chip" size="mini">已申请</button>
      </view>

      <view class="card reward-section">
        <view class="card-title">
          <text>可兑换奖励</text>
          <text>{{ rewardRows.length }} 项</text>
        </view>

        <view v-if="!selectedChild" class="empty-row">
          <text>请先选择一个孩子</text>
        </view>
        <view v-else-if="rewardRows.length === 0" class="empty-row">
          <text>暂无可兑换奖励</text>
        </view>
        <view v-else class="list">
          <view class="row reward-row" v-for="reward in rewardRows" :key="reward.id">
            <view class="row-main">
              <text class="row-title">{{ reward.name }}</text>
              <text class="row-meta">{{ reward.pointsCost }} 积分 · {{ stockText(reward.stock) }}</text>
            </view>
            <view class="row-actions">
              <button
                class="btn action-btn"
                size="mini"
                :class="[reward.actionStatus, { light: reward.actionStatus !== 'available', primary: reward.actionStatus === 'available' }]"
                :disabled="reward.actionStatus !== 'available' || submittingRewardId === reward.id"
                @click="applyProxyReward(reward)"
              >
                {{ rewardButtonText(reward) }}
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
import { applyReward, listRewardRecords, listRewards } from '../../api/reward.js'
import { getCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const currentFamily = ref(null)
const summary = ref({})
const members = ref([])
const rewards = ref([])
const appliedRecords = ref([])
const selectedChildId = ref(0)
const submittingRewardId = ref(0)

const familyName = computed(() => summary.value.familyName || currentFamily.value?.familyName || '当前家庭')
const children = computed(() => members.value.filter((member) => member.roleType === 'CHILD'))
const selectedChild = computed(() => children.value.find((member) => member.id === selectedChildId.value) || null)

const rewardRows = computed(() => {
  const child = selectedChild.value
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
  try {
    await ensureDemoLogin()
    const familyId = family.familyId
    const [summaryData, memberList, rewardList, applied] = await Promise.all([
      getDashboardSummary(familyId),
      listMembers(familyId),
      listRewards(familyId),
      listRewardRecords({ familyId, status: 'APPLIED' })
    ])

    summary.value = summaryData || {}
    members.value = memberList || []
    rewards.value = rewardList || []
    appliedRecords.value = applied || []
    syncSelectedChild()
  } catch (error) {
    if (error.statusCode !== 401 || retried) {
      throw error
    }
    await ensureDemoLogin(true)
    await loadPage(true)
  } finally {
    loading.value = false
  }
}

function syncSelectedChild() {
  if (children.value.some((child) => child.id === selectedChildId.value)) {
    return
  }
  selectedChildId.value = children.value[0]?.id || 0
}

function selectChild(childId) {
  selectedChildId.value = childId
}

async function applyProxyReward(reward) {
  const child = selectedChild.value
  if (!child) {
    uni.showToast({ title: '请先选择孩子', icon: 'none' })
    return
  }

  submittingRewardId.value = reward.id
  try {
    const record = await applyReward({
      familyId: currentFamily.value.familyId,
      rewardId: reward.id,
      memberId: child.id
    })
    addAppliedRewardRecord(record, reward, child)
    uni.showToast({ title: '已为孩子兑换', icon: 'success' })
  } finally {
    submittingRewardId.value = 0
  }
}

function addAppliedRewardRecord(record, reward, child) {
  appliedRecords.value = [
    {
      ...record,
      rewardId: reward.id,
      rewardName: reward.name,
      pointsCost: reward.pointsCost,
      memberId: child.id,
      nickname: child.nickname
    },
    ...appliedRecords.value.filter((item) => item.id !== record.id)
  ]
  rewards.value = rewards.value.map((item) => {
    if (item.id !== reward.id || item.stock < 0) {
      return item
    }
    return {
      ...item,
      stock: Math.max((item.stock || 0) - 1, 0)
    }
  })
  members.value = members.value.map((member) => {
    if (member.id !== child.id) {
      return member
    }
    return {
      ...member,
      currentPoints: Math.max((member.currentPoints || 0) - reward.pointsCost, 0)
    }
  })
}

function rewardButtonText(reward) {
  if (submittingRewardId.value === reward.id) {
    return '处理中'
  }
  const map = {
    available: '兑换',
    applied: '待发放',
    insufficient: '积分不足',
    soldout: '已兑完'
  }
  return map[reward.actionStatus] || '兑换'
}

function stockText(stock) {
  if (stock < 0) {
    return '不限库存'
  }
  if (stock === 0) {
    return '已兑完'
  }
  return `剩余 ${stock} 份`
}

function goBack() {
  uni.navigateBack()
}
</script>

<style>
.proxy-page {
  box-sizing: border-box;
}

.page-header {
  align-items: flex-start;
  display: flex;
  justify-content: space-between;
  margin-bottom: 24rpx;
}

.header-btn {
  margin: 0;
  min-width: 112rpx;
}

.content,
.reward-list,
.reward-main,
.reward-side {
  display: flex;
  flex-direction: column;
}

.content {
  gap: 22rpx;
}

.state-card,
.empty-text {
  color: #707887;
  font-size: 26rpx;
  line-height: 1.5;
  text-align: center;
}

.child-options {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
  margin-top: 22rpx;
}

.points-card {
  align-items: center;
  background: linear-gradient(135deg, #fff2cf 0%, #fff 64%);
  display: flex;
  justify-content: space-between;
}

.points-label,
.points-value,
.points-hint,
.reward-name,
.reward-meta,
.reward-points {
  display: block;
}

.points-label,
.reward-meta,
.points-hint {
  color: #707887;
  font-size: 24rpx;
}

.points-value {
  color: #e9852c;
  font-size: 48rpx;
  font-weight: 800;
  line-height: 1.15;
}

.reward-list {
  gap: 18rpx;
  margin-top: 22rpx;
}

.reward-card {
  align-items: center;
  background: #fffdf8;
  border: 1rpx solid rgba(255, 184, 77, 0.26);
  border-radius: 24rpx;
  display: flex;
  gap: 20rpx;
  justify-content: space-between;
  padding: 22rpx;
}

.reward-main {
  gap: 8rpx;
  min-width: 0;
}

.reward-side {
  align-items: flex-end;
  flex-shrink: 0;
  gap: 10rpx;
}

.reward-name {
  color: #172033;
  font-size: 30rpx;
  font-weight: 800;
}

.reward-points {
  color: #e9852c;
  font-size: 24rpx;
  font-weight: 800;
}

.action-btn {
  margin: 0;
  min-width: 116rpx;
}

.proxy-page {
  padding-left: 40rpx;
  padding-right: 40rpx;
}

.proxy-content {
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

.head-actions {
  display: flex;
  gap: 14rpx;
  margin: -14rpx 0 24rpx;
}

.eyebrow,
.title,
.subtitle,
.section-name,
.metric-main,
.metric-sub,
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

.section-name,
.metric-main,
.metric-sub {
  color: #fff;
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

.chips,
.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 14rpx;
}

.list {
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

.row-meta {
  color: #7b8190;
  font-size: 22rpx;
  line-height: 1.35;
  margin-top: 8rpx;
}

.row-actions {
  align-items: flex-end;
  display: flex;
  flex: 0 0 auto;
  gap: 10rpx;
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

.proxy-page button,
.proxy-page .btn,
.proxy-page .chip {
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

.proxy-page .btn.primary {
  background: #ff7a45;
}

.proxy-page .btn.light,
.proxy-page .chip {
  background: #eef7ff;
  border: 1rpx solid rgba(75, 159, 255, 0.18);
  box-shadow: none;
  color: #2f80ed;
}

.proxy-page .chip {
  background: rgba(255, 255, 255, 0.9);
  border-color: #edf0f5;
  color: #697180;
  min-width: 88rpx;
}

.proxy-page .chip.active {
  background: #4b9fff;
  border-color: #4b9fff;
  box-shadow: 0 16rpx 36rpx rgba(75, 159, 255, 0.2);
  color: #fff;
}

.proxy-page .action-btn.light,
.proxy-page .action-btn.applied,
.proxy-page .action-btn.insufficient,
.proxy-page .action-btn.soldout {
  background: #eef2f7;
  box-shadow: none;
  color: #8a95a5;
}

.proxy-page .points-card {
  align-items: stretch;
  display: block;
}

.proxy-page .points-card .section-name {
  margin-bottom: 10rpx;
}

.proxy-page .points-card .metric-main {
  margin-bottom: 10rpx;
}

.proxy-page .filters {
  margin-top: -6rpx;
}
</style>
