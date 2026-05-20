# 小程序奖励页真实数据接入设计

## 背景

后端已经具备奖励兑换主链路：家长可以创建奖励，家庭成员可以查看有效奖励，孩子可以申请兑换，家长可以发放或驳回兑换申请，孩子可以确认收到已发放奖励。小程序端当前奖励页仍是静态占位页，首页虽然已经展示了待发放奖励数量，但用户无法在小程序内继续处理奖励流程。

本轮只接入奖励页真实数据，让奖励兑换链路在小程序端形成可操作闭环。不新增奖励创建表单，不新增奖励上下架能力，不改后端接口和数据模型。

## 目标

1. 奖励页进入时读取当前家庭，并基于当前家庭调用后端真实接口。
2. 孩子视角可以查看奖励列表、当前积分、已申请记录和已发放待确认记录。
3. 孩子可以对积分足够且有库存的奖励发起兑换申请。
4. 孩子可以对家长已发放的奖励确认收到。
5. 家长视角可以查看奖励列表和待发放申请。
6. 家长可以对待发放申请执行“发放”或“驳回”。
7. 请求失败、登录过期、无家庭、空列表都有明确页面状态或跳转行为。

## 非目标

1. 不实现新建奖励、编辑奖励、上下架奖励。
2. 不实现奖励图片、分类、搜索、分页。
3. 不实现家长代孩子申请奖励。
4. 不新增页面路由；奖励功能仍在 `pages/rewards/index.vue` 内完成。
5. 不修改后端奖励状态机。

## 推荐方案

采用“奖励页单页接入”的方案：新增 `api/reward.js` 封装后端奖励接口，奖励页在 `onShow` 中加载 summary、奖励列表和奖励记录，根据当前角色展示不同操作区。

这个方案最小但完整。它复用上一轮已经建立的 `request.js`、`ensureDemoLogin`、`getCurrentFamily` 和 dashboard summary，不扩大小程序路由，也不提前引入奖励管理表单。

备选方案一是先接个人页成员与积分流水。这个方向也合理，但它偏查询展示，不能继续打通“积分消耗后获得奖励”的核心闭环。

备选方案二是一次性接奖励页、成员页、积分流水和奖励创建。这个范围过大，会把展示、兑换、管理、筛选多个交互混在一轮里，不利于验证和提交。

## 文件边界

- `parent-child-miniprogram/api/reward.js`：新增奖励接口 client，封装列表、记录、申请、发放、驳回、确认收到。
- `parent-child-miniprogram/pages/rewards/index.vue`：从占位页改为真实奖励页，负责加载状态、角色分支、列表展示和操作。
- `parent-child-miniprogram/api/dashboard.js`：不修改，继续使用现有 `getDashboardSummary`。
- `parent-child-miniprogram/utils/request.js`：不修改，继续使用上一轮的 token 和 401 处理。

## 接口使用

### 奖励列表

```http
GET /api/reward/list?familyId=1
```

返回 `Reward[]`：

```json
{
  "id": 1,
  "familyId": 1,
  "name": "周末看电影",
  "pointsCost": 30,
  "stock": -1,
  "status": 1,
  "createdBy": 10
}
```

`stock = -1` 表示不限库存，`stock = 0` 表示已无库存，`stock > 0` 表示剩余库存。

### 奖励记录

```http
GET /api/reward/records?familyId=1&status=APPLIED
GET /api/reward/records?familyId=1&status=DELIVERED
```

返回 `RewardRecordListItem[]`：

```json
{
  "id": 1,
  "familyId": 1,
  "rewardId": 1,
  "rewardName": "周末看电影",
  "memberId": 20,
  "nickname": "小明",
  "pointsCost": 30,
  "status": "APPLIED",
  "applyTime": "2026-05-20T10:00:00+08:00",
  "operateTime": null,
  "operateBy": null
}
```

