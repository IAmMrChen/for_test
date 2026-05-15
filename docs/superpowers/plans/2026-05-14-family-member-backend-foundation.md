# Family Member Backend Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立 MVP 后端第一阶段能力：微信用户识别、家庭创建、多家庭列表、虚拟孩子创建、按角色邀请成员、接受邀请，并落实“一个微信用户在一个家庭内只有一个身份”。

**Architecture:** 沿用当前 `parent-child-api` 的轻量结构和 `chat-api` 的分层风格：`api` 只负责 HTTP 入参出参，`srv` 负责业务和事务，`model` 定义业务模型和枚举，数据库访问使用原生 SQL + `resx.Db.Main`。本计划只覆盖后端第一阶段，任务、奖品、积分流水、前端页面接入单独进入后续计划。

**Tech Stack:** Go 1.25、标准库 `net/http`、`sqlmer`、MySQL、现有 JWT demo、原生 SQL、`go test`。

---

## 范围切分

`draft.md` 覆盖完整 MVP，包含家庭、成员、邀请、任务、奖品、积分、日志、通知和小程序页面。为了保持可测试、可验收，本计划只实现后端基础域：

本计划覆盖：

* 微信用户在服务端的创建/加载。
* 当前用户 JWT 身份进入业务接口。
* 创建家庭，并自动创建家主成员。
* 查询当前用户加入的家庭列表。
* 创建虚拟孩子。
* 创建邀请，邀请时指定角色。
* 接受邀请，并校验同一微信用户在同一家庭内只有一个身份。

本计划不覆盖：

* 任务创建、任务提交、任务审核。
* 奖品创建、兑换、发放、拒绝。
* 积分发放、扣减、流水。
* 微信真实 `code -> openid` 接口调用。
* 小程序页面接入。

后续建议计划：

1. `task-points-backend-plan`：任务、审核、积分发放。
2. `reward-exchange-backend-plan`：奖品、兑换、库存、积分扣减/回退。
3. `miniprogram-api-integration-plan`：小程序接入真实 API。

## 目标文件结构

```text
parent-child-api/
  api/
    routes.go
    apix/
      auth_middleware.go
      json.go
    family/
      family_api.go
    member/
      member_api.go
    invite/
      invite_api.go
  model/
    family_model.go
    member_model.go
    invite_model.go
    user_model.go
  srv/
    family_service.go
    family_service_test.go
    invite_service.go
    invite_service_test.go
    member_service.go
    member_service_test.go
    user_service.go
  init.sql
```

## 约定

* 所有数据库读写优先写在 `srv` 层，使用原生 SQL。
* `api` 层不直接访问 `resx.Db`。
* 所有跨表写入使用事务。
* 所有业务错误先用普通 `panic(fmt.Errorf(...))`，等错误码体系成熟后再统一替换。
* 单元测试优先覆盖纯业务规则；数据库 SQL 通过清晰的原生语句和后续集成测试验证。
* 每个任务完成后运行：

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./...
```

预期：所有包测试通过。

---

### Task 1: 成员角色与状态模型

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\model\member_model.go`
- Create: `D:\projects\github\for_test\parent-child-api\model\member_model_test.go`

- [ ] **Step 1: Write the failing test**

创建 `model/member_model_test.go`：

```go
package model

import "testing"

func TestFamilyRoleIsParentRole(t *testing.T) {
	tests := []struct {
		role FamilyRole
		want bool
	}{
		{FamilyRoleOwner, true},
		{FamilyRoleAdmin, true},
		{FamilyRoleParent, true},
		{FamilyRoleChild, false},
	}

	for _, tt := range tests {
		if got := tt.role.IsParentRole(); got != tt.want {
			t.Fatalf("%s IsParentRole = %v, want %v", tt.role, got, tt.want)
		}
	}
}

func TestFamilyRoleCanManageMembers(t *testing.T) {
	if !FamilyRoleOwner.CanManageMembers() {
		t.Fatal("owner should manage members")
	}
	if !FamilyRoleAdmin.CanManageMembers() {
		t.Fatal("admin should manage members")
	}
	if FamilyRoleParent.CanManageMembers() {
		t.Fatal("parent should not manage parent/admin members")
	}
	if FamilyRoleChild.CanManageMembers() {
		t.Fatal("child should not manage members")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./model
```

Expected: FAIL，提示 `undefined: FamilyRole`。

- [ ] **Step 3: Write minimal implementation**

