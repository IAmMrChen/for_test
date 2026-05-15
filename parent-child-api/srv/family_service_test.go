package srv

import (
	"fmt"
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
