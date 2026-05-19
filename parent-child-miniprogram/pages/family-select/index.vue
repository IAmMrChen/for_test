<template>
  <view class="container">
    <view class="header">
      <text class="title">选择家庭</text>
      <text class="subtitle">进入一个家庭后，就可以查看任务、积分和奖励</text>
    </view>

    <view v-if="loading" class="state-block">
      <text>正在加载家庭...</text>
    </view>

    <view v-else-if="families.length === 0" class="state-block">
      <text class="empty-title">暂无家庭</text>
      <text class="empty-desc">请先在后端创建家庭，或接受家人发来的邀请。</text>
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
import { listFamilies } from '../../api/family.js'
import { setCurrentFamily } from '../../utils/storage.js'

const loading = ref(false)
const families = ref([])

onShow(() => {
  loadFamilies()
})

async function loadFamilies() {
  loading.value = true
  try {
    await ensureDemoLogin()
    families.value = await listFamilies()
  } finally {
    loading.value = false
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