创建 `model/member_model.go`：

```go
package model

type FamilyRole string

const (
	FamilyRoleOwner  FamilyRole = "OWNER"
	FamilyRoleAdmin  FamilyRole = "ADMIN"
	FamilyRoleParent FamilyRole = "PARENT"
	FamilyRoleChild  FamilyRole = "CHILD"
)

func (x FamilyRole) IsParentRole() bool {
	return x == FamilyRoleOwner || x == FamilyRoleAdmin || x == FamilyRoleParent
}

func (x FamilyRole) CanManageMembers() bool {
	return x == FamilyRoleOwner || x == FamilyRoleAdmin
}

type FamilyMemberStatus string

const (
	FamilyMemberStatusActive  FamilyMemberStatus = "ACTIVE"
	FamilyMemberStatusRemoved FamilyMemberStatus = "REMOVED"
)

type FamilyMember struct {
	Id                int64              `json:"id"`
	FamilyId          int64              `json:"familyId"`
	UserId            *int64             `json:"userId"`
	RoleType          FamilyRole         `json:"roleType"`
	Nickname          string             `json:"nickname"`
	CurrentPoints     int                `json:"currentPoints"`
	TotalEarnedPoints int                `json:"totalEarnedPoints"`
	IsVirtual         bool               `json:"isVirtual"`
	Status            FamilyMemberStatus `json:"status"`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./model
```

Expected: PASS。

---

### Task 2: 对齐数据库初始化脚本

**Files:**
- Modify: `D:\projects\github\for_test\init.sql`

- [ ] **Step 1: Add schema changes**

在 `family_members` 表中增加成员状态字段，并补充唯一约束。修改表定义中的索引区域为：

```sql
  `status` ENUM('ACTIVE', 'REMOVED') NOT NULL DEFAULT 'ACTIVE' COMMENT '成员状态: ACTIVE-正常, REMOVED-已移除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_family_user` (`family_id`, `user_id`),
  KEY `idx_family_role` (`family_id`, `role_type`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='家庭成员关系表';
```

在 `point_logs` 表之后新增邀请表：

```sql
-- 9. 家庭邀请表 (Family Invites)
-- 记录家庭成员邀请，邀请时固定目标角色。
CREATE TABLE IF NOT EXISTS `family_invites` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `family_id` BIGINT UNSIGNED NOT NULL COMMENT '家庭ID',
  `inviter_member_id` BIGINT UNSIGNED NOT NULL COMMENT '邀请人成员ID',
  `target_role` ENUM('ADMIN', 'PARENT', 'CHILD') NOT NULL COMMENT '受邀加入角色',
  `token` VARCHAR(128) NOT NULL COMMENT '邀请令牌',
  `status` ENUM('ACTIVE', 'ACCEPTED', 'EXPIRED', 'CANCELED') NOT NULL DEFAULT 'ACTIVE' COMMENT '邀请状态',
  `expires_at` DATETIME NOT NULL COMMENT '过期时间',
  `accepted_by_user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '接受邀请的用户ID',
  `accepted_at` DATETIME DEFAULT NULL COMMENT '接受时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_token` (`token`),
  KEY `idx_family_status` (`family_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='家庭邀请表';
```

- [ ] **Step 2: Verify SQL text contains required objects**

Run:

```powershell
Select-String -Path init.sql -Pattern "family_invites","uk_family_user","status.*ACTIVE"
```

Expected: 输出包含 `family_invites`、`uk_family_user`、`ACTIVE`。

---

### Task 3: 家庭模型

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\model\family_model.go`

- [ ] **Step 1: Add model file**

创建 `model/family_model.go`：

```go
package model

type Family struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	CreatorId int64  `json:"creatorId"`
}

type FamilyCreateRequest struct {
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
}

type FamilyCreateResponse struct {
	Family Family       `json:"family"`
	Member FamilyMember `json:"member"`
}

type FamilyListItem struct {
	FamilyId   int64      `json:"familyId"`
	FamilyName string     `json:"familyName"`
	MemberId   int64      `json:"memberId"`
	RoleType   FamilyRole `json:"roleType"`
	Nickname   string     `json:"nickname"`
}
```

- [ ] **Step 2: Run model tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./model
```

Expected: PASS。

---

### Task 4: 家庭服务创建逻辑

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\srv\family_service.go`

- [ ] **Step 1: Write service implementation with transaction and raw SQL**

创建 `srv/family_service.go`：

```go
package srv

import (
	"fmt"
	"strings"

	"parent-child-api/model"
	"parent-child-api/resx"
)

var FamilyService familyService

type familyService struct{}

func (x familyService) CreateFamily(userId int64, req model.FamilyCreateRequest) model.FamilyCreateResponse {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		panic(fmt.Errorf("family name is required"))
	}

	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = "家主"
	}

	tx := resx.Db.Main.MustCreateTransactionEx()
	defer tx.MustClose()

	insertFamilySQL := `
		INSERT INTO families(name, creator_id)
		VALUES(@p1, @p2)
	`
	tx.MustExecute(insertFamilySQL, name, userId)

	familyId, ok := tx.MustScalarInt64("SELECT LAST_INSERT_ID()")
	if !ok {
		panic(fmt.Errorf("failed to load created family id"))
	}

	insertMemberSQL := `
		INSERT INTO family_members(
			family_id, user_id, role_type, nickname,
			current_points, total_earned_points, is_virtual, created_by, status
		) VALUES(
			@p1, @p2, @p3, @p4,
			0, 0, 0, @p2, @p5
		)
	`
	tx.MustExecute(insertMemberSQL, familyId, userId, model.FamilyRoleOwner, nickname, model.FamilyMemberStatusActive)

	memberId, ok := tx.MustScalarInt64("SELECT LAST_INSERT_ID()")
	if !ok {
		panic(fmt.Errorf("failed to load created family member id"))
	}

	tx.MustCommit()

	return model.FamilyCreateResponse{
		Family: model.Family{
			Id:        familyId,
			Name:      name,
			CreatorId: userId,
		},
		Member: model.FamilyMember{
			Id:        memberId,
			FamilyId:  familyId,
			UserId:    &userId,
			RoleType:  model.FamilyRoleOwner,
			Nickname:  nickname,
			IsVirtual: false,
			Status:    model.FamilyMemberStatusActive,
		},
	}
}

