<template>
  <view class="container">
    <view class="header">
      <text class="title">奖励</text>
      <text class="subtitle">完成任务后，用积分兑换喜欢的奖励</text>
    </view>

    <view class="empty-state">
      <text class="empty-title">奖励列表将在下一轮接入</text>
      <text class="empty-desc">后续会支持查看奖励、申请兑换和确认领取。</text>
    </view>
  </view>
</template>

<style>
.container {
  min-height: 100vh;
  padding: 48px 20px 24px;
  background: #f5f7fa;
  text-align: center;
}

.header {
  display: flex;
  flex-direction: column;
  margin-bottom: 36px;
}

.title {
  color: #1f2937;
  display: block;
  font-size: 26px;
  font-weight: 700;
  margin-bottom: 8px;
}

.subtitle {
  color: #64748b;
  font-size: 14px;
}

.empty-state {
  background: #fff;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 40px 20px;
}

.empty-title {
  color: #1f2937;
  font-size: 16px;
  font-weight: 700;
}

.empty-desc {
  color: #64748b;
  font-size: 14px;
  line-height: 20px;
}
</style>
