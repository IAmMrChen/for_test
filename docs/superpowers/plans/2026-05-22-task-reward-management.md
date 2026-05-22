# 任务与奖品管理 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补齐家长编辑/归档任务、编辑/下架奖品的 MVP 管理闭环。

**Architecture:** 后端沿用现有 `api -> srv -> model` 分层和原生 SQL 风格。任务/奖品管理只更新已有状态字段与基础字段，不新增表结构，不影响历史记录和积分流水。

**Tech Stack:** Go `net/http`、项目现有 `resx.Db.Main` SQL 包、uni-app 小程序 API 封装。

---

## 文件结构

- Modify: `parent-child-api/model/task_model.go`
  - 增加 `TaskUpdateRequest`、`TaskArchiveRequest`。
- Modify: `parent-child-api/model/reward_model.go`
  - 增加 `RewardUpdateRequest`、`RewardOffShelfRequest`。
- Modify: `parent-child-api/srv/task_service.go`
  - 增加 `UpdateTask`、`ArchiveTask` 和主动任务加载辅助函数。
- Modify: `parent-child-api/srv/task_service_test.go`
  - 增加缺少 task id、更新请求规范化测试。
- Modify: `parent-child-api/srv/task_integration_test.go`
  - 增加家长编辑/归档任务、孩子无权限管理任务的集成测试。
- Modify: `parent-child-api/api/task/task_api.go`
  - 增加 `Update`、`Archive` handler。
- Modify: `parent-child-api/model/reward_model.go`
  - 增加奖品管理请求模型。
- Modify: `parent-child-api/srv/reward_service.go`
  - 增加 `UpdateReward`、`OffShelfReward` 和主动奖品加载辅助函数。
- Modify: `parent-child-api/srv/reward_service_test.go`
  - 增加缺少 reward id、更新请求规范化测试。
- Modify: `parent-child-api/srv/reward_integration_test.go`
  - 增加家长编辑/下架奖品、孩子无权限管理奖品的集成测试。
- Modify: `parent-child-api/api/reward/reward_api.go`
  - 增加 `Update`、`OffShelf` handler。
- Modify: `parent-child-api/api/routes.go`
  - 注册四个管理路由。
- Modify: `parent-child-miniprogram/api/task.js`
  - 增加 `updateTask`、`archiveTask`。
- Modify: `parent-child-miniprogram/api/reward.js`
  - 增加 `updateReward`、`offShelfReward`。
- Modify: `parent-child-miniprogram/pages/index/index.vue`
  - 家长任务卡片增加编辑/归档入口。
- Modify: `parent-child-miniprogram/pages/rewards/index.vue`
  - 家长奖品卡片增加编辑/下架入口。

---

### Task 1: 任务管理后端红测

**Files:**
- Modify: `parent-child-api/srv/task_service_test.go`
- Modify: `parent-child-api/srv/task_integration_test.go`

- [ ] **Step 1: 写更新请求规范化单元测试**

在 `task_service_test.go` 增加：

```go
func TestNormalizeTaskUpdateRequest(t *testing.T) {
	req := normalizeTaskUpdateRequest(model.TaskUpdateRequest{
		FamilyId:  1,
		TaskId:    2,
		Title:     " Read ",
		Points:    0,
		CycleType: "",
	})

	if req.Title != "Read" {
		t.Fatalf("title = %q, want Read", req.Title)
	}
	if req.Points != 1 {
		t.Fatalf("points = %d, want 1", req.Points)
	}
	if req.CycleType != model.TaskCycleTypeOnce {
		t.Fatalf("cycleType = %s, want ONCE", req.CycleType)
	}
}
```

- [ ] **Step 2: 写缺少 task id 的单元测试**

在 `task_service_test.go` 增加：

```go
func TestTaskServiceArchiveTaskRejectsMissingTask(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("ArchiveTask should panic")
		}
		if fmt.Sprint(v) != "task id is required" {
			t.Fatalf("panic = %v, want task id is required", v)
		}
	}()

	TaskService.ArchiveTask(1, model.TaskArchiveRequest{FamilyId: 1})
}
```

