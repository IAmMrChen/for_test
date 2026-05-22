# 周期任务提交限制 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `ONCE`、`DAILY`、`WEEKLY` 任务真正按周期限制同一孩子的有效提交记录。

**Architecture:** 后端只改 `TaskService` 的领取/直接提交前置校验，复用现有 `task_records.created_at`、`status` 和任务 `cycle_type`。不改数据库结构，不改 API 入参出参。

**Tech Stack:** Go、MySQL 原生 SQL、项目现有 `resx.Db.Main` 包。

---

## 文件结构

- Modify: `parent-child-api/srv/task_service.go`
  - 替换 `requireNoOpenTaskRecord` 为按任务周期判断的 `requireNoEffectiveTaskRecord`。
  - 增加 `taskRecordCycleCondition` helper。
- Modify: `parent-child-api/srv/task_service_test.go`
  - 增加周期 SQL 条件单元测试。
- Modify: `parent-child-api/srv/task_integration_test.go`
  - 增加每日任务审核通过后同日不可重复领取、驳回后可重试、一次性任务不可重复直接提交的集成测试。

---

### Task 1: 周期限制红测

**Files:**
- Modify: `parent-child-api/srv/task_service_test.go`
- Modify: `parent-child-api/srv/task_integration_test.go`

- [ ] **Step 1: 写周期条件单测**

在 `task_service_test.go` 增加：

