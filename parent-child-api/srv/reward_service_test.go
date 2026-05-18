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
