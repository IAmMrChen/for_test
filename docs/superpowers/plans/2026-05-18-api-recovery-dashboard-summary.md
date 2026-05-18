# API Recovery Dashboard Summary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补齐 API 层统一错误兜底，并提供家庭首页聚合摘要接口，让前端能稳定获取当前家庭、成员积分、任务与奖励待处理数量。

**Architecture:** API 层新增 recover middleware，所有已注册路由统一包裹，避免 service 中的业务 panic 直接中断 HTTP 请求。首页摘要放在独立 dashboard service 中，用原生 SQL 查询必要聚合数据，API 层只负责鉴权、解析 `familyId` 和 JSON 输出。

**Tech Stack:** Go `net/http`、现有 `apix` helper、现有 JWT 鉴权、`sqlmer` 原生 SQL 查询、Go 单元测试与数据库集成测试。

---

## 文件结构

- 新增 `parent-child-api/api/apix/recover_middleware.go`：API 统一 panic 兜底，转换为 JSON 错误响应。
- 新增 `parent-child-api/api/apix/recover_middleware_test.go`：覆盖正常响应和 panic 响应。
- 修改 `parent-child-api/api/routes.go`：所有路由注册统一套上 recover middleware，并注册 dashboard 路由。
- 修改 `parent-child-api/api/routes_test.go`：验证业务 panic 会转为 JSON 错误。
- 新增 `parent-child-api/model/dashboard_model.go`：定义首页摘要响应模型。
- 新增 `parent-child-api/srv/dashboard_service.go`：实现首页聚合查询。
- 新增 `parent-child-api/srv/dashboard_service_test.go`：覆盖基础校验和角色可见性。
- 新增 `parent-child-api/srv/dashboard_integration_test.go`：用真实数据库验证任务、奖励、积分聚合结果。
- 新增 `parent-child-api/api/dashboard/dashboard_api.go`：注册 `GET /api/dashboard/summary`。

---

### Task 1: API panic 兜底

**Files:**
- Create: `parent-child-api/api/apix/recover_middleware_test.go`
- Create: `parent-child-api/api/apix/recover_middleware.go`
- Modify: `parent-child-api/api/routes.go`
- Modify: `parent-child-api/api/routes_test.go`

- [ ] **Step 1: 写失败测试**

在 `parent-child-api/api/apix/recover_middleware_test.go` 增加：

```go
package apix

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithRecoverWritesJsonErrorWhenHandlerPanics(t *testing.T) {
	handler := WithRecover(func(w http.ResponseWriter, r *http.Request) {
		panic(fmt.Errorf("permission denied"))
	})

	recorder := httptest.NewRecorder()
	handler(recorder, httptest.NewRequest(http.MethodGet, "/demo", nil))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !strings.Contains(recorder.Body.String(), `"message":"permission denied"`) {
		t.Fatalf("body = %s, want permission denied message", recorder.Body.String())
	}
}

func TestWithRecoverKeepsNormalResponse(t *testing.T) {
	handler := WithRecover(func(w http.ResponseWriter, r *http.Request) {
		WriteData(w, map[string]string{"hello": "world"})
	})

	recorder := httptest.NewRecorder()
	handler(recorder, httptest.NewRequest(http.MethodGet, "/demo", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `"hello":"world"`) {
		t.Fatalf("body = %s, want normal data", recorder.Body.String())
	}
}
```

- [ ] **Step 2: 运行失败测试**

Run: `go test ./api/apix -run TestWithRecover -count=1`

Expected: 编译失败，提示 `undefined: WithRecover`。

- [ ] **Step 3: 实现最小 middleware**

在 `parent-child-api/api/apix/recover_middleware.go` 增加：

```go
package apix

import (
	"fmt"
	"net/http"
)

func WithRecover(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				WriteError(w, http.StatusBadRequest, recoverMessage(recovered))
			}
		}()

		next(w, r)
	}
}

func recoverMessage(recovered any) string {
	switch value := recovered.(type) {
	case error:
		return value.Error()
	case string:
		return value
	default:
		return fmt.Sprint(value)
	}
}
```

