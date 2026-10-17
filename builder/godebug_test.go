package builder

import "testing"

func TestStartupGODEBUG(t *testing.T) {
	for _, tc := range []struct {
		env, want string
	}{
		{"", ""},
		{"OTHER=x", ""},
		{"GODEBUG=a=1,b=2", "a=1,b=2"},
		{"OTHER=x\x00GODEBUG=a=1\x00LAST=x", "a=1"},
		{"GODEBUG=\x00GODEBUG=a=1", ""},
		{"GODEBUG=a=1\x00GODEBUG=a=2", "a=1"},
		{"NOTGODEBUG=a=1\x00GODEBUG=a=2", "a=2"},
	} {
		if got := startupGODEBUG(tc.env); got != tc.want {
			t.Errorf("startupGODEBUG(%q) = %q, want %q", tc.env, got, tc.want)
		}
	}
}
