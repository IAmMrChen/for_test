# Task Points Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 落地 MVP 的“家长发布任务、孩子提交任务、家长审核并发放积分”后端闭环。

**Architecture:** 继续沿用现有 `parent-child-api` 的分层方式：`api/task` 只负责 HTTP 入参和响应，`srv/task_service.go` 承载任务、审核、积分事务逻辑，`model/task_model.go` 定义 DTO 与状态枚举。数据库继续使用 `resx.Db.Main` + 原生 SQL，积分发放必须在同一事务里更新 `family_members` 与写入 `point_logs`。

**Tech Stack:** Go 1.25、`net/http`、`github.com/bunnier/sqlmer`、MySQL、现有 JWT 鉴权中间件、`go test`。

---

## 范围

本计划只覆盖任务与积分，不覆盖奖品兑换。原因是奖品兑换依赖积分余额、积分流水和孩子身份校验；先把任务发积分闭环做稳，下一份计划再做奖励/兑换。

本计划实现的能力：

* 家长创建任务。
* 家长/孩子查看当前家庭任务列表。
* 孩子提交任务。
* 家长代虚拟孩子提交任务。
* 家长审核任务，通过后给孩子加积分并写积分流水。
* 家长驳回任务。
* 查询任务提交记录。
* 真实数据库集成测试跑通完整链路。

本计划不实现：

* 奖品、兑换、库存扣减。
* 周期任务自动生成实例。
* 防止每日/每周重复提交的复杂规则。MVP 先允许同一任务多次提交，由后续周期规则计划细化。
* 图片/附件上传。
* 操作日志表。

## 既有上下文

当前已经存在：

* `model.FamilyRole`、`model.FamilyMember`、`MemberService.RequireParentRole`。
* `family_members.current_points` 和 `family_members.total_earned_points`。
* `tasks`、`task_records`、`point_logs` 三张表雏形。
* `api` 已按二级职能目录拆分，如 `api/family`、`api/member`、`api/invite`。
* `srv/integration_flow_test.go` 已用 `PARENT_CHILD_DB_INTEGRATION=1` 模式跑通家庭/成员/邀请真实 DB 链路。

需要注意：

* `tasks.created_by`、`task_records.member_id`、`task_records.audit_by` 按“家庭成员 ID”使用。
* `family_members.created_by` 在前一轮已约定为“用户 ID”，不要混用到任务表语义里。
* 任务审核通过时必须事务内完成：
  * `task_records.status = APPROVED`
  * `family_members.current_points += task.points`
  * `family_members.total_earned_points += task.points`
  * `point_logs` 写入正向流水

## 文件结构

```text
parent-child-api/
  api/
    routes.go
    routes_test.go
    task/
      task_api.go
  model/
    task_model.go
    task_model_test.go
  srv/
    member_service.go
    member_service_test.go
    task_service.go
    task_service_test.go
    task_integration_test.go
  init.sql
```

## 接口草案

```text
POST /api/task/create
GET  /api/task/list?familyId=1
POST /api/task/submit
POST /api/task/audit
GET  /api/task/records?familyId=1&status=PENDING
```

请求/响应由 `apix.WriteData` 统一包装。业务错误继续按现有 service 风格 `panic(fmt.Errorf(...))`，后续再统一错误恢复中间件。

---

### Task 1: 任务模型与状态枚举

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\model\task_model.go`
- Create: `D:\projects\github\for_test\parent-child-api\model\task_model_test.go`

- [ ] **Step 1: Write the failing test**

创建 `model/task_model_test.go`：

```go
package model

import "testing"

func TestTaskRecordStatusCanAudit(t *testing.T) {
	if !TaskRecordStatusPending.CanAudit() {
		t.Fatal("pending record should be auditable")
	}
	if TaskRecordStatusApproved.CanAudit() {
		t.Fatal("approved record should not be auditable")
	}
	if TaskRecordStatusRejected.CanAudit() {
		t.Fatal("rejected record should not be auditable")
	}
}

