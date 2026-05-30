<template>
  <view class="sun-page rewards-page">
    <view class="page-head">
      <view>
        <text class="eyebrow">奖励商店</text>
        <text class="title">奖励</text>
        <text class="subtitle">兑换、发放和维护奖励库</text>
      </view>
    </view>

    <view v-if="loading" class="card state-card">
      <text>正在加载奖励...</text>
    </view>

    <view v-else class="rewards-content">
      <view v-if="isParentRole" class="card blue-card proxy-entry" @click="goRewardProxy">
        <view class="proxy-entry-head">
          <text>代孩子兑换奖励</text>
          <button class="btn blue" size="mini" @click.stop="goRewardProxy">进入</button>
        </view>
        <text class="card-copy">奖励较多时进入独立页处理，先选孩子，再选奖励。</text>
      </view>

      <view v-if="isChild" class="card hero-card points-card">
        <text class="section-name">{{ familyName }}</text>
        <text class="metric-main">{{ currentPoints }}</text>
        <text class="metric-sub">当前积分，可用于兑换奖励。</text>
      </view>

      <view v-if="isChild && deliveredRecords.length > 0" class="card blue-card">
        <view class="card-title">
          <text>待确认收到</text>
          <text>{{ deliveredRecords.length }} 项</text>
        </view>
        <view class="list">
          <view class="row" v-for="record in deliveredRecords" :key="record.id">
            <view class="row-main">
              <text class="row-title">{{ record.rewardName }}</text>
              <text class="row-meta">消耗 {{ record.pointsCost }} 积分</text>
            </view>
            <button class="btn green" size="mini" @click="receive(record)">确认</button>
          </view>
        </view>
      </view>

      <view v-if="isParentRole" class="card">
        <view class="card-title">
          <text>待发放申请</text>
          <text>{{ appliedRecords.length }} 项</text>
        </view>
        <view v-if="appliedRecords.length === 0" class="empty-row">
          <text>当前没有待发放奖励</text>
        </view>
        <view v-else class="list">
          <view class="row" v-for="record in appliedRecords" :key="record.id">
            <view class="row-main">
              <text class="row-title">{{ record.nickname || '孩子' }}申请{{ record.rewardName }}</text>
              <text class="row-meta">{{ record.pointsCost }} 积分 · {{ formatTime(record.applyTime) }}</text>
            </view>
            <view class="row-actions">
              <button class="btn primary" size="mini" @click="deliver(record)">发放</button>
              <button class="btn danger" size="mini" @click="reject(record)">驳回</button>
            </view>
          </view>
        </view>
      </view>

      <view class="card reward-library-card">
        <view class="card-title">
          <view>
            <text>奖励库</text>
            <text v-if="!isParentRole" class="card-subtitle">可兑换奖励</text>
          </view>
          <button v-if="isParentRole && !showRewardForm" class="btn primary" size="mini" @click="openRewardForm">新建</button>
        </view>

        <view v-if="rewards.length === 0" class="empty-row">
          <text>暂无可兑换奖励</text>
        </view>
        <view v-else class="list">
          <view class="row reward-row" v-for="reward in rewardRows" :key="reward.id">
            <view class="row-main">
              <text class="row-title">{{ reward.name }}</text>
              <text class="row-meta">{{ reward.pointsCost }} 积分 · {{ stockText(reward.stock) }}</text>
            </view>
            <view v-if="isChild" class="row-actions">
              <button
                class="btn blue exchange-btn"
                size="mini"
                :class="reward.actionStatus"
                :disabled="reward.actionStatus !== 'available'"
                @click="apply(reward)"
              >
                {{ rewardButtonText(reward) }}
              </button>
            </view>
            <view v-if="isParentRole" class="row-actions">
              <button class="btn light" size="mini" @click="editReward(reward)">编辑</button>
              <button class="btn danger" size="mini" @click="offShelfExistingReward(reward)">下架</button>
            </view>
          </view>
        </view>
      </view>

      <view v-if="isParentRole && showRewardForm" class="card hero-card create-panel">
        <view class="card-title">
          <text>{{ editingRewardId ? '编辑奖励' : '新建奖励表单' }}</text>
        </view>
        <view class="create-form">
          <view class="form-row">
            <text class="form-label">奖励名称</text>
            <view class="input-shell">
              <input v-model.trim="rewardForm.name" class="form-input" placeholder="例如：周末电影票" placeholder-class="hero-placeholder" />
            </view>
          </view>
          <view class="form-row">
            <text class="form-label">所需积分</text>
            <view class="input-shell">
              <input v-model="rewardForm.pointsCost" class="form-input" type="number" placeholder="30" placeholder-class="hero-placeholder" />
            </view>
          </view>
          <view class="form-row">
            <text class="form-label">库存</text>
            <view class="input-shell">
              <input v-model="rewardForm.stock" class="form-input" type="number" placeholder="0" placeholder-class="hero-placeholder" />
            </view>
          </view>
          <view class="form-actions">
            <text class="form-hint">库存填 0 表示不限库存</text>
            <view class="form-buttons">
              <button class="btn light" size="mini" :disabled="submittingReward" @click="cancelRewardForm">取消</button>
              <button class="btn light" size="mini" :disabled="submittingReward" @click="createNewReward">
                {{ submittingReward ? '保存中' : '确认保存' }}
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
  applyReward,
  createReward,
  deliverReward,
  listRewardRecords,
  listRewards,
  offShelfReward,
  receiveReward,
  rejectReward,
  updateReward
} from '../../api/reward.js'
import { getCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const currentFamily = ref(null)
const summary = ref({})
const rewards = ref([])
const appliedRecords = ref([])
const deliveredRecords = ref([])
const showRewardForm = ref(false)
const submittingReward = ref(false)
const rewardForm = ref(defaultRewardForm())
const editingRewardId = ref(0)

const roleType = computed(() => summary.value.roleType || currentFamily.value?.roleType || '')
const isChild = computed(() => roleType.value === 'CHILD')
const isParentRole = computed(() => ['OWNER', 'ADMIN', 'PARENT'].includes(roleType.value))
const familyName = computed(() => summary.value.familyName || currentFamily.value?.familyName || '当前家庭')
const currentPoints = computed(() => summary.value.currentPoints || 0)

const rewardRows = computed(() => {
  return rewards.value.map((reward) => {
    const pending = appliedRecords.value.find((record) => record.rewardId === reward.id)
    if (pending) {
      return { ...reward, actionStatus: 'applied' }
    }
    if (reward.stock === 0) {
      return { ...reward, actionStatus: 'soldout' }
    }
    if (currentPoints.value < reward.pointsCost) {
      return { ...reward, actionStatus: 'insufficient' }
    }
    return { ...reward, actionStatus: 'available' }
  })
})

onShow(() => {
  loadRewardPage()
})

async function loadRewardPage(retried = false) {
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
    const [summaryData, rewardList] = await Promise.all([
      getDashboardSummary(familyId),
      listRewards(familyId)
    ])

    summary.value = summaryData || {}
    rewards.value = rewardList || []

    if (summary.value.roleType === 'CHILD') {
      const [applied, delivered] = await Promise.all([
        listRewardRecords({ familyId, status: 'APPLIED' }),
        listRewardRecords({ familyId, status: 'DELIVERED' })
      ])
      appliedRecords.value = applied || []
      deliveredRecords.value = delivered || []
      return
    }

    if (isParentRole.value) {
      const applied = await listRewardRecords({ familyId, status: 'APPLIED' })
      appliedRecords.value = applied || []
      deliveredRecords.value = []
      return
    }

    appliedRecords.value = []
    deliveredRecords.value = []
  } catch (error) {
    if (error.statusCode !== 401 || retried) {
      throw error
    }
    await ensureDemoLogin(true)
    await loadRewardPage(true)
  } finally {
    loading.value = false
  }
}

