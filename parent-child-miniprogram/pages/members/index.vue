<template>
  <view class="sun-page members-page">
    <view class="page-header">
      <view>
        <text class="sun-title">成员管理</text>
        <text class="sun-subtitle">{{ familyName }}</text>
      </view>
      <button class="sun-btn secondary header-btn" size="mini" @click="goBack">返回</button>
    </view>

    <view v-if="loading" class="sun-card state-card">
      <text>正在加载成员...</text>
    </view>

    <view v-else class="content">
      <view v-if="isParentRole" class="action-grid">
        <view class="sun-card action-card">
          <view class="action-header">
            <view>
              <text class="action-title">创建虚拟孩子</text>
              <text class="action-subtitle">为暂时没有微信账号的孩子建立积分身份</text>
            </view>
            <button v-if="!showVirtualChildForm" class="sun-btn action-mini" size="mini" @click="openVirtualChildForm">创建</button>
          </view>

          <view v-if="showVirtualChildForm" class="form-panel">
            <view class="form-row">
              <text class="form-label">孩子昵称</text>
              <input v-model.trim="virtualChildForm.nickname" class="form-input" placeholder="例如：小宝" />
            </view>
            <view class="form-actions">
              <button class="sun-btn secondary action-mini" size="mini" :disabled="submittingVirtualChild" @click="cancelVirtualChildForm">取消</button>
              <button class="sun-btn action-mini" size="mini" :disabled="submittingVirtualChild" @click="submitVirtualChild">
                {{ submittingVirtualChild ? '创建中' : '确认创建' }}
              </button>
            </view>
          </view>
        </view>

        <view class="sun-card action-card">
          <view class="action-header">
            <view>
              <text class="action-title">邀请成员</text>
              <text class="action-subtitle">生成带角色的邀请，加入后自动拥有对应身份</text>
            </view>
            <button v-if="!showInviteCreateForm" class="sun-btn action-mini" size="mini" @click="openInviteCreateForm">邀请</button>
          </view>

          <view v-if="showInviteCreateForm" class="form-panel">
            <view class="form-row">
              <text class="form-label">加入角色</text>
              <view class="role-options">
                <button
                  v-for="role in inviteRoleOptions"
                  :key="role.value"
                  class="sun-chip"
                  :class="{ active: inviteCreateForm.targetRole === role.value }"
                  size="mini"
                  @click="inviteCreateForm.targetRole = role.value"
                >
                  {{ role.label }}
                </button>
              </view>
            </view>
            <view class="form-actions">
              <button class="sun-btn secondary action-mini" size="mini" :disabled="submittingInviteCreate" @click="cancelInviteCreateForm">取消</button>
              <button class="sun-btn action-mini" size="mini" :disabled="submittingInviteCreate" @click="submitCreateInvite">
                {{ submittingInviteCreate ? '生成中' : '生成邀请' }}
              </button>
            </view>
          </view>

          <view v-if="latestInvite" class="invite-result">
            <text class="invite-label">邀请码</text>
            <text class="invite-token">{{ latestInvite.token }}</text>
            <text class="invite-expire">有效期至 {{ formatTime(latestInvite.expiresAt) }}</text>
            <button class="sun-btn secondary copy-btn" size="mini" @click="copyInviteToken">复制邀请码</button>
          </view>
        </view>
      </view>

      <view class="sun-card">
        <view class="sun-card-title">
          <text>家庭成员</text>
          <text class="member-count">{{ members.length }} 人</text>
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
              <button
                v-if="isParentRole && member.isVirtual"
                class="sun-btn secondary bind-btn"
                size="mini"
                @click="createBindInvite(member)"
              >
                邀请关联
              </button>
            </view>
          </view>
        </view>

        <view v-if="latestBindInvite" class="invite-result">
          <text class="invite-label">虚拟孩子绑定码</text>
          <text class="invite-token">{{ latestBindInvite.token }}</text>
          <text class="invite-expire">有效期至 {{ formatTime(latestBindInvite.expiresAt) }}</text>
          <button class="sun-btn secondary copy-btn" size="mini" @click="copyBindInviteToken">复制绑定码</button>
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
import { createInvite } from '../../api/invite.js'
import { createVirtualChild, createVirtualChildBindInvite, listMembers } from '../../api/member.js'
import { getCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const currentFamily = ref(null)
const summary = ref({})
const members = ref([])
const showVirtualChildForm = ref(false)
const submittingVirtualChild = ref(false)
const virtualChildForm = ref(defaultVirtualChildForm())
const showInviteCreateForm = ref(false)
const submittingInviteCreate = ref(false)
const inviteCreateForm = ref(defaultInviteCreateForm())
const latestInvite = ref(null)
const latestBindInvite = ref(null)

const familyName = computed(() => summary.value.familyName || currentFamily.value?.familyName || '当前家庭')
const roleType = computed(() => summary.value.roleType || currentFamily.value?.roleType || '')
const isParentRole = computed(() => ['OWNER', 'ADMIN', 'PARENT'].includes(roleType.value))
const inviteRoleOptions = computed(() => {
  if (roleType.value === 'OWNER') {
    return [
      { label: '管理员', value: 'ADMIN' },
      { label: '家长', value: 'PARENT' },
      { label: '孩子', value: 'CHILD' }
    ]
  }
  if (['ADMIN', 'PARENT'].includes(roleType.value)) {
    return [
      { label: '家长', value: 'PARENT' },
      { label: '孩子', value: 'CHILD' }
    ]
  }
  return []
})

onShow(() => {
  loadMembersPage()
})

async function loadMembersPage(retried = false) {
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
    const [summaryData, memberList] = await Promise.all([
      getDashboardSummary(familyId),
      listMembers(familyId)
    ])
    summary.value = summaryData || {}
    members.value = memberList || []
  } catch (error) {
    if (error.statusCode !== 401 || retried) {
      throw error
    }
    await ensureDemoLogin(true)
    await loadMembersPage(true)
  } finally {
    loading.value = false
  }
}

