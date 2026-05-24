<template>
  <view class="sun-page proxy-page">
    <view class="page-header">
      <view>
        <text class="sun-title">代孩子兑换奖励</text>
        <text class="sun-subtitle">{{ familyName }}</text>
      </view>
      <button class="sun-btn secondary header-btn" size="mini" @click="goBack">返回</button>
    </view>

    <view v-if="loading" class="sun-card state-card">
      <text>正在加载奖励...</text>
    </view>

    <view v-else class="content">
      <view class="sun-card child-card">
        <view class="sun-card-title">
          <text>选择孩子</text>
        </view>
        <view v-if="children.length === 0" class="empty-text">
          <text>暂无孩子，请先到成员管理中创建或邀请孩子</text>
        </view>
        <view v-else class="child-options">
          <button
            v-for="child in children"
            :key="child.id"
            class="sun-chip"
            :class="{ active: selectedChildId === child.id }"
            size="mini"
            @click="selectChild(child.id)"
          >
            {{ child.nickname }}
          </button>
        </view>
      </view>

      <view v-if="selectedChild" class="sun-card points-card">
        <view>
          <text class="points-label">当前积分</text>
          <text class="points-value">{{ selectedChild.currentPoints || 0 }}</text>
        </view>
        <text class="points-hint">为 {{ selectedChild.nickname }} 兑换</text>
      </view>

      <view class="sun-card reward-section">
        <view class="sun-card-title">
          <text>可兑换奖励</text>
        </view>

        <view v-if="!selectedChild" class="empty-text">
          <text>请先选择一个孩子</text>
        </view>
        <view v-else-if="rewardRows.length === 0" class="empty-text">
          <text>暂无可兑换奖励</text>
        </view>
        <view v-else class="reward-list">
          <view class="reward-card" v-for="reward in rewardRows" :key="reward.id">
            <view class="reward-main">
              <text class="reward-name">{{ reward.name }}</text>
              <text class="reward-meta">{{ stockText(reward.stock) }}</text>
            </view>
            <view class="reward-side">
              <text class="reward-points">{{ reward.pointsCost }} 积分</text>
              <button
                class="sun-btn action-btn"
                size="mini"
                :class="{ secondary: reward.actionStatus !== 'available' }"
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
  color: var(--sun-muted);
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
  color: var(--sun-muted);
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
  color: var(--sun-ink);
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
</style>