- [ ] **Step 4: 验证 middleware 测试通过**

Run: `go test ./api/apix -run TestWithRecover -count=1`

Expected: PASS。

- [ ] **Step 5: 路由层接入 recover 并补测试**

在 `parent-child-api/api/routes.go` 中新增 `recoverRoute` helper，所有 `mux.HandleFunc` 的 handler 外层统一包 `recoverRoute(...)`。例如：

```go
func recoverRoute(handler http.HandlerFunc) http.HandlerFunc {
	return apix.WithRecover(handler)
}
```

并将注册改为：

```go
mux.HandleFunc("/api/auth/demo-login", recoverRoute(authapi.DemoLogin))
mux.HandleFunc("/api/demo/me", recoverRoute(apix.WithAuth(jwtSecret, demoapi.Me)))
```

在 `parent-child-api/api/routes_test.go` 增加：

```go
func TestRegisterRoutesRecoversBusinessPanic(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, "secret")

	token, err := ux.GenerateAuthToken("secret", 1001)
	if err != nil {
		t.Fatalf("GenerateAuthToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/family/create", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !strings.Contains(recorder.Body.String(), `"message":"family name is required"`) {
		t.Fatalf("body = %s, want family name error", recorder.Body.String())
	}
}
```

- [ ] **Step 6: 验证 API 路由测试通过**

Run: `go test ./api -run TestRegisterRoutesRecoversBusinessPanic -count=1`

Expected: PASS。

- [ ] **Step 7: 提交**

```bash
git add parent-child-api/api/apix/recover_middleware.go parent-child-api/api/apix/recover_middleware_test.go parent-child-api/api/routes.go parent-child-api/api/routes_test.go
git commit -m "feat: 增加 API 错误兜底"
```

---

### Task 2: 首页摘要 service

**Files:**
- Create: `parent-child-api/model/dashboard_model.go`
- Create: `parent-child-api/srv/dashboard_service_test.go`
- Create: `parent-child-api/srv/dashboard_service.go`
- Create: `parent-child-api/srv/dashboard_integration_test.go`

- [ ] **Step 1: 写模型**

在 `parent-child-api/model/dashboard_model.go` 增加：

```go
package model

type DashboardSummary struct {
	FamilyId                 int64  `json:"familyId"`
	FamilyName               string `json:"familyName"`
	MemberId                 int64  `json:"memberId"`
	RoleType                 string `json:"roleType"`
	Nickname                 string `json:"nickname"`
	CurrentPoints            int    `json:"currentPoints"`
	TotalEarnedPoints        int    `json:"totalEarnedPoints"`
	ActiveTaskCount          int    `json:"activeTaskCount"`
	MyClaimedTaskCount       int    `json:"myClaimedTaskCount"`
	MyPendingTaskCount       int    `json:"myPendingTaskCount"`
	FamilyPendingTaskCount   int    `json:"familyPendingTaskCount"`
	ActiveRewardCount        int    `json:"activeRewardCount"`
	MyAppliedRewardCount     int    `json:"myAppliedRewardCount"`
	FamilyAppliedRewardCount int    `json:"familyAppliedRewardCount"`
}
```

- [ ] **Step 2: 写失败测试**

在 `parent-child-api/srv/dashboard_service_test.go` 增加：

```go
package srv

import (
	"testing"

	"github.com/IAmMrChen/for_test/parent-child-api/model"
)

func TestDashboardSummaryRequiresFamilyId(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil || recovered.(error).Error() != "family id is required" {
			t.Fatalf("panic = %v, want family id is required", recovered)
		}
	}()

	DashboardService.Summary(1001, 0)
}

func TestDashboardCanSeeFamilyPendingOnlyParent(t *testing.T) {
	if !dashboardCanSeeFamilyPending(model.FamilyRoleParent) {
		t.Fatalf("parent should see family pending counters")
	}
	if dashboardCanSeeFamilyPending(model.FamilyRoleChild) {
		t.Fatalf("child should not see family pending counters")
	}
}
```

- [ ] **Step 3: 运行失败测试**

Run: `go test ./srv -run TestDashboard -count=1`

