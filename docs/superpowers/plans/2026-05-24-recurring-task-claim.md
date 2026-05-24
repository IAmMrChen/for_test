# Recurring Task Claim Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让每日、每周循环任务支持“领取一次后持续可提交”，并允许领取端停止自己的领取关系。

**Architecture:** 新增 `task_claims` 表保存孩子对循环任务的持续领取关系，`task_records` 继续保存每次提交、审核和积分发放记录。家长端发布侧仍使用现有“归档”；领取端新增“停止领取”。前端通过 active claims 合成任务按钮状态。

**Tech Stack:** Go `net/http` API、原生 SQL via `resx.Db.Main`、uni-app/Vue 小程序、Node 行为测试、Go 单元与集成测试。

---

## File Structure

- Modify: `init.sql`，新增 `task_claims` 表。
- Modify: `db_schema.md`，补充任务领取关系说明。
- Modify: `parent-child-api/model/task_model.go`，新增领取关系模型、请求和响应类型。
- Modify: `parent-child-api/srv/task_service.go`，实现开始领取、停止领取、列出领取关系、循环任务直接提交校验。
- Modify: `parent-child-api/srv/task_service_test.go`，覆盖 helper 和模型行为。
- Modify: `parent-child-api/srv/task_integration_test.go`，覆盖领取一次、跨周期直接提交、停止领取。
- Modify: `parent-child-api/srv/integration_flow_test.go`，测试数据库辅助建表时补齐 `task_claims`。
- Modify: `parent-child-api/api/task/task_api.go`，新增 HTTP handler。
- Modify: `parent-child-api/api/routes.go`，注册 `GET /api/task/claims`、`POST /api/task/startClaim`、`POST /api/task/stopClaim`。
- Modify: `parent-child-api/api/routes_test.go`，覆盖新增路由参数校验。
- Modify: `parent-child-miniprogram/api/task.js`，新增 `listTaskClaims`、`startTaskClaim`、`stopTaskClaim`。
- Modify: `parent-child-miniprogram/pages/index/index.vue`、`parent-child-miniprogram/pages/tasks/index.vue`，合成循环任务状态。
- Modify: `parent-child-miniprogram/pages/index/index.behavior.test.mjs`、`parent-child-miniprogram/pages/tasks/index.behavior.test.mjs`，覆盖“停止领取”和不重复领取。

## Task 1: Model And Schema

**Files:**
- Modify: `init.sql`
- Modify: `db_schema.md`
- Modify: `parent-child-api/model/task_model.go`
- Test: `parent-child-api/model/task_model_test.go`

- [ ] **Step 1: Write model tests**

Add to `parent-child-api/model/task_model_test.go`:

```go
func TestTaskClaimStatusActive(t *testing.T) {
	if TaskClaimStatusActive != "ACTIVE" {
		t.Fatalf("active status = %s", TaskClaimStatusActive)
	}
	if TaskClaimStatusStopped != "STOPPED" {
		t.Fatalf("stopped status = %s", TaskClaimStatusStopped)
	}
}
```

- [ ] **Step 2: Run model test to verify it fails**

Run:

```powershell
cd parent-child-api
go test ./model -run TestTaskClaimStatusActive -count=1
```

Expected: FAIL with `undefined: TaskClaimStatusActive`.

- [ ] **Step 3: Add model types**

Add to `parent-child-api/model/task_model.go` after `TaskRecordStatus`:

```go
type TaskClaimStatus string

const (
	TaskClaimStatusActive  TaskClaimStatus = "ACTIVE"
	TaskClaimStatusStopped TaskClaimStatus = "STOPPED"
)

type TaskClaim struct {
	Id        int64           `json:"id"`
	FamilyId  int64           `json:"familyId"`
	TaskId    int64           `json:"taskId"`
	MemberId  int64           `json:"memberId"`
	Status    TaskClaimStatus `json:"status"`
	ClaimedAt time.Time       `json:"claimedAt"`
	StoppedAt *time.Time      `json:"stoppedAt"`
}

type TaskClaimRequest struct {
	FamilyId int64 `json:"familyId"`
	TaskId   int64 `json:"taskId"`
	MemberId int64 `json:"memberId"`
}

type TaskClaimListRequest struct {
	FamilyId int64 `json:"familyId"`
}

type TaskClaimListItem struct {
	Id        int64           `json:"id"`
	FamilyId  int64           `json:"familyId"`
	TaskId    int64           `json:"taskId"`
	MemberId  int64           `json:"memberId"`
	Status    TaskClaimStatus `json:"status"`
}
```

