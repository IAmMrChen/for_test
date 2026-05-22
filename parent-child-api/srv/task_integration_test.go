package srv

import (
	"fmt"
	"os"
	"testing"
	"time"

	"parent-child-api/model"
	"parent-child-api/resx"
)

func TestIntegrationTaskPointsFlow(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed + 100
	childUserId := seed + 101

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("task-points-family-%d", seed),
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
		Title:     "Brush teeth",
		Points:    3,
		CycleType: model.TaskCycleTypeDaily,
	})

	claimed := TaskService.ClaimTask(childUserId, model.TaskClaimRequest{
		FamilyId: familyId,
		TaskId:   task.Id,
	})
	if claimed.Status != model.TaskRecordStatusClaimed {
		t.Fatalf("claimed status = %s, want CLAIMED", claimed.Status)
	}

	mustPanicWith(t, "task record already exists", func() {
		TaskService.ClaimTask(childUserId, model.TaskClaimRequest{
			FamilyId: familyId,
			TaskId:   task.Id,
		})
	})

	record := TaskService.SubmitTask(childUserId, model.TaskSubmitRequest{
		FamilyId: familyId,
		RecordId: claimed.Id,
	})
	if record.Status != model.TaskRecordStatusPending {
		t.Fatalf("record status = %s, want PENDING", record.Status)
	}

	audited := TaskService.AuditTask(ownerUserId, model.TaskAuditRequest{
		RecordId: record.Id,
		Approved: true,
	})
	if audited.Status != model.TaskRecordStatusApproved {
		t.Fatalf("audited status = %s, want APPROVED", audited.Status)
	}

	reloaded := MemberService.LoadActiveMember(childUserId, familyId)
	if reloaded == nil {
		t.Fatal("child member not found")
	}
	if reloaded.CurrentPoints != 3 || reloaded.TotalEarnedPoints != 3 {
		t.Fatalf("points = current %d total %d, want 3/3", reloaded.CurrentPoints, reloaded.TotalEarnedPoints)
	}

	count, ok := resx.Db.Main.MustScalarInt(`
		SELECT COUNT(*)
		FROM point_logs
		WHERE family_id=@p1 AND member_id=@p2 AND points=@p3 AND source_type=@p4 AND source_id=@p5
	`, familyId, child.Id, 3, model.PointSourceTypeTask, record.Id)
	if !ok || count == nil || *count != 1 {
		t.Fatalf("point log count = %v, want 1", count)
	}
}

func TestIntegrationTaskUpdateAndArchiveFlow(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed + 300
	childUserId := seed + 301

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("task-management-family-%d", seed),
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
	InviteService.AcceptInvite(childUserId, invite.Token)

	task := TaskService.CreateTask(ownerUserId, model.TaskCreateRequest{
		FamilyId:  familyId,
		Title:     "Old task",
		Points:    5,
		CycleType: model.TaskCycleTypeOnce,
	})

	mustPanicWith(t, "permission denied", func() {
		TaskService.UpdateTask(childUserId, model.TaskUpdateRequest{
			FamilyId:  familyId,
			TaskId:    task.Id,
			Title:     "Child update",
			Points:    6,
			CycleType: model.TaskCycleTypeDaily,
		})
	})

	updated := TaskService.UpdateTask(ownerUserId, model.TaskUpdateRequest{
		FamilyId:  familyId,
		TaskId:    task.Id,
		Title:     "New task",
		Points:    9,
		CycleType: model.TaskCycleTypeWeekly,
	})
	if updated.Title != "New task" || updated.Points != 9 || updated.CycleType != model.TaskCycleTypeWeekly {
		t.Fatalf("updated task = %+v", updated)
	}

	mustPanicWith(t, "permission denied", func() {
		TaskService.ArchiveTask(childUserId, model.TaskArchiveRequest{
			FamilyId: familyId,
			TaskId:   task.Id,
		})
	})

	archived := TaskService.ArchiveTask(ownerUserId, model.TaskArchiveRequest{
		FamilyId: familyId,
		TaskId:   task.Id,
	})
	if archived.Status != model.TaskStatusArchived {
		t.Fatalf("archived status = %d, want archived", archived.Status)
	}

	tasks := TaskService.ListTasks(ownerUserId, familyId)
	for _, item := range tasks {
		if item.Id == task.Id {
			t.Fatalf("archived task should not be listed: %+v", item)
		}
	}

	mustPanicWith(t, "task not found", func() {
		TaskService.ClaimTask(childUserId, model.TaskClaimRequest{
			FamilyId: familyId,
			TaskId:   task.Id,
		})
	})
}
