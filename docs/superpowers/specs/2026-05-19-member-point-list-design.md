# 成员列表与积分流水查询设计

## 背景

当前后端已经完成家庭、成员邀请、任务领取/提交/审核、奖励兑换、首页摘要等 MVP 主链路。任务审核通过时会写入 `point_logs` 正向流水，奖励申请时会写入扣减流水，奖励驳回时会写入退回流水，但客户端还没有统一查询积分流水的接口。

同时，家庭内成员能力目前只有“创建虚拟孩子”和“邀请真实微信用户加入家庭”，缺少成员列表接口。后续任务代提交、奖励代兑换、成员管理、家庭首页切换成员视角，都需要先能稳定拿到家庭成员列表。

本设计只补齐后端查询能力，不处理小程序页面接入，不新增积分手动调整，不修改现有任务和奖励状态流。

## 目标

1. 提供家庭成员列表接口，让已加入家庭的用户可以查看当前家庭内的有效成员。
2. 提供积分流水查询接口，让孩子查看自己的积分变化，让家长查看家庭内孩子或成员的积分变化。
3. 复用现有 `api -> srv -> model` 分层和 `resx.Db.Main` 原生 SQL 查询风格。
4. 保持 MVP 简洁：先按时间倒序返回最近流水，不做分页游标、复杂筛选和导出。

## 非目标

1. 不实现积分手动调整接口。
2. 不实现成员移除、改名、角色变更。
3. 不实现小程序页面真实数据绑定。
4. 不实现分页、时间范围筛选、流水类型多选。
5. 不改造已有 `point_logs` 表结构，除非集成测试辅助建表时需要补齐缺失索引。

## 方案选择

推荐方案：新增轻量查询接口，保持 service 内部权限判断。

备选方案一：只在 dashboard summary 中继续增加积分流水摘要。这个方案不能满足“流水可查”和后续积分页需求，会把首页接口做得过重。

备选方案二：先做完整成员管理和积分管理后台。这个方案覆盖面太大，会引入角色变更、成员移除、积分调整等新流程，不适合当前 MVP 收口。

因此本轮只做两个查询接口：成员列表和积分流水列表。它们是后续前端接入和管理能力的公共基础，但不会扩大写操作面。

## 接口设计

### 家庭成员列表

接口：

```http
GET /api/member/list?familyId=1
Authorization: Bearer <jwt>
```

响应数据：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 10,
      "familyId": 1,
      "userId": 1001,
      "roleType": "OWNER",
      "nickname": "爸爸",
      "currentPoints": 0,
      "totalEarnedPoints": 0,
      "isVirtual": false,
      "status": "ACTIVE"
    }
  ]
}
```

规则：

1. `familyId` 必填且必须大于 0。
2. 当前用户必须是该家庭的有效成员。
3. MVP 阶段家庭内有效成员对所有家庭成员可见。
4. 只返回 `status = ACTIVE` 的成员，按角色和 ID 稳定排序：家主、管理员、家长、孩子，再按 `id ASC`。

### 积分流水列表

接口：

```http
GET /api/point/logs?familyId=1&memberId=10
Authorization: Bearer <jwt>
```

`memberId` 可选。

响应数据：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 88,
      "familyId": 1,
      "memberId": 10,
      "nickname": "小明",
      "points": 5,
      "sourceType": "TASK",
      "sourceId": 66,
      "sourceTitle": "每日阅读",
      "createdAt": "2026-05-19T10:00:00Z"
    }
  ]
}
```

规则：

1. `familyId` 必填且必须大于 0。
2. 当前用户必须是该家庭的有效成员。
3. 孩子只能查看自己的流水，即使传入其他 `memberId` 也返回 `permission denied`。
4. 家长类角色可以查看家庭内全部流水；传入 `memberId` 时只查看目标成员流水。
5. 如果传入 `memberId`，该成员必须属于同一家庭且为有效成员。
6. 按 `point_logs.id DESC` 返回。
7. MVP 暂不分页，默认最多返回 100 条，避免接口被历史数据拖慢。

## 模型设计

新增 `parent-child-api/model/point_model.go`：

