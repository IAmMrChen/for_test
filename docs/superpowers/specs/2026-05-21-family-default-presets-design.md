# 新家庭默认预置任务和奖品设计

## 背景

PRD 的 MVP 必做范围包含“默认预置任务和奖品”。当前服务端 `FamilyService.CreateFamily` 只创建家庭和家主成员，新家庭创建成功后任务列表和奖励列表为空。用户还需要手动发布任务、创建奖励，才能跑通亲子积分闭环。

当前小程序已经支持：

- 创建家庭。
- 家长发布任务。
- 家长创建奖励。
- 家长创建虚拟孩子并代孩子完成任务、兑换奖励。

因此预置数据应放在后端家庭创建流程中完成，让新家庭创建后天然具备一套可体验的任务和奖励。

## 目标

1. 新家庭创建成功时自动生成默认任务。
2. 新家庭创建成功时自动生成默认奖品。
3. 预置数据由家主成员创建，方便后续审计和展示。
4. 预置任务和奖品可以被现有列表接口直接查询到。
5. 不影响已有家庭。

## 非目标

1. 不为已有家庭补数据。
2. 不提供是否创建预置数据的开关。
3. 不新增接口。
4. 不新增模板管理能力。
5. 不在小程序端写死预置数据。

## 方案选择

### 方案 A：前端创建家庭成功后连续调用任务和奖励创建接口

优点是实现简单。缺点是失败时容易出现家庭已创建但预置数据缺失，也会让前端承担业务初始化责任。

### 方案 B：后端在创建家庭事务内插入预置数据

家庭、家主成员、默认任务、默认奖品在同一个事务内完成。成功后新家庭即可使用，失败则整体回滚。

推荐采用方案 B。

### 方案 C：后台异步初始化预置数据

适合大型模板体系，但当前没有异步任务基础设施，MVP 不需要。

## 预置内容

默认任务：

| 标题 | 积分 | 周期 |
| :--- | ---: | :--- |
| 阅读 30 分钟 | 10 | DAILY |
| 整理自己的物品 | 8 | DAILY |
| 主动完成一件家务 | 12 | WEEKLY |

默认奖品：

| 名称 | 所需积分 | 库存 |
| :--- | ---: | ---: |
| 选择一次周末活动 | 50 | -1 |
| 兑换 30 分钟娱乐时间 | 30 | -1 |
| 一份小礼物 | 80 | -1 |

说明：

- 库存 `-1` 表示不限库存，与现有奖励逻辑一致。
- 周期类型使用现有 `TaskCycleType` 枚举。
- 任务状态使用 `TaskStatusActive`。
- 奖品状态使用 `RewardStatusActive`。

## 服务端设计

位置：`parent-child-api/srv/family_service.go`

新增内部结构：

```go
type defaultTaskPreset struct {
	Title     string
	Points    int
	CycleType model.TaskCycleType
}

type defaultRewardPreset struct {
	Name       string
	PointsCost int
	Stock      int
}
```

新增函数：

```go
func defaultTaskPresets() []defaultTaskPreset
func defaultRewardPresets() []defaultRewardPreset
func createDefaultFamilyPresets(tran transactionExecutor, familyId, creatorMemberId int64)
```

其中 `transactionExecutor` 只需要暴露当前事务已使用的 `MustExecute(query string, args ...any) int64` 能力。

`CreateFamily` 流程调整：

1. 创建家庭。
2. 创建家主成员。
3. 拿到 `memberId`。
4. 调用 `createDefaultFamilyPresets(tran, familyId, memberId)`。
5. 提交事务。

## 测试设计

新增单元测试覆盖：

1. `defaultTaskPresets()` 返回 3 条任务。
2. 每条任务标题非空、积分大于 0、周期合法。
3. `defaultRewardPresets()` 返回 3 条奖品。
4. 每条奖品名称非空、所需积分大于 0、库存为 `-1` 或正数。

集成测试覆盖：

1. 创建家庭后，`TaskService.ListTasks(ownerUserId, familyId)` 可以查到默认任务。
2. 创建家庭后，`RewardService.ListRewards(ownerUserId, familyId)` 可以查到默认奖品。
3. 默认任务的 `CreatedBy` 等于家主成员 ID。
4. 默认奖品的 `CreatedBy` 等于家主成员 ID。

## 错误处理

预置数据创建在创建家庭事务内执行。任一插入失败，事务不提交，调用方收到现有 recover 中间件包装后的错误响应。

## 验收标准

1. 新建家庭后首页任务列表不为空。
2. 新建家庭后奖励页奖励列表不为空。
3. 默认任务可以被孩子或家长代虚拟孩子领取、提交、审核。
4. 默认奖品可以被孩子或家长代虚拟孩子申请兑换。
5. 已有家庭数据不被修改。
6. `go test ./... -count=1` 通过。
