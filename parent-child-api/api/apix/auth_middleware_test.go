package apix

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"parent-child-api/ux"
)

func TestWithAuthLoadsBearerToken(t *testing.T) {
	token, err := ux.JwtUtil.SignAuthToken(ux.AuthTokenClaims{
		UserId:   42,
		OpenId:   "openid-42",
		Nickname: "爸爸",
	}, "test-secret", time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	var gotUserId int64
	handler := WithAuth("test-secret", func(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
		gotUserId = token.UserId
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/demo/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNoContent)
	}
	if gotUserId != 42 {
		t.Fatalf("got user id = %d, want 42", gotUserId)
	}
}

func TestWithAuthRejectsMissingToken(t *testing.T) {
	handler := WithAuth("test-secret", func(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
		t.Fatal("handler should not be called without token")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/demo/me", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
}
