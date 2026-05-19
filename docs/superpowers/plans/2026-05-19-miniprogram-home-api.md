# Miniprogram Home API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让小程序具备统一 API 请求基础层，并将家庭选择页和首页从 mock 数据切换为后端真实数据。

**Architecture:** 小程序侧新增 `config`、`utils`、`api` 三层：`config` 管理本地后端地址和 demo 用户，`utils/request.js` 统一处理 token、请求和错误，`api/*.js` 封装后端接口。页面只负责展示、状态合成和调用 API，不直接拼接 URL 或读写 token。

**Tech Stack:** uni-app Vue3、现有 `uni.request`/`uni.storage` API、后端 JWT demoLogin、Go 后端现有接口、Node 语法检查。

---

## 文件结构

- Create: `parent-child-miniprogram/config/env.js`
- Create: `parent-child-miniprogram/utils/storage.js`
- Create: `parent-child-miniprogram/utils/request.js`
- Create: `parent-child-miniprogram/api/auth.js`
- Create: `parent-child-miniprogram/api/family.js`
- Create: `parent-child-miniprogram/api/dashboard.js`
- Create: `parent-child-miniprogram/api/task.js`
- Modify: `parent-child-miniprogram/pages.json`
- Modify: `parent-child-miniprogram/pages/family-select/index.vue`
- Modify: `parent-child-miniprogram/pages/index/index.vue`
- Modify: `parent-child-miniprogram/pages/rewards/index.vue`
- Modify: `parent-child-miniprogram/pages/profile/index.vue`

---

### Task 1: 前端请求与 API 基础层

**Files:**
- Create: `parent-child-miniprogram/config/env.js`
- Create: `parent-child-miniprogram/utils/storage.js`
- Create: `parent-child-miniprogram/utils/request.js`
- Create: `parent-child-miniprogram/api/auth.js`
- Create: `parent-child-miniprogram/api/family.js`
- Create: `parent-child-miniprogram/api/dashboard.js`
- Create: `parent-child-miniprogram/api/task.js`

- [ ] **Step 1: 写请求层和 API 模块**

新增 `parent-child-miniprogram/config/env.js`：

```js
export const API_BASE_URL = 'http://127.0.0.1:8080'

export const DEMO_USER = {
  userId: 1001,
  nickname: '演示用户'
}
```

新增 `parent-child-miniprogram/utils/storage.js`：

```js
const TOKEN_KEY = 'authToken'
const CURRENT_USER_KEY = 'currentUser'
const CURRENT_FAMILY_KEY = 'currentFamily'

export function getToken() {
  return uni.getStorageSync(TOKEN_KEY) || ''
}

export function setToken(token) {
  uni.setStorageSync(TOKEN_KEY, token)
}

export function clearToken() {
  uni.removeStorageSync(TOKEN_KEY)
}

export function getCurrentUser() {
  return uni.getStorageSync(CURRENT_USER_KEY) || null
}

export function setCurrentUser(user) {
  uni.setStorageSync(CURRENT_USER_KEY, user)
}

export function getCurrentFamily() {
  return uni.getStorageSync(CURRENT_FAMILY_KEY) || null
}

export function setCurrentFamily(family) {
  uni.setStorageSync(CURRENT_FAMILY_KEY, family)
}

export function clearCurrentFamily() {
  uni.removeStorageSync(CURRENT_FAMILY_KEY)
}
```

新增 `parent-child-miniprogram/utils/request.js`：

```js
import { API_BASE_URL } from '../config/env.js'
import { getToken } from './storage.js'

export function request(options) {
  const { url, method = 'GET', data = {}, auth = true } = options
  const header = {
    ...(options.header || {})
  }
  const token = getToken()
  if (auth && token) {
    header.Authorization = `Bearer ${token}`
  }

  return new Promise((resolve, reject) => {
    uni.request({
      url: `${API_BASE_URL}${url}`,
      method,
      data,
      header,
      success(res) {
        if (res.statusCode === 401) {
          const error = new Error('登录已过期')
          error.statusCode = 401
          reject(error)
          return
        }
        const body = res.data || {}
        if (res.statusCode >= 200 && res.statusCode < 300 && body.code === 0) {
          resolve(body.data)
          return
        }
        reject(new Error(body.message || '请求失败'))
      },
      fail() {
        reject(new Error('网络请求失败'))
      }
    })
  }).catch((error) => {
    uni.showToast({
      title: error.message || '请求失败',
      icon: 'none'
    })
    throw error
  })
}
```

