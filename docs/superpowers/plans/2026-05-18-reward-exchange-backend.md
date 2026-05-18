# Reward Exchange Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 落地 MVP 的“家长创建奖品、孩子申请兑换、家长发放或驳回、孩子确认领取”的后端闭环。

**Architecture:** 继续沿用 `parent-child-api` 的分层方式：`api/reward` 只负责 HTTP 入参和响应，`srv/reward_service.go` 承载奖品、兑换、库存、积分扣减/回退的事务逻辑，`model/reward_model.go` 定义 DTO 与状态枚举。数据库继续使用 `resx.Db.Main` + 原生 SQL；兑换申请、扣积分、扣库存、写 `point_logs` 必须在同一事务内完成，驳回时积分和库存回退也必须在同一事务内完成。

**Tech Stack:** Go 1.25, net/http, github.com/bunnier/sqlmer, MySQL, 原生 SQL, go test。

---

## 范围

本计划只覆盖奖品与兑换后端，不覆盖前端页面、通知、默认奖品初始化、复杂库存预警和兑换备注上传。

需要完成：

* 家长创建奖品。
* 家长和孩子查看家庭内上架奖品。
* 孩子申请兑换奖品，家长可代虚拟孩子申请兑换。
* 申请兑换时原子扣减孩子积分；有限库存奖品同步扣减库存。
* 家长发放奖品，将兑换记录从 `APPLIED` 改为 `DELIVERED`。
* 家长驳回兑换，将兑换记录从 `APPLIED` 改为 `REJECTED`，并回退积分和库存。
* 孩子确认领取，将兑换记录从 `DELIVERED` 改为 `RECEIVED`。
* 家长可查看全家兑换记录，孩子只能查看自己的兑换记录。
* 集成测试覆盖“赚积分 -> 申请兑换 -> 发放 -> 确认领取”和“申请兑换 -> 驳回退款”两个核心分支。

不做：

* 奖品编辑和上下架接口。
* 默认奖品库初始化。
* 奖品图片。
* 兑换申请备注。
* 多阶段物流或履约明细。

## 文件结构

```text
parent-child-api/
  api/
    routes.go
    routes_test.go
    reward/
      reward_api.go
  model/
    reward_model.go
    reward_model_test.go
  srv/
    reward_service.go
    reward_service_test.go
    reward_integration_test.go
    integration_flow_test.go
  init.sql
```

## API

```text
POST /api/reward/create
GET  /api/reward/list?familyId=1
POST /api/reward/apply
POST /api/reward/deliver
POST /api/reward/reject
POST /api/reward/receive
GET  /api/reward/records?familyId=1&status=APPLIED
```

---

### Task 1: 奖品模型与状态枚举

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\model\reward_model.go`
- Create: `D:\projects\github\for_test\parent-child-api\model\reward_model_test.go`

- [ ] **Step 1: Write the failing test**

创建 `model/reward_model_test.go`：

```go
package model

import "testing"

