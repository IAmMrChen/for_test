<template>
  <view class="container">
    <!-- 顶部导航栏模拟 -->
    <view class="navbar">
      <view class="current-family">
        <text class="greeting">Hi, {{ currentRole === 'CHILD' ? '大宝' : '爸爸' }} 👋</text>
        <text class="family-name">{{ currentFamily?.name || '幸福一家人' }}</text>
      </view>
      <view class="points-card" v-if="isChild">
        <text class="points-label">当前积分</text>
        <view class="points-value">350 <text class="unit">分</text></view>
      </view>
    </view>

    <!-- 角色视角切换 (开发调试用) -->
    <view class="debug-toggle">
      <button size="mini" @click="toggleRole">切换视角: {{ isChild ? '孩子' : '家长' }}</button>
    </view>

    <!-- 孩子视角：任务列表 -->
    <view v-if="isChild" class="content-area">
      <view class="section-title">今日任务</view>
      <view class="task-list">
        <view class="task-item" v-for="task in childTasks" :key="task.id">
          <view class="task-info">
            <view class="task-title">{{ task.title }}</view>
            <view class="task-reward">+{{ task.points }} 积分</view>
          </view>
          <button 
            class="action-btn" 
            :class="task.status"
            @click="submitTask(task)"
          >
            {{ task.status === 'pending' ? '去完成' : (task.status === 'reviewing' ? '审核中' : '已完成') }}
          </button>
        </view>
      </view>
    </view>

    <!-- 家长视角：待办看板 -->
    <view v-else class="content-area">
      <view class="dashboard-stats">
        <view class="stat-item">
          <view class="stat-num text-orange">3</view>
          <view class="stat-desc">待审核</view>
        </view>
        <view class="stat-item">
          <view class="stat-num text-green">1</view>
          <view class="stat-desc">待发奖</view>
        </view>
      </view>

      <view class="section-title">待办事项</view>
      <view class="audit-list">
        <view class="audit-item" v-for="audit in pendingAudits" :key="audit.id">
          <view class="audit-header">
            <text class="child-name">{{ audit.childName }}</text>
            <text class="time">{{ audit.time }}</text>
          </view>
          <view class="audit-content">
            <text>完成了任务：</text><text style="font-weight: bold;">{{ audit.taskName }}</text>
          </view>
          <view class="audit-actions">
            <button class="btn-reject" size="mini" type="warn" plain>驳回</button>
            <button class="btn-approve" size="mini" type="primary">通过</button>
          </view>
        </view>
      </view>
    </view>

  </view>
</template>

<script setup>
import { ref, computed } from 'vue';
import { onShow } from '@dcloudio/uni-app';

const currentFamily = ref(null);
const currentRole = ref('CHILD'); // CHILD or PARENT

onShow(() => {
  const stored = uni.getStorageSync('currentFamily');
  if (stored) {
    currentFamily.value = stored;
    currentRole.value = stored.roleType === 'CHILD' ? 'CHILD' : 'PARENT';
  }
});

const isChild = computed(() => currentRole.value === 'CHILD');

const toggleRole = () => {
  currentRole.value = currentRole.value === 'CHILD' ? 'PARENT' : 'CHILD';
};

// 模拟数据
const childTasks = ref([
  { id: 1, title: '按时刷牙', points: 5, status: 'pending' },
  { id: 2, title: '整理书包', points: 10, status: 'reviewing' },
  { id: 3, title: '阅读30分钟', points: 20, status: 'finished' }
]);

const pendingAudits = ref([
  { id: 101, childName: '大宝', taskName: '扫地', time: '10:30' },
  { id: 102, childName: '大宝', taskName: '背古诗', time: 'Yesterday' }
]);

const submitTask = (task) => {
  if (task.status !== 'pending') return;
  uni.showToast({ title: '已提交，等爸爸妈妈审核', icon: 'success' });
  task.status = 'reviewing';
};
</script>

<style>
.container {
  min-height: 100vh;
  background-color: #f5f7fa;
}

/* 导航栏 */
.navbar {
  background: #fff;
  padding: 44px 20px 20px 20px; /* 适配异形屏大致高度 */
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}
.current-family {
  display: flex;
  flex-direction: column;
}
.greeting {
  font-size: 20px;
  font-weight: bold;
  color: #333;
}
.family-name {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 2px;
}

.points-card {
  text-align: right;
  display: flex;
  flex-direction: column;
}
.points-label {
  font-size: 10px;
  color: #999;
}
.points-value {
  font-size: 20px;
  font-weight: bold;
  color: #f59e0b;
}
.unit {
  font-size: 12px;
}

.debug-toggle {
  padding: 10px;
  text-align: center;
  background: #ffe4e6;
  margin-bottom: 10px;
}

.content-area {
  padding: 20px;
}
.section-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
  color: #333;
}

/* 孩子任务列表 */
.task-item {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.task-title {
  font-size: 16px;
  font-weight: 500;
}
.task-reward {
  font-size: 12px;
  color: #f59e0b;
  margin-top: 4px;
}
.action-btn {
  font-size: 12px;
  padding: 0 16px;
  height: 32px;
  line-height: 32px;
  border-radius: 20px;
  background: #3b82f6;
  color: #fff;
  border: none;
  margin: 0;
}
.action-btn.reviewing {
  background: #fbbf24;
}
.action-btn.finished {
  background: #e2e8f0;
  color: #94a3b8;
}

/* 家长看板 */
.dashboard-stats {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
}
.stat-item {
  flex: 1;
  background: #fff;
  padding: 16px;
  border-radius: 12px;
  text-align: center;
}
.stat-num {
  font-size: 24px;
  font-weight: bold;
}
.text-orange { color: #f97316; }
.text-green { color: #10b981; }
.stat-desc {
  font-size: 12px;
  color: #64748b;
  margin-top: 4px;
}

.audit-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.audit-item {
  background: #fff;
  padding: 16px;
  border-radius: 12px;
}
.audit-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 12px;
  color: #64748b;
}
.child-name {
  font-weight: bold;
  color: #333;
}
.audit-content {
  margin-bottom: 12px;
  font-size: 14px;
}
.audit-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}
</style>