- [ ] **Step 4: Add database schema**

Add to `init.sql` after `task_records`:

```sql
CREATE TABLE IF NOT EXISTS `task_claims` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '领取关系ID',
  `family_id` BIGINT UNSIGNED NOT NULL COMMENT '家庭ID',
  `task_id` BIGINT UNSIGNED NOT NULL COMMENT '任务ID',
  `member_id` BIGINT UNSIGNED NOT NULL COMMENT '孩子成员ID',
  `status` ENUM('ACTIVE', 'STOPPED') NOT NULL DEFAULT 'ACTIVE' COMMENT '领取状态',
  `claimed_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '领取时间',
  `stopped_at` DATETIME DEFAULT NULL COMMENT '停止领取时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_claim_member` (`task_id`, `member_id`),
  KEY `idx_claim_family_member_status` (`family_id`, `member_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='循环任务持续领取关系';
```

Update `db_schema.md` with a `task_claims` table and relation from `family_members` and `tasks`.

- [ ] **Step 5: Run model test**

Run:

```powershell
cd parent-child-api
go test ./model -run TestTaskClaimStatusActive -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add init.sql db_schema.md parent-child-api/model/task_model.go parent-child-api/model/task_model_test.go
git commit -m "feat: 增加循环任务领取关系模型"
```

## Task 2: Task Claim Service

**Files:**
- Modify: `parent-child-api/srv/task_service.go`
- Modify: `parent-child-api/srv/task_service_test.go`

- [ ] **Step 1: Write helper tests**

Add to `parent-child-api/srv/task_service_test.go`:

```go
func TestRecurringTaskRequiresActiveClaim(t *testing.T) {
	task := model.Task{CycleType: model.TaskCycleTypeDaily}
	if !taskNeedsActiveClaim(task) {
		t.Fatal("daily task should require active claim for child direct submit")
	}
	task.CycleType = model.TaskCycleTypeWeekly
	if !taskNeedsActiveClaim(task) {
		t.Fatal("weekly task should require active claim for child direct submit")
	}
	task.CycleType = model.TaskCycleTypeOnce
	if taskNeedsActiveClaim(task) {
		t.Fatal("once task should not require active claim")
	}
}
```

- [ ] **Step 2: Run helper test to verify it fails**

Run:

```powershell
cd parent-child-api
go test ./srv -run TestRecurringTaskRequiresActiveClaim -count=1
```

Expected: FAIL with `undefined: taskNeedsActiveClaim`.

- [ ] **Step 3: Implement service helpers**

Add to `parent-child-api/srv/task_service.go`:

```go
func taskNeedsActiveClaim(task model.Task) bool {
	return task.CycleType == model.TaskCycleTypeDaily || task.CycleType == model.TaskCycleTypeWeekly
}

func (x taskService) requireActiveTaskClaim(familyId, taskId, memberId int64) {
	const sql = `
		SELECT COUNT(1)
		FROM task_claims
		WHERE family_id=@p1
			AND task_id=@p2
			AND member_id=@p3
			AND status=@p4
	`
	count, ok := resx.Db.Main.MustScalarInt(sql, familyId, taskId, memberId, model.TaskClaimStatusActive)
	if !ok || count == nil || *count == 0 {
		panic(fmt.Errorf("task claim is required"))
	}
}
```

- [ ] **Step 4: Add public service methods**

Add methods to `taskService`:

```go
func (x taskService) StartTaskClaim(operatorUserId int64, req model.TaskClaimRequest) model.TaskClaim {
	task := x.loadActiveTask(req.FamilyId, req.TaskId)
	if !taskNeedsActiveClaim(task) {
		panic(fmt.Errorf("only recurring task can start claim"))
	}
	target := x.resolveTaskTargetMember(operatorUserId, req.FamilyId, req.MemberId)

	sql := `
		INSERT INTO task_claims(family_id, task_id, member_id, status, claimed_at, stopped_at)
		VALUES(@p1, @p2, @p3, @p4, NOW(), NULL)
		ON DUPLICATE KEY UPDATE status=@p4, claimed_at=NOW(), stopped_at=NULL
	`
	resx.Db.Main.MustExecute(sql, req.FamilyId, req.TaskId, target.Id, model.TaskClaimStatusActive)
	return x.loadTaskClaim(req.FamilyId, req.TaskId, target.Id)
}

func (x taskService) StopTaskClaim(operatorUserId int64, req model.TaskClaimRequest) model.TaskClaim {
	task := x.loadActiveTask(req.FamilyId, req.TaskId)
	if !taskNeedsActiveClaim(task) {
		panic(fmt.Errorf("only recurring task can stop claim"))
	}
	target := x.resolveTaskTargetMember(operatorUserId, req.FamilyId, req.MemberId)
	sql := `
		UPDATE task_claims
		SET status=@p1, stopped_at=NOW()
		WHERE family_id=@p2 AND task_id=@p3 AND member_id=@p4 AND status=@p5
	`
	affected := resx.Db.Main.MustExecute(sql, model.TaskClaimStatusStopped, req.FamilyId, req.TaskId, target.Id, model.TaskClaimStatusActive)
	if affected == 0 {
		panic(fmt.Errorf("active task claim not found"))
	}
	return x.loadTaskClaim(req.FamilyId, req.TaskId, target.Id)
}

func (x taskService) ListTaskClaims(userId int64, req model.TaskClaimListRequest) []model.TaskClaimListItem {
	if req.FamilyId == 0 {
		panic(fmt.Errorf("family id is required"))
	}
	operator := MemberService.LoadActiveMember(userId, req.FamilyId)
	if operator == nil {
		panic(fmt.Errorf("permission denied"))
	}
	sql := `
		SELECT id, family_id, task_id, member_id, status
		FROM task_claims
		WHERE family_id=@p1 AND status=@p2
	`
	args := []any{req.FamilyId, model.TaskClaimStatusActive}
	if !operator.RoleType.IsParentRole() {
		sql += " AND member_id=@p3"
		args = append(args, operator.Id)
	}
	return resx.Db.Main.MustListOf(model.TaskClaimListItem{}, sql, args...).([]model.TaskClaimListItem)
}
```

Add `loadTaskClaim`:

```go
func (x taskService) loadTaskClaim(familyId, taskId, memberId int64) model.TaskClaim {
	const sql = `
		SELECT id, family_id, task_id, member_id, status, claimed_at, stopped_at
		FROM task_claims
		WHERE family_id=@p1 AND task_id=@p2 AND member_id=@p3
	`
	claim := &model.TaskClaim{}
	ok := resx.Db.Main.MustGetStruct(claim, sql, familyId, taskId, memberId)
	if !ok {
		panic(fmt.Errorf("task claim not found"))
	}
	return *claim
}
```

- [ ] **Step 5: Update direct submit rule**

In `SubmitTask`, after `task := x.loadActiveTask(req.FamilyId, req.TaskId)` and target resolution, require an active claim when the operator is the child:

```go
operator := MemberService.LoadActiveMember(operatorUserId, req.FamilyId)
if taskNeedsActiveClaim(task) && operator != nil && operator.Id == target.Id {
	x.requireActiveTaskClaim(req.FamilyId, task.Id, target.Id)
}
```

For parent proxy submit, keep the existing direct-submit flow and call `StartTaskClaim` before insert when no active claim exists.

- [ ] **Step 6: Run helper tests**

Run:

```powershell
cd parent-child-api
go test ./srv -run TestRecurringTaskRequiresActiveClaim -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```powershell
git add parent-child-api/srv/task_service.go parent-child-api/srv/task_service_test.go
git commit -m "feat: 增加循环任务领取服务"
```

## Task 3: HTTP API And Routes

**Files:**
- Modify: `parent-child-api/api/task/task_api.go`
- Modify: `parent-child-api/api/routes.go`
- Modify: `parent-child-api/api/routes_test.go`

- [ ] **Step 1: Write route test**

Add to `parent-child-api/api/routes_test.go`:

```go
func TestRegisterRoutesTaskClaimsRejectsInvalidFamilyId(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, "secret")
	req := httptest.NewRequest(http.MethodGet, "/api/task/claims?familyId=bad", nil)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest && res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want bad request or unauthorized", res.Code)
	}
}
```

- [ ] **Step 2: Run route test to verify it fails**

Run:

```powershell
cd parent-child-api
go test ./api -run TestRegisterRoutesTaskClaimsRejectsInvalidFamilyId -count=1
```

Expected: FAIL because route is not registered or handler missing.

- [ ] **Step 3: Add API handlers**

Add to `parent-child-api/api/task/task_api.go`:

```go
func (x TaskApi) StartClaim(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.TaskClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.TaskService.StartTaskClaim(token.UserId, req))
}

