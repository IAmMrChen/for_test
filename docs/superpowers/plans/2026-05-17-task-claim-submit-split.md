# Task Claim Submit Split Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 MVP 任务流程从“提交即创建待审核记录”补齐为“领取任务 -> 提交任务 -> 家长审核”的清晰状态流。

**Architecture:** 继续沿用 `parent-child-api` 的 `model -> srv -> api` 分层。`task_records.status` 增加 `CLAIMED`，`ClaimTask` 创建已领取记录，`SubmitTask` 优先按 `recordId` 把已领取记录提交为 `PENDING`，同时保留按 `taskId` 直接提交的兼容路径。重复领取/提交通过 service 查询同一孩子同一任务下 `CLAIMED/PENDING` 记录来拦截。

**Tech Stack:** Go 1.25, net/http, github.com/bunnier/sqlmer, MySQL, 原生 SQL, go test。

---

## 范围

本计划只处理任务领取和提交语义，不处理奖励兑换，不实现每日/每周周期窗口去重，也不引入任务完成证明上传。

需要完成：

* 新增任务记录状态 `CLAIMED`。
* 新增领取任务接口 `POST /api/task/claim`。
* `SubmitTask` 支持 `recordId` 提交已领取记录。
* `SubmitTask` 保留旧的 `taskId` 直接提交，方便家长代虚拟孩子和现有调用继续可用。
* 同一孩子同一任务不能重复存在 `CLAIMED/PENDING` 记录。
* 审核只能审核 `PENDING` 记录，不能审核 `CLAIMED`。
* 集成测试覆盖领取、提交、审核、积分入账完整流程。

## 文件结构

```text
parent-child-api/
  api/
    routes.go
    routes_test.go
    task/task_api.go
  model/
    task_model.go
    task_model_test.go
  srv/
    task_service.go
    task_service_test.go
    task_integration_test.go
    integration_flow_test.go
  init.sql
```

## API

```text
POST /api/task/claim
POST /api/task/submit
```

`POST /api/task/claim` 请求体：

```json
{
  "familyId": 1,
  "taskId": 2,
  "memberId": 3
}
```

`memberId` 为空时表示当前孩子自己领取；家长代虚拟孩子领取时必须传目标虚拟孩子的 `memberId`。

`POST /api/task/submit` 请求体：

```json
{
  "familyId": 1,
  "recordId": 9,
  "submitRemark": "done"
}
```

兼容旧请求：

```json
{
  "familyId": 1,
  "taskId": 2,
  "memberId": 3,
  "submitRemark": "done"
}
```

---

### Task 1: 模型状态与 DTO

**Files:**
- Modify: `D:\projects\github\for_test\parent-child-api\model\task_model.go`
- Modify: `D:\projects\github\for_test\parent-child-api\model\task_model_test.go`

- [ ] **Step 1: Write the failing test**

在 `model/task_model_test.go` 中增加：

```go
func TestTaskRecordStatusCanSubmit(t *testing.T) {
	if !TaskRecordStatusClaimed.CanSubmit() {
		t.Fatal("claimed record should be submittable")
	}
	if TaskRecordStatusPending.CanSubmit() {
		t.Fatal("pending record should not be submittable")
	}
	if TaskRecordStatusApproved.CanSubmit() {
		t.Fatal("approved record should not be submittable")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./model
```

Expected: FAIL with `undefined: TaskRecordStatusClaimed`.

- [ ] **Step 3: Implement minimal model changes**

在 `model/task_model.go` 中：

```go
const (
	TaskRecordStatusClaimed  TaskRecordStatus = "CLAIMED"
	TaskRecordStatusPending  TaskRecordStatus = "PENDING"
	TaskRecordStatusApproved TaskRecordStatus = "APPROVED"
	TaskRecordStatusRejected TaskRecordStatus = "REJECTED"
)

func (x TaskRecordStatus) CanSubmit() bool {
	return x == TaskRecordStatusClaimed
}

type TaskClaimRequest struct {
	FamilyId int64 `json:"familyId"`
	TaskId   int64 `json:"taskId"`
	MemberId int64 `json:"memberId"`
}

type TaskSubmitRequest struct {
	FamilyId     int64  `json:"familyId"`
	RecordId     int64  `json:"recordId"`
	TaskId       int64  `json:"taskId"`
	MemberId     int64  `json:"memberId"`
	SubmitRemark string `json:"submitRemark"`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./model
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add parent-child-api/model/task_model.go parent-child-api/model/task_model_test.go
git commit -m "feat: add claimed task record status"
```