func TestRewardRecordStatusTransitions(t *testing.T) {
	if !RewardRecordStatusApplied.CanOperate() {
		t.Fatal("applied record should be operable by parent")
	}
	if RewardRecordStatusDelivered.CanOperate() {
		t.Fatal("delivered record should not be parent-operable")
	}
	if !RewardRecordStatusDelivered.CanReceive() {
		t.Fatal("delivered record should be receivable by child")
	}
	if RewardRecordStatusApplied.CanReceive() {
		t.Fatal("applied record should not be receivable")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./model
```

Expected: FAIL with `undefined: RewardRecordStatusApplied`.

- [ ] **Step 3: Write minimal implementation**

创建 `model/reward_model.go`：

```go
package model

import "time"

type RewardStatus int

const (
	RewardStatusOffShelf RewardStatus = 0
	RewardStatusActive   RewardStatus = 1
)

type RewardRecordStatus string

const (
	RewardRecordStatusApplied  RewardRecordStatus = "APPLIED"
	RewardRecordStatusDelivered RewardRecordStatus = "DELIVERED"
	RewardRecordStatusReceived RewardRecordStatus = "RECEIVED"
	RewardRecordStatusRejected RewardRecordStatus = "REJECTED"
)

func (x RewardRecordStatus) CanOperate() bool {
	return x == RewardRecordStatusApplied
}

func (x RewardRecordStatus) CanReceive() bool {
	return x == RewardRecordStatusDelivered
}

type Reward struct {
	Id         int64        `json:"id"`
	FamilyId   int64        `json:"familyId"`
	Name       string       `json:"name"`
	PointsCost int          `json:"pointsCost"`
	Stock      int          `json:"stock"`
	Status     RewardStatus `json:"status"`
	CreatedBy  int64        `json:"createdBy"`
}

type RewardCreateRequest struct {
	FamilyId   int64  `json:"familyId"`
	Name       string `json:"name"`
	PointsCost int    `json:"pointsCost"`
	Stock      int    `json:"stock"`
}

type RewardApplyRequest struct {
	FamilyId int64 `json:"familyId"`
	RewardId int64 `json:"rewardId"`
	MemberId int64 `json:"memberId"`
}

type RewardRecordOperateRequest struct {
	RecordId int64 `json:"recordId"`
}

type RewardRecordListRequest struct {
	FamilyId int64              `json:"familyId"`
	Status   RewardRecordStatus `json:"status"`
}

type RewardRecord struct {
	Id          int64              `json:"id"`
	FamilyId    int64              `json:"familyId"`
	RewardId    int64              `json:"rewardId"`
	MemberId    int64              `json:"memberId"`
	PointsCost  int                `json:"pointsCost"`
	Status      RewardRecordStatus `json:"status"`
	ApplyTime   time.Time          `json:"applyTime"`
	OperateTime *time.Time         `json:"operateTime"`
	OperateBy   *int64             `json:"operateBy"`
}

type RewardRecordListItem struct {
	Id          int64              `json:"id"`
	FamilyId    int64              `json:"familyId"`
	RewardId    int64              `json:"rewardId"`
	RewardName  string             `json:"rewardName"`
	MemberId    int64              `json:"memberId"`
	Nickname    string             `json:"nickname"`
	PointsCost  int                `json:"pointsCost"`
	Status      RewardRecordStatus `json:"status"`
	ApplyTime   time.Time          `json:"applyTime"`
	OperateTime *time.Time         `json:"operateTime"`
	OperateBy   *int64             `json:"operateBy"`
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
git add parent-child-api/model/reward_model.go parent-child-api/model/reward_model_test.go
git commit -m "feat: add reward models"
```

---

### Task 2: Schema 对齐与集成迁移

**Files:**
- Modify: `D:\projects\github\for_test\init.sql`
- Modify: `D:\projects\github\for_test\parent-child-api\srv\integration_flow_test.go`

- [ ] **Step 1: Update schema indexes**

在 `rewards` 表补充：

```sql
KEY `idx_family_status` (`family_id`, `status`)
```

在 `reward_records` 表补充：

```sql
KEY `idx_reward_member_status` (`reward_id`, `member_id`, `status`),
KEY `idx_family_status` (`family_id`, `status`)
```

- [ ] **Step 2: Extend integration schema**

在 `ensureIntegrationSchema` 中确保 `rewards`、`reward_records` 存在；如果不存在，按 `init.sql` 的字段创建。再补充以下索引：

```go
if !integrationIndexExists("rewards", "idx_family_status") {
	resx.Db.Main.MustExecute("CREATE INDEX idx_family_status ON rewards(family_id, status)")
}
if !integrationIndexExists("reward_records", "idx_reward_member_status") {
	resx.Db.Main.MustExecute("CREATE INDEX idx_reward_member_status ON reward_records(reward_id, member_id, status)")
}
if !integrationIndexExists("reward_records", "idx_family_status") {
	resx.Db.Main.MustExecute("CREATE INDEX idx_family_status ON reward_records(family_id, status)")
}
```

- [ ] **Step 3: Extend cleanup**

在 `cleanupIntegrationFamily` 中删除顺序补充：

```go
resx.Db.Main.MustExecute("DELETE FROM reward_records WHERE family_id=@p1", familyId)
resx.Db.Main.MustExecute("DELETE FROM rewards WHERE family_id=@p1", familyId)
```

放在删除 `family_members` 之前。

- [ ] **Step 4: Verify schema text**

Run:

```powershell
Select-String -Path init.sql -Pattern "idx_reward_member_status","idx_family_status"
```

Expected: output contains reward indexes.

- [ ] **Step 5: Commit**

```powershell
git add init.sql parent-child-api/srv/integration_flow_test.go
git commit -m "feat: prepare reward exchange schema"
```

---

### Task 3: 奖品创建与列表服务

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\srv\reward_service.go`
- Create: `D:\projects\github\for_test\parent-child-api\srv\reward_service_test.go`

- [ ] **Step 1: Write the failing test**

创建 `srv/reward_service_test.go`：

```go
package srv

import (
	"fmt"
	"testing"

	"parent-child-api/model"
	"parent-child-api/resx"
)

func TestRewardServiceCreateRewardPanicsWhenNameIsEmpty(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("CreateReward should panic")
		}
		if fmt.Sprint(v) != "reward name is required" {
			t.Fatalf("panic = %v, want reward name is required", v)
		}
	}()

	RewardService.CreateReward(1, model.RewardCreateRequest{FamilyId: 1, Name: " "})
}

func TestNormalizeRewardCreateRequest(t *testing.T) {
	req := normalizeRewardCreateRequest(model.RewardCreateRequest{
		FamilyId:   1,
		Name:       " Ice cream ",
		PointsCost: 0,
		Stock:      0,
	})

	if req.Name != "Ice cream" {
		t.Fatalf("name = %q, want Ice cream", req.Name)
	}
	if req.PointsCost != 1 {
		t.Fatalf("pointsCost = %d, want 1", req.PointsCost)
	}
	if req.Stock != -1 {
		t.Fatalf("stock = %d, want -1 unlimited stock", req.Stock)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: FAIL with `undefined: RewardService`.

- [ ] **Step 3: Implement create and list**

创建 `srv/reward_service.go`：

```go
package srv

import (
	"fmt"
	"strings"

	"parent-child-api/model"
	"parent-child-api/resx"
)

var RewardService rewardService

type rewardService struct{}

func normalizeRewardCreateRequest(req model.RewardCreateRequest) model.RewardCreateRequest {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		panic(fmt.Errorf("reward name is required"))
	}
	if req.PointsCost <= 0 {
		req.PointsCost = 1
	}
	if req.Stock == 0 {
		req.Stock = -1
	}
	return req
}

func (x rewardService) CreateReward(operatorUserId int64, req model.RewardCreateRequest) model.Reward {
	req = normalizeRewardCreateRequest(req)
	operator := MemberService.RequireParentRole(operatorUserId, req.FamilyId)

	const sql = `
		INSERT INTO rewards(family_id, name, points_cost, stock, status, created_by)
		VALUES(@p1, @p2, @p3, @p4, @p5, @p6)
	`
	resx.Db.Main.MustExecute(sql, req.FamilyId, req.Name, req.PointsCost, req.Stock, model.RewardStatusActive, operator.Id)

	rewardIdValue, ok := resx.Db.Main.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || rewardIdValue == nil {
		panic(fmt.Errorf("failed to load created reward id"))
	}

	return model.Reward{
		Id:         int64(*rewardIdValue),
		FamilyId:   req.FamilyId,
		Name:       req.Name,
		PointsCost: req.PointsCost,
		Stock:      req.Stock,
		Status:     model.RewardStatusActive,
		CreatedBy:  operator.Id,
	}
}

func (x rewardService) ListRewards(userId int64, familyId int64) []model.Reward {
	if MemberService.LoadActiveMember(userId, familyId) == nil {
		panic(fmt.Errorf("permission denied"))
	}

	const sql = `
		SELECT id
			, family_id
			, name
			, points_cost
			, stock
			, status
			, created_by
		FROM rewards
		WHERE family_id=@p1 AND status=@p2
		ORDER BY id DESC
	`
	return resx.Db.Main.MustListOf(model.Reward{}, sql, familyId, model.RewardStatusActive).([]model.Reward)
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
git add parent-child-api/srv/reward_service.go parent-child-api/srv/reward_service_test.go
git commit -m "feat: add reward create list service"
```

---

### Task 4: 兑换申请事务

**Files:**
- Modify: `D:\projects\github\for_test\parent-child-api\srv\reward_service.go`
- Modify: `D:\projects\github\for_test\parent-child-api\srv\reward_service_test.go`

- [ ] **Step 1: Write the failing test**

在 `srv/reward_service_test.go` 增加：

```go
func TestRewardServiceApplyRejectsMissingFamily(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("ApplyReward should panic")
		}
		if fmt.Sprint(v) != "family id is required" {
			t.Fatalf("panic = %v, want family id is required", v)
		}
	}()

	RewardService.ApplyReward(1, model.RewardApplyRequest{RewardId: 1})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: FAIL with `RewardService.ApplyReward undefined`.

- [ ] **Step 3: Implement apply**

在 `srv/reward_service.go` 增加：

```go
func (x rewardService) ApplyReward(operatorUserId int64, req model.RewardApplyRequest) model.RewardRecord {
	if req.FamilyId == 0 {
		panic(fmt.Errorf("family id is required"))
	}
	if req.RewardId == 0 {
		panic(fmt.Errorf("reward id is required"))
	}

	target := resolveRewardTargetMember(operatorUserId, req.FamilyId, req.MemberId)

	tran := resx.Db.Main.MustCreateTransactionEx()
	defer tran.MustClose()

	reward := loadActiveRewardForUpdate(tran, req.FamilyId, req.RewardId)
	if reward.Stock == 0 {
		panic(fmt.Errorf("reward stock is not enough"))
	}
	if reward.Stock > 0 {
		affected := tran.MustExecute(`
			UPDATE rewards
			SET stock=stock-1
			WHERE id=@p1 AND family_id=@p2 AND status=@p3 AND stock>0
		`, reward.Id, reward.FamilyId, model.RewardStatusActive)
		if affected != 1 {
			panic(fmt.Errorf("reward stock is not enough"))
		}
	}

	affected := tran.MustExecute(`
		UPDATE family_members
		SET current_points=current_points-@p1
		WHERE id=@p2 AND family_id=@p3 AND role_type=@p4 AND status=@p5 AND current_points>=@p1
	`, reward.PointsCost, target.Id, req.FamilyId, model.FamilyRoleChild, model.FamilyMemberStatusActive)
	if affected != 1 {
		panic(fmt.Errorf("points is not enough"))
	}

	tran.MustExecute(`
		INSERT INTO reward_records(family_id, reward_id, member_id, points_cost, status)
		VALUES(@p1, @p2, @p3, @p4, @p5)
	`, req.FamilyId, reward.Id, target.Id, reward.PointsCost, model.RewardRecordStatusApplied)

	recordIdValue, ok := tran.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || recordIdValue == nil {
		panic(fmt.Errorf("failed to load created reward record id"))
	}
	recordId := int64(*recordIdValue)

	tran.MustExecute(`
		INSERT INTO point_logs(family_id, member_id, points, source_type, source_id)
		VALUES(@p1, @p2, @p3, @p4, @p5)
	`, req.FamilyId, target.Id, -reward.PointsCost, model.PointSourceTypeReward, recordId)

	tran.MustCommit()

	return model.RewardRecord{
		Id:         recordId,
		FamilyId:   req.FamilyId,
		RewardId:   reward.Id,
		MemberId:   target.Id,
		PointsCost: reward.PointsCost,
		Status:     model.RewardRecordStatusApplied,
	}
}
```

新增 helpers：

```go
func resolveRewardTargetMember(operatorUserId, familyId, memberId int64) model.FamilyMember {
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
		panic(fmt.Errorf("only child can apply reward"))
	}
	return target
}

func loadActiveRewardForUpdate(db structGetter, familyId, rewardId int64) model.Reward {
	const sql = `
		SELECT id
			, family_id
			, name
			, points_cost
			, stock
			, status
			, created_by
		FROM rewards
		WHERE id=@p1 AND family_id=@p2 AND status=@p3
		FOR UPDATE
	`

	reward := &model.Reward{}
	ok := db.MustGetStruct(reward, sql, rewardId, familyId, model.RewardStatusActive)
	if !ok {
		panic(fmt.Errorf("reward not found"))
	}
	return *reward
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
git add parent-child-api/srv/reward_service.go parent-child-api/srv/reward_service_test.go
git commit -m "feat: apply reward exchange"
```

---

### Task 5: 发放、驳回、确认领取

**Files:**
- Modify: `D:\projects\github\for_test\parent-child-api\srv\reward_service.go`
- Modify: `D:\projects\github\for_test\parent-child-api\srv\reward_service_test.go`

- [ ] **Step 1: Write failing tests**

在 `srv/reward_service_test.go` 增加：

```go
func TestRewardOperateStatusFromRejectFlag(t *testing.T) {
	if rewardOperateStatus(true) != model.RewardRecordStatusDelivered {
		t.Fatal("deliver flag should map to DELIVERED")
	}
	if rewardOperateStatus(false) != model.RewardRecordStatusRejected {
		t.Fatal("reject flag should map to REJECTED")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: FAIL with `undefined: rewardOperateStatus`.

- [ ] **Step 3: Implement operations**

在 `srv/reward_service.go` 增加：

```go
func rewardOperateStatus(delivered bool) model.RewardRecordStatus {
	if delivered {
		return model.RewardRecordStatusDelivered
	}
	return model.RewardRecordStatusRejected
}

func (x rewardService) DeliverReward(operatorUserId int64, req model.RewardRecordOperateRequest) model.RewardRecord {
	return x.operateRewardRecord(operatorUserId, req.RecordId, true)
}

func (x rewardService) RejectReward(operatorUserId int64, req model.RewardRecordOperateRequest) model.RewardRecord {
	return x.operateRewardRecord(operatorUserId, req.RecordId, false)
}

func (x rewardService) operateRewardRecord(operatorUserId int64, recordId int64, delivered bool) model.RewardRecord {
	if recordId == 0 {
		panic(fmt.Errorf("record id is required"))
	}

	tran := resx.Db.Main.MustCreateTransactionEx()
	defer tran.MustClose()

	record := loadAppliedRewardRecordForUpdate(tran, recordId)
	operator := MemberService.RequireParentRole(operatorUserId, record.FamilyId)
	nextStatus := rewardOperateStatus(delivered)

	affected := tran.MustExecute(`
		UPDATE reward_records
		SET status=@p1, operate_time=NOW(), operate_by=@p2
		WHERE id=@p3 AND status=@p4
	`, nextStatus, operator.Id, record.Id, model.RewardRecordStatusApplied)
	if affected != 1 {
		panic(fmt.Errorf("reward record has been changed"))
	}

	if !delivered {
		tran.MustExecute(`
			UPDATE family_members
			SET current_points=current_points+@p1
			WHERE id=@p2 AND family_id=@p3 AND role_type=@p4 AND status=@p5
		`, record.PointsCost, record.MemberId, record.FamilyId, model.FamilyRoleChild, model.FamilyMemberStatusActive)

		reward := loadRewardForOperate(tran, record.RewardId, record.FamilyId)
		if reward.Stock >= 0 {
			tran.MustExecute(`
				UPDATE rewards
				SET stock=stock+1
				WHERE id=@p1 AND family_id=@p2
			`, record.RewardId, record.FamilyId)
		}

		tran.MustExecute(`
			INSERT INTO point_logs(family_id, member_id, points, source_type, source_id)
			VALUES(@p1, @p2, @p3, @p4, @p5)
		`, record.FamilyId, record.MemberId, record.PointsCost, model.PointSourceTypeReward, record.Id)
	}

	tran.MustCommit()
	record.Status = nextStatus
	record.OperateBy = &operator.Id
	return record
}

func (x rewardService) ReceiveReward(userId int64, req model.RewardRecordOperateRequest) model.RewardRecord {
	if req.RecordId == 0 {
		panic(fmt.Errorf("record id is required"))
	}

	tran := resx.Db.Main.MustCreateTransactionEx()
	defer tran.MustClose()

	record := loadDeliveredRewardRecordForUpdate(tran, req.RecordId)
	member := MemberService.LoadActiveMember(userId, record.FamilyId)
	if member == nil || member.Id != record.MemberId {
		panic(fmt.Errorf("permission denied"))
	}

	affected := tran.MustExecute(`
		UPDATE reward_records
		SET status=@p1
		WHERE id=@p2 AND status=@p3
	`, model.RewardRecordStatusReceived, record.Id, model.RewardRecordStatusDelivered)
	if affected != 1 {
		panic(fmt.Errorf("reward record has been changed"))
	}

	tran.MustCommit()
	record.Status = model.RewardRecordStatusReceived
	return record
}
```

增加加载函数：

```go
func loadAppliedRewardRecordForUpdate(db structGetter, recordId int64) model.RewardRecord {
	return loadRewardRecordForUpdate(db, recordId, model.RewardRecordStatusApplied)
}

func loadDeliveredRewardRecordForUpdate(db structGetter, recordId int64) model.RewardRecord {
	return loadRewardRecordForUpdate(db, recordId, model.RewardRecordStatusDelivered)
}

func loadRewardRecordForUpdate(db structGetter, recordId int64, status model.RewardRecordStatus) model.RewardRecord {
	const sql = `
		SELECT id
			, family_id
			, reward_id
			, member_id
			, points_cost
			, status
			, apply_time
			, operate_time
			, operate_by
		FROM reward_records
		WHERE id=@p1 AND status=@p2
		FOR UPDATE
	`
	record := &model.RewardRecord{}
	ok := db.MustGetStruct(record, sql, recordId, status)
	if !ok {
		panic(fmt.Errorf("reward record not found"))
	}
	return *record
}

func loadRewardForOperate(db structGetter, rewardId int64, familyId int64) model.Reward {
	const sql = `
		SELECT id
			, family_id
			, name
			, points_cost
			, stock
			, status
			, created_by
		FROM rewards
		WHERE id=@p1 AND family_id=@p2
	`
	reward := &model.Reward{}
	ok := db.MustGetStruct(reward, sql, rewardId, familyId)
	if !ok {
		panic(fmt.Errorf("reward not found"))
	}
	return *reward
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
git add parent-child-api/srv/reward_service.go parent-child-api/srv/reward_service_test.go
git commit -m "feat: operate reward exchange records"
```

---

### Task 6: 兑换记录列表

**Files:**
- Modify: `D:\projects\github\for_test\parent-child-api\srv\reward_service.go`

- [ ] **Step 1: Add list records implementation**

在 `srv/reward_service.go` 增加：

```go
func (x rewardService) ListRewardRecords(userId int64, req model.RewardRecordListRequest) []model.RewardRecordListItem {
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
			, r.reward_id
			, w.name AS reward_name
			, r.member_id
			, m.nickname
			, r.points_cost
			, r.status
			, r.apply_time
			, r.operate_time
			, r.operate_by
		FROM reward_records r
		INNER JOIN rewards w ON w.id = r.reward_id
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

	return resx.Db.Main.MustListOf(model.RewardRecordListItem{}, sql, args...).([]model.RewardRecordListItem)
}
```

- [ ] **Step 2: Run compile check**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./srv
```

Expected: PASS.

- [ ] **Step 3: Commit**

```powershell
git add parent-child-api/srv/reward_service.go
git commit -m "feat: list reward exchange records"
```

---

### Task 7: 奖品 API 与路由

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\api\reward\reward_api.go`
- Modify: `D:\projects\github\for_test\parent-child-api\api\routes.go`
- Modify: `D:\projects\github\for_test\parent-child-api\api\routes_test.go`

- [ ] **Step 1: Write route test**

在 `api/routes_test.go` 增加：

```go
func TestRegisterRoutesRewardCreateRejectsInvalidJSON(t *testing.T) {
	const jwtSecret = "test-secret"
	token, err := ux.JwtUtil.SignAuthToken(ux.AuthTokenClaims{UserId: 1}, jwtSecret, time.Hour)
	if err != nil {
		t.Fatalf("SignAuthToken error = %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, jwtSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/reward/create", strings.NewReader("{"))
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
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./api -run TestRegisterRoutesRewardCreateRejectsInvalidJSON -count=1
```

Expected: FAIL because `/api/reward/create` returns 404.

- [ ] **Step 3: Add reward API**

创建 `api/reward/reward_api.go`：

```go
package rewardapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"parent-child-api/api/apix"
	"parent-child-api/model"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type RewardApi struct{}

func (x RewardApi) Create(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.RewardCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.RewardService.CreateReward(token.UserId, req))
}

func (x RewardApi) List(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}
	apix.WriteData(w, srv.RewardService.ListRewards(token.UserId, familyId))
}

func (x RewardApi) Apply(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.RewardApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.RewardService.ApplyReward(token.UserId, req))
}

func (x RewardApi) Deliver(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.RewardRecordOperateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.RewardService.DeliverReward(token.UserId, req))
}

func (x RewardApi) Reject(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.RewardRecordOperateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.RewardService.RejectReward(token.UserId, req))
}