后端已经按角色过滤记录：孩子只能看到自己的记录；家长类角色可以看到家庭内记录。

### 操作接口

```http
POST /api/reward/apply
POST /api/reward/deliver
POST /api/reward/reject
POST /api/reward/receive
```

申请兑换传：

```json
{
  "familyId": 1,
  "rewardId": 1
}
```

发放、驳回、确认收到传：

```json
{
  "recordId": 1
}
```

## 页面行为

### 通用加载

1. `onShow` 读取 `getCurrentFamily()`。
2. 无当前家庭时跳转到 `/pages/family-select/index`。
3. 调用 `ensureDemoLogin()`。
4. 并行加载：
   - `getDashboardSummary(familyId)`
   - `listRewards(familyId)`
5. 孩子角色继续加载：
   - `listRewardRecords({ familyId, status: 'APPLIED' })`
   - `listRewardRecords({ familyId, status: 'DELIVERED' })`
6. 家长角色继续加载：
   - `listRewardRecords({ familyId, status: 'APPLIED' })`
7. 遇到 401 时强制 `ensureDemoLogin(true)` 并重试一次。

### 孩子视角

顶部展示：

- 当前家庭名
- 当前积分
- 已申请待发放数量

奖励列表展示：

- 奖励名称
- 所需积分
- 库存文案
- 操作按钮

按钮规则：

- 已存在 `APPLIED` 记录：显示“待发放”，禁用。
- 积分不足：显示“积分不足”，禁用。
- 库存为 0：显示“已兑完”，禁用。
- 其他情况：显示“兑换”，点击调用 `/api/reward/apply`。

已发放待确认区域展示 `DELIVERED` 记录，按钮为“确认收到”，点击调用 `/api/reward/receive`。

### 家长视角

顶部展示：

- 当前家庭名
- 有效奖励数量
- 待发放申请数量

奖励列表只展示，不提供创建入口：

- 奖励名称
- 所需积分
- 库存文案

待发放申请区域展示 `APPLIED` 记录：

- 孩子昵称
- 奖励名称
- 消耗积分
- 申请时间
- “驳回”按钮调用 `/api/reward/reject`
- “发放”按钮调用 `/api/reward/deliver`

## 状态与错误

1. 加载中展示页面级 loading。
2. 奖励列表为空时展示“暂无可兑换奖励”。
3. 家长待发放为空时展示“当前没有待发放奖励”。
4. 孩子已发放待确认为空时不展示该区块，避免页面噪音。
5. 操作成功后展示 toast，并重新加载页面数据。
6. 非 401 请求错误沿用 `request.js` 的 toast。
7. 401 自动重新登录只重试一次；重试仍失败则把错误交给请求层处理。

## 测试策略

1. 前端语法检查：对 `api/*.js` 执行 `node --check`。
2. 页面源码扫描：检查奖励页不含明显乱码片段。
3. 后端回归：执行 `go test ./... -count=1`，确保奖励状态机和既有接口仍通过。
4. 手工验证路径：
   - 孩子进入奖励页可以看到当前积分和奖励列表。
   - 孩子积分不足时按钮禁用。
   - 孩子兑换成功后出现待发放状态。
   - 家长进入奖励页可以看到待发放申请。
   - 家长发放后，孩子可以看到“确认收到”。
   - 家长驳回后，孩子积分会由后端退回，页面刷新后积分更新。

## 后续扩展

1. 家长新建奖励表单。
2. 奖励图片、分类、库存展示优化。
3. 个人页接入成员列表和积分流水。
4. 奖励记录历史筛选。

## 自检

1. 范围聚焦：只做奖励页真实数据接入，不扩展奖励管理后台。
2. 接口一致：所有 API 路径和字段均来自当前后端代码。
3. 角色清晰：孩子负责申请和确认收到，家长负责发放和驳回。
4. 状态闭环：申请、发放、驳回、确认收到都有页面入口。