function defaultVirtualChildForm() {
  return {
    nickname: ''
  }
}

function openVirtualChildForm() {
  showVirtualChildForm.value = true
}

function cancelVirtualChildForm() {
  showVirtualChildForm.value = false
  virtualChildForm.value = defaultVirtualChildForm()
}

async function submitVirtualChild() {
  if (submittingVirtualChild.value) {
    return
  }

  const nickname = virtualChildForm.value.nickname.trim()
  if (!nickname) {
    uni.showToast({ title: '请填写孩子昵称', icon: 'none' })
    return
  }

  submittingVirtualChild.value = true
  try {
    const member = await createVirtualChild({
      familyId: currentFamily.value.familyId,
      nickname
    })
    members.value = [member, ...members.value]
    uni.showToast({ title: '虚拟孩子已创建', icon: 'success' })
    showVirtualChildForm.value = false
    virtualChildForm.value = defaultVirtualChildForm()
  } finally {
    submittingVirtualChild.value = false
  }
}

function defaultInviteCreateForm() {
  return {
    targetRole: 'CHILD'
  }
}

function openInviteCreateForm() {
  showInviteCreateForm.value = true
  latestInvite.value = null
  if (!inviteRoleOptions.value.some((role) => role.value === inviteCreateForm.value.targetRole)) {
    inviteCreateForm.value.targetRole = inviteRoleOptions.value[0]?.value || ''
  }
}

function cancelInviteCreateForm() {
  showInviteCreateForm.value = false
  inviteCreateForm.value = defaultInviteCreateForm()
}

async function submitCreateInvite() {
  if (submittingInviteCreate.value) {
    return
  }

  const targetRole = inviteCreateForm.value.targetRole
  if (!inviteRoleOptions.value.some((role) => role.value === targetRole)) {
    uni.showToast({ title: '请选择邀请角色', icon: 'none' })
    return
  }

  submittingInviteCreate.value = true
  try {
    latestInvite.value = await createInvite({
      familyId: currentFamily.value.familyId,
      targetRole
    })
    showInviteCreateForm.value = false
    uni.showToast({ title: '邀请已生成', icon: 'success' })
  } finally {
    submittingInviteCreate.value = false
  }
}