func (x RewardApi) Receive(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.RewardRecordOperateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.RewardService.ReceiveReward(token.UserId, req))
}

func (x RewardApi) Records(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}
	req := model.RewardRecordListRequest{
		FamilyId: familyId,
		Status:   model.RewardRecordStatus(r.URL.Query().Get("status")),
	}
	apix.WriteData(w, srv.RewardService.ListRewardRecords(token.UserId, req))
}
```

- [ ] **Step 4: Register routes**

在 `api/routes.go` 增加 import：

```go
rewardapi "parent-child-api/api/reward"
```

注册：

```go
rewardApi := rewardapi.RewardApi{}
mux.HandleFunc("POST /api/reward/create", apix.WithAuth(jwtSecret, rewardApi.Create))
mux.HandleFunc("GET /api/reward/list", apix.WithAuth(jwtSecret, rewardApi.List))
mux.HandleFunc("POST /api/reward/apply", apix.WithAuth(jwtSecret, rewardApi.Apply))
mux.HandleFunc("POST /api/reward/deliver", apix.WithAuth(jwtSecret, rewardApi.Deliver))
mux.HandleFunc("POST /api/reward/reject", apix.WithAuth(jwtSecret, rewardApi.Reject))
mux.HandleFunc("POST /api/reward/receive", apix.WithAuth(jwtSecret, rewardApi.Receive))
mux.HandleFunc("GET /api/reward/records", apix.WithAuth(jwtSecret, rewardApi.Records))
```

- [ ] **Step 5: Run tests**

Run:

```powershell
$env:GOCACHE='D:\projects\github\for_test\parent-child-api\.gocache'; go test ./...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add parent-child-api/api/reward/reward_api.go parent-child-api/api/routes.go parent-child-api/api/routes_test.go
git commit -m "feat: add reward exchange api"
```

---

### Task 8: 集成测试

**Files:**
- Create: `D:\projects\github\for_test\parent-child-api\srv\reward_integration_test.go`

- [ ] **Step 1: Add integration test**

创建 `srv/reward_integration_test.go`：

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

func TestIntegrationRewardExchangeFlow(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed + 200
	childUserId := seed + 201

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("reward-family-%d", seed),
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

	resx.Db.Main.MustExecute(`
		UPDATE family_members
		SET current_points=@p1, total_earned_points=@p1
		WHERE id=@p2
	`, 10, child.Id)

	reward := RewardService.CreateReward(ownerUserId, model.RewardCreateRequest{
		FamilyId:   familyId,
		Name:       "Ice cream",
		PointsCost: 4,
		Stock:      2,
	})

	record := RewardService.ApplyReward(childUserId, model.RewardApplyRequest{
		FamilyId: familyId,
		RewardId: reward.Id,
	})
	if record.Status != model.RewardRecordStatusApplied {
		t.Fatalf("record status = %s, want APPLIED", record.Status)
	}

	reloaded := MemberService.LoadActiveMember(childUserId, familyId)
	if reloaded == nil || reloaded.CurrentPoints != 6 {
		t.Fatalf("current points = %+v, want 6", reloaded)
	}

	delivered := RewardService.DeliverReward(ownerUserId, model.RewardRecordOperateRequest{RecordId: record.Id})
	if delivered.Status != model.RewardRecordStatusDelivered {
		t.Fatalf("delivered status = %s, want DELIVERED", delivered.Status)
	}

	received := RewardService.ReceiveReward(childUserId, model.RewardRecordOperateRequest{RecordId: record.Id})
	if received.Status != model.RewardRecordStatusReceived {
		t.Fatalf("received status = %s, want RECEIVED", received.Status)
	}

	rejectReward := RewardService.CreateReward(ownerUserId, model.RewardCreateRequest{
		FamilyId:   familyId,
		Name:       "Movie",
		PointsCost: 3,
		Stock:      1,
	})
	rejectRecord := RewardService.ApplyReward(childUserId, model.RewardApplyRequest{
		FamilyId: familyId,
		RewardId: rejectReward.Id,
	})
	rejected := RewardService.RejectReward(ownerUserId, model.RewardRecordOperateRequest{RecordId: rejectRecord.Id})
	if rejected.Status != model.RewardRecordStatusRejected {
		t.Fatalf("rejected status = %s, want REJECTED", rejected.Status)
	}

	reloaded = MemberService.LoadActiveMember(childUserId, familyId)
	if reloaded == nil || reloaded.CurrentPoints != 6 {
		t.Fatalf("current points after reject = %+v, want 6", reloaded)
	}

	negativeCount, ok := resx.Db.Main.MustScalarInt(`
		SELECT COUNT(*)
		FROM point_logs
		WHERE family_id=@p1 AND member_id=@p2 AND points=-4 AND source_type=@p3 AND source_id=@p4
	`, familyId, child.Id, model.PointSourceTypeReward, record.Id)
	if !ok || negativeCount == nil || *negativeCount != 1 {
		t.Fatalf("negative point log count = %v, want 1", negativeCount)
	}

	refundCount, ok := resx.Db.Main.MustScalarInt(`
		SELECT COUNT(*)
		FROM point_logs
		WHERE family_id=@p1 AND member_id=@p2 AND points=3 AND source_type=@p3 AND source_id=@p4
	`, familyId, child.Id, model.PointSourceTypeReward, rejectRecord.Id)
	if !ok || refundCount == nil || *refundCount != 1 {
		t.Fatalf("refund point log count = %v, want 1", refundCount)
	}
}
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
go test ./srv -run TestIntegrationRewardExchangeFlow -count=1 -v
```

