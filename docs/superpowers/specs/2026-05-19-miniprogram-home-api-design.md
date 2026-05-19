# 小程序首页真实数据接入设计

## 背景

后端 MVP 主链路已经具备家庭、成员、任务、奖励、首页摘要和积分流水接口。小程序当前仍以 mock 数据为主，且 `pages.json`、首页、家庭选择页、奖励页、个人页存在明显乱码文案。继续在 mock 页面上叠加功能会让问题变得难以维护。

本轮设计采用“先打通基础接入层，再接首页”的方式：新增前端 API 基础设施，修复核心页面中文文案，并让首页基于后端真实数据展示当前家庭摘要、任务列表和待办记录。奖励页、个人页和积分流水页的真实数据绑定放到后续计划。

## 目标

1. 建立小程序侧统一请求层，封装 `baseUrl`、JWT、业务错误和网络错误。
2. 建立小程序侧 API 模块，先覆盖登录、家庭列表、首页摘要、任务列表、任务记录、任务领取、任务提交、任务审核。
3. 修复 `pages.json`、家庭选择页、首页、奖励页、个人页的乱码文案，保证页面能被正常理解。
4. 家庭选择页使用真实 `family/list` 数据，选择后写入 `currentFamily`。
5. 首页使用真实 `dashboard/summary`、`task/list`、`task/records` 数据。
6. 保留 demo 登录机制，方便本地和微信开发者工具中快速验证。

## 非目标

1. 不接入真实微信 `code -> openid` 登录。
2. 不新增后端接口。
3. 不实现奖励页真实数据绑定。
4. 不实现个人页积分流水绑定。
5. 不引入 npm 构建链或第三方状态管理库。
6. 不改动 `unpackage/dist` 编译产物。

## 方案选择

推荐方案：轻量 API client + 首页垂直接入。

备选方案一：一次性接入全部页面。这个方案容易把乱码修复、接口封装、奖励兑换、积分流水和成员管理混在一起，验证面过大。

备选方案二：只修复乱码，不接接口。这个方案能改善观感，但无法推进 MVP 可用闭环。

推荐方案能在不扩大范围的前提下，让小程序从静态 mock 走向可验证的真实数据。首页是最佳切入点，因为它会同时用到家庭、角色、任务、待审核、待发奖和积分摘要。

## 文件结构

新增：

- `parent-child-miniprogram/config/env.js`：保存本地 API 地址和 demo 登录默认用户。
- `parent-child-miniprogram/utils/storage.js`：统一 token、当前用户、当前家庭的本地缓存读写。
- `parent-child-miniprogram/utils/request.js`：统一请求封装。
- `parent-child-miniprogram/api/auth.js`：demo 登录。
- `parent-child-miniprogram/api/family.js`：家庭列表。
- `parent-child-miniprogram/api/dashboard.js`：首页摘要。
- `parent-child-miniprogram/api/task.js`：任务与任务记录。

修改：

- `parent-child-miniprogram/pages.json`：修复导航和 tab 文案。
- `parent-child-miniprogram/pages/family-select/index.vue`：从后端加载家庭列表，支持 demo 登录。
- `parent-child-miniprogram/pages/index/index.vue`：绑定首页真实数据。
- `parent-child-miniprogram/pages/rewards/index.vue`：只修复文案与空状态。
- `parent-child-miniprogram/pages/profile/index.vue`：只修复文案和当前家庭展示。

## 请求与登录设计

### 环境配置

`config/env.js` 暴露：

```js
export const API_BASE_URL = 'http://127.0.0.1:8080'

export const DEMO_USER = {
  userId: 1001,
  nickname: '演示用户'
}
```

后续如果需要切换局域网 IP，只改 `API_BASE_URL`。

### 本地缓存

`utils/storage.js` 负责：

- `getToken()` / `setToken(token)` / `clearToken()`
- `getCurrentUser()` / `setCurrentUser(user)`
- `getCurrentFamily()` / `setCurrentFamily(family)` / `clearCurrentFamily()`

缓存 key 统一命名：

- `authToken`
- `currentUser`
- `currentFamily`

### 请求封装

`utils/request.js` 负责：

1. 拼接 `API_BASE_URL + url`。
2. 自动带上 `Authorization: Bearer <token>`。
3. 解析后端统一响应结构：
   - `code === 0` 返回 `data`。
   - `code !== 0` 抛出业务错误。
4. 网络失败时显示 `uni.showToast({ title: '网络请求失败', icon: 'none' })`。
5. 业务失败时显示后端 `message`。

不在本轮做自动刷新 token，因为后端当前是 demo JWT 机制。

## API 模块设计

### auth API

```js
export function demoLogin(payload = DEMO_USER) {
  return request({
    url: '/api/auth/demoLogin',
    method: 'POST',
    data: payload,
    auth: false
  })
}
```

家庭选择页进入时，如果没有 token，则自动 demo 登录并缓存 token。

### family API

```js
export function listFamilies() {
  return request({ url: '/api/family/list' })
}
```

返回后端 `FamilyListItem`。前端选择家庭时保存完整 item，并兼容字段：

- `familyId`
- `familyName`
- `memberId`
- `roleType`
- `nickname`

### dashboard API

