<template>
  <view class="sun-page family-select-page">
    <view class="header">
      <view class="header-copy">
        <text class="title">选择家庭</text>
        <text class="subtitle">进入一个家庭后，就可以查看任务、积分和奖励</text>
      </view>
      <button class="create-top-btn" size="mini" @click="openCreateForm">创建</button>
    </view>

    <view class="create-family-card">
      <view class="create-copy">
        <text class="create-card-title">创建一个新家庭</text>
        <text class="create-card-desc">创建人默认昵称为家主，之后可以邀请家人加入。</text>
      </view>
      <view v-if="showCreateForm" class="create-form">
        <view class="form-row">
          <text class="form-label">家庭名称</text>
          <input v-model.trim="familyForm.name" class="form-input create-input" placeholder="例如：温暖小家" />
        </view>
        <view class="form-actions">
          <button class="plain-action-btn create-cancel-btn" size="mini" :disabled="submittingFamily" @click="cancelCreateFamily">取消</button>
          <button class="primary-action-btn create-confirm-btn" size="mini" :disabled="submittingFamily" @click="submitCreateFamily">
            {{ submittingFamily ? '创建中' : '确认创建' }}
          </button>
        </view>
      </view>
      <button v-else class="primary-action-btn create-expand-btn" size="mini" @click="openCreateForm">开始创建</button>
    </view>

    <view class="invite-entry">
      <button class="text-action-btn" size="mini" @click="toggleInviteForm">
        {{ showInviteForm ? '收起邀请码' : '已有邀请码？' }}
      </button>
    </view>

    <view v-if="showInviteForm" class="entry-panel">
      <view class="form-row">
        <text class="form-label">邀请码</text>
        <input v-model.trim="inviteForm.token" class="form-input" placeholder="输入家人发来的邀请码" />
      </view>
      <view class="form-actions">
        <button class="plain-action-btn" size="mini" :disabled="submittingInvite" @click="cancelAcceptInvite">取消</button>
        <button class="primary-action-btn" size="mini" :disabled="submittingInvite" @click="submitAcceptInvite">
          {{ submittingInvite ? '加入中' : '确认加入' }}
        </button>
      </view>
    </view>

    <view v-if="loading" class="state-block">
      <text>正在加载家庭...</text>
    </view>

    <view v-else-if="families.length === 0" class="state-block empty-state">
      <text class="empty-title">暂无家庭</text>
      <text class="empty-desc">创建一个家庭，或通过家人发来的邀请加入家庭。</text>
      <button class="empty-create-btn" size="mini" @click="openCreateForm">创建家庭</button>
      <button class="retry-btn" size="mini" @click="loadFamilies">重新加载</button>
    </view>

    <view v-else class="family-list">
      <view
        class="family-card"
        v-for="family in families"
        :key="family.familyId"
        @click="selectFamily(family)"
      >
        <view class="card-left">
          <view class="avatar-placeholder">{{ familyInitial(family) }}</view>
          <view class="info">
            <text class="family-name">{{ family.familyName }}</text>
            <view class="role-badge" :class="family.roleType.toLowerCase()">
              <text>{{ roleName(family.roleType) }}</text>
            </view>
          </view>
        </view>
        <text class="enter-btn">进入</text>
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

function toggleCreateForm() {
  showCreateForm.value = !showCreateForm.value
  if (showCreateForm.value) {
    showInviteForm.value = false
  }
}

function toggleInviteForm() {
  showInviteForm.value = !showInviteForm.value
  if (showInviteForm.value) {
    showCreateForm.value = false
  }
}