func (x familyService) ListFamilies(userId int64) []model.FamilyListItem {
	const sql = `
		SELECT f.id AS family_id
			, f.name AS family_name
			, m.id AS member_id
			, m.role_type
			, m.nickname
		FROM family_members m
		INNER JOIN families f ON f.id = m.family_id
		WHERE m.user_id=@p1 AND m.status=@p2
		ORDER BY m.id DESC
	`

	return resx.Db.Main.MustListOf(new(model.FamilyListItem), sql, userId, model.FamilyMemberStatusActive).([]model.FamilyListItem)
}
```

- [ ] **Step 2: Run compile check**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: PASS。如果 `MustScalarInt64` 不存在，替换为 `MustScalarInt` 并转换为 `int64`。

---

### Task 5: 家庭 API

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\api\family\family_api.go`
- Modify: `D:\projects\github\for_test\parent-child-api\api\routes.go`

- [ ] **Step 1: Add family API**

创建 `api/family/family_api.go`：

```go
package familyapi

import (
	"encoding/json"
	"net/http"

	"parent-child-api/api/apix"
	"parent-child-api/model"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type FamilyApi struct{}

func (x FamilyApi) Create(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.FamilyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	apix.WriteData(w, srv.FamilyService.CreateFamily(token.UserId, req))
}

func (x FamilyApi) List(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	apix.WriteData(w, srv.FamilyService.ListFamilies(token.UserId))
}
```

- [ ] **Step 2: Register routes**

修改 `api/routes.go`：

```go
package api

import (
	"net/http"

	"parent-child-api/api/apix"
	authapi "parent-child-api/api/auth"
	demoapi "parent-child-api/api/demo"
	familyapi "parent-child-api/api/family"
)

func RegisterRoutes(mux *http.ServeMux, jwtSecret string) {
	authApi := authapi.AuthApi{JwtSecret: jwtSecret}
	demoApi := demoapi.DemoApi{}
	familyApi := familyapi.FamilyApi{}

	mux.HandleFunc("POST /api/auth/demoLogin", authApi.DemoLogin)
	mux.HandleFunc("GET /api/demo/me", apix.WithAuth(jwtSecret, demoApi.Me))
	mux.HandleFunc("POST /api/family/create", apix.WithAuth(jwtSecret, familyApi.Create))
	mux.HandleFunc("GET /api/family/list", apix.WithAuth(jwtSecret, familyApi.List))
}
```