---

### Task 2: Schema 与集成迁移辅助

**Files:**
- Modify: `D:\projects\github\for_test\init.sql`
- Modify: `D:\projects\github\for_test\parent-child-api\srv\integration_flow_test.go`

- [ ] **Step 1: Update schema**

将 `task_records.status` enum 改为：

```sql
ENUM('CLAIMED', 'PENDING', 'APPROVED', 'REJECTED')
```

在 `ensureIntegrationSchema` 中增加：

```go
resx.Db.Main.MustExecute(`
	ALTER TABLE task_records
	MODIFY COLUMN status ENUM('CLAIMED', 'PENDING', 'APPROVED', 'REJECTED') NOT NULL DEFAULT 'PENDING'
`)
```

执行前使用 `integrationEnumColumnContains("task_records", "status", "CLAIMED")` 判断是否需要迁移。

- [ ] **Step 2: Add enum helper**

在 `integration_flow_test.go` 中增加：

```go
func integrationEnumColumnContains(table string, column string, value string) bool {
	return integrationCount(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA=DATABASE()
			AND TABLE_NAME=@p1
			AND COLUMN_NAME=@p2
			AND COLUMN_TYPE LIKE @p3
	`, table, column, "%'"+value+"'%") > 0
}
```

- [ ] **Step 3: Verify SQL text**

Run:

```powershell
Select-String -Path init.sql -Pattern "CLAIMED"
```

Expected: output contains `CLAIMED`.

- [ ] **Step 4: Commit**

```powershell
git add init.sql parent-child-api/srv/integration_flow_test.go
git commit -m "feat: support claimed task record schema"
```

---

### Task 3: TaskService 领取任务

**Files:**
- Modify: `D:\projects\github\for_test\parent-child-api\srv\task_service.go`
- Modify: `D:\projects\github\for_test\parent-child-api\srv\task_service_test.go`

- [ ] **Step 1: Write the failing test**

在 `srv/task_service_test.go` 中增加：

```go
func TestTaskServiceClaimTaskRejectsMissingFamily(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("ClaimTask should panic")
		}
		if fmt.Sprint(v) != "family id is required" {
			t.Fatalf("panic = %v, want family id is required", v)
		}
	}()

	TaskService.ClaimTask(1, model.TaskClaimRequest{TaskId: 1})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: FAIL with `TaskService.ClaimTask undefined`.

- [ ] **Step 3: Implement claim**

在 `srv/task_service.go` 中增加：

```go
func (x taskService) ClaimTask(operatorUserId int64, req model.TaskClaimRequest) model.TaskRecord {
	if req.FamilyId == 0 {
		panic(fmt.Errorf("family id is required"))
	}
	if req.TaskId == 0 {
		panic(fmt.Errorf("task id is required"))
	}

	target := x.resolveTaskTargetMember(operatorUserId, req.FamilyId, req.MemberId)
	task := x.loadActiveTask(req.FamilyId, req.TaskId)
	x.requireNoOpenTaskRecord(req.FamilyId, task.Id, target.Id)

	const sql = `
		INSERT INTO task_records(family_id, task_id, member_id, status)
		VALUES(@p1, @p2, @p3, @p4)
	`
	resx.Db.Main.MustExecute(sql, req.FamilyId, task.Id, target.Id, model.TaskRecordStatusClaimed)

	recordIdValue, ok := resx.Db.Main.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || recordIdValue == nil {
		panic(fmt.Errorf("failed to load created task record id"))
	}

	return model.TaskRecord{
		Id:       int64(*recordIdValue),
		FamilyId: req.FamilyId,
		TaskId:   task.Id,
		MemberId: target.Id,
		Status:   model.TaskRecordStatusClaimed,
	}
}
```

新增 helper：

