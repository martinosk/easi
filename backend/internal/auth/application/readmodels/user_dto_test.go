package readmodels

import "testing"

func TestUserDTOIsActiveAdmin(t *testing.T) {
	cases := []struct {
		name   string
		role   string
		status string
		want   bool
	}{
		{"active admin", "admin", "active", true},
		{"disabled admin", "admin", "disabled", false},
		{"active member", "member", "active", false},
		{"disabled member", "member", "disabled", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := UserDTO{Role: tc.role, Status: tc.status}
			if got := u.IsActiveAdmin(); got != tc.want {
				t.Fatalf("IsActiveAdmin() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestUserDTOActiveRole(t *testing.T) {
	var missing *UserDTO
	cases := []struct {
		name string
		user *UserDTO
		want string
	}{
		{"active user keeps the role", &UserDTO{Role: "architect", Status: "active"}, "architect"},
		{"disabled user has no role", &UserDTO{Role: "architect", Status: "disabled"}, ""},
		{"missing user has no role", missing, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.user.ActiveRole(); got != tc.want {
				t.Fatalf("ActiveRole() = %q, want %q", got, tc.want)
			}
		})
	}
}
