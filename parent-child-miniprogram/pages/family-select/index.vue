<template>
  <view class="sun-page family-select-page">
    <view class="page-head">
      <view>
        <text class="eyebrow">欢迎回来</text>
        <text class="title">选择家庭</text>
        <text class="subtitle">进入家庭后查看任务、积分和奖励</text>
      </view>
    </view>

    <view class="card hero-card create-family-card">
      <text class="card-heading">创建一个新家庭</text>
      <text class="card-copy">创建人默认昵称为家主，之后可以邀请家人加入。</text>
      <view v-if="showCreateForm" class="form-panel">
        <view class="input-shell">
          <input v-model.trim="familyForm.name" class="form-input hero-input" placeholder="家庭名称" placeholder-class="hero-placeholder" />
        </view>
        <view class="create-actions">
          <text class="create-preview">{{ familyForm.name || '温暖小家' }}</text>
          <button class="btn light" size="mini" :disabled="submittingFamily" @click="submitCreateFamily">
            {{ submittingFamily ? '创建中' : '确认创建' }}
          </button>
        </view>
      </view>
      <button v-else class="btn light create-expand-btn" size="mini" @click="openCreateForm">开始创建</button>
    </view>

    <view class="card">
      <view class="card-title">
        <text>我的家庭</text>
        <text class="card-count">{{ families.length }} 个</text>
      </view>

      <view v-if="loading" class="state-row">
        <text>正在加载家庭...</text>
      </view>
      <view v-else-if="families.length === 0" class="state-row empty-state">
        <text class="empty-title">暂无家庭</text>
        <text class="empty-desc">创建一个家庭，或通过家人发来的邀请加入家庭。</text>
        <button class="btn primary" size="mini" @click="openCreateForm">创建家庭</button>
        <button class="btn light" size="mini" @click="loadFamilies">重新加载</button>
      </view>
      <view v-else class="family-list">
        <view
          class="family-card"
          v-for="family in families"
          :key="family.familyId"
          @click="selectFamily(family)"
        >
          <view class="family-left">
            <view class="avatar">{{ familyInitial(family) }}</view>
            <view class="family-info">
              <text class="row-title">{{ family.familyName }}</text>
              <text class="row-meta">{{ roleName(family.roleType) }} · {{ family.memberCount || 1 }} 位成员</text>
            </view>
          </view>
          <button class="btn blue" size="mini" @click.stop="selectFamily(family)">进入</button>
        </view>
      </view>
    </view>

    <view class="card blue-card invite-card" :class="{ expanded: showInviteForm }">
      <view class="card-title">
        <text>已有邀请码？</text>
        <button class="btn light" size="mini" @click="toggleInviteForm">{{ showInviteForm ? '收起' : '展开' }}</button>
      </view>
      <text v-if="showInviteForm" class="card-copy muted">邀请码作为低权重入口，不占据首屏主体。</text>

      <view v-if="showInviteForm" class="form-panel invite-form">
        <view class="input-shell normal-input">
          <input v-model.trim="inviteForm.token" class="form-input" placeholder="输入家人发来的邀请码" />
        </view>
        <view class="form-actions">
          <view class="form-buttons">
            <button class="btn light" size="mini" :disabled="submittingInvite" @click="cancelAcceptInvite">取消</button>
            <button class="btn primary" size="mini" :disabled="submittingInvite" @click="submitAcceptInvite">
              {{ submittingInvite ? '加入中' : '确认加入' }}
            </button>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'

