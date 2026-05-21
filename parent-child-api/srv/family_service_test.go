package srv

import (
	"fmt"
	"strings"
	"testing"

	"parent-child-api/model"
	"parent-child-api/resx"
)

func TestFamilyServiceCreateFamilyPanicsWhenNameIsEmpty(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("CreateFamily should panic")
		}

		err, ok := v.(error)
		if !ok {
			t.Fatalf("panic = %T(%v), want error", v, v)
		}
		if fmt.Sprint(err) != "family name is required" {
			t.Fatalf("panic error = %v, want family name is required", err)
		}
	}()

	FamilyService.CreateFamily(1, model.FamilyCreateRequest{Name: " \t\n "})
}

func TestNormalizeFamilyCreateRequestUsesDefaultNickname(t *testing.T) {
	name, nickname := normalizeFamilyCreateRequest(model.FamilyCreateRequest{
		Name:     " Home ",
		Nickname: " \t\n ",
	})

	if name != "Home" {
		t.Fatalf("name = %q, want Home", name)
	}
	if nickname != "\u5bb6\u957f" {
		t.Fatalf("nickname = %q, want %q", nickname, "\u5bb6\u957f")
	}
}

func TestDefaultTaskPresetsAreValid(t *testing.T) {
	presets := defaultTaskPresets()
	if len(presets) != 3 {
		t.Fatalf("len(defaultTaskPresets()) = %d, want 3", len(presets))
	}

	for _, preset := range presets {
		if strings.TrimSpace(preset.Title) == "" {
			t.Fatalf("task preset title is empty: %+v", preset)
		}
		if preset.Points <= 0 {
			t.Fatalf("task preset points = %d, want > 0", preset.Points)
		}
		if !preset.CycleType.Valid() {
			t.Fatalf("task preset cycle type = %s, want valid", preset.CycleType)
		}
	}
}

func TestDefaultRewardPresetsAreValid(t *testing.T) {
	presets := defaultRewardPresets()
	if len(presets) != 3 {
		t.Fatalf("len(defaultRewardPresets()) = %d, want 3", len(presets))
	}

	for _, preset := range presets {
		if strings.TrimSpace(preset.Name) == "" {
			t.Fatalf("reward preset name is empty: %+v", preset)
		}
		if preset.PointsCost <= 0 {
			t.Fatalf("reward preset points cost = %d, want > 0", preset.PointsCost)
		}
		if preset.Stock != -1 && preset.Stock <= 0 {
			t.Fatalf("reward preset stock = %d, want -1 or > 0", preset.Stock)
		}
	}
}
