<template>
  <view class="container">
    <view class="user-card">
      <view class="avatar">👨‍👩‍👧‍👦</view>
      <view class="info">
        <text class="nickname">微信用户</text>
        <text class="current-family">当前: {{ currentFamily?.name }}</text>
      </view>
    </view>

    <view class="menu-list">
      <view class="menu-item" @click="switchFamily">
        <view class="menu-left">
          <text class="icon">🏠</text>
          <text>切换家庭</text>
        </view>
        <text class="arrow">></text>
      </view>
      
      <view class="menu-item">
        <view class="menu-left">
          <text class="icon">📝</text>
          <text>家庭日志</text>
        </view>
        <text class="arrow">></text>
      </view>
      
      <view class="menu-item">
        <view class="menu-left">
          <text class="icon">⚙️</text>
          <text>设置</text>
        </view>
        <text class="arrow">></text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue';
import { onShow } from '@dcloudio/uni-app';

const currentFamily = ref(null);

onShow(() => {
  currentFamily.value = uni.getStorageSync('currentFamily') || { name: '未选择' };
});

const switchFamily = () => {
  uni.reLaunch({
    url: '/pages/family-select/index'
  });
};
</script>

<style>
.container {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}
.user-card {
  background: #fff;
  padding: 30px 20px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}
.avatar {
  font-size: 48px;
  background: #f0f9ff;
  width: 80px;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}
.info {
  display: flex;
  flex-direction: column;
}
.nickname {
  font-size: 20px;
  font-weight: bold;
  margin-bottom: 4px;
}
.current-family {
  font-size: 14px;
  color: #64748b;
  background: #f1f5f9;
  padding: 2px 8px;
  border-radius: 10px;
}

.menu-list {
  background: #fff;
  border-radius: 16px;
  padding: 0 16px;
}
.menu-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  border-bottom: 1px solid #f1f5f9;
}
.menu-item:last-child {
  border-bottom: none;
}
.menu-left {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 16px;
  color: #333;
}
.icon {
  font-size: 20px;
}
.arrow {
  color: #cbd5e1;
}
</style>
