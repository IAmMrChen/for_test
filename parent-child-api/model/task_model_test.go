package model

import "testing"

func TestTaskRecordStatusCanAudit(t *testing.T) {
	if !TaskRecordStatusPending.CanAudit() {
		t.Fatal("pending record should be auditable")
	}
	if TaskRecordStatusApproved.CanAudit() {
		t.Fatal("approved record should not be auditable")
	}
	if TaskRecordStatusRejected.CanAudit() {
		t.Fatal("rejected record should not be auditable")
	}
}

func TestTaskRecordStatusCanSubmit(t *testing.T) {
	if !TaskRecordStatusClaimed.CanSubmit() {
		t.Fatal("claimed record should be submittable")
	}
	if TaskRecordStatusPending.CanSubmit() {
		t.Fatal("pending record should not be submittable")
	}
	if TaskRecordStatusApproved.CanSubmit() {
		t.Fatal("approved record should not be submittable")
	}
}

func TestTaskCycleTypeValid(t *testing.T) {
	if !TaskCycleTypeOnce.Valid() || !TaskCycleTypeDaily.Valid() || !TaskCycleTypeWeekly.Valid() {
		t.Fatal("known cycle types should be valid")
	}
	if TaskCycleType("MONTHLY").Valid() {
		t.Fatal("unknown cycle type should be invalid")
	}
}
