package constants

import "testing"

func TestBatchIllegalTransitionsRejected(t *testing.T) {
	illegal := []struct{ from, to BatchStatus }{
		{BatchStatusRunning, BatchStatusDraft},
		{BatchStatusHold, BatchStatusDraft},
		{BatchStatusRework, BatchStatusDraft},
		{BatchStatusDraft, BatchStatusReleased},
	}
	for _, tt := range illegal {
		if tt.from.CanTransitionTo(tt.to) {
			t.Fatalf("illegal transition %s -> %s should be rejected", tt.from, tt.to)
		}
		if err := ValidateBatchTransition(tt.from, tt.to); err == nil {
			t.Fatalf("ValidateBatchTransition(%s, %s) should return error", tt.from, tt.to)
		}
	}
}
