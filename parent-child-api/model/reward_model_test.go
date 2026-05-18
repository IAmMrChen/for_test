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
