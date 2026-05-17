package srv

import (
	"fmt"
	"testing"

	"parent-child-api/model"
	"parent-child-api/resx"
)

func TestTaskServiceCreateTaskPanicsWhenTitleIsEmpty(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("CreateTask should panic")
		}
		if fmt.Sprint(v) != "task title is required" {
			t.Fatalf("panic = %v, want task title is required", v)
		}
	}()

	TaskService.CreateTask(1, model.TaskCreateRequest{FamilyId: 1, Title: " "})
}

func TestNormalizeTaskCreateRequest(t *testing.T) {
	req := normalizeTaskCreateRequest(model.TaskCreateRequest{
		FamilyId:  1,
		Title:     " Brush teeth ",
		Points:    0,
		CycleType: "",
	})

	if req.Title != "Brush teeth" {
		t.Fatalf("title = %q, want Brush teeth", req.Title)
	}
	if req.Points != 1 {
		t.Fatalf("points = %d, want 1", req.Points)
	}
	if req.CycleType != model.TaskCycleTypeOnce {
		t.Fatalf("cycleType = %s, want ONCE", req.CycleType)
	}
}

func TestResolveSubmitMemberRejectsMissingFamily(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("SubmitTask should panic")
		}
		if fmt.Sprint(v) != "family id is required" {
			t.Fatalf("panic = %v, want family id is required", v)
		}
	}()

	TaskService.SubmitTask(1, model.TaskSubmitRequest{TaskId: 1})
}

func TestTaskAuditStatusFromApprovedFlag(t *testing.T) {
	if taskAuditStatus(true) != model.TaskRecordStatusApproved {
		t.Fatal("approved flag should map to APPROVED")
	}
	if taskAuditStatus(false) != model.TaskRecordStatusRejected {
		t.Fatal("false approved flag should map to REJECTED")
	}
}

func TestTaskRecordListRejectsMissingFamily(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("ListTaskRecords should panic")
		}
		if fmt.Sprint(v) != "family id is required" {
			t.Fatalf("panic = %v, want family id is required", v)
		}
	}()

	TaskService.ListTaskRecords(1, model.TaskRecordListRequest{})
}
