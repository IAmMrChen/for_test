# 虚拟孩子绑定真实账号 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 支持家长生成虚拟孩子绑定邀请，真实微信用户接受后绑定到原虚拟孩子成员。

**Architecture:** 新增独立 `family_child_bind_invites` 表和绑定邀请模型。后端沿用 `api -> srv -> model` 分层和原生 SQL；接受绑定时只更新原 `family_members` 行，不创建新成员。

**Tech Stack:** Go、MySQL、项目现有 `resx.Db.Main`、uni-app Vue 3。

---

## 文件结构

- Modify: `init.sql`
  - 增加 `family_child_bind_invites` 表。
- Modify: `parent-child-api/model/family_child_model.go`
  - 增加绑定邀请请求和响应模型。
- Modify: `parent-child-api/srv/member_service.go`
  - 增加创建绑定邀请、接受绑定邀请、加载虚拟孩子、加载绑定邀请方法。
- Modify: `parent-child-api/srv/member_service_test.go`
  - 增加缺少 member id、insert 参数单元测试。
- Modify: `parent-child-api/srv/integration_flow_test.go`
  - 集成 schema 增加绑定邀请表。
- Modify: `parent-child-api/srv/member_integration_test.go`
  - 增加绑定邀请完整流程测试。
- Modify: `parent-child-api/api/member/member_api.go`
  - 增加两个 handler。
- Modify: `parent-child-api/api/routes.go`
  - 注册两个路由。
- Modify: `parent-child-miniprogram/api/member.js`
  - 增加绑定邀请 API 封装。
- Modify: `parent-child-miniprogram/pages/profile/index.vue`
  - 虚拟孩子卡片增加邀请关联入口。
- Modify: `parent-child-miniprogram/pages/family-select/index.vue`
  - 增加接受绑定 token 入口。

---

### Task 1: 后端红测

**Files:**
- Modify: `parent-child-api/srv/member_service_test.go`
- Create: `parent-child-api/srv/member_integration_test.go`

- [ ] **Step 1: 单元红测 - 缺少 member id**

在 `member_service_test.go` 增加：

```go
func TestMemberServiceCreateVirtualChildBindInviteRejectsMissingMember(t *testing.T) {
	resx.Db = nil

	mustPanicWith(t, "member id is required", func() {
		MemberService.CreateVirtualChildBindInvite(1, model.VirtualChildBindInviteCreateRequest{FamilyId: 1})
	})
}
```

- [ ] **Step 2: 单元红测 - insert 参数**

在 `member_service_test.go` 增加：

```go
func TestVirtualChildBindInviteInsertArgs(t *testing.T) {
	expiresAt := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	args := virtualChildBindInviteInsertArgs(9, 10, 11, "token", expiresAt)

	if args[0] != int64(9) || args[1] != int64(10) || args[2] != int64(11) {
		t.Fatalf("ids args = %+v", args[:3])
	}
	if args[3] != "token" {
		t.Fatalf("token arg = %v, want token", args[3])
	}
	if args[4] != model.FamilyInviteStatusActive {
		t.Fatalf("status arg = %v, want ACTIVE", args[4])
	}
	if args[5] != expiresAt {
		t.Fatalf("expires arg = %v, want %v", args[5], expiresAt)
	}
}
```

同时给 `member_service_test.go` 增加 `time` import。

- [ ] **Step 3: 集成红测**

创建 `member_integration_test.go`：

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

