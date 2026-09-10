package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDenyIntakeWriteForPlainUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		role string
		want bool
	}{
		{name: "plain user denied", role: "user", want: true},
		{name: "empty role denied", role: "", want: true},
		{name: "lawyer allowed", role: "lawyer", want: false},
		{name: "assistant allowed", role: "assistant", want: false},
		{name: "intake assistant allowed", role: "intake_assistant", want: false},
		{name: "director allowed", role: "director", want: false},
		{name: "conflict officer denied", role: "conflict_officer", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/case-intakes", nil)
			c.Set("role", tt.role)
			if got := denyIntakeWriteForPlainUser(c); got != tt.want {
				t.Fatalf("role %q: denied = %v, want %v", tt.role, got, tt.want)
			}
			if tt.want && w.Code != http.StatusForbidden {
				t.Fatalf("role %q: status = %d, want 403", tt.role, w.Code)
			}
		})
	}
}