function copyInviteToken() {
  if (!latestInvite.value?.token) {
    return
  }

  uni.setClipboardData({
    data: latestInvite.value.token,
    success() {
      uni.showToast({ title: '邀请码已复制', icon: 'success' })
    },
    fail() {
      uni.showToast({ title: '复制失败，请手动复制', icon: 'none' })
    }
  })
}

async function createBindInvite(member) {
  latestBindInvite.value = await createVirtualChildBindInvite({
    familyId: currentFamily.value.familyId,
    memberId: member.id
  })
  uni.showToast({ title: '绑定码已生成', icon: 'success' })
}

function copyBindInviteToken() {
  if (!latestBindInvite.value?.token) {
    return
  }

  uni.setClipboardData({
    data: latestBindInvite.value.token,
    success() {
      uni.showToast({ title: '绑定码已复制', icon: 'success' })
    },
    fail() {
      uni.showToast({ title: '复制失败，请手动复制', icon: 'none' })
    }
  })
}

function goBack() {
  uni.navigateBack()
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
.members-page {
  box-sizing: border-box;
}

.page-header {
  align-items: flex-start;
  display: flex;
  justify-content: space-between;
  margin-bottom: 24rpx;
}

.header-btn,
.action-mini,
.copy-btn,
.bind-btn {
  margin: 0;
}

.content,
.action-grid,
.form-panel,
.member-list,
.member-main,
.member-side,
.invite-result {
  display: flex;
  flex-direction: column;
}

.content,
.action-grid {
  gap: 22rpx;
}

.action-header,
.member-card {
  align-items: center;
  display: flex;
  justify-content: space-between;
  gap: 18rpx;
}

.action-title,
.action-subtitle,
.member-name,
.member-score-label,
.invite-label,
.invite-expire,
.invite-token {
  display: block;
}

.action-title,
.member-name {
  color: var(--sun-ink);
  font-size: 30rpx;
  font-weight: 800;
}

.action-subtitle,
.member-score-label,
.invite-label,
.invite-expire,
.empty-text,
.member-count {
  color: var(--sun-muted);
  font-size: 24rpx;
  line-height: 1.45;
}

.form-panel {
  gap: 20rpx;
  margin-top: 24rpx;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.form-label {
  color: var(--sun-ink);
  font-size: 24rpx;
  font-weight: 800;
}

.form-input {
  background: #f8fafc;
  border-radius: 18rpx;
  color: var(--sun-ink);
  font-size: 28rpx;
  height: 78rpx;
  padding: 0 22rpx;
}

.form-actions {
  display: flex;
  gap: 16rpx;
  justify-content: flex-end;
}

.role-options {
  display: flex;
  flex-wrap: wrap;
  gap: 14rpx;
}

.member-list {
  gap: 18rpx;
  margin-top: 22rpx;
}

.member-card {
  background: #fffdf8;
  border: 1rpx solid rgba(255, 184, 77, 0.22);
  border-radius: 24rpx;
  padding: 22rpx;
}

.member-main {
  gap: 10rpx;
  min-width: 0;
}

.member-side {
  align-items: flex-end;
  flex-shrink: 0;
  gap: 6rpx;
}

.member-score {
  color: #e9852c;
  font-size: 34rpx;
  font-weight: 800;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
}

.role-tag,
.virtual-tag {
  border-radius: 999rpx;
  font-size: 22rpx;
  font-weight: 700;
  padding: 6rpx 14rpx;
}

.role-tag {
  background: #eef7ff;
  color: var(--sun-sky);
}

.virtual-tag {
  background: #fff2cf;
  color: #c26b18;
}

.invite-result {
  background: #f8fafc;
  border-radius: 20rpx;
  gap: 10rpx;
  margin-top: 22rpx;
  padding: 20rpx;
}

.invite-token {
  color: var(--sun-ink);
  font-size: 26rpx;
  font-weight: 800;
  word-break: break-all;
}

.state-card,
.empty-text {
  text-align: center;
}
</style>