```go
func (x taskService) resolveTaskTargetMember(operatorUserId, familyId, memberId int64) model.FamilyMember {
	operator := MemberService.LoadActiveMember(operatorUserId, familyId)
	if operator == nil {
		panic(fmt.Errorf("permission denied"))
	}

	target := *operator
	if memberId != 0 && memberId != operator.Id {
		member := MemberService.LoadActiveMemberById(memberId)
		if member == nil || member.FamilyId != familyId || member.RoleType != model.FamilyRoleChild {
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
	return target
}

func (x taskService) requireNoOpenTaskRecord(familyId, taskId, memberId int64) {
	const sql = `
		SELECT COUNT(*)
		FROM task_records
		WHERE family_id=@p1
			AND task_id=@p2
			AND member_id=@p3
			AND status IN (@p4, @p5)
	`
	count, ok := resx.Db.Main.MustScalarInt(sql, familyId, taskId, memberId, model.TaskRecordStatusClaimed, model.TaskRecordStatusPending)
	if ok && count != nil && *count > 0 {
		panic(fmt.Errorf("task record already exists"))
	}
}
```

- [ ] **Step 4: Refactor SubmitTask to use target helper**

将 `SubmitTask` 中现有成员解析替换为：

```go
target := x.resolveTaskTargetMember(operatorUserId, req.FamilyId, req.MemberId)
```

直接提交创建 `PENDING` 记录前调用：

```go
x.requireNoOpenTaskRecord(req.FamilyId, task.Id, target.Id)
```

- [ ] **Step 5: Run tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add parent-child-api/srv/task_service.go parent-child-api/srv/task_service_test.go
git commit -m "feat: add task claim service"
```

---

### Task 4: SubmitTask 支持 recordId

**Files:**
- Modify: `D:\projects\github\for_test\parent-child-api\srv\task_service.go`
- Modify: `D:\projects\github\for_test\parent-child-api\srv\task_service_test.go`

- [ ] **Step 1: Write the failing test**

在 `srv/task_service_test.go` 中增加：

```go
func TestTaskServiceSubmitTaskRejectsMissingRecordAndTask(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("SubmitTask should panic")
		}
		if fmt.Sprint(v) != "task id or record id is required" {
			t.Fatalf("panic = %v, want task id or record id is required", v)
		}
	}()

	TaskService.SubmitTask(1, model.TaskSubmitRequest{FamilyId: 1})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: FAIL because current code panics with `task id is required`.

- [ ] **Step 3: Implement recordId submit path**

在 `SubmitTask` 开头保留 `familyId` 校验后改为：

```go
if req.RecordId != 0 {
	return x.submitClaimedTaskRecord(operatorUserId, req)
}
if req.TaskId == 0 {
	panic(fmt.Errorf("task id or record id is required"))
}
```

增加：

```go
func (x taskService) submitClaimedTaskRecord(operatorUserId int64, req model.TaskSubmitRequest) model.TaskRecord {
	tran := resx.Db.Main.MustCreateTransactionEx()
	defer tran.MustClose()

	record := loadClaimedTaskRecordForUpdate(tran, req.RecordId)
	operator := MemberService.LoadActiveMember(operatorUserId, record.FamilyId)
	if operator == nil {
		panic(fmt.Errorf("permission denied"))
	}
	if operator.Id != record.MemberId {
		member := MemberService.LoadActiveMemberById(record.MemberId)
		if member == nil || member.FamilyId != record.FamilyId || !memberCanSubmitForChild(*operator, member.IsVirtual) {
			panic(fmt.Errorf("permission denied"))
		}
	}

	submitRemark := strings.TrimSpace(req.SubmitRemark)
	affected := tran.MustExecute(`
		UPDATE task_records
		SET status=@p1, submit_remark=@p2, submit_time=NOW()
		WHERE id=@p3 AND status=@p4
	`, model.TaskRecordStatusPending, submitRemark, req.RecordId, model.TaskRecordStatusClaimed)
	if affected != 1 {
		panic(fmt.Errorf("task record has been changed"))
	}
	tran.MustCommit()

	record.Status = model.TaskRecordStatusPending
	record.SubmitRemark = submitRemark
	return record
}

func loadClaimedTaskRecordForUpdate(db structGetter, recordId int64) model.TaskRecord {
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
	ok := db.MustGetStruct(record, sql, recordId, model.TaskRecordStatusClaimed)
	if !ok {
		panic(fmt.Errorf("task record not found"))
	}
	return *record
}
```

