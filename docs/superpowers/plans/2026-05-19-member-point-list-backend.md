# Member Point List Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 MVP 后端补齐家庭成员列表与积分流水查询接口。

**Architecture:** 继续沿用 `api -> srv -> model` 分层，API 只负责鉴权、query 参数解析和 JSON 输出，权限与查询语义放在 service 层。成员列表扩展现有 `MemberService`；积分流水使用新的 `PointService` 和 `api/point` 二级目录，SQL 查询继续使用 `resx.Db.Main` 原生语句。

**Tech Stack:** Go `net/http`、现有 JWT 鉴权 middleware、现有 API recover middleware、`sqlmer` 原生 SQL 查询、Go 单元测试与数据库集成测试。

---

## 文件结构

- 修改 `parent-child-api/srv/member_service.go`：新增 `ListMembers(userId, familyId int64)`。
- 修改 `parent-child-api/srv/member_service_test.go`：新增成员列表基础校验测试。
- 修改 `parent-child-api/api/member/member_api.go`：新增 `List` API 方法。
- 修改 `parent-child-api/api/routes.go`：注册 `GET /api/member/list` 和 `GET /api/point/logs`。
- 修改 `parent-child-api/api/routes_test.go`：新增两个 query 参数错误测试。
- 新增 `parent-child-api/model/point_model.go`：定义积分流水查询请求与响应项。
- 新增 `parent-child-api/srv/point_service.go`：实现积分流水查询、孩子/家长权限和来源标题拼接。
- 新增 `parent-child-api/srv/point_service_test.go`：覆盖基础校验和孩子权限 helper。
- 新增 `parent-child-api/api/point/point_api.go`：新增积分流水 HTTP API。
- 新增 `parent-child-api/srv/point_integration_test.go`：用真实数据库验证成员列表、积分流水、权限隔离。

---

### Task 1: 成员列表 service 与 API

**Files:**
- Modify: `parent-child-api/srv/member_service_test.go`
- Modify: `parent-child-api/srv/member_service.go`
- Modify: `parent-child-api/api/member/member_api.go`
- Modify: `parent-child-api/api/routes.go`
- Modify: `parent-child-api/api/routes_test.go`

- [ ] **Step 1: 写 service 失败测试**

在 `parent-child-api/srv/member_service_test.go` 增加：

```go
func TestMemberListRequiresFamilyId(t *testing.T) {
	mustPanicWith(t, "family id is required", func() {
		MemberService.ListMembers(1001, 0)
	})
}
```

- [ ] **Step 2: 运行失败测试**

Run: `go test ./srv -run TestMemberListRequiresFamilyId -count=1`

Expected: 编译失败，提示 `MemberService.ListMembers undefined`。

- [ ] **Step 3: 实现 `ListMembers`**

在 `parent-child-api/srv/member_service.go` 增加：