func (x TaskApi) StopClaim(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.TaskClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.TaskService.StopTaskClaim(token.UserId, req))
}

func (x TaskApi) Claims(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}
	apix.WriteData(w, srv.TaskService.ListTaskClaims(token.UserId, model.TaskClaimListRequest{FamilyId: familyId}))
}
```

- [ ] **Step 4: Register routes**

Add to `parent-child-api/api/routes.go` near task routes:

```go
mux.HandleFunc("GET /api/task/claims", recoverRoute(apix.WithAuth(jwtSecret, taskApi.Claims)))
mux.HandleFunc("POST /api/task/startClaim", recoverRoute(apix.WithAuth(jwtSecret, taskApi.StartClaim)))
mux.HandleFunc("POST /api/task/stopClaim", recoverRoute(apix.WithAuth(jwtSecret, taskApi.StopClaim)))
```

- [ ] **Step 5: Run route test**

Run:

```powershell
cd parent-child-api
go test ./api -run TestRegisterRoutesTaskClaimsRejectsInvalidFamilyId -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add parent-child-api/api/task/task_api.go parent-child-api/api/routes.go parent-child-api/api/routes_test.go
git commit -m "feat: 增加循环任务领取接口"
```

## Task 4: Integration Coverage

**Files:**
- Modify: `parent-child-api/srv/integration_flow_test.go`
- Modify: `parent-child-api/srv/task_integration_test.go`

- [ ] **Step 1: Add integration table helper**

In `ensureIntegrationSchema`, create `task_claims` if missing:

```go
if !integrationTableExists("task_claims") {
	resx.Db.Main.MustExecute(`
		CREATE TABLE task_claims (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			family_id BIGINT UNSIGNED NOT NULL,
			task_id BIGINT UNSIGNED NOT NULL,
			member_id BIGINT UNSIGNED NOT NULL,
			status ENUM('ACTIVE', 'STOPPED') NOT NULL DEFAULT 'ACTIVE',
			claimed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			stopped_at DATETIME DEFAULT NULL,
			UNIQUE KEY uk_task_claim_member (task_id, member_id),
			KEY idx_claim_family_member_status (family_id, member_id, status)
		)
	`)
}
```

- [ ] **Step 2: Add integration test**

Add to `parent-child-api/srv/task_integration_test.go`:

```go
func TestIntegrationRecurringTaskClaimCanSubmitAfterFirstClaim(t *testing.T) {
	requireIntegration(t)
	ensureIntegrationSchema()
	ownerUserId, childUserId, familyId, child := createTaskIntegrationFamily(t, "recurring-claim")
	task := TaskService.CreateTask(ownerUserId, model.TaskCreateRequest{
		FamilyId: familyId,
		Title: "Daily reading",
		Points: 3,
		CycleType: model.TaskCycleTypeDaily,
	})
	claim := TaskService.StartTaskClaim(childUserId, model.TaskClaimRequest{FamilyId: familyId, TaskId: task.Id})
	if claim.MemberId != child.Id || claim.Status != model.TaskClaimStatusActive {
		t.Fatalf("claim = %+v, want active child claim", claim)
	}
	record := TaskService.SubmitTask(childUserId, model.TaskSubmitRequest{FamilyId: familyId, TaskId: task.Id})
	if record.Status != model.TaskRecordStatusPending {
		t.Fatalf("record status = %s, want pending", record.Status)
	}
	stopped := TaskService.StopTaskClaim(childUserId, model.TaskClaimRequest{FamilyId: familyId, TaskId: task.Id})
	if stopped.Status != model.TaskClaimStatusStopped {
		t.Fatalf("stopped status = %s, want stopped", stopped.Status)
	}
}
```

- [ ] **Step 3: Run integration test**

Run:

```powershell
cd parent-child-api
$env:PARENT_CHILD_DB_INTEGRATION='1'; go test ./srv -run TestIntegrationRecurringTaskClaimCanSubmitAfterFirstClaim -count=1 -v
```

Expected: PASS with integration database configured.

- [ ] **Step 4: Commit**

```powershell
git add parent-child-api/srv/integration_flow_test.go parent-child-api/srv/task_integration_test.go
git commit -m "test: 覆盖循环任务持续领取流程"
```

## Task 5: Mini Program API And State

**Files:**
- Modify: `parent-child-miniprogram/api/task.js`
- Modify: `parent-child-miniprogram/pages/tasks/index.vue`
- Modify: `parent-child-miniprogram/pages/index/index.vue`
- Modify: `parent-child-miniprogram/pages/tasks/index.behavior.test.mjs`
- Modify: `parent-child-miniprogram/pages/index/index.behavior.test.mjs`

- [ ] **Step 1: Update behavior tests**

Add assertions:

```js
assert.ok(pageSource.includes('listTaskClaims'), '任务页应加载循环任务领取关系')
assert.ok(pageSource.includes('stopTaskClaim'), '任务页应支持停止领取循环任务')
assert.ok(pageSource.includes('停止领取'), '任务页应展示停止领取入口')
```

Add to home test:

```js
assert.ok(source.includes('listTaskClaims'), '首页代孩子任务应合成持续领取状态')
assert.ok(source.includes('completedOnceTasksHidden'), '首页应隐藏已完成一次性任务')
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```powershell
node parent-child-miniprogram/pages/tasks/index.behavior.test.mjs
node parent-child-miniprogram/pages/index/index.behavior.test.mjs
```