- [ ] **Step 3: 写任务集成测试**

在 `task_integration_test.go` 增加测试，使用现有集成测试 helper 风格：

```go
func TestTaskServiceUpdateAndArchiveTask(t *testing.T) {
	fixture := createTaskIntegrationFixture(t)

	task := TaskService.CreateTask(fixture.ParentUserId, model.TaskCreateRequest{
		FamilyId:  fixture.FamilyId,
		Title:     "Old",
		Points:    5,
		CycleType: model.TaskCycleTypeOnce,
	})

	updated := TaskService.UpdateTask(fixture.ParentUserId, model.TaskUpdateRequest{
		FamilyId:  fixture.FamilyId,
		TaskId:    task.Id,
		Title:     "New",
		Points:    9,
		CycleType: model.TaskCycleTypeWeekly,
	})
	if updated.Title != "New" || updated.Points != 9 || updated.CycleType != model.TaskCycleTypeWeekly {
		t.Fatalf("updated task = %+v", updated)
	}

	archived := TaskService.ArchiveTask(fixture.ParentUserId, model.TaskArchiveRequest{
		FamilyId: fixture.FamilyId,
		TaskId:   task.Id,
	})
	if archived.Status != model.TaskStatusArchived {
		t.Fatalf("archived status = %d, want archived", archived.Status)
	}

	tasks := TaskService.ListTasks(fixture.ParentUserId, fixture.FamilyId)
	for _, item := range tasks {
		if item.Id == task.Id {
			t.Fatalf("archived task should not be listed: %+v", item)
		}
	}
}
```

- [ ] **Step 4: 跑红测**

Run:

```powershell
go test ./srv -run "TestNormalizeTaskUpdateRequest|TestTaskServiceArchiveTaskRejectsMissingTask|TestTaskServiceUpdateAndArchiveTask" -count=1
```

Working directory: `parent-child-api`

Expected: FAIL，原因是 `TaskUpdateRequest`、`TaskArchiveRequest` 或 `UpdateTask`、`ArchiveTask` 未定义。

---

### Task 2: 任务管理后端实现

**Files:**
- Modify: `parent-child-api/model/task_model.go`
- Modify: `parent-child-api/srv/task_service.go`
- Modify: `parent-child-api/api/task/task_api.go`
- Modify: `parent-child-api/api/routes.go`

- [ ] **Step 1: 增加任务请求模型**

在 `task_model.go` 中 `TaskCreateRequest` 后增加：

```go
type TaskUpdateRequest struct {
	FamilyId  int64         `json:"familyId"`
	TaskId    int64         `json:"taskId"`
	Title     string        `json:"title"`
	Points    int           `json:"points"`
	CycleType TaskCycleType `json:"cycleType"`
}

type TaskArchiveRequest struct {
	FamilyId int64 `json:"familyId"`
	TaskId   int64 `json:"taskId"`
}
```

- [ ] **Step 2: 增加任务更新规范化**

在 `task_service.go` 中 `normalizeTaskCreateRequest` 后增加：

```go
func normalizeTaskUpdateRequest(req model.TaskUpdateRequest) model.TaskUpdateRequest {
	normalized := normalizeTaskCreateRequest(model.TaskCreateRequest{
		FamilyId:  req.FamilyId,
		Title:     req.Title,
		Points:    req.Points,
		CycleType: req.CycleType,
	})
	req.Title = normalized.Title
	req.Points = normalized.Points
	req.CycleType = normalized.CycleType
	return req
}
```

- [ ] **Step 3: 实现 `UpdateTask`**

在 `task_service.go` 中 `CreateTask` 后增加：

```go
func (x taskService) UpdateTask(operatorUserId int64, req model.TaskUpdateRequest) model.Task {
	if req.TaskId == 0 {
		panic(fmt.Errorf("task id is required"))
	}
	req = normalizeTaskUpdateRequest(req)
	MemberService.RequireParentRole(operatorUserId, req.FamilyId)

	affected := resx.Db.Main.MustExecute(`
		UPDATE tasks
		SET title=@p1, points=@p2, cycle_type=@p3
		WHERE id=@p4 AND family_id=@p5 AND status=@p6
	`, req.Title, req.Points, req.CycleType, req.TaskId, req.FamilyId, model.TaskStatusActive)
	if affected != 1 {
		panic(fmt.Errorf("task not found"))
	}
	return x.loadActiveTask(req.FamilyId, req.TaskId)
}
```