- [ ] **Step 3: Run tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./...
```

Expected: PASS。

---

### Task 6: 虚拟孩子创建

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\model\family_child_model.go`
- Create: `D:\projects\github\for_test\parent-child-api\api\member\member_api.go`
- Modify: `D:\projects\github\for_test\parent-child-api\srv\member_service.go`
- Modify: `D:\projects\github\for_test\parent-child-api\api\routes.go`

- [ ] **Step 1: Add request model**

创建 `model/family_child_model.go`：

```go
package model

type VirtualChildCreateRequest struct {
	FamilyId int64  `json:"familyId"`
	Nickname string `json:"nickname"`
}
```

- [ ] **Step 2: Add member service**

创建 `srv/member_service.go`：

```go
package srv

import (
	"fmt"
	"strings"

	"parent-child-api/model"
	"parent-child-api/resx"
)

var MemberService memberService

type memberService struct{}

func (x memberService) LoadActiveMember(userId, familyId int64) *model.FamilyMember {
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
		WHERE family_id=@p1 AND user_id=@p2 AND status=@p3
	`

	member := &model.FamilyMember{}
	ok := resx.Db.Main.MustGetStruct(member, sql, familyId, userId, model.FamilyMemberStatusActive)
	if !ok {
		return nil
	}
	return member
}

func (x memberService) RequireParentRole(userId, familyId int64) model.FamilyMember {
	member := x.LoadActiveMember(userId, familyId)
	if member == nil || !member.RoleType.IsParentRole() {
		panic(fmt.Errorf("permission denied"))
	}
	return *member
}

func (x memberService) CreateVirtualChild(operatorUserId int64, req model.VirtualChildCreateRequest) model.FamilyMember {
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		panic(fmt.Errorf("nickname is required"))
	}

	operator := x.RequireParentRole(operatorUserId, req.FamilyId)

	const sql = `
		INSERT INTO family_members(
			family_id, user_id, role_type, nickname,
			current_points, total_earned_points, is_virtual, created_by, status
		) VALUES(
			@p1, NULL, @p2, @p3,
			0, 0, 1, @p4, @p5
		)
	`
	resx.Db.Main.MustExecute(sql, req.FamilyId, model.FamilyRoleChild, nickname, operator.Id, model.FamilyMemberStatusActive)

	memberId, ok := resx.Db.Main.MustScalarInt64("SELECT LAST_INSERT_ID()")
	if !ok {
		panic(fmt.Errorf("failed to load created virtual child id"))
	}

	return model.FamilyMember{
		Id:        memberId,
		FamilyId:  req.FamilyId,
		RoleType:  model.FamilyRoleChild,
		Nickname:  nickname,
		IsVirtual: true,
		Status:    model.FamilyMemberStatusActive,
	}
}
```

- [ ] **Step 3: Add member API**

创建 `api/member/member_api.go`：

```go
package memberapi

import (
	"encoding/json"
	"net/http"

	"parent-child-api/api/apix"
	"parent-child-api/model"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type MemberApi struct{}

func (x MemberApi) CreateVirtualChild(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.VirtualChildCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	apix.WriteData(w, srv.MemberService.CreateVirtualChild(token.UserId, req))
}
```

- [ ] **Step 4: Register route**

在 `api/routes.go` 中引入 `memberapi "parent-child-api/api/member"`，并在 `RegisterRoutes` 中加入：

```go
memberApi := memberapi.MemberApi{}
mux.HandleFunc("POST /api/member/createVirtualChild", apix.WithAuth(jwtSecret, memberApi.CreateVirtualChild))
```

- [ ] **Step 5: Run tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./...
```

Expected: PASS。如果 `MustScalarInt64` 不存在，按 Task 4 的处理方式替换。

---

### Task 7: 邀请模型与服务

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\model\invite_model.go`
- Create: `D:\projects\github\for_test\parent-child-api\srv\invite_service.go`

- [ ] **Step 1: Add invite model**

创建 `model/invite_model.go`：

```go
package model

import "time"

type FamilyInviteStatus string

const (
	FamilyInviteStatusActive   FamilyInviteStatus = "ACTIVE"
	FamilyInviteStatusAccepted FamilyInviteStatus = "ACCEPTED"
	FamilyInviteStatusExpired  FamilyInviteStatus = "EXPIRED"
	FamilyInviteStatusCanceled FamilyInviteStatus = "CANCELED"
)

type FamilyInviteCreateRequest struct {
	FamilyId   int64      `json:"familyId"`
	TargetRole FamilyRole `json:"targetRole"`
}

type FamilyInviteAcceptRequest struct {
	Token string `json:"token"`
}

type FamilyInvite struct {
	Id               int64              `json:"id"`
	FamilyId         int64              `json:"familyId"`
	InviterMemberId  int64              `json:"inviterMemberId"`
	TargetRole       FamilyRole         `json:"targetRole"`
	Token            string             `json:"token"`
	Status           FamilyInviteStatus `json:"status"`
	ExpiresAt        time.Time          `json:"expiresAt"`
	AcceptedByUserId *int64             `json:"acceptedByUserId"`
	AcceptedAt       *time.Time         `json:"acceptedAt"`
}
```

- [ ] **Step 2: Add invite service**

创建 `srv/invite_service.go`：

```go
package srv

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"parent-child-api/model"
	"parent-child-api/resx"
)