async function apply(reward) {
  const record = await applyReward({
    familyId: currentFamily.value.familyId,
    rewardId: reward.id
  })
  addAppliedRewardRecord(record, reward)
  uni.showToast({ title: '已申请兑换', icon: 'success' })
}

async function deliver(record) {
  await deliverReward({ recordId: record.id })
  removeAppliedRewardRecord(record.id)
  uni.showToast({ title: '已发放', icon: 'success' })
}

async function reject(record) {
  await rejectReward({ recordId: record.id })
  removeAppliedRewardRecord(record.id)
  uni.showToast({ title: '已驳回', icon: 'success' })
}

async function receive(record) {
  await receiveReward({ recordId: record.id })
  removeDeliveredRewardRecord(record.id)
  uni.showToast({ title: '已确认收到', icon: 'success' })
}

function defaultRewardForm() {
  return {
    name: '',
    pointsCost: 30,
    stock: 0
  }
}

function resetRewardForm() {
  rewardForm.value = defaultRewardForm()
  editingRewardId.value = 0
}

function openRewardForm() {
  editingRewardId.value = 0
  rewardForm.value = defaultRewardForm()
  showRewardForm.value = true
}

function cancelRewardForm() {
  showRewardForm.value = false
  resetRewardForm()
}

function normalizePositiveInteger(value) {
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) ? parsed : 0
}

