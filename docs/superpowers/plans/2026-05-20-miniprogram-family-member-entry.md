# 小程序家庭与成员入口 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在小程序补齐创建家庭、接受邀请、创建虚拟孩子、生成角色邀请的 MVP 入口。

**Architecture:** 家庭选择页负责家庭外路径：创建家庭、输入邀请码加入家庭。个人页负责家庭内成员动作：创建虚拟孩子、创建角色邀请。服务端接口已存在，本计划只新增小程序 API 封装和页面交互，不改 Go 后端业务逻辑。

**Tech Stack:** uni-app、Vue 3 `<script setup>`、现有 `utils/request.js`、现有 Go 服务端接口。

---

## 文件结构

- Modify: `parent-child-miniprogram/api/family.js`
  - 增加 `createFamily(data)`。
- Modify: `parent-child-miniprogram/api/member.js`
  - 增加 `createVirtualChild(data)`。
- Create: `parent-child-miniprogram/api/invite.js`
  - 增加 `createInvite(data)` 和 `acceptInvite(data)`。
- Modify: `parent-child-miniprogram/pages/family-select/index.vue`
  - 增加创建家庭表单。
  - 增加输入邀请码加入家庭表单。
  - 创建或加入成功后写入当前家庭并跳首页。
- Modify: `parent-child-miniprogram/pages/profile/index.vue`
  - 家长角色增加创建虚拟孩子表单。
  - 家长角色增加生成邀请表单。
  - 生成邀请后展示 token、过期时间和复制按钮。

## 行为边界

- `OWNER`、`ADMIN`、`PARENT` 可以创建虚拟孩子和创建邀请。
- `OWNER` 可以邀请 `ADMIN`、`PARENT`、`CHILD`。
- `ADMIN` 和 `PARENT` 可以邀请 `PARENT`、`CHILD`。
- `CHILD` 不显示成员管理表单。
- 创建家庭时昵称可以为空，空值交给服务端使用默认昵称。
- 接受邀请后重新拉取家庭列表，用返回成员的 `familyId` 定位完整家庭信息。
- 本次不做成员移除、角色变更、邀请取消、邀请列表、虚拟孩子绑定真实账号。

---

### Task 1: 增加家庭与邀请 API 封装

**Files:**
- Modify: `parent-child-miniprogram/api/family.js`
- Modify: `parent-child-miniprogram/api/member.js`
- Create: `parent-child-miniprogram/api/invite.js`

- [ ] **Step 1: 修改家庭 API**

在 `parent-child-miniprogram/api/family.js` 末尾加入：

```js
export function createFamily(data) {
  return request({
    url: '/api/family/create',
    method: 'POST',
    data
  })
}
```

- [ ] **Step 2: 修改成员 API**

在 `parent-child-miniprogram/api/member.js` 末尾加入：

```js
export function createVirtualChild(data) {
  return request({
    url: '/api/member/createVirtualChild',
    method: 'POST',
    data
  })
}
```

- [ ] **Step 3: 新增邀请 API**

创建 `parent-child-miniprogram/api/invite.js`：

```js
import { request } from '../utils/request.js'

export function createInvite(data) {
  return request({
    url: '/api/invite/create',
    method: 'POST',
    data
  })
}

export function acceptInvite(data) {
  return request({
    url: '/api/invite/accept',
    method: 'POST',
    data
  })
}
```

- [ ] **Step 4: 运行 JS 语法检查**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
```

Expected: 命令退出码为 `0`，无 `SyntaxError`。

- [ ] **Step 5: 提交 API 封装**

Run:

```powershell
git add parent-child-miniprogram\api\family.js parent-child-miniprogram\api\member.js parent-child-miniprogram\api\invite.js
git commit -m "feat: 增加小程序家庭邀请接口"
```

---

### Task 2: 家庭选择页增加创建与加入入口

**Files:**
- Modify: `parent-child-miniprogram/pages/family-select/index.vue`

- [ ] **Step 1: 修改导入**

把家庭 API 导入改成：

```js
import { createFamily, listFamilies } from '../../api/family.js'
import { acceptInvite } from '../../api/invite.js'
```

保留：

```js
import { ensureDemoLogin } from '../../api/auth.js'
import { setCurrentFamily } from '../../utils/storage.js'
```

- [ ] **Step 2: 增加页面状态**

在 `families` 状态后加入：

```js
const showCreateForm = ref(false)
const showInviteForm = ref(false)
const submittingFamily = ref(false)
const submittingInvite = ref(false)
const familyForm = ref(defaultFamilyForm())
const inviteForm = ref(defaultInviteForm())
```

- [ ] **Step 3: 调整空家庭提示和新增操作区模板**

把空家庭态中的说明文案：

```vue
<text class="empty-desc">请先在后端创建家庭，或接受家人发来的邀请。</text>
```

改成：

```vue
<text class="empty-desc">创建一个家庭，或输入家人发来的邀请码加入家庭。</text>
```

在 `header` 区块之后、`loading` 区块之前插入：

```vue
    <view class="entry-actions">
      <button class="entry-btn primary" size="mini" @click="toggleCreateForm">创建家庭</button>
      <button class="entry-btn plain" size="mini" @click="toggleInviteForm">输入邀请码</button>
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
```

- [ ] **Step 4: 增加表单函数**

在 `selectFamily(family)` 函数之前加入：

```js
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
```

- [ ] **Step 5: 增加样式**

在 `<style>` 中追加：

```css
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
```

- [ ] **Step 6: 运行页面检查**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\family-select\index.vue
```

Expected: 退出码为 `1`，没有乱码特征命中。

- [ ] **Step 7: 提交家庭选择页**

Run:

```powershell
git add parent-child-miniprogram\pages\family-select\index.vue
git commit -m "feat: 增加家庭创建与加入入口"
```

