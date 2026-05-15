<template>
  <view class="container">
    <view class="header">
      <text class="title">选择要进入的家庭</text>
      <text class="subtitle">请选择您当前的身份</text>
    </view>

    <view class="family-list">
      <!-- 模拟数据：家庭列表 -->
      <view 
        class="family-card" 
        v-for="(family, index) in families" 
        :key="index"
        @click="selectFamily(family)"
      >
        <view class="card-left">
          <view class="avatar-placeholder">{{ family.name[0] }}</view>
          <view class="info">
            <text class="family-name">{{ family.name }}</text>
            <view class="role-badge" :class="family.roleType.toLowerCase()">
              <text>{{ getRoleName(family.roleType) }}</text>
            </view>
          </view>
        </view>
        <view class="card-right">
          <text class="enter-btn">进入</text>
        </view>
      </view>

      <!-- 创建新家庭入口 -->
      <view class="create-card" @click="createNewFamily">
        <text class="plus-icon">+</text>
        <text>创建新家庭</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue';

// 模拟数据
const families = ref([
  { id: 1, name: "幸福一家人", roleType: "OWNER", roleName: "爸爸" },
  { id: 2, name: "外婆家", roleType: "PARENT", roleName: "舅舅" },
  { id: 3, name: "测试家庭", roleType: "CHILD", roleName: "大宝" }
]);

const getRoleName = (type) => {
  const map = {
    OWNER: '家主',
    ADMIN: '管理员',
    PARENT: '家长',
    CHILD: '孩子'
  };
  return map[type] || type;
};

const selectFamily = (family) => {
  // 存储 familyId 到本地，并跳转首页
  uni.setStorageSync('currentFamily', family);
  uni.switchTab({
    url: '/pages/index/index'
  });
};

const createNewFamily = () => {
  uni.showToast({ title: '去创建家庭页面', icon: 'none' });
};
</script>

<style>
.container {
  padding: 40px 20px;
}
.header {
  margin-bottom: 40px;
  display: flex;
  flex-direction: column;
}
.title {
  font-size: 24px;
  font-weight: bold;
  color: #2c3e50;
  margin-bottom: 8px;
}
.subtitle {
  font-size: 14px;
  color: #7f8c8d;
}
.family-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.family-card {
  background: #fff;
  border-radius: 16px;
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
}
.card-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.avatar-placeholder {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: #e0f2fe;
  color: #0284c7;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 18px;
}
.info {
  display: flex;
  flex-direction: column;
}
.family-name {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
}
.role-badge {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  width: fit-content;
  display: flex;
}
.role-badge.owner { background: #fef3c7; color: #d97706; }
.role-badge.parent { background: #d1fae5; color: #059669; }
.role-badge.child { background: #fee2e2; color: #dc2626; }

.enter-btn {
  font-size: 14px;
  color: #94a3b8;
}

.create-card {
  border: 2px dashed #cbd5e1;
  border-radius: 16px;
  padding: 20px;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  color: #64748b;
  font-weight: 500;
}
</style>