```go
func TestTaskRecordCycleCondition(t *testing.T) {
	cases := []struct {
		name      string
		cycleType model.TaskCycleType
		want      string
	}{
		{name: "once", cycleType: model.TaskCycleTypeOnce, want: ""},
		{name: "daily", cycleType: model.TaskCycleTypeDaily, want: "DATE(created_at)=CURRENT_DATE()"},
		{name: "weekly", cycleType: model.TaskCycleTypeWeekly, want: "YEARWEEK(created_at, 1)=YEARWEEK(CURRENT_DATE(), 1)"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := taskRecordCycleCondition(tt.cycleType); got != tt.want {
				t.Fatalf("condition = %q, want %q", got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: 写每日任务同日不可重复领取集成测试**

在 `task_integration_test.go` 增加测试：

```go
func TestIntegrationDailyTaskRejectsSameDayEffectiveRecord(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed + 500
	childUserId := seed + 501

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("daily-task-family-%d", seed),
		Nickname: "owner",
	})
	familyId := family.Family.Id
	t.Cleanup(func() {
		cleanupIntegrationFamily(t, familyId)
	})

	invite := InviteService.CreateInvite(ownerUserId, model.FamilyInviteCreateRequest{
		FamilyId:   familyId,
		TargetRole: model.FamilyRoleChild,
	})
	InviteService.AcceptInvite(childUserId, invite.Token)

	task := TaskService.CreateTask(ownerUserId, model.TaskCreateRequest{
		FamilyId:  familyId,
		Title:     "Daily reading",
		Points:    3,
		CycleType: model.TaskCycleTypeDaily,
	})

	first := TaskService.SubmitTask(childUserId, model.TaskSubmitRequest{
		FamilyId: familyId,
		TaskId:   task.Id,
	})
	TaskService.AuditTask(ownerUserId, model.TaskAuditRequest{
		RecordId: first.Id,
		Approved: true,
	})

	mustPanicWith(t, "task record already exists", func() {
		TaskService.ClaimTask(childUserId, model.TaskClaimRequest{
			FamilyId: familyId,
			TaskId:   task.Id,
		})
	})
}
```

- [ ] **Step 3: 写驳回后同日可重试集成测试**

在同一文件增加：

```go
func TestIntegrationDailyTaskAllowsRetryAfterRejected(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed + 510
	childUserId := seed + 511

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("daily-task-retry-family-%d", seed),
		Nickname: "owner",
	})
	familyId := family.Family.Id
	t.Cleanup(func() {
		cleanupIntegrationFamily(t, familyId)
	})

	invite := InviteService.CreateInvite(ownerUserId, model.FamilyInviteCreateRequest{
		FamilyId:   familyId,
		TargetRole: model.FamilyRoleChild,
	})
	InviteService.AcceptInvite(childUserId, invite.Token)

	task := TaskService.CreateTask(ownerUserId, model.TaskCreateRequest{
		FamilyId:  familyId,
		Title:     "Daily retry",
		Points:    3,
		CycleType: model.TaskCycleTypeDaily,
	})

	first := TaskService.SubmitTask(childUserId, model.TaskSubmitRequest{
		FamilyId: familyId,
		TaskId:   task.Id,
	})
	TaskService.AuditTask(ownerUserId, model.TaskAuditRequest{
		RecordId: first.Id,
		Approved: false,
	})

	second := TaskService.ClaimTask(childUserId, model.TaskClaimRequest{
		FamilyId: familyId,
		TaskId:   task.Id,
	})
	if second.Status != model.TaskRecordStatusClaimed {
		t.Fatalf("second status = %s, want CLAIMED", second.Status)
	}
}
```

- [ ] **Step 4: 跑红测**

Run:

```powershell
$env:PARENT_CHILD_DB_INTEGRATION='1'; go test ./srv -run "TestTaskRecordCycleCondition|TestIntegrationDailyTaskRejectsSameDayEffectiveRecord|TestIntegrationDailyTaskAllowsRetryAfterRejected" -count=1
```

Working directory: `parent-child-api`

Expected: FAIL。单元测试因 `taskRecordCycleCondition` 未定义失败，或集成测试因重复领取未被阻止失败。

---

### Task 2: 周期限制实现

**Files:**
- Modify: `parent-child-api/srv/task_service.go`

- [ ] **Step 1: 增加周期 SQL helper**

在 `resolveTaskTargetMember` 后增加：

```go
func taskRecordCycleCondition(cycleType model.TaskCycleType) string {
	switch cycleType {
	case model.TaskCycleTypeDaily:
		return "DATE(created_at)=CURRENT_DATE()"
	case model.TaskCycleTypeWeekly:
		return "YEARWEEK(created_at, 1)=YEARWEEK(CURRENT_DATE(), 1)"
	default:
		return ""
	}
}
```

- [ ] **Step 2: 替换校验方法**

将 `requireNoOpenTaskRecord` 替换为：

```go
func (x taskService) requireNoEffectiveTaskRecord(familyId int64, task model.Task, memberId int64) {
	sql := `
		SELECT COUNT(*)
		FROM task_records
		WHERE family_id=@p1
			AND task_id=@p2
			AND member_id=@p3
			AND status IN (@p4, @p5, @p6)
	`
	args := []any{
		familyId,
		task.Id,
		memberId,
		model.TaskRecordStatusClaimed,
		model.TaskRecordStatusPending,
		model.TaskRecordStatusApproved,
	}

	if condition := taskRecordCycleCondition(task.CycleType); condition != "" {
		sql += " AND " + condition
	}

	count, ok := resx.Db.Main.MustScalarInt(sql, args...)
	if ok && count != nil && *count > 0 {
		panic(fmt.Errorf("task record already exists"))
	}
}
```

- [ ] **Step 3: 更新调用点**

在 `ClaimTask` 和直接提交分支中，将：

```go
x.requireNoOpenTaskRecord(req.FamilyId, task.Id, target.Id)
```

替换为：

```go
x.requireNoEffectiveTaskRecord(req.FamilyId, task, target.Id)
```

- [ ] **Step 4: 跑周期测试**

Run:

```powershell
$env:PARENT_CHILD_DB_INTEGRATION='1'; go test ./srv -run "TestTaskRecordCycleCondition|TestIntegrationDailyTaskRejectsSameDayEffectiveRecord|TestIntegrationDailyTaskAllowsRetryAfterRejected" -count=1
```

Working directory: `parent-child-api`

Expected: PASS。

---

### Task 3: 全量验证和提交

**Files:**
- Verify: `parent-child-api/srv/task_service.go`
- Verify: `parent-child-api/srv/task_service_test.go`
- Verify: `parent-child-api/srv/task_integration_test.go`

- [ ] **Step 1: 后端全量测试**

Run:

```powershell
go test ./... -count=1
```

Working directory: `parent-child-api`

Expected: exit 0。

- [ ] **Step 2: 提交实现**

```powershell
git add parent-child-api\srv\task_service.go parent-child-api\srv\task_service_test.go parent-child-api\srv\task_integration_test.go
git commit -m "feat: 增加周期任务提交限制"
```

## 自查

- 领取和直接提交都调用新周期限制。
- `APPROVED` 计入有效记录，避免重复发放积分。
- `REJECTED` 不计入有效记录，允许修正后重试。
- 没有新增数据库字段。