```go
func (x memberService) ListMembers(userId, familyId int64) []model.FamilyMember {
	if familyId == 0 {
		panic(fmt.Errorf("family id is required"))
	}
	if x.LoadActiveMember(userId, familyId) == nil {
		panic(fmt.Errorf("permission denied"))
	}

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
		WHERE family_id=@p1 AND status=@p2
		ORDER BY CASE role_type
			WHEN 'OWNER' THEN 1
			WHEN 'ADMIN' THEN 2
			WHEN 'PARENT' THEN 3
			WHEN 'CHILD' THEN 4
			ELSE 5
		END, id ASC
	`

	return resx.Db.Main.MustListOf(model.FamilyMember{}, sql, familyId, model.FamilyMemberStatusActive).([]model.FamilyMember)
}
```

- [ ] **Step 4: 验证 service 测试通过**

Run: `go test ./srv -run TestMemberListRequiresFamilyId -count=1`

Expected: PASS。

- [ ] **Step 5: 写 API 失败测试**

在 `parent-child-api/api/routes_test.go` 增加：

```go
func TestRegisterRoutesMemberListRejectsInvalidFamilyId(t *testing.T) {
	const jwtSecret = "test-secret"
	token, err := ux.JwtUtil.SignAuthToken(ux.AuthTokenClaims{UserId: 1}, jwtSecret, time.Hour)
	if err != nil {
		t.Fatalf("SignAuthToken error = %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, jwtSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/member/list?familyId=bad", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}

	want := `{"code":400,"message":"invalid familyId"}`
	if strings.TrimSpace(res.Body.String()) != want {
		t.Fatalf("body = %s, want %s", res.Body.String(), want)
	}
}
```

- [ ] **Step 6: 运行 API 失败测试**

Run: `go test ./api -run TestRegisterRoutesMemberListRejectsInvalidFamilyId -count=1`

Expected: FAIL，当前路由未注册，返回 404。

- [ ] **Step 7: 实现成员列表 API 与路由**

在 `parent-child-api/api/member/member_api.go` 增加 import：

```go
import "strconv"
```

并增加方法：

```go
func (x MemberApi) List(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}

	apix.WriteData(w, srv.MemberService.ListMembers(token.UserId, familyId))
}
```

在 `parent-child-api/api/routes.go` 注册：

```go
mux.HandleFunc("GET /api/member/list", recoverRoute(apix.WithAuth(jwtSecret, memberApi.List)))
```

- [ ] **Step 8: 验证 API 测试通过**

Run: `go test ./api -run TestRegisterRoutesMemberListRejectsInvalidFamilyId -count=1`

Expected: PASS。

- [ ] **Step 9: 提交**

```bash
git add parent-child-api/srv/member_service.go parent-child-api/srv/member_service_test.go parent-child-api/api/member/member_api.go parent-child-api/api/routes.go parent-child-api/api/routes_test.go
git commit -m "feat: 增加家庭成员列表接口"
```

---

### Task 2: 积分流水模型与 service

**Files:**
- Create: `parent-child-api/model/point_model.go`
- Create: `parent-child-api/srv/point_service_test.go`
- Create: `parent-child-api/srv/point_service.go`

- [ ] **Step 1: 写积分模型**

在 `parent-child-api/model/point_model.go` 增加：

```go
package model

import "time"

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

- [ ] **Step 2: 写 service 失败测试**

在 `parent-child-api/srv/point_service_test.go` 增加：

```go
package srv

import (
	"testing"

	"parent-child-api/model"
)

func TestPointLogListRequiresFamilyId(t *testing.T) {
	mustPanicWith(t, "family id is required", func() {
		PointService.ListPointLogs(1001, model.PointLogListRequest{})
	})
}

func TestPointLogMemberScopeForChild(t *testing.T) {
	member := model.FamilyMember{Id: 10, RoleType: model.FamilyRoleChild}

	if got := pointLogScopedMemberId(member, 0); got != 10 {
		t.Fatalf("scoped member id = %d, want 10", got)
	}
	if got := pointLogScopedMemberId(member, 10); got != 10 {
		t.Fatalf("scoped member id = %d, want 10", got)
	}
	mustPanicWith(t, "permission denied", func() {
		pointLogScopedMemberId(member, 11)
	})
}

func TestPointLogMemberScopeForParent(t *testing.T) {
	member := model.FamilyMember{Id: 10, RoleType: model.FamilyRoleParent}

	if got := pointLogScopedMemberId(member, 0); got != 0 {
		t.Fatalf("scoped member id = %d, want 0", got)
	}
	if got := pointLogScopedMemberId(member, 11); got != 11 {
		t.Fatalf("scoped member id = %d, want 11", got)
	}
}
```

- [ ] **Step 3: 运行失败测试**

Run: `go test ./srv -run TestPointLog -count=1`

Expected: 编译失败，提示 `PointService` 和 `pointLogScopedMemberId` 未定义。

- [ ] **Step 4: 实现积分流水 service**

在 `parent-child-api/srv/point_service.go` 增加：