- [ ] **Step 4: 实现 `ArchiveTask`**

在 `task_service.go` 中 `UpdateTask` 后增加：

```go
func (x taskService) ArchiveTask(operatorUserId int64, req model.TaskArchiveRequest) model.Task {
	if req.FamilyId == 0 {
		panic(fmt.Errorf("family id is required"))
	}
	if req.TaskId == 0 {
		panic(fmt.Errorf("task id is required"))
	}
	MemberService.RequireParentRole(operatorUserId, req.FamilyId)

	task := x.loadActiveTask(req.FamilyId, req.TaskId)
	affected := resx.Db.Main.MustExecute(`
		UPDATE tasks
		SET status=@p1
		WHERE id=@p2 AND family_id=@p3 AND status=@p4
	`, model.TaskStatusArchived, req.TaskId, req.FamilyId, model.TaskStatusActive)
	if affected != 1 {
		panic(fmt.Errorf("task not found"))
	}
	task.Status = model.TaskStatusArchived
	return task
}
```

- [ ] **Step 5: 增加 API handler 和路由**

在 `task_api.go` 增加 `Update`、`Archive`，结构与 `Create` 一致。

在 `routes.go` 注册：

```go
mux.HandleFunc("POST /api/task/update", recoverRoute(apix.WithAuth(jwtSecret, taskApi.Update)))
mux.HandleFunc("POST /api/task/archive", recoverRoute(apix.WithAuth(jwtSecret, taskApi.Archive)))
```

- [ ] **Step 6: 跑任务相关测试**

Run:

```powershell
go test ./srv -run "TestNormalizeTaskUpdateRequest|TestTaskServiceArchiveTaskRejectsMissingTask|TestTaskServiceUpdateAndArchiveTask" -count=1
```

Working directory: `parent-child-api`

Expected: PASS。

- [ ] **Step 7: 提交任务管理后端**

```powershell
git add parent-child-api\model\task_model.go parent-child-api\srv\task_service.go parent-child-api\srv\task_service_test.go parent-child-api\srv\task_integration_test.go parent-child-api\api\task\task_api.go parent-child-api\api\routes.go
git commit -m "feat: 增加任务编辑和归档接口"
```

---

### Task 3: 奖品管理后端红测与实现

**Files:**
- Modify: `parent-child-api/model/reward_model.go`
- Modify: `parent-child-api/srv/reward_service.go`
- Modify: `parent-child-api/srv/reward_service_test.go`
- Modify: `parent-child-api/srv/reward_integration_test.go`
- Modify: `parent-child-api/api/reward/reward_api.go`
- Modify: `parent-child-api/api/routes.go`

- [ ] **Step 1: 写奖品红测**

在 `reward_service_test.go` 增加 `TestNormalizeRewardUpdateRequest` 和 `TestRewardServiceOffShelfRewardRejectsMissingReward`。

在 `reward_integration_test.go` 增加 `TestRewardServiceUpdateAndOffShelfReward`，验证编辑后字段更新，下架后 `ListRewards` 不再返回该奖品。

- [ ] **Step 2: 跑红测**

Run:

```powershell
go test ./srv -run "TestNormalizeRewardUpdateRequest|TestRewardServiceOffShelfRewardRejectsMissingReward|TestRewardServiceUpdateAndOffShelfReward" -count=1
```

Working directory: `parent-child-api`

Expected: FAIL，原因是请求模型或服务方法未定义。

- [ ] **Step 3: 增加奖品请求模型**

在 `reward_model.go` 中 `RewardCreateRequest` 后增加：