var InviteService inviteService

type inviteService struct{}

func (x inviteService) CreateInvite(operatorUserId int64, req model.FamilyInviteCreateRequest) model.FamilyInvite {
	if req.TargetRole != model.FamilyRoleAdmin &&
		req.TargetRole != model.FamilyRoleParent &&
		req.TargetRole != model.FamilyRoleChild {
		panic(fmt.Errorf("invalid target role: %s", req.TargetRole))
	}

	operator := MemberService.RequireParentRole(operatorUserId, req.FamilyId)
	if req.TargetRole == model.FamilyRoleAdmin && operator.RoleType != model.FamilyRoleOwner {
		panic(fmt.Errorf("only owner can invite admin"))
	}

	token := newInviteToken()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	const sql = `
		INSERT INTO family_invites(
			family_id, inviter_member_id, target_role, token, status, expires_at
		) VALUES(
			@p1, @p2, @p3, @p4, @p5, @p6
		)
	`
	resx.Db.Main.MustExecute(sql, req.FamilyId, operator.Id, req.TargetRole, token, model.FamilyInviteStatusActive, expiresAt)

	id, ok := resx.Db.Main.MustScalarInt64("SELECT LAST_INSERT_ID()")
	if !ok {
		panic(fmt.Errorf("failed to load created invite id"))
	}

	return model.FamilyInvite{
		Id:              id,
		FamilyId:        req.FamilyId,
		InviterMemberId: operator.Id,
		TargetRole:      req.TargetRole,
		Token:           token,
		Status:          model.FamilyInviteStatusActive,
		ExpiresAt:       expiresAt,
	}
}

func (x inviteService) AcceptInvite(userId int64, token string) model.FamilyMember {
	invite := x.loadActiveInvite(token)
	if invite.ExpiresAt.Before(time.Now()) {
		panic(fmt.Errorf("invite expired"))
	}

	existing := MemberService.LoadActiveMember(userId, invite.FamilyId)
	if existing != nil {
		panic(fmt.Errorf("user already has a family identity"))
	}

	tx := resx.Db.Main.MustCreateTransactionEx()
	defer tx.MustClose()

	insertMemberSQL := `
		INSERT INTO family_members(
			family_id, user_id, role_type, nickname,
			current_points, total_earned_points, is_virtual, created_by, status
		) VALUES(
			@p1, @p2, @p3, '', 0, 0, 0, @p4, @p5
		)
	`
	tx.MustExecute(insertMemberSQL, invite.FamilyId, userId, invite.TargetRole, invite.InviterMemberId, model.FamilyMemberStatusActive)

	memberId, ok := tx.MustScalarInt64("SELECT LAST_INSERT_ID()")
	if !ok {
		panic(fmt.Errorf("failed to load accepted member id"))
	}

	updateInviteSQL := `
		UPDATE family_invites
		SET status=@p1, accepted_by_user_id=@p2, accepted_at=NOW()
		WHERE id=@p3 AND status=@p4
	`
	affected := tx.MustExecute(updateInviteSQL, model.FamilyInviteStatusAccepted, userId, invite.Id, model.FamilyInviteStatusActive)
	if affected != 1 {
		panic(fmt.Errorf("invite has been changed"))
	}

	tx.MustCommit()

	return model.FamilyMember{
		Id:        memberId,
		FamilyId:  invite.FamilyId,
		UserId:    &userId,
		RoleType:  invite.TargetRole,
		IsVirtual: false,
		Status:    model.FamilyMemberStatusActive,
	}
}

