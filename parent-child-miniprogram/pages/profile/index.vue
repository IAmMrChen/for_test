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
        <view v-if="isParentRole" class="manage-panel">
          <view class="manage-header">
            <view>
              <text class="manage-title">虚拟孩子</text>
              <text class="manage-subtitle">先为没有微信账号的孩子建立积分身份</text>
            </view>
            <button v-if="!showVirtualChildForm" class="small-primary-btn" size="mini" @click="openVirtualChildForm">创建</button>
          </view>

          <view v-if="showVirtualChildForm" class="manage-form">
            <view class="form-row">
              <text class="form-label">孩子昵称</text>
              <input v-model.trim="virtualChildForm.nickname" class="form-input" placeholder="例如：小宝" />
            </view>
            <view class="form-actions">
              <button class="plain-action-btn" size="mini" :disabled="submittingVirtualChild" @click="cancelVirtualChildForm">取消</button>
              <button class="primary-action-btn" size="mini" :disabled="submittingVirtualChild" @click="submitVirtualChild">
                {{ submittingVirtualChild ? '创建中' : '确认创建' }}
              </button>
            </view>
          </view>
        </view>

        <view v-if="isParentRole" class="manage-panel">
          <view class="manage-header">
            <view>
              <text class="manage-title">邀请成员</text>
              <text class="manage-subtitle">创建一个带角色的邀请码</text>
            </view>
            <button v-if="!showInviteCreateForm" class="small-primary-btn" size="mini" @click="openInviteCreateForm">邀请</button>
          </view>

          <view v-if="showInviteCreateForm" class="manage-form">
            <view class="form-row">
              <text class="form-label">加入角色</text>
              <view class="role-options">
                <button
                  v-for="role in inviteRoleOptions"
                  :key="role.value"
                  class="role-option-btn"
                  :class="{ active: inviteCreateForm.targetRole === role.value }"
                  size="mini"
                  @click="inviteCreateForm.targetRole = role.value"
                >
                  {{ role.label }}
                </button>
              </view>
            </view>
            <view class="form-actions">
              <button class="plain-action-btn" size="mini" :disabled="submittingInviteCreate" @click="cancelInviteCreateForm">取消</button>
              <button class="primary-action-btn" size="mini" :disabled="submittingInviteCreate" @click="submitCreateInvite">
                {{ submittingInviteCreate ? '生成中' : '生成邀请码' }}
              </button>
            </view>
          </view>

          <view v-if="latestInvite" class="invite-result">
            <text class="invite-label">邀请码</text>
            <text class="invite-token">{{ latestInvite.token }}</text>
            <text class="invite-expire">有效期至 {{ formatTime(latestInvite.expiresAt) }}</text>
            <button class="copy-btn" size="mini" @click="copyInviteToken">复制邀请码</button>
          </view>
        </view>

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
              <button
                v-if="isParentRole && member.isVirtual"
                class="link-action-btn"
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
          <button class="copy-btn" size="mini" @click="copyBindInviteToken">复制绑定码</button>
        </view>
      </view>

      <view class="section">
        <view class="entry-card" @click="goPointLogs">
          <view>
            <text class="entry-title">积分流水</text>
            <text class="entry-subtitle">查看积分收入、兑换和调整记录</text>
          </view>
          <text class="entry-arrow">进入</text>
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

const familyName = computed(() => summary.value.familyName || currentFamily.value?.familyName || '未选择家庭')
const nickname = computed(() => summary.value.nickname || '我的')
const roleType = computed(() => summary.value.roleType || currentFamily.value?.roleType || '')
const isParentRole = computed(() => ['OWNER', 'ADMIN', 'PARENT'].includes(roleType.value))
const currentPoints = computed(() => summary.value.currentPoints || 0)
const totalEarnedPoints = computed(() => summary.value.totalEarnedPoints || 0)
const avatarText = computed(() => (nickname.value || '我').slice(0, 1))
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
    await loadProfile(true)
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
    await createVirtualChild({
      familyId: currentFamily.value.familyId,
      nickname
    })
    uni.showToast({ title: '虚拟孩子已创建', icon: 'success' })
    showVirtualChildForm.value = false
    virtualChildForm.value = defaultVirtualChildForm()
    await loadProfile()
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
    uni.showToast({ title: '邀请码已生成', icon: 'success' })
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

function switchFamily() {
  uni.reLaunch({
    url: '/pages/family-select/index'
  })
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
.entry-subtitle {
  color: #64748b;
  font-size: 12px;
}

.content,
.section,
.member-list {
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

.member-list {
  gap: 12px;
}

.member-card,
.entry-card,
.state-block {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
}

.member-card,
.entry-card {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.member-main {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.member-name,
.entry-title {
  color: #111827;
  font-size: 15px;
  font-weight: 700;
}

.entry-title,
.entry-subtitle {
  display: block;
}

.entry-arrow {
  color: #94a3b8;
  font-size: 13px;
}

.tag-row {
  align-items: center;
  display: flex;
  gap: 6px;
}

.role-tag,
.virtual-tag {
  border-radius: 4px;
  font-size: 11px;
  padding: 2px 6px;
}

.role-tag {
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

.state-block {
  color: #64748b;
  font-size: 14px;
  text-align: center;
}

.state-block.compact {
  padding: 20px 16px;
}

.manage-panel {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
}

.manage-header {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.manage-title,
.manage-subtitle {
  display: block;
}

.manage-title {
  color: #111827;
  font-size: 15px;
  font-weight: 700;
}

.manage-subtitle {
  color: #64748b;
  font-size: 12px;
  margin-top: 4px;
}

.manage-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-top: 16px;
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

.role-options {
  display: flex;
  gap: 8px;
}

.role-option-btn {
  background: #f8fafc;
  border: 1px solid #dbe3ef;
  border-radius: 16px;
  color: #475569;
  font-size: 12px;
  line-height: 30px;
  margin: 0;
  padding: 0 14px;
}

.role-option-btn.active {
  background: #eff6ff;
  border-color: #2563eb;
  color: #2563eb;
}

.invite-result {
  background: #f8fafc;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 14px;
  padding: 12px;
}

.invite-label,
.invite-expire {
  color: #64748b;
  font-size: 12px;
}

.invite-token {
  color: #111827;
  font-size: 13px;
  font-weight: 700;
  word-break: break-all;
}

.copy-btn {
  align-self: flex-start;
  background: #fff;
  border: 1px solid #cbd5e1;
  border-radius: 16px;
  color: #475569;
  font-size: 12px;
  line-height: 30px;
  margin: 4px 0 0;
  padding: 0 14px;
}

.link-action-btn {
  background: #eff6ff;
  border: none;
  border-radius: 14px;
  color: #2563eb;
  font-size: 11px;
  line-height: 26px;
  margin: 6px 0 0;
  padding: 0 10px;
}
</style>