func TestTaskCycleTypeValid(t *testing.T) {
	if !TaskCycleTypeOnce.Valid() || !TaskCycleTypeDaily.Valid() || !TaskCycleTypeWeekly.Valid() {
		t.Fatal("known cycle types should be valid")
	}
	if TaskCycleType("MONTHLY").Valid() {
		t.Fatal("unknown cycle type should be invalid")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./model
```

Expected: FAIL with `undefined: TaskRecordStatusPending`.

- [ ] **Step 3: Write minimal implementation**

创建 `model/task_model.go`：

```go
package model

import "time"

type TaskCycleType string

const (
	TaskCycleTypeOnce   TaskCycleType = "ONCE"
	TaskCycleTypeDaily  TaskCycleType = "DAILY"
	TaskCycleTypeWeekly TaskCycleType = "WEEKLY"
)

func (x TaskCycleType) Valid() bool {
	return x == TaskCycleTypeOnce || x == TaskCycleTypeDaily || x == TaskCycleTypeWeekly
}

type TaskStatus int

const (
	TaskStatusArchived TaskStatus = 0
	TaskStatusActive   TaskStatus = 1
)

type TaskRecordStatus string

const (
	TaskRecordStatusPending  TaskRecordStatus = "PENDING"
	TaskRecordStatusApproved TaskRecordStatus = "APPROVED"
	TaskRecordStatusRejected TaskRecordStatus = "REJECTED"
)

func (x TaskRecordStatus) CanAudit() bool {
	return x == TaskRecordStatusPending
}

type PointSourceType string

const (
	PointSourceTypeTask   PointSourceType = "TASK"
	PointSourceTypeReward PointSourceType = "REWARD"
	PointSourceTypeAdjust PointSourceType = "ADJUST"
)

type Task struct {
	Id        int64         `json:"id"`
	FamilyId  int64         `json:"familyId"`
	Title     string        `json:"title"`
	Points    int           `json:"points"`
	CycleType TaskCycleType `json:"cycleType"`
	Status    TaskStatus    `json:"status"`
	CreatedBy int64         `json:"createdBy"`
}

type TaskCreateRequest struct {
	FamilyId   int64         `json:"familyId"`
	Title      string        `json:"title"`
	Points     int           `json:"points"`
	CycleType  TaskCycleType `json:"cycleType"`
}

type TaskListRequest struct {
	FamilyId int64 `json:"familyId"`
}

type TaskSubmitRequest struct {
	FamilyId     int64  `json:"familyId"`
	TaskId       int64  `json:"taskId"`
	MemberId     int64  `json:"memberId"`
	SubmitRemark string `json:"submitRemark"`
}

type TaskAuditRequest struct {
	RecordId    int64  `json:"recordId"`
	Approved    bool   `json:"approved"`
	AuditRemark string `json:"auditRemark"`
}

type TaskRecordListRequest struct {
	FamilyId int64            `json:"familyId"`
	Status   TaskRecordStatus `json:"status"`
}

type TaskRecord struct {
	Id          int64            `json:"id"`
	FamilyId    int64            `json:"familyId"`
	TaskId      int64            `json:"taskId"`
	MemberId    int64            `json:"memberId"`
	Status      TaskRecordStatus `json:"status"`
	SubmitRemark string           `json:"submitRemark"`
	SubmitTime  time.Time        `json:"submitTime"`
	AuditTime   *time.Time       `json:"auditTime"`
	AuditBy     *int64           `json:"auditBy"`
	AuditRemark string           `json:"auditRemark"`
}

type TaskRecordListItem struct {
	Id          int64            `json:"id"`
	FamilyId    int64            `json:"familyId"`
	TaskId      int64            `json:"taskId"`
	TaskTitle   string           `json:"taskTitle"`
	Points      int              `json:"points"`
	MemberId    int64            `json:"memberId"`
	Nickname    string           `json:"nickname"`
	Status      TaskRecordStatus `json:"status"`
	SubmitRemark string           `json:"submitRemark"`
	SubmitTime  time.Time        `json:"submitTime"`
	AuditTime   *time.Time       `json:"auditTime"`
	AuditBy     *int64           `json:"auditBy"`
	AuditRemark string           `json:"auditRemark"`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./model
```

Expected: PASS.

---

### Task 2: 对齐任务表与积分表 schema

**Files:**
- Modify: `D:\projects\github\for_test\init.sql`

- [ ] **Step 1: Update schema**

在 `tasks` 表中补充更适合查询的索引：

```sql
KEY `idx_family_status` (`family_id`, `status`)
```

在 `task_records` 表中补充以下字段与索引：

```sql
  `submit_remark` VARCHAR(255) DEFAULT NULL COMMENT '提交说明',
  KEY `idx_task_member_status` (`task_id`, `member_id`, `status`),
  KEY `idx_family_status` (`family_id`, `status`)
```

在 `point_logs` 表中补充家庭内查询与幂等辅助索引：

```sql
  KEY `idx_family_member` (`family_id`, `member_id`),
  KEY `idx_source` (`source_type`, `source_id`)
```

注意不要删除已有字段；如果需要替换重复索引，先确认不会丢失查询能力。

- [ ] **Step 2: Verify SQL text contains required objects**

Run:

```powershell
Select-String -Path init.sql -Pattern "submit_remark","idx_task_member_status","idx_source","idx_family_status"
```

Expected: output contains all four patterns.

---

### Task 3: 成员权限辅助方法

**Files:**
- Modify: `D:\projects\github\for_test\parent-child-api\srv\member_service.go`
- Modify: `D:\projects\github\for_test\parent-child-api\srv\member_service_test.go`

- [ ] **Step 1: Write the failing test**

在 `srv/member_service_test.go` 增加：

```go
func TestMemberRoleHelpers(t *testing.T) {
	parent := model.FamilyMember{RoleType: model.FamilyRoleParent}
	child := model.FamilyMember{RoleType: model.FamilyRoleChild}

	if !memberCanSubmitForChild(parent, true) {
		t.Fatal("parent should submit for virtual child")
	}
	if memberCanSubmitForChild(child, true) {
		t.Fatal("child should not submit for virtual child through parent helper")
	}
	if memberCanSubmitForChild(parent, false) {
		t.Fatal("parent should not submit for real child through this helper")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: FAIL with `undefined: memberCanSubmitForChild`.

- [ ] **Step 3: Implement helpers**

在 `srv/member_service.go` 增加：

```go
func (x memberService) LoadActiveMemberById(memberId int64) *model.FamilyMember {
	const sql = `
		SELECT id
			, family_id
			, user_id
			, role_type
			, nickname
			, current_points
			, total_earned_points
			, is_virtual
			, status
		FROM family_members
		WHERE id=@p1 AND status=@p2
	`

	member := &model.FamilyMember{}
	ok := resx.Db.Main.MustGetStruct(member, sql, memberId, model.FamilyMemberStatusActive)
	if !ok {
		return nil
	}
	return member
}

func memberCanSubmitForChild(operator model.FamilyMember, targetIsVirtual bool) bool {
	return targetIsVirtual && operator.RoleType.IsParentRole()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: PASS.

---

### Task 4: 任务服务 - 创建与列表

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\srv\task_service.go`
- Create: `D:\projects\github\for_test\parent-child-api\srv\task_service_test.go`

- [ ] **Step 1: Write the failing test**

创建 `srv/task_service_test.go`：

```go
package srv

import (
	"fmt"
	"testing"

	"parent-child-api/model"
	"parent-child-api/resx"
)

func TestTaskServiceCreateTaskPanicsWhenTitleIsEmpty(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("CreateTask should panic")
		}
		if fmt.Sprint(v) != "task title is required" {
			t.Fatalf("panic = %v, want task title is required", v)
		}
	}()

	TaskService.CreateTask(1, model.TaskCreateRequest{FamilyId: 1, Title: " "})
}

func TestNormalizeTaskCreateRequest(t *testing.T) {
	req := normalizeTaskCreateRequest(model.TaskCreateRequest{
		FamilyId:  1,
		Title:     " Brush teeth ",
		Points:    0,
		CycleType: "",
	})

	if req.Title != "Brush teeth" {
		t.Fatalf("title = %q, want Brush teeth", req.Title)
	}
	if req.Points != 1 {
		t.Fatalf("points = %d, want 1", req.Points)
	}
	if req.CycleType != model.TaskCycleTypeOnce {
		t.Fatalf("cycleType = %s, want ONCE", req.CycleType)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: FAIL with `undefined: TaskService`.

- [ ] **Step 3: Implement create and list**

创建 `srv/task_service.go`：

```go
package srv

import (
	"fmt"
	"strings"

	"parent-child-api/model"
	"parent-child-api/resx"
)

var TaskService taskService

type taskService struct{}

func normalizeTaskCreateRequest(req model.TaskCreateRequest) model.TaskCreateRequest {
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		panic(fmt.Errorf("task title is required"))
	}
	if req.Points <= 0 {
		req.Points = 1
	}
	if req.CycleType == "" {
		req.CycleType = model.TaskCycleTypeOnce
	}
	if !req.CycleType.Valid() {
		panic(fmt.Errorf("invalid task cycle type: %s", req.CycleType))
	}
	return req
}

func (x taskService) CreateTask(operatorUserId int64, req model.TaskCreateRequest) model.Task {
	req = normalizeTaskCreateRequest(req)
	operator := MemberService.RequireParentRole(operatorUserId, req.FamilyId)

	const sql = `
		INSERT INTO tasks(family_id, title, points, cycle_type, status, created_by)
		VALUES(@p1, @p2, @p3, @p4, @p5, @p6)
	`
	resx.Db.Main.MustExecute(sql, req.FamilyId, req.Title, req.Points, req.CycleType, model.TaskStatusActive, operator.Id)

	taskIdValue, ok := resx.Db.Main.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || taskIdValue == nil {
		panic(fmt.Errorf("failed to load created task id"))
	}

	return model.Task{
		Id:        int64(*taskIdValue),
		FamilyId:  req.FamilyId,
		Title:     req.Title,
		Points:    req.Points,
		CycleType: req.CycleType,
		Status:    model.TaskStatusActive,
		CreatedBy: operator.Id,
	}
}

func (x taskService) ListTasks(userId int64, familyId int64) []model.Task {
	if MemberService.LoadActiveMember(userId, familyId) == nil {
		panic(fmt.Errorf("permission denied"))
	}

	const sql = `
		SELECT id
			, family_id
			, title
			, points
			, cycle_type
			, status
			, created_by
		FROM tasks
		WHERE family_id=@p1 AND status=@p2
		ORDER BY id DESC
	`

	return resx.Db.Main.MustListOf(model.Task{}, sql, familyId, model.TaskStatusActive).([]model.Task)
}
```

- [ ] **Step 4: Run tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: PASS.

---

### Task 5: 任务服务 - 提交任务

**Files:**
- Modify: `D:\projects\github\for_test\parent-child-api\srv\task_service.go`
- Modify: `D:\projects\github\for_test\parent-child-api\srv\task_service_test.go`

- [ ] **Step 1: Write the failing test**

在 `srv/task_service_test.go` 增加：

```go
func TestResolveSubmitMemberRejectsMissingFamily(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("SubmitTask should panic")
		}
		if fmt.Sprint(v) != "family id is required" {
			t.Fatalf("panic = %v, want family id is required", v)
		}
	}()

	TaskService.SubmitTask(1, model.TaskSubmitRequest{TaskId: 1})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: FAIL with `TaskService.SubmitTask undefined`.

- [ ] **Step 3: Implement submit**

在 `srv/task_service.go` 增加：

```go
func (x taskService) SubmitTask(operatorUserId int64, req model.TaskSubmitRequest) model.TaskRecord {
	if req.FamilyId == 0 {
		panic(fmt.Errorf("family id is required"))
	}
	if req.TaskId == 0 {
		panic(fmt.Errorf("task id is required"))
	}

	operator := MemberService.LoadActiveMember(operatorUserId, req.FamilyId)
	if operator == nil {
		panic(fmt.Errorf("permission denied"))
	}

	target := *operator
	if req.MemberId != 0 && req.MemberId != operator.Id {
		member := MemberService.LoadActiveMemberById(req.MemberId)
		if member == nil || member.FamilyId != req.FamilyId || member.RoleType != model.FamilyRoleChild {
			panic(fmt.Errorf("target child not found"))
		}
		if !memberCanSubmitForChild(*operator, member.IsVirtual) {
			panic(fmt.Errorf("permission denied"))
		}
		target = *member
	}

	if target.RoleType != model.FamilyRoleChild {
		panic(fmt.Errorf("only child can submit task"))
	}

	task := x.loadActiveTask(req.FamilyId, req.TaskId)

	const sql = `
		INSERT INTO task_records(family_id, task_id, member_id, status, submit_remark)
		VALUES(@p1, @p2, @p3, @p4, @p5)
	`
	resx.Db.Main.MustExecute(sql, req.FamilyId, task.Id, target.Id, model.TaskRecordStatusPending, strings.TrimSpace(req.SubmitRemark))

	recordIdValue, ok := resx.Db.Main.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || recordIdValue == nil {
		panic(fmt.Errorf("failed to load created task record id"))
	}

	return model.TaskRecord{
		Id:       int64(*recordIdValue),
		FamilyId: req.FamilyId,
		TaskId:   task.Id,
		MemberId: target.Id,
		Status:   model.TaskRecordStatusPending,
		SubmitRemark: strings.TrimSpace(req.SubmitRemark),
	}
}

func (x taskService) loadActiveTask(familyId, taskId int64) model.Task {
	const sql = `
		SELECT id
			, family_id
			, title
			, points
			, cycle_type
			, status
			, created_by
		FROM tasks
		WHERE id=@p1 AND family_id=@p2 AND status=@p3
	`

	task := &model.Task{}
	ok := resx.Db.Main.MustGetStruct(task, sql, taskId, familyId, model.TaskStatusActive)
	if !ok {
		panic(fmt.Errorf("task not found"))
	}
	return *task
}
```

- [ ] **Step 4: Run tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: PASS.

---

### Task 6: 任务服务 - 审核与积分事务

**Files:**
- Modify: `D:\projects\github\for_test\parent-child-api\srv\task_service.go`
- Modify: `D:\projects\github\for_test\parent-child-api\srv\task_service_test.go`

- [ ] **Step 1: Write the failing test**

在 `srv/task_service_test.go` 增加：

```go
func TestTaskAuditStatusFromApprovedFlag(t *testing.T) {
	if taskAuditStatus(true) != model.TaskRecordStatusApproved {
		t.Fatal("approved flag should map to APPROVED")
	}
	if taskAuditStatus(false) != model.TaskRecordStatusRejected {
		t.Fatal("false approved flag should map to REJECTED")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: FAIL with `undefined: taskAuditStatus`.

- [ ] **Step 3: Implement audit**

在 `srv/task_service.go` 增加：

```go
func taskAuditStatus(approved bool) model.TaskRecordStatus {
	if approved {
		return model.TaskRecordStatusApproved
	}
	return model.TaskRecordStatusRejected
}

func (x taskService) AuditTask(operatorUserId int64, req model.TaskAuditRequest) model.TaskRecord {
	if req.RecordId == 0 {
		panic(fmt.Errorf("record id is required"))
	}

	tran := resx.Db.Main.MustCreateTransactionEx()
	defer tran.MustClose()

	record := loadPendingTaskRecordForUpdate(tran, req.RecordId)
	operator := MemberService.RequireParentRole(operatorUserId, record.FamilyId)
	task := loadTaskForAudit(tran, record.TaskId, record.FamilyId)

	nextStatus := taskAuditStatus(req.Approved)
	updateSql := `
		UPDATE task_records
		SET status=@p1, audit_time=NOW(), audit_by=@p2, audit_remark=@p3
		WHERE id=@p4 AND status=@p5
	`
	affected := tran.MustExecute(updateSql, nextStatus, operator.Id, req.AuditRemark, req.RecordId, model.TaskRecordStatusPending)
	if affected != 1 {
		panic(fmt.Errorf("task record has been changed"))
	}

	if req.Approved {
		tran.MustExecute(`
			UPDATE family_members
			SET current_points=current_points+@p1
				, total_earned_points=total_earned_points+@p1
			WHERE id=@p2 AND family_id=@p3 AND role_type=@p4 AND status=@p5
		`, task.Points, record.MemberId, record.FamilyId, model.FamilyRoleChild, model.FamilyMemberStatusActive)

		tran.MustExecute(`
			INSERT INTO point_logs(family_id, member_id, points, source_type, source_id)
			VALUES(@p1, @p2, @p3, @p4, @p5)
		`, record.FamilyId, record.MemberId, task.Points, model.PointSourceTypeTask, record.Id)
	}

	tran.MustCommit()

	record.Status = nextStatus
	record.AuditBy = &operator.Id
	record.AuditRemark = req.AuditRemark
	return record
}

func loadPendingTaskRecordForUpdate(db structGetter, recordId int64) model.TaskRecord {
	const sql = `
		SELECT id
			, family_id
			, task_id
			, member_id
			, status
			, IFNULL(submit_remark, '') AS submit_remark
			, submit_time
			, audit_time
			, audit_by
			, IFNULL(audit_remark, '') AS audit_remark
		FROM task_records
		WHERE id=@p1 AND status=@p2
		FOR UPDATE
	`

	record := &model.TaskRecord{}
	ok := db.MustGetStruct(record, sql, recordId, model.TaskRecordStatusPending)
	if !ok {
		panic(fmt.Errorf("task record not found"))
	}
	return *record
}

func loadTaskForAudit(db structGetter, taskId int64, familyId int64) model.Task {
	const sql = `
		SELECT id
			, family_id
			, title
			, points
			, cycle_type
			, status
			, created_by
		FROM tasks
		WHERE id=@p1 AND family_id=@p2
	`

	task := &model.Task{}
	ok := db.MustGetStruct(task, sql, taskId, familyId)
	if !ok {
		panic(fmt.Errorf("task not found"))
	}
	return *task
}
```

Note: this step reuses the `structGetter` interface already defined in `srv/invite_service.go`.

- [ ] **Step 4: Run tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: PASS.

---

### Task 7: 任务记录列表

**Files:**
- Modify: `D:\projects\github\for_test\parent-child-api\srv\task_service.go`

- [ ] **Step 1: Add list records implementation**

在 `srv/task_service.go` 增加：

```go
func (x taskService) ListTaskRecords(userId int64, req model.TaskRecordListRequest) []model.TaskRecordListItem {
	if req.FamilyId == 0 {
		panic(fmt.Errorf("family id is required"))
	}

	member := MemberService.LoadActiveMember(userId, req.FamilyId)
	if member == nil {
		panic(fmt.Errorf("permission denied"))
	}

	sql := `
		SELECT r.id
			, r.family_id
			, r.task_id
			, t.title AS task_title
			, t.points
			, r.member_id
			, m.nickname
			, r.status
			, IFNULL(r.submit_remark, '') AS submit_remark
			, r.submit_time
			, r.audit_time
			, r.audit_by
			, IFNULL(r.audit_remark, '') AS audit_remark
		FROM task_records r
		INNER JOIN tasks t ON t.id = r.task_id
		INNER JOIN family_members m ON m.id = r.member_id
		WHERE r.family_id=@p1
	`
	args := []any{req.FamilyId}

	if !member.RoleType.IsParentRole() {
		sql += " AND r.member_id=@p2"
		args = append(args, member.Id)
	}
	if req.Status != "" {
		sql += fmt.Sprintf(" AND r.status=@p%d", len(args)+1)
		args = append(args, req.Status)
	}
	sql += " ORDER BY r.id DESC"

	return resx.Db.Main.MustListOf(model.TaskRecordListItem{}, sql, args...).([]model.TaskRecordListItem)
}
```

- [ ] **Step 2: Run compile check**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: PASS.

---

### Task 8: 任务 API 与路由

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\api\task\task_api.go`
- Modify: `D:\projects\github\for_test\parent-child-api\api\routes.go`
- Modify: `D:\projects\github\for_test\parent-child-api\api\routes_test.go`

- [ ] **Step 1: Write route test**

在 `api/routes_test.go` 增加：

```go
func TestRegisterRoutesTaskCreateRejectsInvalidJSON(t *testing.T) {
	const jwtSecret = "test-secret"
	token, err := ux.JwtUtil.SignAuthToken(ux.AuthTokenClaims{UserId: 1}, jwtSecret, time.Hour)
	if err != nil {
		t.Fatalf("SignAuthToken error = %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, jwtSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/task/create", strings.NewReader("{"))
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./api -run TestRegisterRoutesTaskCreateRejectsInvalidJSON -count=1
```

Expected: FAIL because `/api/task/create` returns 404.

- [ ] **Step 3: Add task API**

创建 `api/task/task_api.go`：

```go
package taskapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"parent-child-api/api/apix"
	"parent-child-api/model"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type TaskApi struct{}

func (x TaskApi) Create(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.TaskCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.TaskService.CreateTask(token.UserId, req))
}

func (x TaskApi) List(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}
	apix.WriteData(w, srv.TaskService.ListTasks(token.UserId, familyId))
}

func (x TaskApi) Submit(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.TaskSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.TaskService.SubmitTask(token.UserId, req))
}

func (x TaskApi) Audit(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.TaskAuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.TaskService.AuditTask(token.UserId, req))
}

func (x TaskApi) Records(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}

	req := model.TaskRecordListRequest{
		FamilyId: familyId,
		Status:   model.TaskRecordStatus(r.URL.Query().Get("status")),
	}
	apix.WriteData(w, srv.TaskService.ListTaskRecords(token.UserId, req))
}
```

- [ ] **Step 4: Register routes**

修改 `api/routes.go`：

```go
import taskapi "parent-child-api/api/task"
```

在 `RegisterRoutes` 中增加：

```go
taskApi := taskapi.TaskApi{}
mux.HandleFunc("POST /api/task/create", apix.WithAuth(jwtSecret, taskApi.Create))
mux.HandleFunc("GET /api/task/list", apix.WithAuth(jwtSecret, taskApi.List))
mux.HandleFunc("POST /api/task/submit", apix.WithAuth(jwtSecret, taskApi.Submit))
mux.HandleFunc("POST /api/task/audit", apix.WithAuth(jwtSecret, taskApi.Audit))
mux.HandleFunc("GET /api/task/records", apix.WithAuth(jwtSecret, taskApi.Records))
```

- [ ] **Step 5: Run tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./...
```

Expected: PASS.

---

### Task 9: 任务与积分真实数据库集成测试

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\srv\task_integration_test.go`
- Modify: `D:\projects\github\for_test\parent-child-api\srv\integration_flow_test.go`

- [ ] **Step 1: Extend integration schema helper**

把 `integration_flow_test.go` 中的 `ensureIntegrationSchema` 扩展为可补齐：

```sql
ALTER TABLE task_records ADD COLUMN submit_remark VARCHAR(255) DEFAULT NULL
CREATE INDEX idx_task_member_status ON task_records(task_id, member_id, status)
CREATE INDEX idx_family_status ON task_records(family_id, status)
CREATE INDEX idx_family_status ON tasks(family_id, status)
CREATE INDEX idx_family_member ON point_logs(family_id, member_id)
CREATE INDEX idx_source ON point_logs(source_type, source_id)
```

每个 `ALTER` / `CREATE INDEX` 前必须先用 `information_schema` 检查字段或索引是否存在，避免重复执行报错。

- [ ] **Step 2: Add integration test**

创建 `srv/task_integration_test.go`：

```go
package srv

import (
	"fmt"
	"os"
	"testing"
	"time"

	"parent-child-api/model"
	"parent-child-api/resx"
)

func TestIntegrationTaskPointsFlow(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed + 100
	childUserId := seed + 101

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("task-points-family-%d", seed),
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
	child := InviteService.AcceptInvite(childUserId, invite.Token)

	task := TaskService.CreateTask(ownerUserId, model.TaskCreateRequest{
		FamilyId:  familyId,
		Title:     "Brush teeth",
		Points:    3,
		CycleType: model.TaskCycleTypeDaily,
	})

	record := TaskService.SubmitTask(childUserId, model.TaskSubmitRequest{
		FamilyId: familyId,
		TaskId:   task.Id,
	})
	if record.Status != model.TaskRecordStatusPending {
		t.Fatalf("record status = %s, want PENDING", record.Status)
	}

	audited := TaskService.AuditTask(ownerUserId, model.TaskAuditRequest{
		RecordId: record.Id,
		Approved: true,
	})
	if audited.Status != model.TaskRecordStatusApproved {
		t.Fatalf("audited status = %s, want APPROVED", audited.Status)
	}

	reloaded := MemberService.LoadActiveMember(childUserId, familyId)
	if reloaded == nil {
		t.Fatal("child member not found")
	}
	if reloaded.CurrentPoints != 3 || reloaded.TotalEarnedPoints != 3 {
		t.Fatalf("points = current %d total %d, want 3/3", reloaded.CurrentPoints, reloaded.TotalEarnedPoints)
	}

	count, ok := resx.Db.Main.MustScalarInt(`
		SELECT COUNT(*)
		FROM point_logs
		WHERE family_id=@p1 AND member_id=@p2 AND points=@p3 AND source_type=@p4 AND source_id=@p5
	`, familyId, child.Id, 3, model.PointSourceTypeTask, record.Id)
	if !ok || count == nil || *count != 1 {
		t.Fatalf("point log count = %v, want 1", count)
	}
}
```

- [ ] **Step 3: Run normal tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./...
```

Expected: PASS, integration tests skipped by default.

- [ ] **Step 4: Run database integration test**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'
$env:PARENT_CHILD_DB_INTEGRATION='1'
go test ./srv -run TestIntegrationTaskPointsFlow -count=1 -v
```

Expected: PASS.

---

## Self-Review

Spec coverage:

* 家长发布任务：Task 4 + Task 8。
* 孩子提交任务：Task 5 + Task 8。
* 家长审核任务：Task 6 + Task 8。
* 审核通过发放积分：Task 6，事务内更新成员积分并写 `point_logs`。
* 真实孩子与虚拟孩子：Task 5 支持真实孩子自己提交，也支持家长代虚拟孩子提交。
* 数据库验证：Task 9。

Placeholder scan:

* 无 `TBD`、`TODO`、`later`。
* 奖品兑换明确排除在本计划外，不是遗漏。

Type consistency:

* `TaskCycleType` 与 SQL `tasks.cycle_type` 的 `ONCE/DAILY/WEEKLY` 一致。
* `TaskRecordStatus` 与 SQL `task_records.status` 的 `PENDING/APPROVED/REJECTED` 一致。
* `PointSourceTypeTask` 与 SQL `point_logs.source_type` 的 `TASK` 一致。
* `created_by` 在 `tasks` 中按成员 ID 使用；`family_members.created_by` 的用户 ID 语义不被复用到任务表。