Expected: FAIL because new API and markers are missing.

- [ ] **Step 3: Add API client**

Add to `parent-child-miniprogram/api/task.js`:

```js
export function listTaskClaims(familyId) {
  return request({
    url: '/api/task/claims',
    data: { familyId }
  })
}

export function startTaskClaim(data) {
  return request({
    url: '/api/task/startClaim',
    method: 'POST',
    data
  })
}

export function stopTaskClaim(data) {
  return request({
    url: '/api/task/stopClaim',
    method: 'POST',
    data
  })
}
```

- [ ] **Step 4: Update task page state**

In `pages/tasks/index.vue`, import the new functions and add `taskClaims = ref([])`. Load claims with tasks and records. In `taskRows`, treat an active claim as `viewStatus: 'claimed'` for recurring tasks when no current `PENDING` or current-cycle `APPROVED` exists.

Add a helper:

```js
function hasActiveClaim(task) {
  return taskClaims.value.some((claim) => claim.taskId === task.id && claim.memberId === currentMemberId.value)
}
```

Add action:

```js
async function stopTaskClaimForTask(task) {
  const familyId = currentFamily.value.familyId
  await stopTaskClaim({ familyId, taskId: task.id })
  taskClaims.value = taskClaims.value.filter((claim) => claim.taskId !== task.id)
  uni.showToast({ title: '已停止领取', icon: 'success' })
}
```