新增 `parent-child-miniprogram/api/auth.js`：

```js
import { DEMO_USER } from '../config/env.js'
import { request } from '../utils/request.js'
import { getToken, setCurrentUser, setToken } from '../utils/storage.js'

export async function demoLogin(payload = DEMO_USER) {
  const result = await request({
    url: '/api/auth/demoLogin',
    method: 'POST',
    data: payload,
    auth: false
  })
  setToken(result.token)
  setCurrentUser(result.user)
  return result
}

export async function ensureDemoLogin() {
  if (getToken()) {
    return
  }
  await demoLogin()
}
```

新增 `parent-child-miniprogram/api/family.js`：

```js
import { request } from '../utils/request.js'

export function listFamilies() {
  return request({ url: '/api/family/list' })
}
```

新增 `parent-child-miniprogram/api/dashboard.js`：

```js
import { request } from '../utils/request.js'

export function getDashboardSummary(familyId) {
  return request({
    url: '/api/dashboard/summary',
    data: { familyId }
  })
}
```

新增 `parent-child-miniprogram/api/task.js`：

```js
import { request } from '../utils/request.js'

export function listTasks(familyId) {
  return request({
    url: '/api/task/list',
    data: { familyId }
  })
}

export function listTaskRecords(params) {
  return request({
    url: '/api/task/records',
    data: params
  })
}

export function claimTask(data) {
  return request({
    url: '/api/task/claim',
    method: 'POST',
    data
  })
}

export function submitTask(data) {
  return request({
    url: '/api/task/submit',
    method: 'POST',
    data
  })
}

export function auditTask(data) {
  return request({
    url: '/api/task/audit',
    method: 'POST',
    data
  })
}
```

- [ ] **Step 2: 运行语法检查**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { Get-Content -Raw $_.FullName | node --input-type=module --check }
```

Expected: exit 0。

- [ ] **Step 3: 提交**

```bash
git add parent-child-miniprogram/config/env.js parent-child-miniprogram/utils/storage.js parent-child-miniprogram/utils/request.js parent-child-miniprogram/api/auth.js parent-child-miniprogram/api/family.js parent-child-miniprogram/api/dashboard.js parent-child-miniprogram/api/task.js
git commit -m "feat: 增加小程序请求基础层"
```

---

### Task 2: 家庭选择页真实数据

**Files:**
- Modify: `parent-child-miniprogram/pages.json`
- Modify: `parent-child-miniprogram/pages/family-select/index.vue`

- [ ] **Step 1: 修复 `pages.json` 文案**

将 `parent-child-miniprogram/pages.json` 改为：

```json
{
  "pages": [
    {
      "path": "pages/family-select/index",
      "style": {
        "navigationBarTitleText": "选择家庭",
        "navigationStyle": "custom"
      }
    },
    {
      "path": "pages/index/index",
      "style": {
        "navigationBarTitleText": "首页",
        "navigationStyle": "custom"
      }
    },
    {
      "path": "pages/rewards/index",
      "style": {
        "navigationBarTitleText": "奖励",
        "navigationStyle": "custom"
      }
    },
    {
      "path": "pages/profile/index",
      "style": {
        "navigationBarTitleText": "我的",
        "navigationStyle": "custom"
      }
    }
  ],
  "globalStyle": {
    "navigationBarTextStyle": "black",
    "navigationBarTitleText": "亲子积分",
    "navigationBarBackgroundColor": "#F8F8F8",
    "backgroundColor": "#F8F8F8"
  },
  "tabBar": {
    "color": "#94a3b8",
    "selectedColor": "#3b82f6",
    "borderStyle": "black",
    "backgroundColor": "#ffffff",
    "list": [
      {
        "pagePath": "pages/index/index",
        "text": "首页",
        "iconPath": "static/logo.png",
        "selectedIconPath": "static/logo.png"
      },
      {
        "pagePath": "pages/rewards/index",
        "text": "奖励",
        "iconPath": "static/logo.png",
        "selectedIconPath": "static/logo.png"
      },
      {
        "pagePath": "pages/profile/index",
        "text": "我的",
        "iconPath": "static/logo.png",
        "selectedIconPath": "static/logo.png"
      }
    ]
  }
}
```

- [ ] **Step 2: 重写家庭选择页**

将 `parent-child-miniprogram/pages/family-select/index.vue` 改为一个真实数据页：

- 进入页面调用 `ensureDemoLogin()`。
- 调用 `listFamilies()`。
- 展示加载、空状态、错误重试。
- 点击家庭后使用 `setCurrentFamily(family)` 并 `uni.switchTab({ url: '/pages/index/index' })`。

- [ ] **Step 3: 验证页面源码不含明显乱码**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages.json parent-child-miniprogram\pages\family-select\index.vue
```

