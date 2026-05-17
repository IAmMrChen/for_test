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

func TestTaskCycleTypeValid(t *testing.T) {
	if !TaskCycleTypeOnce.Valid() || !TaskCycleTypeDaily.Valid() || !TaskCycleTypeWeekly.Valid() {
		t.Fatal("known cycle types should be valid")
	}
	if TaskCycleType("MONTHLY").Valid() {
		t.Fatal("unknown cycle type should be invalid")
	}
}
