package model

import "testing"

func TestOpenBatchCountAccurate(t *testing.T) {
	l := PackagingLine{Batches: []ProductionBatch{
		{Status: "running"}, {Status: "released"}, {Status: "draft"},
	}}
	if got := l.OpenBatchCount(); got != 2 {
		t.Fatalf("OpenBatchCount = %d, want 2", got)
	}
}
