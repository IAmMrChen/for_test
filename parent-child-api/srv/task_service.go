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
	submitRemark := strings.TrimSpace(req.SubmitRemark)

	const sql = `
		INSERT INTO task_records(family_id, task_id, member_id, status, submit_remark)
		VALUES(@p1, @p2, @p3, @p4, @p5)
	`
	resx.Db.Main.MustExecute(sql, req.FamilyId, task.Id, target.Id, model.TaskRecordStatusPending, submitRemark)

	recordIdValue, ok := resx.Db.Main.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || recordIdValue == nil {
		panic(fmt.Errorf("failed to load created task record id"))
	}

	return model.TaskRecord{
		Id:           int64(*recordIdValue),
		FamilyId:     req.FamilyId,
		TaskId:       task.Id,
		MemberId:     target.Id,
		Status:       model.TaskRecordStatusPending,
		SubmitRemark: submitRemark,
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
	auditRemark := strings.TrimSpace(req.AuditRemark)
	affected := tran.MustExecute(`
		UPDATE task_records
		SET status=@p1, audit_time=NOW(), audit_by=@p2, audit_remark=@p3
		WHERE id=@p4 AND status=@p5
	`, nextStatus, operator.Id, auditRemark, req.RecordId, model.TaskRecordStatusPending)
	if affected != 1 {
		panic(fmt.Errorf("task record has been changed"))
	}

	if req.Approved {
		affected = tran.MustExecute(`
			UPDATE family_members
			SET current_points=current_points+@p1
				, total_earned_points=total_earned_points+@p1
			WHERE id=@p2 AND family_id=@p3 AND role_type=@p4 AND status=@p5
		`, task.Points, record.MemberId, record.FamilyId, model.FamilyRoleChild, model.FamilyMemberStatusActive)
		if affected != 1 {
			panic(fmt.Errorf("target child not found"))
		}

		tran.MustExecute(`
			INSERT INTO point_logs(family_id, member_id, points, source_type, source_id)
			VALUES(@p1, @p2, @p3, @p4, @p5)
		`, record.FamilyId, record.MemberId, task.Points, model.PointSourceTypeTask, record.Id)
	}

	tran.MustCommit()

	record.Status = nextStatus
	record.AuditBy = &operator.Id
	record.AuditRemark = auditRemark
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