- [ ] **Step 4: Run tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add parent-child-api/srv/task_service.go parent-child-api/srv/task_service_test.go
git commit -m "feat: submit claimed task records"
```

---

### Task 5: API 与路由

**Files:**
- Modify: `D:\projects\github\for_test\parent-child-api\api\task\task_api.go`
- Modify: `D:\projects\github\for_test\parent-child-api\api\routes.go`
- Modify: `D:\projects\github\for_test\parent-child-api\api\routes_test.go`

- [ ] **Step 1: Write route test**

在 `api/routes_test.go` 中增加：

```go
func TestRegisterRoutesTaskClaimRejectsInvalidJSON(t *testing.T) {
	const jwtSecret = "test-secret"
	token, err := ux.JwtUtil.SignAuthToken(ux.AuthTokenClaims{UserId: 1}, jwtSecret, time.Hour)
	if err != nil {
		t.Fatalf("SignAuthToken error = %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, jwtSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/task/claim", strings.NewReader("{"))
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
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./api -run TestRegisterRoutesTaskClaimRejectsInvalidJSON -count=1
```

Expected: FAIL because `/api/task/claim` returns 404.

- [ ] **Step 3: Add API handler**

在 `api/task/task_api.go` 中增加：

```go
func (x TaskApi) Claim(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.TaskClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.TaskService.ClaimTask(token.UserId, req))
}
```

在 `api/routes.go` 中注册：

```go
mux.HandleFunc("POST /api/task/claim", apix.WithAuth(jwtSecret, taskApi.Claim))
```

- [ ] **Step 4: Run tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add parent-child-api/api/task/task_api.go parent-child-api/api/routes.go parent-child-api/api/routes_test.go
git commit -m "feat: add task claim api"
```

---

### Task 6: 集成测试

**Files:**
- Modify: `D:\projects\github\for_test\parent-child-api\srv\task_integration_test.go`

- [ ] **Step 1: Update integration test**

在 `TestIntegrationTaskPointsFlow` 中将直接提交改成领取后提交：

```go
claimed := TaskService.ClaimTask(childUserId, model.TaskClaimRequest{
	FamilyId: familyId,
	TaskId:   task.Id,
})
if claimed.Status != model.TaskRecordStatusClaimed {
	t.Fatalf("claimed status = %s, want CLAIMED", claimed.Status)
}

record := TaskService.SubmitTask(childUserId, model.TaskSubmitRequest{
	FamilyId: familyId,
	RecordId: claimed.Id,
})
if record.Status != model.TaskRecordStatusPending {
	t.Fatalf("record status = %s, want PENDING", record.Status)
}
```

增加重复领取断言：

```go
mustPanicWith(t, "task record already exists", func() {
	TaskService.ClaimTask(childUserId, model.TaskClaimRequest{
		FamilyId: familyId,
		TaskId:   task.Id,
	})
})
```

- [ ] **Step 2: Run normal tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 3: Run database integration test**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'
$env:PARENT_CHILD_DB_INTEGRATION='1'
go test ./srv -run TestIntegrationTaskPointsFlow -count=1 -v
```

Expected: PASS.

- [ ] **Step 4: Commit**

```powershell
git add parent-child-api/srv/task_integration_test.go
git commit -m "test: cover task claim submit flow"
```

---

## Self-Review

Spec coverage:

* 领取任务：Task 3 + Task 5。
* 提交已领取任务：Task 4。
* 兼容旧直接提交：Task 4 保留 `taskId` 路径。
* 防重复 `CLAIMED/PENDING`：Task 3。
* 审核仍只处理 `PENDING`：现有 `loadPendingTaskRecordForUpdate` 保持不变，Task 1 明确 `CLAIMED` 不可审核。
* 集成验证：Task 6。

Placeholder scan:

* 没有占位词。
* 没有依赖未定义的类型或方法；新类型和新方法均在前置任务定义。

Type consistency:

* `TaskRecordStatusClaimed` 与 SQL `CLAIMED` 一致。
* `TaskClaimRequest.memberId` 与 `TaskSubmitRequest.memberId` 语义一致，均用于家长代虚拟孩子。
* `recordId` 只用于提交已领取记录，不用于审核以外的旧接口。
