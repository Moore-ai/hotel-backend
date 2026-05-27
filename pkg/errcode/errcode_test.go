package errcode

import "testing"

func TestConflictMessageIsActionable(t *testing.T) {
	got := Message(ErrConflict)
	if got == "conflict" || got == "" {
		t.Fatalf("ErrConflict message = %q, want actionable business message", got)
	}
}