func TestIntegrationVirtualChildBindInviteFlow(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed + 600
	childUserId := seed + 601
	anotherUserId := seed + 602

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("virtual-bind-family-%d", seed),
		Nickname: "owner",
	})
	familyId := family.Family.Id
	t.Cleanup(func() {
		cleanupIntegrationFamily(t, familyId)
	})

	virtualChild := MemberService.CreateVirtualChild(ownerUserId, model.VirtualChildCreateRequest{
		FamilyId: familyId,
		Nickname: "virtual-child",
	})
	resx.Db.Main.MustExecute(`
		UPDATE family_members
		SET current_points=@p1, total_earned_points=@p1
		WHERE id=@p2
	`, 12, virtualChild.Id)

	bindInvite := MemberService.CreateVirtualChildBindInvite(ownerUserId, model.VirtualChildBindInviteCreateRequest{
		FamilyId: familyId,
		MemberId: virtualChild.Id,
	})
	if bindInvite.Token == "" || bindInvite.ChildMemberId != virtualChild.Id {
		t.Fatalf("bind invite = %+v", bindInvite)
	}

	bound := MemberService.AcceptVirtualChildBindInvite(childUserId, bindInvite.Token)
	if bound.Id != virtualChild.Id {
		t.Fatalf("bound member id = %d, want original %d", bound.Id, virtualChild.Id)
	}
	if bound.UserId == nil || *bound.UserId != childUserId {
		t.Fatalf("bound user id = %v, want %d", bound.UserId, childUserId)
	}
	if bound.IsVirtual {
		t.Fatalf("bound child should not be virtual: %+v", bound)
	}
	if bound.CurrentPoints != 12 || bound.TotalEarnedPoints != 12 {
		t.Fatalf("bound points = current %d total %d, want 12/12", bound.CurrentPoints, bound.TotalEarnedPoints)
	}

	secondInvite := InviteService.CreateInvite(ownerUserId, model.FamilyInviteCreateRequest{
		FamilyId:   familyId,
		TargetRole: model.FamilyRoleChild,
	})
	mustPanicWith(t, "user already has a family identity", func() {
		InviteService.AcceptInvite(childUserId, secondInvite.Token)
	})

	existingParentInvite := InviteService.CreateInvite(ownerUserId, model.FamilyInviteCreateRequest{
		FamilyId:   familyId,
		TargetRole: model.FamilyRoleParent,
	})
	InviteService.AcceptInvite(anotherUserId, existingParentInvite.Token)

	secondVirtualChild := MemberService.CreateVirtualChild(ownerUserId, model.VirtualChildCreateRequest{
		FamilyId: familyId,
		Nickname: "second-virtual-child",
	})
	secondBindInvite := MemberService.CreateVirtualChildBindInvite(ownerUserId, model.VirtualChildBindInviteCreateRequest{
		FamilyId: familyId,
		MemberId: secondVirtualChild.Id,
	})
	mustPanicWith(t, "user already has a family identity", func() {
		MemberService.AcceptVirtualChildBindInvite(anotherUserId, secondBindInvite.Token)
	})
}
```

- [ ] **Step 4: 跑红测**

Run:

```powershell
$env:PARENT_CHILD_DB_INTEGRATION='1'; go test ./srv -run "TestMemberServiceCreateVirtualChildBindInviteRejectsMissingMember|TestVirtualChildBindInviteInsertArgs|TestIntegrationVirtualChildBindInviteFlow" -count=1
```

Working directory: `parent-child-api`

Expected: FAIL，原因是请求模型和服务方法未定义。

---

### Task 2: 后端实现

**Files:**
- Modify: `init.sql`
- Modify: `parent-child-api/model/family_child_model.go`
- Modify: `parent-child-api/srv/member_service.go`
- Modify: `parent-child-api/srv/integration_flow_test.go`
- Modify: `parent-child-api/api/member/member_api.go`
- Modify: `parent-child-api/api/routes.go`

- [ ] **Step 1: 增加模型**

在 `family_child_model.go` 增加：

```go
type VirtualChildBindInviteCreateRequest struct {
	FamilyId int64 `json:"familyId"`
	MemberId int64 `json:"memberId"`
}

type VirtualChildBindInviteAcceptRequest struct {
	Token string `json:"token"`
}