function normalizeNonNegativeInteger(value) {
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) ? parsed : -1
}

function validateRewardForm() {
  const name = rewardForm.value.name.trim()
  const pointsCost = normalizePositiveInteger(rewardForm.value.pointsCost)
  const stock = normalizeNonNegativeInteger(rewardForm.value.stock)

  if (!name) {
    uni.showToast({ title: '请填写奖励名称', icon: 'none' })
    return null
  }
  if (pointsCost <= 0) {
    uni.showToast({ title: '所需积分必须大于 0', icon: 'none' })
    return null
  }
  if (stock < 0) {
    uni.showToast({ title: '库存不能小于 0', icon: 'none' })
    return null
  }

  return {
    name,
    pointsCost,
    stock
  }
}

async function createNewReward() {
  if (submittingReward.value) {
    return
  }

  const payload = validateRewardForm()
  if (!payload) {
    return
  }

  submittingReward.value = true
  try {
    const familyId = currentFamily.value.familyId
    if (editingRewardId.value) {
      await updateReward({
        ...payload,
        familyId,
        rewardId: editingRewardId.value
      })
      uni.showToast({ title: '奖励已更新', icon: 'success' })
    } else {
      await createReward({
        ...payload,
        familyId
      })
      uni.showToast({ title: '奖励已创建', icon: 'success' })
    }
    showRewardForm.value = false
    resetRewardForm()
    await loadRewardPage()
  } finally {
    submittingReward.value = false
  }
}

function editReward(reward) {
  editingRewardId.value = reward.id
  rewardForm.value = {
    name: reward.name,
    pointsCost: reward.pointsCost,
    stock: reward.stock < 0 ? 0 : reward.stock
  }
  showRewardForm.value = true
}

async function offShelfExistingReward(reward) {
  const confirmed = await new Promise((resolve) => {
    uni.showModal({
      title: '确认下架',
      content: `下架后孩子将不能再兑换“${reward.name}”`,
      success: (res) => resolve(res.confirm)
    })
  })
  if (!confirmed) {
    return
  }

  await offShelfReward({
    familyId: currentFamily.value.familyId,
    rewardId: reward.id
  })
  uni.showToast({ title: '奖励已下架', icon: 'success' })
  await loadRewardPage()
}