Expected: 编译失败，提示 `undefined: DashboardService` 和 `undefined: dashboardCanSeeFamilyPending`。

- [ ] **Step 4: 实现 service**

在 `parent-child-api/srv/dashboard_service.go` 增加：

```go
package srv

import (
	"fmt"

	"github.com/IAmMrChen/for_test/parent-child-api/model"
	"github.com/IAmMrChen/for_test/parent-child-api/resx"
)

var DashboardService dashboardService

type dashboardService struct{}

func (s dashboardService) Summary(userId int64, familyId int64) model.DashboardSummary {
	if familyId <= 0 {
		panic(fmt.Errorf("family id is required"))
	}

	member := MemberService.LoadActiveMember(userId, familyId)
	if member == nil {
		panic(fmt.Errorf("permission denied"))
	}

	family := &model.Family{}
	ok := resx.Db.Main.MustGetStruct(family, `SELECT id, name, creator_id FROM families WHERE id = @p1`, familyId)
	if !ok {
		panic(fmt.Errorf("family not found"))
	}

	summary := model.DashboardSummary{
		FamilyId:          family.Id,
		FamilyName:        family.Name,
		MemberId:          member.Id,
		RoleType:          member.RoleType,
		Nickname:          member.Nickname,
		CurrentPoints:     member.CurrentPoints,
		TotalEarnedPoints: member.TotalEarnedPoints,
	}

	summary.ActiveTaskCount = mustScalarCount(resx.Db.Main, `SELECT COUNT(1) FROM tasks WHERE family_id = @p1 AND status = @p2`, familyId, model.TaskStatusActive)
	summary.MyClaimedTaskCount = mustScalarCount(resx.Db.Main, `SELECT COUNT(1) FROM task_records WHERE family_id = @p1 AND member_id = @p2 AND status = @p3`, familyId, member.Id, model.TaskRecordStatusClaimed)
	summary.MyPendingTaskCount = mustScalarCount(resx.Db.Main, `SELECT COUNT(1) FROM task_records WHERE family_id = @p1 AND member_id = @p2 AND status = @p3`, familyId, member.Id, model.TaskRecordStatusPendingAudit)
	summary.ActiveRewardCount = mustScalarCount(resx.Db.Main, `SELECT COUNT(1) FROM rewards WHERE family_id = @p1 AND status = @p2`, familyId, model.RewardStatusActive)
	summary.MyAppliedRewardCount = mustScalarCount(resx.Db.Main, `SELECT COUNT(1) FROM reward_records WHERE family_id = @p1 AND member_id = @p2 AND status = @p3`, familyId, member.Id, model.RewardRecordStatusApplied)

	if dashboardCanSeeFamilyPending(member.RoleType) {
		summary.FamilyPendingTaskCount = mustScalarCount(resx.Db.Main, `SELECT COUNT(1) FROM task_records WHERE family_id = @p1 AND status = @p2`, familyId, model.TaskRecordStatusPendingAudit)
		summary.FamilyAppliedRewardCount = mustScalarCount(resx.Db.Main, `SELECT COUNT(1) FROM reward_records WHERE family_id = @p1 AND status = @p2`, familyId, model.RewardRecordStatusApplied)
	}

	return summary
}

func dashboardCanSeeFamilyPending(roleType string) bool {
	return roleType == model.FamilyRoleParent
}
```

- [ ] **Step 5: 验证 service 单元测试通过**

Run: `go test ./srv -run TestDashboard -count=1`

Expected: PASS。

- [ ] **Step 6: 写数据库集成测试**

在 `parent-child-api/srv/dashboard_integration_test.go` 增加一个 `TestIntegrationDashboardSummaryFlow`，复用已有集成测试初始化方式，完整串起创建家庭、虚拟孩子、发布任务、领取任务、提交任务、家长审核、发布奖励、孩子申请奖励，然后断言：