---

### Task 3: 个人页增加虚拟孩子入口

**Files:**
- Modify: `parent-child-miniprogram/pages/profile/index.vue`

- [ ] **Step 1: 修改成员 API 导入**

把成员 API 导入改成：

```js
import { createVirtualChild, listMembers } from '../../api/member.js'
```

- [ ] **Step 2: 增加虚拟孩子表单状态**

在现有 `pointLogs` 状态后加入：

```js
const showVirtualChildForm = ref(false)
const submittingVirtualChild = ref(false)
const virtualChildForm = ref(defaultVirtualChildForm())
```

确认页面已有或新增：

```js
const isParentRole = computed(() => ['OWNER', 'ADMIN', 'PARENT'].includes(roleType.value))
```

- [ ] **Step 3: 在成员区顶部插入表单**

在“家庭成员”标题下方、成员空态之前插入：

```vue
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
```

- [ ] **Step 4: 增加虚拟孩子函数**

在 `loadProfile(retried = false)` 函数之后加入：

```js
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
```

- [ ] **Step 5: 增加样式**

在 `<style>` 中追加：

```css
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
```

- [ ] **Step 6: 运行页面检查**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\profile\index.vue
```

Expected: 退出码为 `1`，没有乱码特征命中。

- [ ] **Step 7: 提交虚拟孩子入口**

Run:

```powershell
git add parent-child-miniprogram\pages\profile\index.vue
git commit -m "feat: 增加虚拟孩子创建入口"
```

---

### Task 4: 个人页增加角色邀请入口

**Files:**
- Modify: `parent-child-miniprogram/pages/profile/index.vue`

- [ ] **Step 1: 增加邀请 API 导入**

在 API 导入区加入：

```js
import { createInvite } from '../../api/invite.js'
```

- [ ] **Step 2: 增加邀请状态**

在虚拟孩子表单状态后加入：

```js
const showInviteCreateForm = ref(false)
const submittingInviteCreate = ref(false)
const inviteCreateForm = ref(defaultInviteCreateForm())
const latestInvite = ref(null)
```

在计算属性区加入：

```js
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
```

- [ ] **Step 3: 在成员区插入邀请表单**

在虚拟孩子 `manage-panel` 之后插入：

```vue
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
```

- [ ] **Step 4: 增加邀请函数**

在 `submitVirtualChild()` 后加入：

```js
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
```

- [ ] **Step 5: 增加邀请样式**

在 `<style>` 中追加：

```css
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
```

- [ ] **Step 6: 运行页面检查**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\profile\index.vue
```

Expected: 退出码为 `1`，没有乱码特征命中。

- [ ] **Step 7: 提交邀请入口**

Run:

```powershell
git add parent-child-miniprogram\pages\profile\index.vue
git commit -m "feat: 增加家庭成员邀请入口"
```

---

### Task 5: 集成验证

**Files:**
- Verify: `parent-child-miniprogram/api/family.js`
- Verify: `parent-child-miniprogram/api/member.js`
- Verify: `parent-child-miniprogram/api/invite.js`
- Verify: `parent-child-miniprogram/pages/family-select/index.vue`
- Verify: `parent-child-miniprogram/pages/profile/index.vue`
- Verify: `parent-child-api`

- [ ] **Step 1: 运行前端 JS 语法检查**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
```

Expected: 命令退出码为 `0`，无 `SyntaxError`。

- [ ] **Step 2: 扫描页面乱码特征**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\family-select\index.vue parent-child-miniprogram\pages\profile\index.vue
```

Expected: 退出码为 `1`，没有乱码特征命中。

- [ ] **Step 3: 运行服务端测试**

Run:

```powershell
go test ./... -count=1
```

Working directory: `parent-child-api`

Expected: 所有 package 显示 `ok` 或 `?`，命令退出码为 `0`。如果沙箱因为 Go build cache 无权限失败，使用同一命令申请提升权限后重跑。

- [ ] **Step 4: 人工路径验证**

在微信开发者工具或 H5 预览中验证：

1. 没有家庭的用户进入家庭选择页，可以看到“创建家庭”和“输入邀请码”。
2. 创建家庭时家庭名称为空，会提示“请填写家庭名称”。
3. 填写家庭名称后创建成功，自动进入首页。
4. 家主进入个人页，可以创建虚拟孩子，成员列表刷新并展示“虚拟账号”。
5. 家主可以生成管理员、家长、孩子邀请码。
6. 家长或管理员只能生成家长、孩子邀请码。
7. 孩子进入个人页看不到创建虚拟孩子和邀请成员入口。
8. 另一个微信用户输入有效邀请码后加入家庭，自动进入首页。
9. 同一个微信用户重复接受同一家庭的邀请码时，前端展示服务端错误提示。

- [ ] **Step 5: 提交验证修正**

如果集成验证产生修正，按实际修改文件提交：

```powershell
git add parent-child-miniprogram\api\family.js parent-child-miniprogram\api\member.js parent-child-miniprogram\api\invite.js parent-child-miniprogram\pages\family-select\index.vue parent-child-miniprogram\pages\profile\index.vue
git commit -m "fix: 修正家庭成员入口"
```

如果没有产生修正，不创建空提交。

---

## 自检结果

- 设计规格中的创建家庭、接受邀请、创建虚拟孩子、生成角色邀请、角色可见性、基础校验和提交禁用均在 Task 1 至 Task 5 覆盖。
- 本计划不修改服务端模型，不新增小程序路由，不实现成员移除、角色变更、邀请取消或虚拟孩子绑定。
- 字段名与现有服务端保持一致：`name`、`nickname`、`familyId`、`targetRole`、`token`。
- 验证命令覆盖小程序 API 语法、页面乱码扫描和服务端测试。
