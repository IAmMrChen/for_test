package srv

import (
	"fmt"
	"strings"

	"parent-child-api/model"
	"parent-child-api/resx"
)

var FamilyService familyService

type familyService struct{}

const defaultFamilyNickname = "\u5bb6\u957f"

type defaultTaskPreset struct {
	Title     string
	Points    int
	CycleType model.TaskCycleType
}

type defaultRewardPreset struct {
	Name       string
	PointsCost int
	Stock      int
}

type familyPresetExecutor interface {
	MustExecute(query string, args ...any) int64
}

func defaultTaskPresets() []defaultTaskPreset {
	return []defaultTaskPreset{
		{Title: "阅读 30 分钟", Points: 10, CycleType: model.TaskCycleTypeDaily},
		{Title: "整理自己的物品", Points: 8, CycleType: model.TaskCycleTypeDaily},
		{Title: "主动完成一件家务", Points: 12, CycleType: model.TaskCycleTypeWeekly},
	}
}

func defaultRewardPresets() []defaultRewardPreset {
	return []defaultRewardPreset{
		{Name: "选择一次周末活动", PointsCost: 50, Stock: -1},
		{Name: "兑换 30 分钟娱乐时间", PointsCost: 30, Stock: -1},
		{Name: "一份小礼物", PointsCost: 80, Stock: -1},
	}
}

func createDefaultFamilyPresets(db familyPresetExecutor, familyId, creatorMemberId int64) {
	const insertTaskSql = `
		INSERT INTO tasks(family_id, title, points, cycle_type, status, created_by)
		VALUES(@p1, @p2, @p3, @p4, @p5, @p6)
	`
	for _, preset := range defaultTaskPresets() {
		db.MustExecute(
			insertTaskSql,
			familyId,
			preset.Title,
			preset.Points,
			preset.CycleType,
			model.TaskStatusActive,
			creatorMemberId,
		)
	}

	const insertRewardSql = `
		INSERT INTO rewards(family_id, name, points_cost, stock, status, created_by)
		VALUES(@p1, @p2, @p3, @p4, @p5, @p6)
	`
	for _, preset := range defaultRewardPresets() {
		db.MustExecute(
			insertRewardSql,
			familyId,
			preset.Name,
			preset.PointsCost,
			preset.Stock,
			model.RewardStatusActive,
			creatorMemberId,
		)
	}
}

func normalizeFamilyCreateRequest(req model.FamilyCreateRequest) (name string, nickname string) {
	name = strings.TrimSpace(req.Name)
	if name == "" {
		panic(fmt.Errorf("family name is required"))
	}

	nickname = strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = defaultFamilyNickname
	}

	return name, nickname
}

func (x familyService) CreateFamily(userId int64, req model.FamilyCreateRequest) model.FamilyCreateResponse {
	name, nickname := normalizeFamilyCreateRequest(req)

	tran := resx.Db.Main.MustCreateTransactionEx()
	defer tran.MustClose()

	const insertFamilySql = `
		INSERT INTO families(name, creator_id)
		VALUES (@p1, @p2)
	`
	tran.MustExecute(insertFamilySql, name, userId)

	familyIdValue, ok := tran.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || familyIdValue == nil {
		panic(fmt.Errorf("failed to load created family id"))
	}
	familyId := int64(*familyIdValue)

	const insertMemberSql = `
		INSERT INTO family_members(
			family_id
			, user_id
			, role_type
			, nickname
			, current_points
			, total_earned_points
			, is_virtual
			, created_by
			, status
		)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9)
	`
	tran.MustExecute(
		insertMemberSql,
		familyId,
		userId,
		model.FamilyRoleOwner,
		nickname,
		0,
		0,
		0,
		userId,
		model.FamilyMemberStatusActive,
	)

	memberIdValue, ok := tran.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || memberIdValue == nil {
		panic(fmt.Errorf("failed to load created family member id"))
	}
	memberId := int64(*memberIdValue)

	createDefaultFamilyPresets(tran, familyId, memberId)

	tran.MustCommit()

	return model.FamilyCreateResponse{
		Family: model.Family{
			Id:        familyId,
			Name:      name,
			CreatorId: userId,
		},
		Member: model.FamilyMember{
			Id:                memberId,
			FamilyId:          familyId,
			UserId:            &userId,
			RoleType:          model.FamilyRoleOwner,
			Nickname:          nickname,
			CurrentPoints:     0,
			TotalEarnedPoints: 0,
			IsVirtual:         false,
			Status:            model.FamilyMemberStatusActive,
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

	return resx.Db.Main.MustListOf(model.FamilyListItem{}, sql, userId, model.FamilyMemberStatusActive).([]model.FamilyListItem)
}
