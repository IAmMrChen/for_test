# 小程序家庭与成员入口设计

## 背景

当前服务端已经提供家庭与成员相关核心接口：

- `POST /api/family/create`：创建家庭，并把当前微信用户设为家主。
- `GET /api/family/list`：列出当前用户已加入的家庭。
- `POST /api/member/createVirtualChild`：在家庭内创建虚拟孩子。
- `GET /api/member/list`：列出家庭成员。
- `POST /api/invite/create`：创建指定角色的邀请码。
- `POST /api/invite/accept`：当前微信用户接受邀请码并加入家庭。

小程序当前已经可以选择家庭、查看成员、查看积分流水，但还缺少创建家庭、创建虚拟孩子、发起真实微信用户邀请、接受邀请的入口。这个缺口会影响 PRD 中“虚拟孩子和真实孩子需求并存”的 MVP 验收。

## 目标

1. 用户没有家庭时，可以直接在小程序创建家庭，而不是依赖后端准备数据。
2. 家长角色可以在个人页创建虚拟孩子。
3. 家长角色可以在个人页创建邀请，并在创建时选择被邀请人的角色。
4. 被邀请微信用户可以在家庭选择页输入邀请码加入家庭。
5. 一个微信用户在一个家庭内只会有一个身份，继续依赖服务端校验。

## 非目标

1. 不做成员移除。
2. 不做成员角色变更。
3. 不做邀请取消、邀请列表、邀请过期管理。
4. 不做微信分享卡片自动拉起接受邀请。
5. 不做虚拟孩子绑定真实微信账号。
6. 不调整服务端数据模型。

## 方案选择

### 方案 A：只在个人页做成员管理

优点是入口集中，个人页已经展示成员列表。缺点是无家庭用户没有机会创建家庭或接受邀请，MVP 新用户路径断开。

### 方案 B：家庭选择页处理入门路径，个人页处理家庭内成员动作

家庭选择页负责“创建家庭”和“输入邀请码加入家庭”；个人页负责“创建虚拟孩子”和“生成邀请”。这和用户心智一致：还没进家庭时先解决加入/创建家庭，进入家庭后再管理成员。

推荐采用方案 B。

### 方案 C：新增独立家庭管理页

可扩展性最好，但当前小程序页面数量少，新增路由会带来导航、跳转和状态同步成本。MVP 阶段暂不需要。

## 页面设计

### 家庭选择页

位置：`parent-child-miniprogram/pages/family-select/index.vue`

新增两个轻量入口：

1. 创建家庭
   - 字段：家庭名称、我在家庭中的昵称。
   - 家庭名称必填。
   - 昵称可选，空值交给后端使用默认值“家长”。
   - 创建成功后，把新家庭写入 `setCurrentFamily`，并跳转首页。

2. 输入邀请码加入家庭
   - 字段：邀请码。
   - 邀请码必填。
   - 接受成功后，把返回成员所属家庭设为当前家庭。
   - 由于 `acceptInvite` 当前只返回 `FamilyMember`，其中没有 `familyName`，前端接受成功后应重新调用 `listFamilies()`，在结果中按 `familyId` 找到新家庭，再写入 `setCurrentFamily`。
   - 如果找不到对应家庭，则提示用户重新加载家庭列表，不跳转。

空家庭态不再提示“请先在后端创建家庭”，改成引导用户创建家庭或输入邀请码。

### 个人页

位置：`parent-child-miniprogram/pages/profile/index.vue`

仅家长角色显示家庭管理入口。家长角色定义沿用现有前端规则：

- `OWNER`
- `ADMIN`
- `PARENT`

新增两个能力：

1. 创建虚拟孩子
   - 字段：孩子昵称。
   - 昵称必填。
   - 提交成功后刷新成员列表。

2. 生成邀请
   - 字段：邀请角色。
   - `OWNER` 可邀请 `ADMIN`、`PARENT`、`CHILD`。
   - `ADMIN` 和 `PARENT` 可邀请 `PARENT`、`CHILD`。
   - `CHILD` 不显示入口。
   - 生成成功后展示邀请码 `token`、过期时间和“复制邀请码”按钮。
   - 复制失败时仍保留 token 文本，方便用户手动复制。

## API 封装

新增或补充以下小程序 API 方法：

### `parent-child-miniprogram/api/family.js`

```js
export function createFamily(data) {
  return request({
    url: '/api/family/create',
    method: 'POST',
    data
  })
}
```

### `parent-child-miniprogram/api/member.js`

```js
export function createVirtualChild(data) {
  return request({
    url: '/api/member/createVirtualChild',
    method: 'POST',
    data
  })
}
```

### `parent-child-miniprogram/api/invite.js`

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

## 数据映射

创建家庭接口返回：

```json
{
  "family": {
    "id": 1,
    "name": "陈家",
    "creatorId": 1
  },
  "member": {
    "id": 1,
    "familyId": 1,
    "roleType": "OWNER",
    "nickname": "爸爸"
  }
}
```

前端写入当前家庭时统一转成家庭列表结构：

```js
{
  familyId: response.family.id,
  familyName: response.family.name,
  memberId: response.member.id,
  roleType: response.member.roleType,
  nickname: response.member.nickname
}
```

接受邀请接口返回 `FamilyMember`，前端先调用 `listFamilies()`，再按 `member.familyId` 定位完整家庭信息。

## 校验与错误处理

1. 家庭名称为空：提示“请填写家庭名称”。
2. 邀请码为空：提示“请填写邀请码”。
3. 虚拟孩子昵称为空：提示“请填写孩子昵称”。
4. 未选择邀请角色：提示“请选择邀请角色”。
5. 提交中按钮禁用，避免重复提交。
6. 接口错误沿用 `request.js` 的统一错误提示。
7. `401` 沿用当前页面的 `ensureDemoLogin(true)` 重试风格。

## 验收标准

1. 新用户进入家庭选择页后，可以创建家庭并自动进入首页。
2. 新用户可以输入有效邀请码加入家庭，并自动进入首页。
3. 家长进入个人页后，可以创建虚拟孩子，成员列表刷新并显示“虚拟账号”标签。
4. 家主可以创建管理员、家长、孩子邀请。
5. 普通家长和管理员可以创建家长、孩子邀请，但不能创建管理员邀请。
6. 孩子进入个人页后看不到创建虚拟孩子和生成邀请入口。
7. 同一个微信用户重复接受同一家庭邀请时，服务端拒绝，前端展示错误提示。
8. 所有新增表单提交时都有基础必填校验和提交中禁用状态。

## 实施顺序

1. 增加小程序 API 封装。
2. 家庭选择页增加创建家庭和接受邀请入口。
3. 个人页增加创建虚拟孩子入口。
4. 个人页增加生成邀请入口。
5. 跑前端语法检查、乱码扫描和服务端测试。