```go
childSummary := DashboardService.Summary(childUserId, family.Id)
if childSummary.CurrentPoints != 5 {
	t.Fatalf("child current points = %d, want 5", childSummary.CurrentPoints)
}
if childSummary.MyAppliedRewardCount != 1 {
	t.Fatalf("child applied rewards = %d, want 1", childSummary.MyAppliedRewardCount)
}
if childSummary.FamilyAppliedRewardCount != 0 {
	t.Fatalf("child family applied rewards = %d, want 0", childSummary.FamilyAppliedRewardCount)
}

parentSummary := DashboardService.Summary(parentUserId, family.Id)
if parentSummary.FamilyAppliedRewardCount != 1 {
	t.Fatalf("parent family applied rewards = %d, want 1", parentSummary.FamilyAppliedRewardCount)
}
```

- [ ] **Step 7: 运行数据库集成测试**

Run: `$env:PARENT_CHILD_DB_INTEGRATION='1'; go test ./srv -run TestIntegrationDashboardSummaryFlow -count=1 -v`

Expected: PASS。

- [ ] **Step 8: 提交**

```bash
git add parent-child-api/model/dashboard_model.go parent-child-api/srv/dashboard_service.go parent-child-api/srv/dashboard_service_test.go parent-child-api/srv/dashboard_integration_test.go
git commit -m "feat: 增加首页摘要服务"
```

---

### Task 3: 首页摘要 API

**Files:**
- Create: `parent-child-api/api/dashboard/dashboard_api.go`
- Modify: `parent-child-api/api/routes.go`
- Modify: `parent-child-api/api/routes_test.go`

- [ ] **Step 1: 写失败测试**

在 `parent-child-api/api/routes_test.go` 增加：

```go
func TestRegisterRoutesDashboardSummaryRejectsInvalidFamilyId(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, "secret")

	token, err := ux.GenerateAuthToken("secret", 1001)
	if err != nil {
		t.Fatalf("GenerateAuthToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/summary?familyId=bad", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !strings.Contains(recorder.Body.String(), `"message":"family id is required"`) {
		t.Fatalf("body = %s, want family id error", recorder.Body.String())
	}
}
```

- [ ] **Step 2: 运行失败测试**

Run: `go test ./api -run TestRegisterRoutesDashboardSummaryRejectsInvalidFamilyId -count=1`

Expected: FAIL，当前路由不存在或返回 404。

- [ ] **Step 3: 实现 dashboard API**

在 `parent-child-api/api/dashboard/dashboard_api.go` 增加：

```go
package dashboardapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/IAmMrChen/for_test/parent-child-api/api/apix"
	"github.com/IAmMrChen/for_test/parent-child-api/srv"
	"github.com/IAmMrChen/for_test/parent-child-api/ux"
)

func Summary(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId <= 0 {
		panic(fmt.Errorf("family id is required"))
	}

	apix.WriteData(w, srv.DashboardService.Summary(token.UserId, familyId))
}
```

在 `parent-child-api/api/routes.go` 引入 dashboard API，并注册：

```go
mux.HandleFunc("/api/dashboard/summary", recoverRoute(apix.WithAuth(jwtSecret, dashboardapi.Summary)))
```

- [ ] **Step 4: 验证 API 测试通过**

Run: `go test ./api -run TestRegisterRoutesDashboardSummaryRejectsInvalidFamilyId -count=1`

Expected: PASS。

- [ ] **Step 5: 全量验证**

Run: `go test ./... -count=1`

Expected: PASS。

Run: `$env:PARENT_CHILD_DB_INTEGRATION='1'; go test ./srv -run TestIntegrationDashboardSummaryFlow -count=1 -v`

Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add parent-child-api/api/dashboard/dashboard_api.go parent-child-api/api/routes.go parent-child-api/api/routes_test.go
git commit -m "feat: 增加首页摘要接口"
```

---

## 自检

- Spec coverage: 覆盖 API 业务异常兜底、首页摘要模型、service 聚合、HTTP API 和数据库集成验证。
- Placeholder scan: 无 `TBD`、`TODO`、`implement later`；所有关键步骤包含明确文件、代码和命令。
- Type consistency: `DashboardSummary`、`DashboardService.Summary`、`dashboardCanSeeFamilyPending`、`WithRecover` 命名在各任务中保持一致。
