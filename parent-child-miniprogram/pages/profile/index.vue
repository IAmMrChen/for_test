<template>
  <view class="sun-page profile-page">
    <view class="sun-card user-card">
      <view class="avatar">{{ avatarText }}</view>
      <view class="info">
        <text class="nickname">{{ nickname }}</text>
        <text class="current-family">{{ familyName }} · {{ roleName(roleType) }}</text>
      </view>
      <button class="sun-btn secondary switch-btn" size="mini" @click="switchFamily">切换</button>
    </view>

    <view class="points-row">
      <view class="sun-card point-card">
        <text class="point-value">{{ currentPoints }}</text>
        <text class="point-label">当前积分</text>
      </view>
      <view class="sun-card point-card">
        <text class="point-value total">{{ totalEarnedPoints }}</text>
        <text class="point-label">累计积分</text>
      </view>
    </view>

    <view v-if="loading" class="sun-card state-block">
      <text>正在加载个人页...</text>
    </view>

    <view v-else class="content">
      <view v-if="isParentRole" class="entry-card sun-card" @click="goMembers">
        <view>
          <text class="entry-title">家庭成员</text>
          <text class="entry-subtitle">管理孩子、邀请家人和关联虚拟孩子</text>
        </view>
        <button class="sun-btn secondary entry-btn" size="mini" @click.stop="goMembers">进入</button>
      </view>

      <view class="entry-card sun-card" @click="goPointLogs">
        <view>
          <text class="entry-title">积分流水</text>
          <text class="entry-subtitle">查看积分收入、兑换和调整记录</text>
        </view>
        <button class="sun-btn secondary entry-btn" size="mini" @click.stop="goPointLogs">进入</button>
      </view>
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'

import { ensureDemoLogin } from '../../api/auth.js'
import { getDashboardSummary } from '../../api/dashboard.js'
import { getCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const currentFamily = ref(null)
const summary = ref({})

const familyName = computed(() => summary.value.familyName || currentFamily.value?.familyName || '未选择家庭')
const nickname = computed(() => summary.value.nickname || '我的')
const roleType = computed(() => summary.value.roleType || currentFamily.value?.roleType || '')
const isParentRole = computed(() => ['OWNER', 'ADMIN', 'PARENT'].includes(roleType.value))
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
    summary.value = (await getDashboardSummary(family.familyId)) || {}
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

function goMembers() {
  uni.navigateTo({ url: '/pages/members/index' })
}

function goPointLogs() {
  uni.navigateTo({ url: '/pages/points/index' })
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
</script>

<style>
.profile-page {
  box-sizing: border-box;
}

.user-card {
  align-items: center;
  display: flex;
  gap: 22rpx;
}

.avatar {
  align-items: center;
  background: linear-gradient(135deg, #ffcf6c 0%, #ff8f5d 100%);
  border-radius: 50%;
  color: #fff;
  display: flex;
  flex-shrink: 0;
  font-size: 34rpx;
  font-weight: 800;
  height: 96rpx;
  justify-content: center;
  width: 96rpx;
}

.info {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 8rpx;
  min-width: 0;
}

.nickname,
.current-family,
.point-value,
.point-label,
.entry-title,
.entry-subtitle {
  display: block;
}

.nickname {
  color: var(--sun-ink);
  font-size: 34rpx;
  font-weight: 800;
}

.current-family,
.point-label,
.entry-subtitle,
.state-block {
  color: var(--sun-muted);
  font-size: 24rpx;
  line-height: 1.45;
}

.switch-btn,
.entry-btn {
  margin: 0;
}

.points-row {
  display: grid;
  gap: 18rpx;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.point-card {
  margin-bottom: 0;
  text-align: center;
}

.point-value {
  color: #e9852c;
  font-size: 46rpx;
  font-weight: 800;
  line-height: 1.15;
}

.point-value.total {
  color: var(--sun-sky);
}

.content {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
  margin-top: 22rpx;
}

.entry-card {
  align-items: center;
  display: flex;
  gap: 18rpx;
  justify-content: space-between;
}

.entry-title {
  color: var(--sun-ink);
  font-size: 30rpx;
  font-weight: 800;
}

.state-block {
  margin-top: 22rpx;
  text-align: center;
}
</style>
