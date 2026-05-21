# 新家庭默认预置任务和奖品 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新家庭创建成功时自动生成默认任务和默认奖品，让用户创建家庭后即可体验积分闭环。

**Architecture:** 在 `FamilyService.CreateFamily` 的同一个数据库事务中创建家庭、家主成员、默认任务和默认奖品。预置内容只在服务端定义，不新增接口，不修改前端。

**Tech Stack:** Go、现有 `resx.Db.Main` 原生 SQL 风格、MySQL 集成测试。

---

## 文件结构

- Modify: `parent-child-api/srv/family_service.go`
  - 增加默认任务和默认奖品 preset 定义。
  - 在创建家庭事务中插入默认任务和奖品。
- Modify: `parent-child-api/srv/family_service_test.go`
  - 增加 preset 单元测试。
- Modify: `parent-child-api/srv/integration_flow_test.go`
  - 在家庭成员邀请集成测试中补默认任务和奖品断言。

## 预置内容

默认任务：

```go
[]defaultTaskPreset{
	{Title: "阅读 30 分钟", Points: 10, CycleType: model.TaskCycleTypeDaily},
	{Title: "整理自己的物品", Points: 8, CycleType: model.TaskCycleTypeDaily},
	{Title: "主动完成一件家务", Points: 12, CycleType: model.TaskCycleTypeWeekly},
}
```

默认奖品：

```go
[]defaultRewardPreset{
	{Name: "选择一次周末活动", PointsCost: 50, Stock: -1},
	{Name: "兑换 30 分钟娱乐时间", PointsCost: 30, Stock: -1},
	{Name: "一份小礼物", PointsCost: 80, Stock: -1},
}
```

---

### Task 1: 为预置定义补单元测试

**Files:**
- Modify: `parent-child-api/srv/family_service_test.go`

- [ ] **Step 1: 增加默认任务测试**

在 `family_service_test.go` 末尾追加：

```go
func TestDefaultTaskPresetsAreValid(t *testing.T) {
	presets := defaultTaskPresets()
	if len(presets) != 3 {
		t.Fatalf("len(defaultTaskPresets()) = %d, want 3", len(presets))
	}

	for _, preset := range presets {
		if strings.TrimSpace(preset.Title) == "" {
			t.Fatalf("task preset title is empty: %+v", preset)
		}
		if preset.Points <= 0 {
			t.Fatalf("task preset points = %d, want > 0", preset.Points)
		}
		if !preset.CycleType.Valid() {
			t.Fatalf("task preset cycle type = %s, want valid", preset.CycleType)
		}
	}
}
```

- [ ] **Step 2: 增加默认奖品测试**

继续追加：

```go
func TestDefaultRewardPresetsAreValid(t *testing.T) {
	presets := defaultRewardPresets()
	if len(presets) != 3 {
		t.Fatalf("len(defaultRewardPresets()) = %d, want 3", len(presets))
	}

	for _, preset := range presets {
		if strings.TrimSpace(preset.Name) == "" {
			t.Fatalf("reward preset name is empty: %+v", preset)
		}
		if preset.PointsCost <= 0 {
			t.Fatalf("reward preset points cost = %d, want > 0", preset.PointsCost)
		}
		if preset.Stock != -1 && preset.Stock <= 0 {
			t.Fatalf("reward preset stock = %d, want -1 or > 0", preset.Stock)
		}
	}
}
```

- [ ] **Step 3: 补 `strings` 导入**

把导入区从：

```go
import (
	"fmt"
	"testing"

	"parent-child-api/model"
	"parent-child-api/resx"
)
```

改成：

```go
import (
	"fmt"
	"strings"
	"testing"

	"parent-child-api/model"
	"parent-child-api/resx"
)
```

- [ ] **Step 4: 运行测试确认失败**

Run:

```powershell
go test ./srv -run "TestDefault(Task|Reward)PresetsAreValid" -count=1
```

Working directory: `parent-child-api`

Expected: 失败，错误为 `undefined: defaultTaskPresets` 或 `undefined: defaultRewardPresets`。

---

### Task 2: 实现预置数据定义和插入

**Files:**
- Modify: `parent-child-api/srv/family_service.go`

- [ ] **Step 1: 增加结构和接口**

在 `const defaultFamilyNickname = "\u5bb6\u957f"` 后加入：

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

type familyPresetExecutor interface {
	MustExecute(query string, args ...any) int64
}
```

- [ ] **Step 2: 增加 preset 函数**

在结构定义后加入：

```go
func defaultTaskPresets() []defaultTaskPreset {
	return []defaultTaskPreset{
		{Title: "阅读 30 分钟", Points: 10, CycleType: model.TaskCycleTypeDaily},
		{Title: "整理自己的物品", Points: 8, CycleType: model.TaskCycleTypeDaily},
		{Title: "主动完成一件家务", Points: 12, CycleType: model.TaskCycleTypeWeekly},
	}
}