```go
package srv

import (
	"fmt"

	"parent-child-api/model"
	"parent-child-api/resx"
)

const pointLogListLimit = 100

var PointService pointService

type pointService struct{}

func (x pointService) ListPointLogs(userId int64, req model.PointLogListRequest) []model.PointLogListItem {
	if req.FamilyId == 0 {
		panic(fmt.Errorf("family id is required"))
	}

	operator := MemberService.LoadActiveMember(userId, req.FamilyId)
	if operator == nil {
		panic(fmt.Errorf("permission denied"))
	}

	memberId := pointLogScopedMemberId(*operator, req.MemberId)
	if memberId != 0 {
		target := MemberService.LoadActiveMemberById(memberId)
		if target == nil || target.FamilyId != req.FamilyId {
			panic(fmt.Errorf("target member not found"))
		}
	}

	sql := `
		SELECT p.id
			, p.family_id
			, p.member_id
			, m.nickname
			, p.points
			, p.source_type
			, p.source_id
			, CASE p.source_type
				WHEN 'TASK' THEN IFNULL(t.title, '')
				WHEN 'REWARD' THEN IFNULL(w.name, '')
				ELSE ''
			END AS source_title
			, p.created_at
		FROM point_logs p
		INNER JOIN family_members m ON m.id = p.member_id
		LEFT JOIN task_records tr ON tr.id = p.source_id AND p.source_type = @p2
		LEFT JOIN tasks t ON t.id = tr.task_id
		LEFT JOIN reward_records rr ON rr.id = p.source_id AND p.source_type = @p3
		LEFT JOIN rewards w ON w.id = rr.reward_id
		WHERE p.family_id = @p1
	`
	args := []any{req.FamilyId, model.PointSourceTypeTask, model.PointSourceTypeReward}

	if memberId != 0 {
		sql += " AND p.member_id=@p4"
		args = append(args, memberId)
	}
	sql += " ORDER BY p.id DESC LIMIT 100"

	return resx.Db.Main.MustListOf(model.PointLogListItem{}, sql, args...).([]model.PointLogListItem)
}

func pointLogScopedMemberId(operator model.FamilyMember, requestedMemberId int64) int64 {
	if operator.RoleType.IsParentRole() {
		return requestedMemberId
	}
	if requestedMemberId == 0 || requestedMemberId == operator.Id {
		return operator.Id
	}
	panic(fmt.Errorf("permission denied"))
}
```

- [ ] **Step 5: 验证 service 测试通过**

Run: `go test ./srv -run TestPointLog -count=1`

Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add parent-child-api/model/point_model.go parent-child-api/srv/point_service.go parent-child-api/srv/point_service_test.go
git commit -m "feat: 增加积分流水查询服务"
```

---

### Task 3: 积分流水 API

**Files:**
- Create: `parent-child-api/api/point/point_api.go`
- Modify: `parent-child-api/api/routes.go`
- Modify: `parent-child-api/api/routes_test.go`

- [ ] **Step 1: 写 API 失败测试**

在 `parent-child-api/api/routes_test.go` 增加：

```go
func TestRegisterRoutesPointLogsRejectsInvalidFamilyId(t *testing.T) {
	const jwtSecret = "test-secret"
	token, err := ux.JwtUtil.SignAuthToken(ux.AuthTokenClaims{UserId: 1}, jwtSecret, time.Hour)
	if err != nil {
		t.Fatalf("SignAuthToken error = %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, jwtSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/point/logs?familyId=bad", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}

	want := `{"code":400,"message":"invalid familyId"}`
	if strings.TrimSpace(res.Body.String()) != want {
		t.Fatalf("body = %s, want %s", res.Body.String(), want)
	}
}
```

- [ ] **Step 2: 运行失败测试**

Run: `go test ./api -run TestRegisterRoutesPointLogsRejectsInvalidFamilyId -count=1`

Expected: FAIL，当前路由未注册，返回 404。

- [ ] **Step 3: 实现积分流水 API**

在 `parent-child-api/api/point/point_api.go` 增加：

```go
package pointapi

import (
	"net/http"
	"strconv"

	"parent-child-api/api/apix"
	"parent-child-api/model"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type PointApi struct{}

func (x PointApi) Logs(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}

	memberId, err := parseOptionalInt64(r.URL.Query().Get("memberId"))
	if err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid memberId")
		return
	}

	apix.WriteData(w, srv.PointService.ListPointLogs(token.UserId, model.PointLogListRequest{
		FamilyId: familyId,
		MemberId: memberId,
	}))
}

func parseOptionalInt64(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}
```

在 `parent-child-api/api/routes.go` 引入：

```go
pointapi "parent-child-api/api/point"
```

初始化：

```go
pointApi := pointapi.PointApi{}
```

注册：

```go
mux.HandleFunc("GET /api/point/logs", recoverRoute(apix.WithAuth(jwtSecret, pointApi.Logs)))
```

- [ ] **Step 4: 验证 API 测试通过**

Run: `go test ./api -run TestRegisterRoutesPointLogsRejectsInvalidFamilyId -count=1`

Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add parent-child-api/api/point/point_api.go parent-child-api/api/routes.go parent-child-api/api/routes_test.go
git commit -m "feat: 增加积分流水接口"
```

---

### Task 4: 数据库集成测试与全量验证