import { ensureDemoLogin } from '../../api/auth.js'
import { createFamily, listFamilies } from '../../api/family.js'
import { acceptInvite } from '../../api/invite.js'
import { setCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const families = ref([])
const showCreateForm = ref(false)
const showInviteForm = ref(false)
const submittingFamily = ref(false)
const submittingInvite = ref(false)
const familyForm = ref(defaultFamilyForm())
const inviteForm = ref(defaultInviteForm())

onShow(() => {
  loadFamilies()
})

async function loadFamilies() {
  loading.value = true
  try {
    await ensureDemoLogin()
    families.value = (await listFamilies()) || []
  } catch (error) {
    if (error.statusCode !== 401) {
      throw error
    }
    await ensureDemoLogin(true)
    families.value = (await listFamilies()) || []
  } finally {
    loading.value = false
  }
}

function defaultFamilyForm() {
  return {
    name: ''
  }
}

function defaultInviteForm() {
  return {
    token: ''
  }
}

function openCreateForm() {
  showCreateForm.value = true
  showInviteForm.value = false
}

function toggleInviteForm() {
  showInviteForm.value = !showInviteForm.value
  if (showInviteForm.value) {
    showCreateForm.value = false
  }
}

function cancelAcceptInvite() {
  showInviteForm.value = false
  inviteForm.value = defaultInviteForm()
}

function familyFromCreateResponse(response) {
  return {
    familyId: response.family.id,
    familyName: response.family.name,
    memberId: response.member.id,
    roleType: response.member.roleType,
    nickname: response.member.nickname
  }
}

async function submitCreateFamily() {
  if (submittingFamily.value) {
    return
  }

  const name = familyForm.value.name.trim()
  if (!name) {
    uni.showToast({ title: '请填写家庭名称', icon: 'none' })
    return
  }

  submittingFamily.value = true
  try {
    const response = await createFamily({ name, nickname: '家主' })
    const family = familyFromCreateResponse(response)
    setCurrentFamily(family)
    uni.showToast({ title: '家庭已创建', icon: 'success' })
    uni.switchTab({ url: '/pages/index/index' })
  } finally {
    submittingFamily.value = false
  }
}

async function submitAcceptInvite() {
  if (submittingInvite.value) {
    return
  }

  const token = inviteForm.value.token.trim()
  if (!token) {
    uni.showToast({ title: '请填写邀请码', icon: 'none' })
    return
  }

  submittingInvite.value = true
  try {
    const member = await acceptInvite({ token })
    const nextFamilies = (await listFamilies()) || []
    families.value = nextFamilies
    const family = nextFamilies.find((item) => item.familyId === member.familyId)
    if (!family) {
      uni.showToast({ title: '已加入，请重新加载家庭', icon: 'none' })
      return
    }
    setCurrentFamily(family)
    uni.showToast({ title: '已加入家庭', icon: 'success' })
    uni.switchTab({ url: '/pages/index/index' })
  } finally {
    submittingInvite.value = false
  }
}

function selectFamily(family) {
  setCurrentFamily(family)
  uni.switchTab({
    url: '/pages/index/index'
  })
}

function familyInitial(family) {
  return (family.familyName || '家').slice(0, 1)
}

function roleName(roleType) {
  const map = {
    OWNER: '家主',
    ADMIN: '管理员',
    PARENT: '家长',
    CHILD: '孩子'
  }
  return map[roleType] || roleType
}
</script>

<style>
.family-select-page {
  padding-left: 32rpx;
  padding-right: 32rpx;
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
.card-heading,
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

.title {
  color: #172033;
  font-size: 50rpx;
  font-weight: 950;
  line-height: 1.12;
}

.subtitle {
  color: #707887;
  font-size: 24rpx;
  line-height: 1.5;
  margin-top: 12rpx;
}

.card {
  background: rgba(255, 255, 255, 0.9);
  border: 1rpx solid rgba(255, 184, 77, 0.3);
  border-radius: 36rpx;
  box-shadow: 0 24rpx 56rpx rgba(43, 42, 40, 0.08);
  margin-bottom: 24rpx;
  padding: 26rpx;
}

.hero-card {
  background: linear-gradient(135deg, #ffb84d 0%, #ff7a45 68%);
  border-color: rgba(255, 255, 255, 0.36);
  box-shadow: 0 32rpx 64rpx rgba(255, 122, 69, 0.22);
  color: #fff;
}

.blue-card {
  background: linear-gradient(135deg, #eef7ff 0%, #fff 100%);
  border-color: rgba(75, 159, 255, 0.24);
}

.invite-card {
  box-sizing: border-box;
  display: block;
  padding-bottom: 18rpx;
  padding-top: 18rpx;
  width: auto;
}

.invite-card.expanded {
  padding-bottom: 26rpx;
  padding-top: 26rpx;
}

.invite-card .card-title {
  margin-bottom: 0;
}

.invite-card.expanded .card-title {
  margin-bottom: 18rpx;
}

.invite-card .muted {
  font-size: 20rpx;
  line-height: 1.2;
  margin-top: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.invite-card.expanded .muted {
  font-size: 24rpx;
  line-height: 1.5;
  margin-top: 12rpx;
}

.card-title {
  align-items: center;
  color: #172033;
  display: flex;
  font-size: 28rpx;
  font-weight: 950;
  justify-content: space-between;
  margin-bottom: 18rpx;
}

.card-count,
.muted {
  color: #707887;
}

.card-heading {
  color: #fff;
  font-size: 30rpx;
  font-weight: 950;
}

.card-copy {
  color: rgba(255, 255, 255, 0.92);
  font-size: 24rpx;
  line-height: 1.5;
  margin-top: 12rpx;
}

.form-panel {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
  margin-top: 24rpx;
}

.input-shell {
  align-items: center;
  background: rgba(255, 255, 255, 0.22);
  border: 1rpx solid rgba(255, 255, 255, 0.42);
  border-radius: 24rpx;
  display: flex;
  min-height: 78rpx;
  padding: 0 22rpx;
}

.normal-input {
  background: #fff;
  border-color: rgba(75, 159, 255, 0.18);
  box-shadow: inset 0 0 0 1rpx rgba(237, 240, 245, 0.72);
}

.form-input {
  color: #172033;
  flex: 1;
  font-size: 26rpx;
  min-height: 72rpx;
}

.hero-input {
  color: #fff;
  font-weight: 800;
}

.hero-placeholder {
  color: rgba(255, 255, 255, 0.86);
}

.create-actions,
.form-actions {
  align-items: center;
  display: flex;
  gap: 18rpx;
  justify-content: space-between;
}

.invite-form .form-actions {
  justify-content: flex-end;
}

.form-buttons {
  display: flex;
  gap: 14rpx;
  justify-content: flex-end;
}

.create-preview {
  color: #fff;
  font-size: 24rpx;
  font-weight: 800;
}

.family-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.family-card {
  align-items: center;
  background: rgba(255, 255, 255, 0.86);
  border: 1rpx solid #edf0f5;
  border-radius: 30rpx;
  display: flex;
  gap: 20rpx;
  justify-content: space-between;
  min-height: 132rpx;
  padding: 22rpx;
}

.family-left {
  align-items: center;
  display: flex;
  flex: 1;
  gap: 22rpx;
  min-width: 0;
}

.avatar {
  align-items: center;
  background: linear-gradient(135deg, #ffcc69 0%, #ff8f5d 100%);
  border-radius: 50%;
  color: #fff;
  display: flex;
  flex-shrink: 0;
  font-size: 34rpx;
  font-weight: 950;
  height: 84rpx;
  justify-content: center;
  width: 84rpx;
}

.family-info {
  min-width: 0;
}

.row-title {
  color: #172033;
  font-size: 28rpx;
  font-weight: 950;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.row-meta {
  color: #707887;
  font-size: 22rpx;
  line-height: 1.35;
  margin-top: 8rpx;
}

.state-row {
  align-items: center;
  color: #707887;
  display: flex;
  flex-direction: column;
  font-size: 26rpx;
  gap: 16rpx;
  padding: 34rpx 20rpx;
  text-align: center;
}

.empty-title {
  color: #172033;
  font-size: 30rpx;
  font-weight: 900;
}

.empty-desc {
  color: #707887;
  font-size: 24rpx;
  line-height: 1.5;
}

.btn {
  align-items: center;
  border-radius: 999rpx;
  display: inline-flex;
  font-size: 24rpx;
  font-weight: 900;
  justify-content: center;
  line-height: 60rpx;
  min-height: 60rpx;
  min-width: 104rpx;
  padding: 0 24rpx;
}

.family-select-page button {
  min-height: 60rpx;
  border-radius: 999rpx;
  font-size: 24rpx;
  font-weight: 900;
  line-height: 60rpx;
  padding: 0 24rpx;
}

.btn.primary {
  background: #ff7a45;
  color: #fff;
}

.btn.blue {
  background: #4b9fff;
  color: #fff;
}

.btn.light {
  background: #eef7ff;
  border: 1rpx solid rgba(75, 159, 255, 0.18);
  box-shadow: none;
  color: #2f80ed;
}

.create-expand-btn {
  margin-top: 24rpx;
}
</style>
