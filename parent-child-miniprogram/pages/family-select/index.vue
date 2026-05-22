<template>
  <view class="container">
    <view class="header">
      <text class="title">选择家庭</text>
      <text class="subtitle">进入一个家庭后，就可以查看任务、积分和奖励</text>
    </view>

    <view class="entry-actions">
      <button class="entry-btn primary" size="mini" @click="toggleCreateForm">创建家庭</button>
      <button class="entry-btn plain" size="mini" @click="toggleInviteForm">输入邀请码</button>
      <button class="entry-btn plain" size="mini" @click="toggleBindForm">关联孩子</button>
    </view>

    <view v-if="showCreateForm" class="entry-panel">
      <view class="form-row">
        <text class="form-label">家庭名称</text>
        <input v-model.trim="familyForm.name" class="form-input" placeholder="例如：陈家" />
      </view>
      <view class="form-row">
        <text class="form-label">我的昵称</text>
        <input v-model.trim="familyForm.nickname" class="form-input" placeholder="例如：爸爸" />
      </view>
      <view class="form-actions">
        <button class="plain-action-btn" size="mini" :disabled="submittingFamily" @click="cancelCreateFamily">取消</button>
        <button class="primary-action-btn" size="mini" :disabled="submittingFamily" @click="submitCreateFamily">
          {{ submittingFamily ? '创建中' : '确认创建' }}
        </button>
      </view>
    </view>

    <view v-if="showBindForm" class="entry-panel">
      <view class="form-row">
        <text class="form-label">虚拟孩子绑定码</text>
        <input v-model.trim="bindForm.token" class="form-input" placeholder="输入家长发来的绑定码" />
      </view>
      <view class="form-actions">
        <button class="plain-action-btn" size="mini" :disabled="submittingBind" @click="cancelAcceptBind">取消</button>
        <button class="primary-action-btn" size="mini" :disabled="submittingBind" @click="submitAcceptBind">
          {{ submittingBind ? '绑定中' : '确认绑定' }}
        </button>
      </view>
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

    <view v-else-if="families.length === 0" class="state-block">
      <text class="empty-title">暂无家庭</text>
      <text class="empty-desc">创建一个家庭，或输入家人发来的邀请码加入家庭。</text>
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
import { acceptVirtualChildBindInvite } from '../../api/member.js'
import { setCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const families = ref([])
const showCreateForm = ref(false)
const showInviteForm = ref(false)
const showBindForm = ref(false)
const submittingFamily = ref(false)
const submittingInvite = ref(false)
const submittingBind = ref(false)
const familyForm = ref(defaultFamilyForm())
const inviteForm = ref(defaultInviteForm())
const bindForm = ref(defaultBindForm())

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
    name: '',
    nickname: ''
  }
}

function defaultInviteForm() {
  return {
    token: ''
  }
}

function defaultBindForm() {
  return {
    token: ''
  }
}

function toggleCreateForm() {
  showCreateForm.value = !showCreateForm.value
  if (showCreateForm.value) {
    showInviteForm.value = false
    showBindForm.value = false
  }
}

function toggleInviteForm() {
  showInviteForm.value = !showInviteForm.value
  if (showInviteForm.value) {
    showCreateForm.value = false
    showBindForm.value = false
  }
}

function toggleBindForm() {
  showBindForm.value = !showBindForm.value
  if (showBindForm.value) {
    showCreateForm.value = false
    showInviteForm.value = false
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

function cancelAcceptBind() {
  showBindForm.value = false
  bindForm.value = defaultBindForm()
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
  const nickname = familyForm.value.nickname.trim()
  if (!name) {
    uni.showToast({ title: '请填写家庭名称', icon: 'none' })
    return
  }

  submittingFamily.value = true
  try {
    const response = await createFamily({ name, nickname })
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

async function submitAcceptBind() {
  if (submittingBind.value) {
    return
  }

  const token = bindForm.value.token.trim()
  if (!token) {
    uni.showToast({ title: '请填写绑定码', icon: 'none' })
    return
  }

  submittingBind.value = true
  try {
    const member = await acceptVirtualChildBindInvite({ token })
    const nextFamilies = (await listFamilies()) || []
    families.value = nextFamilies
    const family = nextFamilies.find((item) => item.familyId === member.familyId)
    if (!family) {
      uni.showToast({ title: '已绑定，请重新加载家庭', icon: 'none' })
      return
    }
    setCurrentFamily(family)
    uni.showToast({ title: '已绑定孩子账号', icon: 'success' })
    uni.switchTab({ url: '/pages/index/index' })
  } finally {
    submittingBind.value = false
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
  display: flex;
  flex-direction: column;
  margin-bottom: 28px;
}

.entry-actions {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.entry-btn {
  border-radius: 16px;
  font-size: 12px;
  line-height: 32px;
  margin: 0;
  padding: 0 16px;
}

.entry-btn.primary,
.primary-action-btn {
  background: #2563eb;
  border: none;
  color: #fff;
}

.entry-btn.plain,
.plain-action-btn {
  background: #fff;
  border: 1px solid #cbd5e1;
  color: #475569;
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

.empty-title {
  color: #1f2937;
  font-size: 18px;
  font-weight: 700;
}

.empty-desc {
  font-size: 14px;
  line-height: 20px;
}

.retry-btn {
  background: #3b82f6;
  border: none;
  color: #fff;
  margin-top: 8px;
}
</style>