```go
type RewardUpdateRequest struct {
	FamilyId   int64  `json:"familyId"`
	RewardId   int64  `json:"rewardId"`
	Name       string `json:"name"`
	PointsCost int    `json:"pointsCost"`
	Stock      int    `json:"stock"`
}

type RewardOffShelfRequest struct {
	FamilyId int64 `json:"familyId"`
	RewardId int64 `json:"rewardId"`
}
```

- [ ] **Step 4: 实现服务方法**

在 `reward_service.go` 增加 `normalizeRewardUpdateRequest`、`UpdateReward`、`OffShelfReward`。规则与创建一致，编辑/下架只允许家长操作本家庭 active 奖品。

- [ ] **Step 5: 增加 API handler 和路由**

在 `reward_api.go` 增加 `Update`、`OffShelf`。

在 `routes.go` 注册：

```go
mux.HandleFunc("POST /api/reward/update", recoverRoute(apix.WithAuth(jwtSecret, rewardApi.Update)))
mux.HandleFunc("POST /api/reward/offShelf", recoverRoute(apix.WithAuth(jwtSecret, rewardApi.OffShelf)))
```

- [ ] **Step 6: 跑奖品相关测试**

Run:

```powershell
go test ./srv -run "TestNormalizeRewardUpdateRequest|TestRewardServiceOffShelfRewardRejectsMissingReward|TestRewardServiceUpdateAndOffShelfReward" -count=1
```

Working directory: `parent-child-api`

Expected: PASS。

- [ ] **Step 7: 提交奖品管理后端**

```powershell
git add parent-child-api\model\reward_model.go parent-child-api\srv\reward_service.go parent-child-api\srv\reward_service_test.go parent-child-api\srv\reward_integration_test.go parent-child-api\api\reward\reward_api.go parent-child-api\api\routes.go
git commit -m "feat: 增加奖品编辑和下架接口"
```

---

### Task 4: 小程序管理入口

**Files:**
- Modify: `parent-child-miniprogram/api/task.js`
- Modify: `parent-child-miniprogram/api/reward.js`
- Modify: `parent-child-miniprogram/pages/index/index.vue`
- Modify: `parent-child-miniprogram/pages/rewards/index.vue`

- [ ] **Step 1: 增加 API 封装**

`task.js` 增加 `updateTask(data)`、`archiveTask(data)`。

`reward.js` 增加 `updateReward(data)`、`offShelfReward(data)`。

- [ ] **Step 2: 首页任务卡片增加编辑/归档**

家长视角任务卡片增加两个按钮。编辑使用现有创建表单字段填充，提交时调用 `updateTask`；归档二次确认后调用 `archiveTask` 并刷新列表。

- [ ] **Step 3: 奖励页奖品卡片增加编辑/下架**

家长视角奖品卡片增加两个按钮。编辑使用现有创建表单字段填充，提交时调用 `updateReward`；下架二次确认后调用 `offShelfReward` 并刷新列表。

- [ ] **Step 4: 前端语法验证**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
```

Expected: exit 0。

- [ ] **Step 5: 乱码扫描**

Run:

```powershell
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\index\index.vue parent-child-miniprogram\pages\rewards\index.vue
```

Expected: exit 1，无匹配。

- [ ] **Step 6: 提交小程序入口**

```powershell
git add parent-child-miniprogram\api\task.js parent-child-miniprogram\api\reward.js parent-child-miniprogram\pages\index\index.vue parent-child-miniprogram\pages\rewards\index.vue
git commit -m "feat: 增加小程序任务奖品管理入口"
```

---

### Task 5: 全量验证

**Files:**
- Verify: backend and miniprogram changed files.

- [ ] **Step 1: 跑后端全量测试**

Run:

```powershell
go test ./... -count=1
```

Working directory: `parent-child-api`

Expected: exit 0。

- [ ] **Step 2: 跑前端 JS 语法检查**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
```

Expected: exit 0。

- [ ] **Step 3: 查看 git 状态**

Run:

```powershell
git status --short
```

Expected: clean。

## 自查

- 设计要求均有任务覆盖：任务编辑/归档、奖品编辑/下架、API、小程序入口、验证。
- 无新增表结构。
- 管理权限统一走家长角色。
- 历史记录不删除，只隐藏 active 列表。