- [ ] **Step 5: Update home proxy state**

In `pages/index/index.vue`, load `listTaskClaims(familyId)` for parent and child views. For parent proxy, filter claims by selected child. Hide current-cycle completed rows and completed once rows through a computed marker named `completedOnceTasksHidden`.

- [ ] **Step 6: Run mini program behavior tests**

Run:

```powershell
node parent-child-miniprogram/pages/tasks/index.behavior.test.mjs
node parent-child-miniprogram/pages/index/index.behavior.test.mjs
```

Expected: PASS.

- [ ] **Step 7: Commit**

```powershell
git add parent-child-miniprogram/api/task.js parent-child-miniprogram/pages/tasks/index.vue parent-child-miniprogram/pages/index/index.vue parent-child-miniprogram/pages/tasks/index.behavior.test.mjs parent-child-miniprogram/pages/index/index.behavior.test.mjs
git commit -m "feat: 接入循环任务持续领取状态"
```

## Task 6: Full Verification

**Files:**
- No source changes expected.

- [ ] **Step 1: Run Go tests**

Run:

```powershell
cd parent-child-api
go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 2: Run mini program behavior tests**

Run:

```powershell
node parent-child-miniprogram/pages/tasks/index.behavior.test.mjs
node parent-child-miniprogram/pages/index/index.behavior.test.mjs
```

Expected: PASS.

- [ ] **Step 3: Commit only if verification updates files**

If no files changed, do not commit.

## Self-Review

- Spec coverage: Covers recurring active claim, stop claim, parent archive boundary, one-time task hidden behavior, and parent proxy state.
- Placeholder scan: No unresolved markers or vague “handle later” steps.
- Type consistency: Uses `TaskClaimStatusActive`, `TaskClaimRequest`, `TaskClaimListItem`, `listTaskClaims`, `startTaskClaim`, and `stopTaskClaim` consistently.