```js
export function getDashboardSummary(familyId) {
  return request({ url: '/api/dashboard/summary', data: { familyId } })
}
```

首页使用 summary 决定当前用户角色、积分、待审核任务数量、待发奖数量、已领取任务数量、审核中任务数量。

### task API

```js
export function listTasks(familyId) {
  return request({ url: '/api/task/list', data: { familyId } })
}

export function listTaskRecords(params) {
  return request({ url: '/api/task/records', data: params })
}

export function claimTask(data) {
  return request({ url: '/api/task/claim', method: 'POST', data })
}

export function submitTask(data) {
  return request({ url: '/api/task/submit', method: 'POST', data })
}

export function auditTask(data) {
  return request({ url: '/api/task/audit', method: 'POST', data })
}
```

## 页面行为设计

### 家庭选择页

进入页面时：

1. 调用 `ensureDemoLogin()`。
2. 调用 `listFamilies()`。
3. 如果有家庭：
   - 展示家庭卡片。
   - 点击家庭后写入 `currentFamily`，跳转首页。
4. 如果没有家庭：
   - 展示空状态“暂无家庭，请先在后端创建或接受邀请”。

本轮不在前端创建家庭，因为需要表单和角色流程，会另开计划。

### 首页

进入页面时：

1. 读取 `currentFamily`。
2. 如果没有当前家庭，跳转家庭选择页。
3. 使用 `currentFamily.familyId` 请求 `dashboard/summary`。
4. 请求 `task/list`。
5. 根据角色加载不同记录：
   - 孩子：加载自己的 `CLAIMED` 和 `PENDING` 记录，用于任务按钮状态。
   - 家长类角色：加载 `PENDING` 任务记录，用于待审核列表。

孩子视角：

- 顶部展示家庭名、昵称、当前积分。
- 任务列表展示任务名、积分、状态按钮。
- 没有打开记录的任务显示“领取”。
- 已领取任务显示“提交”。
- 审核中任务显示“审核中”。
- 点击“领取”调用 `task/claim`。
- 点击“提交”调用 `task/submit`。

家长视角：

- 顶部展示家庭名和角色。
- 统计卡展示待审核任务、待发奖、可用任务。
- 待办列表展示待审核任务记录。
- 点击“通过”调用 `task/audit`，`approved=true`。
- 点击“驳回”调用 `task/audit`，`approved=false`。

### 奖励页

本轮只修复文案：

- 标题：“奖励”
- 副标题：“完成任务后，用积分兑换喜欢的奖励”
- 空状态：“奖励列表将在下一轮接入”

### 个人页

本轮只修复文案：

- 展示“我的”
- 展示当前家庭名称。
- 提供“切换家庭”入口。
- 其他菜单使用“家庭成员”“积分流水”，但点击只提示“下一轮接入”。

## 状态映射

### 角色文案

- `OWNER`：家主
- `ADMIN`：管理员
- `PARENT`：家长
- `CHILD`：孩子

家长类角色判断：

```js
['OWNER', 'ADMIN', 'PARENT'].includes(roleType)
```

### 任务按钮状态

根据任务列表和任务记录合成：

- 没有 `CLAIMED/PENDING` 记录：`claimable`，按钮“领取”
- 有 `CLAIMED` 记录：`claimed`，按钮“提交”
- 有 `PENDING` 记录：`pending`，按钮“审核中”

已通过和已驳回记录不影响当前任务按钮，后续周期任务规则再细化。

## 错误处理

1. 没有 token：自动 demo 登录。
2. token 失效或后端返回 401：清理 token，重新 demo 登录一次，再重试当前页面初始化。
3. 当前家庭不存在：清理 `currentFamily`，跳转家庭选择页。
4. 接口返回业务错误：toast 展示后端 `message`。
5. 列表为空：展示空状态，不显示 mock 数据。

## 验证策略

由于小程序目录当前没有 `package.json`，本轮不引入 npm 单元测试。验证以静态检查和手动运行为主：

1. 使用 `rg` 确认源文件中不再出现明显乱码片段。
2. 使用 `node --check` 检查新增 `.js` API 文件语法。
3. 使用后端已有测试确认接口仍通过：
   - `go test ./... -count=1`
4. 手动验收路径：
   - 启动后端。
   - 在 HBuilderX 或微信开发者工具打开小程序。
   - 进入家庭选择页，自动 demo 登录。
   - 选择家庭后进入首页。
   - 孩子角色可查看任务、领取、提交。
   - 家长角色可查看待审核任务并审核。

## 实施边界

本轮完成后，小程序首页可以使用真实数据参与 MVP 主闭环，但仍保留以下后续工作：

1. 奖励页真实列表、申请兑换、确认领取。
2. 个人页积分流水和成员列表。
3. 家庭创建、邀请、创建虚拟孩子的前端表单。
4. 真实微信登录。
5. 小程序端自动化测试或更完整的构建脚本。

## 自检

1. 范围清晰：只做基础 API 层、首页绑定和文案修复。
2. 不引入新后端能力：所有数据来自现有接口。
3. 不引入构建链：保持当前 HBuilderX/uni-app 项目形态。
4. 错误处理明确：覆盖 token、业务错误、空状态和当前家庭缺失。
