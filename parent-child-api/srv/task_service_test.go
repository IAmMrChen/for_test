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

func TestTaskRecordCycleCondition(t *testing.T) {
	cases := []struct {
		name      string
		cycleType model.TaskCycleType
		want      string
	}{
		{name: "once", cycleType: model.TaskCycleTypeOnce, want: ""},
		{name: "daily", cycleType: model.TaskCycleTypeDaily, want: "DATE(created_at)=CURRENT_DATE()"},
		{name: "weekly", cycleType: model.TaskCycleTypeWeekly, want: "YEARWEEK(created_at, 1)=YEARWEEK(CURRENT_DATE(), 1)"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := taskRecordCycleCondition(tt.cycleType); got != tt.want {
				t.Fatalf("condition = %q, want %q", got, tt.want)
			}
		})
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

func TestTaskServiceSubmitTaskRejectsMissingRecordAndTask(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("SubmitTask should panic")
		}
		if fmt.Sprint(v) != "task id or record id is required" {
			t.Fatalf("panic = %v, want task id or record id is required", v)
		}
	}()

	TaskService.SubmitTask(1, model.TaskSubmitRequest{FamilyId: 1})
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

func TestTaskServiceClaimTaskRejectsMissingFamily(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("ClaimTask should panic")
		}
		if fmt.Sprint(v) != "family id is required" {
			t.Fatalf("panic = %v, want family id is required", v)
		}
	}()

	TaskService.ClaimTask(1, model.TaskClaimRequest{TaskId: 1})
}

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