Expected: PASS.

- [ ] **Step 4: Commit**

```powershell
git add parent-child-api/srv/reward_integration_test.go
git commit -m "test: cover reward exchange flow"
```

---

## Self-Review

Spec coverage:

* 家长创建奖品：Task 3 + Task 7。
* 奖品列表：Task 3 + Task 7。
* 孩子申请兑换：Task 4 + Task 7。
* 家长代虚拟孩子兑换：Task 4 复用 `memberId` + `memberCanSubmitForChild`。
* 积分扣减、库存扣减、兑换记录、积分流水事务：Task 4。
* 家长发放：Task 5 + Task 7。
* 家长驳回、积分回退、库存回退、积分流水：Task 5。
* 孩子确认领取：Task 5 + Task 7。
* 兑换记录列表：Task 6 + Task 7。
* 真实数据库链路：Task 8。

Placeholder scan:

* 没有占位词。
* 没有依赖未定义的类型或方法；新类型和方法均在前置任务定义。

Type consistency:

* `RewardRecordStatus` 与 SQL `reward_records.status` 的 `APPLIED/DELIVERED/RECEIVED/REJECTED` 一致。
* `RewardStatusActive` 与 SQL `rewards.status=1` 一致。
* `PointSourceTypeReward` 已在任务积分计划中定义，复用于兑换扣减和驳回退款。
* `created_by` 在 `rewards` 中按家庭成员 ID 使用，和 `tasks.created_by` 保持一致。