```go
type PointLogListRequest struct {
	FamilyId int64 `json:"familyId"`
	MemberId int64 `json:"memberId"`
}

type PointLogListItem struct {
	Id          int64           `json:"id"`
	FamilyId    int64           `json:"familyId"`
	MemberId    int64           `json:"memberId"`
	Nickname    string          `json:"nickname"`
	Points      int             `json:"points"`
	SourceType  PointSourceType `json:"sourceType"`
	SourceId    int64           `json:"sourceId"`
	SourceTitle string          `json:"sourceTitle"`
	CreatedAt   time.Time       `json:"createdAt"`
}
```

成员列表复用现有 `model.FamilyMember`，不新增 DTO。这样可以直接暴露积分字段，满足任务代提交、奖励代兑换和成员选择的基础展示需要。

## Service 设计

### MemberService.ListMembers

签名：

```go
func (x memberService) ListMembers(userId int64, familyId int64) []model.FamilyMember
```

流程：

1. 校验 `familyId`。
2. 调用 `MemberService.LoadActiveMember(userId, familyId)` 校验当前用户属于家庭。
3. 用原生 SQL 查询 `family_members`。
4. 仅返回 `ACTIVE` 成员。

### PointService.ListPointLogs

新增 `parent-child-api/srv/point_service.go`：

```go
var PointService pointService

type pointService struct{}

func (x pointService) ListPointLogs(userId int64, req model.PointLogListRequest) []model.PointLogListItem
```

流程：

1. 校验 `familyId`。
2. 加载当前用户在家庭中的有效成员身份。
3. 如果当前用户是孩子：
   - 未传 `memberId` 时自动限定为自己的 `member.Id`。
   - 传了自己的 `memberId` 时允许。
   - 传了他人的 `memberId` 时拒绝。
4. 如果当前用户是家长类角色：
   - 未传 `memberId` 时查询家庭全部流水。
   - 传 `memberId` 时校验目标成员属于同一家庭且有效。
5. 查询 `point_logs`，关联 `family_members` 得到昵称。
6. 根据 `source_type` 左关联 `tasks` 或 `rewards` 得到 `sourceTitle`：
   - `TASK` 使用任务标题。
   - `REWARD` 使用奖励名称。
   - 其他类型返回空字符串。

## 权限模型

沿用当前角色约定：

1. `OWNER`、`ADMIN`、`PARENT` 属于家长类角色。
2. `CHILD` 是孩子角色。
3. 成员列表：家庭内有效成员均可查看。
4. 积分流水：孩子只能看自己；家长类角色可以看家庭内所有成员。

该设计不依赖前端隐藏按钮，权限全部在 service 层校验。

## 错误处理

使用现有 service panic + API recover 机制：

1. 缺少或非法 `familyId`：`family id is required`。
2. 当前用户不属于家庭：`permission denied`。
3. 孩子查看他人流水：`permission denied`。
4. 目标成员不存在或不属于家庭：`target member not found`。
5. API query 参数解析失败：返回 400 JSON 错误。

## 测试策略

1. 单元测试：
   - `MemberService.ListMembers` 缺少 `familyId` 会 panic。
   - `PointService.ListPointLogs` 缺少 `familyId` 会 panic。
   - 孩子查看他人流水的权限 helper 返回拒绝。
2. API 路由测试：
   - `GET /api/member/list?familyId=bad` 返回 400。
   - `GET /api/point/logs?familyId=bad` 返回 400。
3. 数据库集成测试：
   - 创建家庭。
   - 邀请孩子。
   - 创建任务并审核通过，产生正向积分流水。
   - 创建奖励并申请，产生扣减流水。
   - 家长查看全部流水，能看到任务和奖励两条。
   - 孩子查看自己的流水，能看到自己的两条。
   - 孩子传入家长成员 ID 查询时被拒绝。
   - 成员列表包含家主和孩子。

## 实施边界

本轮完成后，前端可以拿到：

1. 家庭成员列表，用于成员选择、展示当前孩子积分。
2. 当前用户可见的积分流水，用于积分页和个人页。

下一轮更适合处理小程序侧 API client 和页面真实数据绑定，因为届时家庭列表、首页摘要、任务、奖励、成员和积分流水的后端基础都已经齐了。

## 自检

1. 明确性：本文没有未定义的后续补充实现项。
2. 无范围漂移：没有引入成员写操作、积分调整、前端接入或分页系统。
3. 类型一致：`PointLogListRequest`、`PointLogListItem`、`MemberService.ListMembers`、`PointService.ListPointLogs` 命名在设计中保持一致。
4. 权限明确：成员列表和积分流水分别定义了可见范围，并明确孩子不能查看他人流水。
