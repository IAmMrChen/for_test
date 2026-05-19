<template>
  <view class="container">
    <view class="user-card">
      <view class="avatar">我</view>
      <view class="info">
        <text class="nickname">我的</text>
        <text class="current-family">当前家庭：{{ familyName }}</text>
      </view>
    </view>

    <view class="menu-list">
      <view class="menu-item" @click="switchFamily">
        <view class="menu-left">
          <text class="icon">换</text>
          <text>切换家庭</text>
        </view>
        <text class="arrow">></text>
      </view>

      <view class="menu-item" @click="showComingSoon">
        <view class="menu-left">
          <text class="icon">家</text>
          <text>家庭成员</text>
        </view>
        <text class="arrow">></text>
      </view>

      <view class="menu-item" @click="showComingSoon">
        <view class="menu-left">
          <text class="icon">分</text>
          <text>积分流水</text>
        </view>
        <text class="arrow">></text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'

import { getCurrentFamily } from '../../utils/storage.js'

const currentFamily = ref(null)
const familyName = computed(() => currentFamily.value?.familyName || '未选择家庭')

onShow(() => {
  currentFamily.value = getCurrentFamily()
})

function switchFamily() {
  uni.reLaunch({
    url: '/pages/family-select/index'
  })
}

function showComingSoon() {
  uni.showToast({
    title: '下一轮接入',
    icon: 'none'
  })
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
  gap: 16px;
  margin-bottom: 24px;
  padding: 28px 20px;
}

.avatar {
  align-items: center;
  background: #dbeafe;
  border-radius: 50%;
  color: #2563eb;
  display: flex;
  font-size: 22px;
  font-weight: 700;
  height: 64px;
  justify-content: center;
  width: 64px;
}

.info {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.nickname {
  color: #1f2937;
  font-size: 20px;
  font-weight: 700;
}

.current-family {
  background: #f1f5f9;
  border-radius: 10px;
  color: #64748b;
  font-size: 14px;
  padding: 3px 8px;
}

.menu-list {
  background: #fff;
  border-radius: 12px;
  padding: 0 16px;
}

.menu-item {
  align-items: center;
  border-bottom: 1px solid #f1f5f9;
  display: flex;
  justify-content: space-between;
  padding: 16px 0;
}

.menu-item:last-child {
  border-bottom: none;
}

.menu-left {
  align-items: center;
  color: #333;
  display: flex;
  font-size: 16px;
  gap: 12px;
}

.icon {
  align-items: center;
  background: #eff6ff;
  border-radius: 50%;
  color: #2563eb;
  display: flex;
  font-size: 14px;
  height: 28px;
  justify-content: center;
  width: 28px;
}

.arrow {
  color: #cbd5e1;
}
</style>