func (x inviteService) loadActiveInvite(token string) model.FamilyInvite {
	const sql = `
		SELECT id
			, family_id
			, inviter_member_id
			, target_role
			, token
			, status
			, expires_at
			, accepted_by_user_id
			, accepted_at
		FROM family_invites
		WHERE token=@p1 AND status=@p2
	`

	invite := &model.FamilyInvite{}
	ok := resx.Db.Main.MustGetStruct(invite, sql, token, model.FamilyInviteStatusActive)
	if !ok {
		panic(fmt.Errorf("invite not found"))
	}
	return *invite
}

func newInviteToken() string {
	var data [32]byte
	if _, err := rand.Read(data[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data[:])
}
```

- [ ] **Step 3: Run tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./...
```

Expected: PASS。如果 `MustScalarInt64` 不存在，按前面任务替换。

---

### Task 8: 邀请 API

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\api\invite\invite_api.go`
- Modify: `D:\projects\github\for_test\parent-child-api\api\routes.go`

- [ ] **Step 1: Add invite API**

创建 `api/invite/invite_api.go`：

```go
package inviteapi

import (
	"encoding/json"
	"net/http"

	"parent-child-api/api/apix"
	"parent-child-api/model"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type InviteApi struct{}

func (x InviteApi) Create(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.FamilyInviteCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	apix.WriteData(w, srv.InviteService.CreateInvite(token.UserId, req))
}

func (x InviteApi) Accept(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.FamilyInviteAcceptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	apix.WriteData(w, srv.InviteService.AcceptInvite(token.UserId, req.Token))
}
```

- [ ] **Step 2: Register invite routes**

在 `api/routes.go` 中引入 `inviteapi "parent-child-api/api/invite"`，并在 `RegisterRoutes` 中加入：

```go
inviteApi := inviteapi.InviteApi{}
mux.HandleFunc("POST /api/invite/create", apix.WithAuth(jwtSecret, inviteApi.Create))
mux.HandleFunc("POST /api/invite/accept", apix.WithAuth(jwtSecret, inviteApi.Accept))
```

- [ ] **Step 3: Run tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./...
```

Expected: PASS。

---

### Task 9: 自检和手动接口验收

**Files:**
- No source changes unless verification exposes a compile issue.

- [ ] **Step 1: Run full backend tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./...
```

Expected: PASS。

- [ ] **Step 2: Start API server**

Run:

```powershell
go run .
```

Expected: 服务监听 `:8080`。如果数据库连接失败，先确认 `parent-child-api/conf.toml` 的 `[db].Main` 是否可访问。

- [ ] **Step 3: Get demo token**

Run:

```powershell
$login = Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/auth/demoLogin -ContentType 'application/json' -Body '{"userId":1,"openid":"demo-openid","nickname":"爸爸"}'
$login.data.token
```

Expected: 输出一段 JWT。

- [ ] **Step 4: Verify auth endpoint**

Run:

```powershell
$headers = @{ Authorization = "Bearer $($login.data.token)" }
Invoke-RestMethod -Method Get -Uri http://localhost:8080/api/demo/me -Headers $headers
```

Expected: 返回当前用户信息。

- [ ] **Step 5: Verify family create**

Run:

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/family/create -Headers $headers -ContentType 'application/json' -Body '{"name":"快乐一家人","nickname":"爸爸"}'
```

Expected: 返回家庭和家主成员。

- [ ] **Step 6: Verify family list**

Run:

```powershell
Invoke-RestMethod -Method Get -Uri http://localhost:8080/api/family/list -Headers $headers
```

Expected: 返回包含“快乐一家人”的家庭列表。

## Self-Review

Spec coverage:

* 覆盖 `draft.md` 的用户识别、创建家庭、多家庭、创建虚拟孩子、邀请真实微信用户、邀请时指定角色、单家庭单身份。
* 未覆盖任务、奖品、积分、日志、通知和小程序接入；这些已在范围切分中列为后续计划。

Placeholder scan:

* 未发现占位性质的待补内容。
* 所有新增文件都有明确路径和代码片段。

Type consistency:

* `FamilyRole` 在模型、服务、API 中统一使用。
* `FamilyMemberStatus` 和 `FamilyInviteStatus` 在 SQL 和 Go 代码中保持字符串枚举一致。
* API 路由均通过 `apix.WithAuth` 接入当前 JWT 用户。