Expected: exit 1，无匹配。

- [ ] **Step 4: 提交**

```bash
git add parent-child-miniprogram/pages.json parent-child-miniprogram/pages/family-select/index.vue
git commit -m "feat: 接入家庭选择真实数据"
```

---

### Task 3: 首页真实数据与任务操作

**Files:**
- Modify: `parent-child-miniprogram/pages/index/index.vue`

- [ ] **Step 1: 重写首页脚本和模板**

将 `parent-child-miniprogram/pages/index/index.vue` 改为：

- `onShow` 中读取 `getCurrentFamily()`。
- 无当前家庭则 `uni.reLaunch({ url: '/pages/family-select/index' })`。
- 调用 `getDashboardSummary(familyId)` 和 `listTasks(familyId)`。
- 孩子角色调用 `listTaskRecords({ familyId, status: 'CLAIMED' })` 与 `listTaskRecords({ familyId, status: 'PENDING' })` 合成按钮状态。
- 家长角色调用 `listTaskRecords({ familyId, status: 'PENDING' })` 展示待审核列表。
- `claimTask`、`submitTask`、`auditTask` 操作成功后刷新页面数据。

- [ ] **Step 2: 验证首页源码不含明显乱码**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\index\index.vue
```

Expected: exit 1，无匹配。

- [ ] **Step 3: 提交**

```bash
git add parent-child-miniprogram/pages/index/index.vue
git commit -m "feat: 接入首页真实数据"
```

---

### Task 4: 奖励页和个人页文案修复

**Files:**
- Modify: `parent-child-miniprogram/pages/rewards/index.vue`
- Modify: `parent-child-miniprogram/pages/profile/index.vue`

- [ ] **Step 1: 修复奖励页**

将奖励页改为清晰空状态：

- 标题：“奖励”
- 副标题：“完成任务后，用积分兑换喜欢的奖励”
- 空状态：“奖励列表将在下一轮接入”

- [ ] **Step 2: 修复个人页**

将个人页改为：

- 标题区域展示“我的”和当前家庭名称。
- “切换家庭”点击跳转家庭选择页。
- “家庭成员”“积分流水”点击 toast：“下一轮接入”。

- [ ] **Step 3: 验证页面源码不含明显乱码**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\rewards\index.vue parent-child-miniprogram\pages\profile\index.vue
```

Expected: exit 1，无匹配。

- [ ] **Step 4: 提交**

```bash
git add parent-child-miniprogram/pages/rewards/index.vue parent-child-miniprogram/pages/profile/index.vue
git commit -m "fix: 修复小程序页面文案"
```

---

### Task 5: 收尾验证

**Files:**
- No source changes expected.

- [ ] **Step 1: 检查新增 JS 语法**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { Get-Content -Raw $_.FullName | node --input-type=module --check }
```

Expected: exit 0。

- [ ] **Step 2: 检查小程序源码乱码**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages.json parent-child-miniprogram\pages\family-select\index.vue parent-child-miniprogram\pages\index\index.vue parent-child-miniprogram\pages\rewards\index.vue parent-child-miniprogram\pages\profile\index.vue
```

Expected: exit 1，无匹配。

- [ ] **Step 3: 运行后端全量测试**

Run:

```powershell
go test ./... -count=1
```

Workdir: `parent-child-api`

Expected: PASS。

- [ ] **Step 4: 检查工作区状态**

Run:

```bash
git status --short --branch
```

Expected: `## main`

---

## 自检

- Spec coverage: 覆盖前端请求基础层、API 模块、家庭选择页真实数据、首页真实数据、任务操作、奖励页和个人页文案修复。
- Placeholder scan: 无未定义的后续补充实现项。
- Type consistency: `ensureDemoLogin`、`listFamilies`、`getDashboardSummary`、`listTasks`、`listTaskRecords`、`claimTask`、`submitTask`、`auditTask` 命名与 spec 一致。
- Scope control: 未接入奖励页真实数据、个人页积分流水、家庭创建表单或真实微信登录。