func defaultRewardPresets() []defaultRewardPreset {
	return []defaultRewardPreset{
		{Name: "选择一次周末活动", PointsCost: 50, Stock: -1},
		{Name: "兑换 30 分钟娱乐时间", PointsCost: 30, Stock: -1},
		{Name: "一份小礼物", PointsCost: 80, Stock: -1},
	}
}
```

- [ ] **Step 3: 增加创建默认数据函数**

在 `defaultRewardPresets()` 后加入：

```go
func createDefaultFamilyPresets(db familyPresetExecutor, familyId, creatorMemberId int64) {
	const insertTaskSql = `
		INSERT INTO tasks(family_id, title, points, cycle_type, status, created_by)
		VALUES(@p1, @p2, @p3, @p4, @p5, @p6)
	`
	for _, preset := range defaultTaskPresets() {
		db.MustExecute(
			insertTaskSql,
			familyId,
			preset.Title,
			preset.Points,
			preset.CycleType,
			model.TaskStatusActive,
			creatorMemberId,
		)
	}

	const insertRewardSql = `
		INSERT INTO rewards(family_id, name, points_cost, stock, status, created_by)
		VALUES(@p1, @p2, @p3, @p4, @p5, @p6)
	`
	for _, preset := range defaultRewardPresets() {
		db.MustExecute(
			insertRewardSql,
			familyId,
			preset.Name,
			preset.PointsCost,
			preset.Stock,
			model.RewardStatusActive,
			creatorMemberId,
		)
	}
}
```

- [ ] **Step 4: 在创建家庭事务中调用**

在 `memberId := int64(*memberIdValue)` 后、`tran.MustCommit()` 前加入：

```go
	createDefaultFamilyPresets(tran, familyId, memberId)
```

- [ ] **Step 5: 运行 preset 单元测试**

Run:

```powershell
go test ./srv -run "TestDefault(Task|Reward)PresetsAreValid" -count=1
```

Working directory: `parent-child-api`

Expected: 测试通过。

- [ ] **Step 6: 提交实现**

Run:

```powershell
git add parent-child-api\srv\family_service.go parent-child-api\srv\family_service_test.go
git commit -m "feat: 增加新家庭预置任务和奖品"
```

---

### Task 3: 补集成测试断言

**Files:**
- Modify: `parent-child-api/srv/integration_flow_test.go`

- [ ] **Step 1: 在创建家庭后断言默认任务和奖品**

在 `familyId := family.Family.Id` 和 `t.Cleanup(...)` 之后加入：

```go
	defaultTasks := TaskService.ListTasks(ownerUserId, familyId)
	if len(defaultTasks) != len(defaultTaskPresets()) {
		t.Fatalf("default task count = %d, want %d", len(defaultTasks), len(defaultTaskPresets()))
	}
	for _, task := range defaultTasks {
		if task.CreatedBy != family.Member.Id {
			t.Fatalf("default task createdBy = %d, want owner member %d", task.CreatedBy, family.Member.Id)
		}
	}

	defaultRewards := RewardService.ListRewards(ownerUserId, familyId)
	if len(defaultRewards) != len(defaultRewardPresets()) {
		t.Fatalf("default reward count = %d, want %d", len(defaultRewards), len(defaultRewardPresets()))
	}
	for _, reward := range defaultRewards {
		if reward.CreatedBy != family.Member.Id {
			t.Fatalf("default reward createdBy = %d, want owner member %d", reward.CreatedBy, family.Member.Id)
		}
	}
```

- [ ] **Step 2: 运行普通测试**

Run:

```powershell
go test ./... -count=1
```

Working directory: `parent-child-api`

Expected: 非集成测试通过，集成测试在未设置 `PARENT_CHILD_DB_INTEGRATION=1` 时跳过。

- [ ] **Step 3: 提交集成测试**

Run:

```powershell
git add parent-child-api\srv\integration_flow_test.go
git commit -m "test: 增加新家庭预置数据集成断言"
```

---

### Task 4: 最终验证

**Files:**
- Verify: `parent-child-api/srv/family_service.go`
- Verify: `parent-child-api/srv/family_service_test.go`
- Verify: `parent-child-api/srv/integration_flow_test.go`

- [ ] **Step 1: 运行服务端测试**

Run:

```powershell
go test ./... -count=1
```

Working directory: `parent-child-api`

Expected: 所有 package 显示 `ok` 或 `?`，命令退出码为 `0`。如果沙箱因为 Go build cache 无权限失败，使用同一命令申请提升权限后重跑。

- [ ] **Step 2: 人工路径验证**

在小程序中验证：

1. 创建新家庭。
2. 进入首页，能看到默认任务。
3. 进入奖励页，能看到默认奖品。
4. 创建虚拟孩子后，可代孩子领取默认任务。
5. 给虚拟孩子获得积分后，可代孩子兑换默认奖品。

---

## 自检结果

- PRD 中“默认预置任务和奖品”已覆盖。
- 预置数据只在新家庭创建事务内生成，不影响已有家庭。
- 本计划不新增接口，不修改前端。
- 单元测试覆盖 preset 定义，集成测试覆盖创建家庭后的可查询性和 `created_by`。
