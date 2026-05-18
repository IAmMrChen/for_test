package apix

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithRecoverWritesJsonErrorWhenHandlerPanics(t *testing.T) {
	handler := WithRecover(func(w http.ResponseWriter, r *http.Request) {
		panic(fmt.Errorf("permission denied"))
	})

	res := httptest.NewRecorder()
	handler(res, httptest.NewRequest(http.MethodGet, "/demo", nil))

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
	if !strings.Contains(res.Body.String(), `"message":"permission denied"`) {
		t.Fatalf("body = %s, want permission denied message", res.Body.String())
	}
}

func TestWithRecoverKeepsNormalResponse(t *testing.T) {
	handler := WithRecover(func(w http.ResponseWriter, r *http.Request) {
		WriteData(w, map[string]string{"hello": "world"})
	})

	res := httptest.NewRecorder()
	handler(res, httptest.NewRequest(http.MethodGet, "/demo", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if !strings.Contains(res.Body.String(), `"hello":"world"`) {
		t.Fatalf("body = %s, want normal data", res.Body.String())
	}
}