function cancelCreateFamily() {
  showCreateForm.value = false
  familyForm.value = defaultFamilyForm()
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
.container {
  min-height: 100vh;
  padding: 48px 20px 24px;
  background: #f5f7fa;
}

.header {
  align-items: flex-start;
  display: flex;
  gap: 16px;
  justify-content: space-between;
  margin-bottom: 20px;
}

.header-copy {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
}

.title {
  color: #1f2937;
  font-size: 26px;
  font-weight: 700;
  margin-bottom: 8px;
}

.subtitle {
  color: #64748b;
  font-size: 14px;
  line-height: 20px;
}

.create-top-btn,
.empty-create-btn,
.primary-action-btn {
  background: #2563eb;
  border: none;
  color: #fff;
}

.create-top-btn {
  border-radius: 16px;
  flex-shrink: 0;
  font-size: 12px;
  line-height: 32px;
  margin: 2px 0 0;
  padding: 0 14px;
}

.invite-entry {
  display: flex;
  justify-content: flex-end;
  margin: -4px 0 14px;
}

.text-action-btn {
  background: transparent;
  border: none;
  color: #2563eb;
  font-size: 13px;
  line-height: 28px;
  margin: 0;
  padding: 0;
}

.entry-panel {
  background: #fff;
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-bottom: 16px;
  padding: 16px;
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

.family-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.family-card {
  align-items: center;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.06);
  display: flex;
  justify-content: space-between;
  padding: 18px;
}

.card-left {
  align-items: center;
  display: flex;
  gap: 12px;
}

.avatar-placeholder {
  align-items: center;
  background: #dbeafe;
  border-radius: 50%;
  color: #2563eb;
  display: flex;
  font-size: 18px;
  font-weight: 700;
  height: 48px;
  justify-content: center;
  width: 48px;
}

.info {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.family-name {
  color: #111827;
  font-size: 16px;
  font-weight: 600;
}

.role-badge {
  border-radius: 4px;
  display: flex;
  font-size: 12px;
  padding: 2px 8px;
  width: fit-content;
}

.role-badge.owner {
  background: #fef3c7;
  color: #b45309;
}

.role-badge.admin,
.role-badge.parent {
  background: #dcfce7;
  color: #047857;
}

.role-badge.child {
  background: #fee2e2;
  color: #dc2626;
}

.enter-btn {
  color: #94a3b8;
  font-size: 14px;
}

.state-block {
  align-items: center;
  background: #fff;
  border-radius: 12px;
  color: #64748b;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 36px 20px;
  text-align: center;
}

.empty-state {
  margin-top: 4px;
}

.empty-title {
  color: #1f2937;
  font-size: 18px;
  font-weight: 700;
}

.empty-desc {
  font-size: 14px;
  line-height: 20px;
}

.empty-create-btn,
.retry-btn {
  border-radius: 16px;
  font-size: 12px;
  line-height: 32px;
  margin-top: 8px;
  padding: 0 18px;
}

.retry-btn {
  background: #fff;
  border: 1px solid #cbd5e1;
  color: #475569;
  margin-top: 0;
}

.family-select-page .header {
  margin-bottom: 24rpx;
}

.family-select-page .title {
  color: var(--sun-ink);
  font-size: 46rpx;
  font-weight: 800;
}

.family-select-page .subtitle {
  color: var(--sun-muted);
  font-size: 24rpx;
  line-height: 1.55;
}

.family-select-page .entry-panel,
.family-select-page .family-card,
.family-select-page .state-block {
  background: rgba(255, 255, 255, 0.9);
  border: 1rpx solid rgba(255, 184, 77, 0.28);
  border-radius: 28rpx;
  box-shadow: 0 14rpx 34rpx rgba(43, 42, 40, 0.08);
}

.family-select-page .create-family-card {
  background: linear-gradient(135deg, #ffb84d 0%, #ff7a45 68%);
  border-radius: 30rpx;
  box-shadow: 0 18rpx 38rpx rgba(255, 122, 69, 0.24);
  color: #fff;
  margin-bottom: 24rpx;
  padding: 28rpx;
}

.family-select-page .create-copy,
.family-select-page .create-card-title,
.family-select-page .create-card-desc {
  display: block;
}

.family-select-page .create-card-title {
  font-size: 30rpx;
  font-weight: 900;
}

.family-select-page .create-card-desc {
  font-size: 24rpx;
  line-height: 1.5;
  margin-top: 12rpx;
  opacity: 0.92;
}

.family-select-page .create-form {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
  margin-top: 22rpx;
}

.family-select-page .create-input {
  background: rgba(255, 255, 255, 0.24);
  border: 1rpx solid rgba(255, 255, 255, 0.42);
  color: #fff;
}

.family-select-page .create-cancel-btn,
.family-select-page .create-confirm-btn,
.family-select-page .create-expand-btn {
  background: #eef7ff;
  color: var(--sun-sky);
  box-shadow: none;
}

.family-select-page .create-top-btn,
.family-select-page .empty-create-btn,
.family-select-page .primary-action-btn {
  background: var(--sun-action);
  border-radius: 999rpx;
}

.family-select-page .plain-action-btn,
.family-select-page .retry-btn {
  background: #eef7ff;
  border-color: rgba(75, 159, 255, 0.22);
  border-radius: 999rpx;
  color: var(--sun-sky);
}

.family-select-page .avatar-placeholder {
  background: linear-gradient(135deg, #ffcf6c 0%, #ff8f5d 100%);
  color: #fff;
}

.family-select-page .enter-btn,
.family-select-page .text-action-btn {
  color: var(--sun-sky);
}

.family-select-page button {
  min-height: 64rpx;
  border-radius: 999rpx;
  font-size: 24rpx;
  font-weight: 800;
  line-height: 64rpx;
  padding: 0 26rpx;
}

.family-select-page .create-family-card .create-cancel-btn,
.family-select-page .create-family-card .create-confirm-btn,
.family-select-page .create-family-card .create-expand-btn {
  background: #eef7ff;
  color: var(--sun-sky);
  box-shadow: none;
}
</style>