**Files:**
- Create: `parent-child-api/srv/point_integration_test.go`

- [ ] **Step 1: 写数据库集成测试**

在 `parent-child-api/srv/point_integration_test.go` 增加：

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

func TestIntegrationMemberListAndPointLogFlow(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed + 400
	childUserId := seed + 401

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("point-log-family-%d", seed),
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
		Title:     "Read",
		Points:    5,
		CycleType: model.TaskCycleTypeDaily,
	})
	claimed := TaskService.ClaimTask(childUserId, model.TaskClaimRequest{
		FamilyId: familyId,
		TaskId:   task.Id,
	})
	submitted := TaskService.SubmitTask(childUserId, model.TaskSubmitRequest{
		FamilyId: familyId,
		RecordId: claimed.Id,
	})
	TaskService.AuditTask(ownerUserId, model.TaskAuditRequest{
		RecordId: submitted.Id,
		Approved: true,
	})

	reward := RewardService.CreateReward(ownerUserId, model.RewardCreateRequest{
		FamilyId:   familyId,
		Name:       "Sticker",
		PointsCost: 2,
		Stock:      2,
	})
	RewardService.ApplyReward(childUserId, model.RewardApplyRequest{
		FamilyId: familyId,
		RewardId: reward.Id,
	})

	members := MemberService.ListMembers(ownerUserId, familyId)
	if !memberListContains(members, family.Member.Id, model.FamilyRoleOwner) {
		t.Fatalf("member list does not contain owner: %+v", members)
	}
	if !memberListContains(members, child.Id, model.FamilyRoleChild) {
		t.Fatalf("member list does not contain child: %+v", members)
	}

	parentLogs := PointService.ListPointLogs(ownerUserId, model.PointLogListRequest{FamilyId: familyId})
	if len(parentLogs) != 2 {
		t.Fatalf("parent logs length = %d, want 2: %+v", len(parentLogs), parentLogs)
	}
	if parentLogs[0].SourceType != model.PointSourceTypeReward || parentLogs[0].SourceTitle != reward.Name || parentLogs[0].Points != -2 {
		t.Fatalf("reward log = %+v, want reward title and -2 points", parentLogs[0])
	}
	if parentLogs[1].SourceType != model.PointSourceTypeTask || parentLogs[1].SourceTitle != task.Title || parentLogs[1].Points != 5 {
		t.Fatalf("task log = %+v, want task title and 5 points", parentLogs[1])
	}

	childLogs := PointService.ListPointLogs(childUserId, model.PointLogListRequest{FamilyId: familyId})
	if len(childLogs) != 2 {
		t.Fatalf("child logs length = %d, want 2: %+v", len(childLogs), childLogs)
	}

	mustPanicWith(t, "permission denied", func() {
		PointService.ListPointLogs(childUserId, model.PointLogListRequest{
			FamilyId: familyId,
			MemberId: family.Member.Id,
		})
	})
}

func memberListContains(list []model.FamilyMember, memberId int64, role model.FamilyRole) bool {
	for _, item := range list {
		if item.Id == memberId && item.RoleType == role {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: 运行数据库集成测试**

Run: `$env:PARENT_CHILD_DB_INTEGRATION='1'; go test ./srv -run TestIntegrationMemberListAndPointLogFlow -count=1 -v`

Expected: PASS。

- [ ] **Step 3: 全量验证**

Run: `go test ./... -count=1`

Expected: PASS。

Run: `$env:PARENT_CHILD_DB_INTEGRATION='1'; go test ./srv -run "TestIntegration(MemberListAndPointLogFlow|DashboardSummaryFlow|RewardExchangeFlow|TaskPointsFlow|FamilyMemberInviteFlow)" -count=1 -v`

Expected: PASS。

- [ ] **Step 4: 提交**

```bash
git add parent-child-api/srv/point_integration_test.go
git commit -m "test: 覆盖成员列表与积分流水流程"
```

---

## 自检

- Spec coverage: 覆盖成员列表、积分流水列表、孩子/家长权限、来源标题、API 路由、数据库集成验证。
- Placeholder scan: 无未定义的后续补充实现项。
- Type consistency: `PointLogListRequest`、`PointLogListItem`、`MemberService.ListMembers`、`PointService.ListPointLogs`、`pointLogScopedMemberId` 命名保持一致。
- Scope control: 未实现积分调整、分页、成员管理写操作或小程序前端接入。
