package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"parent-child-api/ux"
)

func TestRegisterRoutesRegistersFamilyListBehindAuth(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/family/list", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
}

func TestRegisterRoutesFamilyCreateRejectsInvalidJSON(t *testing.T) {
	const jwtSecret = "test-secret"
	token, err := ux.JwtUtil.SignAuthToken(ux.AuthTokenClaims{UserId: 1}, jwtSecret, time.Hour)
	if err != nil {
		t.Fatalf("SignAuthToken error = %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, jwtSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/family/create", strings.NewReader("{"))
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}

	want := `{"code":400,"message":"invalid json body"}`
	if strings.TrimSpace(res.Body.String()) != want {
		t.Fatalf("body = %s, want %s", res.Body.String(), want)
	}
}

func TestRegisterRoutesMemberCreateVirtualChildRejectsInvalidJSON(t *testing.T) {
	const jwtSecret = "test-secret"
	token, err := ux.JwtUtil.SignAuthToken(ux.AuthTokenClaims{UserId: 1}, jwtSecret, time.Hour)
	if err != nil {
		t.Fatalf("SignAuthToken error = %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, jwtSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/member/createVirtualChild", strings.NewReader("{"))
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}

	want := `{"code":400,"message":"invalid json body"}`
	if strings.TrimSpace(res.Body.String()) != want {
		t.Fatalf("body = %s, want %s", res.Body.String(), want)
	}
}

func TestRegisterRoutesInviteCreateRejectsInvalidJSON(t *testing.T) {
	const jwtSecret = "test-secret"
	token, err := ux.JwtUtil.SignAuthToken(ux.AuthTokenClaims{UserId: 1}, jwtSecret, time.Hour)
	if err != nil {
		t.Fatalf("SignAuthToken error = %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, jwtSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/invite/create", strings.NewReader("{"))
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}

	want := `{"code":400,"message":"invalid json body"}`
	if strings.TrimSpace(res.Body.String()) != want {
		t.Fatalf("body = %s, want %s", res.Body.String(), want)
	}
}

func TestRegisterRoutesInviteAcceptRejectsInvalidJSON(t *testing.T) {
	const jwtSecret = "test-secret"
	token, err := ux.JwtUtil.SignAuthToken(ux.AuthTokenClaims{UserId: 1}, jwtSecret, time.Hour)
	if err != nil {
		t.Fatalf("SignAuthToken error = %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, jwtSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/invite/accept", strings.NewReader("{"))
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}

	want := `{"code":400,"message":"invalid json body"}`
	if strings.TrimSpace(res.Body.String()) != want {
		t.Fatalf("body = %s, want %s", res.Body.String(), want)
	}
}

func TestRegisterRoutesTaskCreateRejectsInvalidJSON(t *testing.T) {
	const jwtSecret = "test-secret"
	token, err := ux.JwtUtil.SignAuthToken(ux.AuthTokenClaims{UserId: 1}, jwtSecret, time.Hour)
	if err != nil {
		t.Fatalf("SignAuthToken error = %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, jwtSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/task/create", strings.NewReader("{"))
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}

	want := `{"code":400,"message":"invalid json body"}`
	if strings.TrimSpace(res.Body.String()) != want {
		t.Fatalf("body = %s, want %s", res.Body.String(), want)
	}
}