function addAppliedRewardRecord(record, reward, child = null) {
  const nextRecord = {
    ...record,
    rewardId: reward.id,
    rewardName: reward.name,
    pointsCost: reward.pointsCost,
    memberId: child?.id || record.memberId,
    nickname: child?.nickname || record.nickname
  }
  appliedRecords.value = [
    nextRecord,
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
  if (!child) {
    summary.value = {
      ...summary.value,
      currentPoints: Math.max((summary.value.currentPoints || 0) - reward.pointsCost, 0)
    }
  }
}

function removeAppliedRewardRecord(recordId) {
  appliedRecords.value = appliedRecords.value.filter((item) => item.id !== recordId)
}

function removeDeliveredRewardRecord(recordId) {
  deliveredRecords.value = deliveredRecords.value.filter((item) => item.id !== recordId)
}

function rewardButtonText(reward) {
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

function goRewardProxy() {
  uni.navigateTo({ url: '/pages/reward-proxy/index' })
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
.record-meta,
.reward-stock {
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
.reward-list {
  display: flex;
  flex-direction: column;
}

.content,
.section {
  gap: 18px;
}

.section-heading,
.proxy-entry {
  align-items: center;
  display: flex;
  justify-content: space-between;
  gap: 14px;
}

.section-title {
  color: #1f2937;
  font-size: 16px;
  font-weight: 700;
}

.section-subtitle,
.proxy-entry-subtitle {
  color: #64748b;
  display: block;
  font-size: 12px;
  margin-top: 4px;
}

.reward-list,
.record-card + .record-card {
  gap: 12px;
}

.reward-card,
.record-card,
.state-block {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
}

.reward-card,
.record-card {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.reward-info,
.reward-side,
.record-main {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.reward-side {
  align-items: flex-end;
}

.reward-name,
.record-title {
  color: #111827;
  font-size: 15px;
  font-weight: 700;
}

.points {
  color: #f59e0b;
  font-size: 13px;
  font-weight: 700;
}

.record-actions {
  display: flex;
  gap: 8px;
}

.manage-actions {
  display: flex;
  gap: 8px;
}

.exchange-btn,
.primary-btn,
.plain-btn {
  border-radius: 16px;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  min-width: 72px;
  padding: 0 12px;
}

.exchange-btn,
.primary-btn {
  background: #3b82f6;
  border: none;
  color: #fff;
}

.plain-btn {
  background: #fff;
  border: 1px solid #cbd5e1;
  color: #475569;
}

.exchange-btn.applied,
.exchange-btn.insufficient,
.exchange-btn.soldout {
  background: #e5e7eb;
  color: #64748b;
}

.state-block {
  color: #64748b;
  font-size: 14px;
  text-align: center;
}

.state-block.compact {
  padding: 20px 16px;
}

.proxy-entry {
  background: linear-gradient(135deg, #fff3d8 0%, #fff 58%);
  border: 1px solid rgba(245, 158, 11, 0.22);
  border-radius: 14px;
  padding: 16px;
}

.proxy-entry-title {
  color: #111827;
  display: block;
  font-size: 16px;
  font-weight: 700;
}

.create-panel {
  background: #fff;
  border-radius: 10px;
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

.create-subtitle,
.form-hint {
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

.danger-action-btn {
  background: #fff;
  border: 1px solid #fecaca;
  border-radius: 16px;
  color: #dc2626;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  min-width: 64px;
  padding: 0 12px;
}

.rewards-page .header {
  margin-bottom: 28rpx;
}

.rewards-page .title {
  color: #172033;
  font-size: 46rpx;
  font-weight: 800;
}

.rewards-page .subtitle,
.rewards-page .summary-label,
.rewards-page .record-meta,
.rewards-page .reward-stock,
.rewards-page .section-subtitle,
.rewards-page .proxy-entry-subtitle {
  color: #707887;
}

.rewards-page .summary-pill,
.rewards-page .reward-card,
.rewards-page .record-card,
.rewards-page .state-block,
.rewards-page .create-panel,
.rewards-page .proxy-entry {
  background: rgba(255, 255, 255, 0.9);
  border: 1rpx solid rgba(255, 184, 77, 0.24);
  border-radius: 28rpx;
  box-shadow: 0 14rpx 34rpx rgba(43, 42, 40, 0.08);
}

.rewards-page .proxy-entry {
  background: linear-gradient(135deg, #fff3d8 0%, #ffffff 62%);
}

.rewards-page .summary-value,
.rewards-page .points {
  color: #e9852c;
}

.rewards-page .exchange-btn,
.rewards-page .primary-btn,
.rewards-page .small-primary-btn,
.rewards-page .primary-action-btn {
  background: #ff7a45;
  border-radius: 999rpx;
}

.rewards-page .plain-btn,
.rewards-page .plain-action-btn {
  background: #eef7ff;
  border-color: rgba(75, 159, 255, 0.22);
  border-radius: 999rpx;
  color: #4b9fff;
}

.rewards-page button {
  min-height: 64rpx;
  border-radius: 999rpx;
  font-size: 24rpx;
  font-weight: 800;
  line-height: 64rpx;
  padding: 0 26rpx;
}

.rewards-page {
  padding-left: 40rpx;
  padding-right: 40rpx;
}

.rewards-content {
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
.card-subtitle,
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

.rewards-page .title {
  color: #172033;
  font-size: 50rpx;
  font-weight: 950;
  letter-spacing: 0;
  line-height: 1.12;
  margin-bottom: 0;
}

.rewards-page .subtitle {
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

.blue-card {
  background: linear-gradient(135deg, #eef7ff 0%, #fff 100%);
  border-color: rgba(75, 159, 255, 0.24);
}

.rewards-page .proxy-entry {
  background: linear-gradient(135deg, #eef7ff 0%, #fff 100%);
  border: 1rpx solid rgba(75, 159, 255, 0.34);
  border-radius: 28rpx;
  box-shadow: 0 18rpx 42rpx rgba(75, 159, 255, 0.08);
  padding: 26rpx 28rpx;
}

.proxy-entry-head {
  align-items: center;
  color: #172033;
  display: flex;
  font-size: 28rpx;
  font-weight: 950;
  gap: 20rpx;
  justify-content: space-between;
  line-height: 1.25;
}

.proxy-entry-head > text {
  flex: 1;
  min-width: 0;
}

.rewards-page .proxy-entry .card-copy {
  color: #707887;
  font-size: 24rpx;
  line-height: 1.45;
  margin: 14rpx 0 0;
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

.card-copy,
.card-subtitle {
  color: #707887;
  font-size: 24rpx;
  line-height: 1.45;
}

.card-copy {
  margin-top: -4rpx;
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

.list,
.create-form {
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
  background: transparent;
  color: #fff;
  flex: 1;
  font-size: 24rpx;
  min-height: 76rpx;
  padding: 0;
}

.hero-placeholder {
  color: rgba(255, 255, 255, 0.72);
}

.form-actions {
  align-items: center;
  display: flex;
  gap: 16rpx;
  justify-content: space-between;
}

.form-hint {
  color: rgba(255, 255, 255, 0.9);
  flex: 1;
  font-size: 22rpx;
  line-height: 1.4;
  margin-top: 0;
}

.form-buttons {
  display: flex;
  flex: 0 0 auto;
  gap: 12rpx;
}

.rewards-page button,
.rewards-page .btn {
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

.rewards-page .btn.primary {
  background: #ff7a45;
}

.rewards-page .btn.blue {
  background: #4b9fff;
  box-shadow: 0 16rpx 32rpx rgba(75, 159, 255, 0.18);
}

.rewards-page .btn.green {
  background: #63c784;
  box-shadow: 0 16rpx 32rpx rgba(99, 199, 132, 0.18);
}

.rewards-page .btn.light {
  background: #eef7ff;
  border: 1rpx solid rgba(75, 159, 255, 0.18);
  box-shadow: none;
  color: #2f80ed;
}

.rewards-page .btn.danger {
  background: #fff1f3;
  border: 1rpx solid rgba(239, 107, 122, 0.2);
  box-shadow: none;
  color: #d95061;
}

.rewards-page .exchange-btn.insufficient,
.rewards-page .exchange-btn.soldout,
.rewards-page .exchange-btn.applied {
  background: #eef2f7;
  box-shadow: none;
  color: #8a95a5;
}

.rewards-page .create-panel.hero-card {
  background: linear-gradient(135deg, #ffb84d 0%, #ff7a45 66%);
  border-color: rgba(255, 255, 255, 0.36);
  box-shadow: 0 32rpx 64rpx rgba(255, 122, 69, 0.22);
}

.rewards-page .create-panel.hero-card .card-title,
.rewards-page .create-panel.hero-card .form-label,
.rewards-page .create-panel.hero-card .form-hint {
  color: #fff;
}

.rewards-page .create-panel .input-shell {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.36);
}

.rewards-page .create-panel .form-input {
  background: transparent;
  color: #fff;
}

</style>
