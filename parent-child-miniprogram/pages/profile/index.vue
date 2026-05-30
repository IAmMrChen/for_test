<template>
  <view class="sun-page profile-page">
    <view class="page-head">
      <view>
        <text class="eyebrow">账户中心</text>
        <text class="title">我的</text>
        <text class="subtitle">家庭、成员和积分记录</text>
      </view>
    </view>

    <view class="head-actions">
      <button class="btn light" size="mini" @click="switchFamily">切换家庭</button>
    </view>

    <view class="card hero-card user-card">
      <view class="family-text">
        <text class="section-name">{{ familyName }}</text>
        <text class="metric-main">{{ currentPoints }}</text>
        <text class="metric-sub">当前积分 · 累计获得 {{ totalEarnedPoints }}</text>
      </view>
      <view class="avatar">{{ avatarText }}</view>
    </view>

    <view v-if="loading" class="card state-card">
      <text>正在加载个人页...</text>
    </view>

    <view v-else class="profile-content">
      <view class="card blue-card entry-card" @click="goPointLogs">
        <view class="entry-main">
          <text class="entry-title">积分流水</text>
          <text class="entry-subtitle">查看积分收入、兑换和调整记录。</text>
        </view>
        <button class="btn blue" size="mini" @click.stop="goPointLogs">查看</button>
      </view>

      <view v-if="isParentRole" class="card entry-card" @click="goMembers">
        <view class="member-section">
          <view class="member-section-head">
            <view>
              <text class="entry-title">家庭成员</text>
              <text class="entry-subtitle">管理孩子、邀请家人和关联虚拟孩子</text>
            </view>
            <button class="btn light" size="mini" @click.stop="goMembers">管理</button>
          </view>

          <view v-if="members.length === 0" class="empty-text">
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
              <view class="member-side">
                <text class="member-score">{{ member.currentPoints || 0 }}</text>
                <text class="member-score-label">当前积分</text>
              </view>
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
import { getCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const currentFamily = ref(null)
const summary = ref({})
const members = ref([])

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
    const summaryData = (await getDashboardSummary(family.familyId)) || {}
    summary.value = summaryData
    if (['OWNER', 'ADMIN', 'PARENT'].includes(summaryData.roleType || family.roleType)) {
      members.value = (await listMembers(family.familyId)) || []
    } else {
      members.value = []
    }
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
  padding-left: 40rpx;
  padding-right: 40rpx;
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
.section-name,
.metric-main,
.metric-sub,
.entry-title,
.entry-subtitle {
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
  line-height: 1.12;
}

.subtitle,
.entry-subtitle {
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
  align-items: center;
  background: linear-gradient(135deg, #ffb84d 0%, #ff7a45 66%);
  border-color: rgba(255, 255, 255, 0.36);
  color: #fff;
  display: flex;
  justify-content: space-between;
}

.blue-card {
  background: linear-gradient(135deg, #eef7ff 0%, #fff 100%);
  border-color: rgba(75, 159, 255, 0.24);
}

.green-card {
  background: linear-gradient(135deg, #effbf3 0%, #fff 100%);
  border-color: rgba(99, 199, 132, 0.26);
}

.section-name,
.metric-main,
.metric-sub {
  color: #fff;
}

.section-name {
  font-size: 28rpx;
  font-weight: 950;
}

.metric-main {
  font-size: 62rpx;
  font-weight: 950;
  line-height: 1;
  margin-top: 14rpx;
}

.metric-sub {
  font-size: 24rpx;
  line-height: 1.35;
  margin-top: 14rpx;
}

.avatar {
  align-items: center;
  background: rgba(255, 255, 255, 0.22);
  border-radius: 50%;
  color: #fff;
  display: flex;
  flex-shrink: 0;
  font-size: 34rpx;
  font-weight: 950;
  height: 96rpx;
  justify-content: center;
  width: 96rpx;
}

.profile-content {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
  margin-top: 24rpx;
}

.entry-card {
  display: block;
}

.entry-main {
  flex: 1;
  min-width: 0;
}

.entry-title {
  color: #172033;
  font-size: 28rpx;
  font-weight: 950;
}

.member-section {
  width: 100%;
}

.member-section-head {
  align-items: center;
  display: flex;
  gap: 16rpx;
  justify-content: space-between;
}

.member-list {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
  margin-top: 22rpx;
}

.member-card {
  align-items: center;
  background: rgba(255, 255, 255, 0.84);
  border: 1rpx solid #edf0f5;
  border-radius: 28rpx;
  display: flex;
  gap: 18rpx;
  justify-content: space-between;
  padding: 22rpx;
}

.member-main {
  min-width: 0;
}

.member-name,
.member-score,
.member-score-label {
  display: block;
}

.member-name {
  color: #172033;
  font-size: 28rpx;
  font-weight: 950;
}

.member-side {
  align-items: flex-end;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  gap: 6rpx;
}

.member-score {
  color: #e9852c;
  font-size: 34rpx;
  font-weight: 950;
  line-height: 1.1;
}

.member-score-label,
.empty-text {
  color: #707887;
  font-size: 24rpx;
  line-height: 1.45;
}

.empty-text {
  margin-top: 22rpx;
  text-align: center;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 10rpx;
}

.role-tag,
.virtual-tag {
  border-radius: 999rpx;
  font-size: 22rpx;
  font-weight: 800;
  padding: 6rpx 14rpx;
}

.role-tag {
  background: #eef7ff;
  color: #4b9fff;
}

.virtual-tag {
  background: #fff2cf;
  color: #c26b18;
}

.state-card {
  color: #7b8190;
  font-size: 26rpx;
  text-align: center;
}

.profile-page button,
.profile-page .btn {
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

.profile-page .btn.blue {
  background: #4b9fff;
  box-shadow: 0 16rpx 32rpx rgba(75, 159, 255, 0.18);
}

.profile-page .btn.light {
  background: #eef7ff;
  border: 1rpx solid rgba(75, 159, 255, 0.18);
  box-shadow: none;
  color: #2f80ed;
}
</style>