type VirtualChildBindInvite struct {
	Id               int64              `json:"id"`
	FamilyId         int64              `json:"familyId"`
	ChildMemberId    int64              `json:"childMemberId"`
	InviterMemberId  int64              `json:"inviterMemberId"`
	Token            string             `json:"token"`
	Status           FamilyInviteStatus `json:"status"`
	ExpiresAt        time.Time          `json:"expiresAt"`
	AcceptedByUserId *int64             `json:"acceptedByUserId"`
	AcceptedAt       *time.Time         `json:"acceptedAt"`
}
```

并增加 `time` import。

- [ ] **Step 2: 增加 schema**

在 `init.sql` 的 `family_invites` 后增加 `family_child_bind_invites` 建表 SQL。

在 `ensureIntegrationSchema` 中增加同样的建表逻辑，并在 `cleanupIntegrationFamily` 中删除该表数据。

- [ ] **Step 3: 实现 service**

在 `member_service.go` 中实现：

- `CreateVirtualChildBindInvite`
- `AcceptVirtualChildBindInvite`
- `virtualChildBindInviteInsertArgs`
- `loadActiveVirtualChildForBind`
- `loadActiveChildBindInviteFrom`
- `claimChildBindInviteSql`

创建绑定邀请复用同包内 `newInviteToken()`。

- [ ] **Step 4: 增加 API 和路由**

`member_api.go` 增加 `CreateVirtualChildBindInvite` 和 `AcceptVirtualChildBindInvite`。

`routes.go` 注册：

```go
mux.HandleFunc("POST /api/member/createVirtualChildBindInvite", recoverRoute(apix.WithAuth(jwtSecret, memberApi.CreateVirtualChildBindInvite)))
mux.HandleFunc("POST /api/member/acceptVirtualChildBindInvite", recoverRoute(apix.WithAuth(jwtSecret, memberApi.AcceptVirtualChildBindInvite)))
```

- [ ] **Step 5: 跑绑定目标测试**

Run:

```powershell
$env:PARENT_CHILD_DB_INTEGRATION='1'; go test ./srv -run "TestMemberServiceCreateVirtualChildBindInviteRejectsMissingMember|TestVirtualChildBindInviteInsertArgs|TestIntegrationVirtualChildBindInviteFlow" -count=1
```

Expected: PASS。

- [ ] **Step 6: 提交后端**

```powershell
git add init.sql parent-child-api\model\family_child_model.go parent-child-api\srv\member_service.go parent-child-api\srv\member_service_test.go parent-child-api\srv\member_integration_test.go parent-child-api\srv\integration_flow_test.go parent-child-api\api\member\member_api.go parent-child-api\api\routes.go
git commit -m "feat: 增加虚拟孩子绑定邀请接口"
```

---

### Task 3: 小程序入口

**Files:**
- Modify: `parent-child-miniprogram/api/member.js`
- Modify: `parent-child-miniprogram/pages/profile/index.vue`
- Modify: `parent-child-miniprogram/pages/family-select/index.vue`

- [ ] **Step 1: 增加 API 封装**

`member.js` 增加：

```js
export function createVirtualChildBindInvite(data) {
  return request({
    url: '/api/member/createVirtualChildBindInvite',
    method: 'POST',
    data
  })
}

export function acceptVirtualChildBindInvite(data) {
  return request({
    url: '/api/member/acceptVirtualChildBindInvite',
    method: 'POST',
    data
  })
}
```

- [ ] **Step 2: 个人中心增加邀请关联**

虚拟孩子 member card 增加“邀请关联”按钮，点击调用 `createVirtualChildBindInvite`，展示 token 和过期时间，并支持复制。

- [ ] **Step 3: 家庭选择页增加接受绑定**

新增一块“关联虚拟孩子”，用户输入 token 后调用 `acceptVirtualChildBindInvite`。成功后设置当前家庭为返回成员所属家庭，并进入首页。

- [ ] **Step 4: 前端验证**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
rg -n "é|å|ç|鐩|閸|æµ|æ¾|鍙|娴" parent-child-miniprogram\pages\profile\index.vue parent-child-miniprogram\pages\family-select\index.vue
```

Expected: `node --check` exit 0；`rg` exit 1。

- [ ] **Step 5: 提交小程序入口**

```powershell
git add parent-child-miniprogram\api\member.js parent-child-miniprogram\pages\profile\index.vue parent-child-miniprogram\pages\family-select\index.vue
git commit -m "feat: 增加小程序虚拟孩子绑定入口"
```

---

### Task 4: 全量验证

**Files:**
- Verify all changed files.

- [ ] **Step 1: 后端全量测试**

Run:

```powershell
go test ./... -count=1
```

Working directory: `parent-child-api`

Expected: exit 0。

- [ ] **Step 2: 前端语法检查**

Run:

```powershell
Get-ChildItem parent-child-miniprogram\config,parent-child-miniprogram\utils,parent-child-miniprogram\api -Filter *.js -Recurse | ForEach-Object { node --check $_.FullName }
```

Expected: exit 0。

- [ ] **Step 3: 查看工作区**

Run:

```powershell
git status --short
```

Expected: clean。

## 自查

- 绑定不创建新成员，只更新原虚拟孩子成员。
- 单家庭单身份规则在接受绑定时执行。
- 绑定后历史记录不迁移。
- 小程序具备生成 token 和接受 token 的最小入口。
