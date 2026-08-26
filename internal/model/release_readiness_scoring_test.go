package model

import (
	"testing"

	"sterile-packaging-release-control/internal/constants"
)

func TestReadyForReleaseRespectsState(t *testing.T) {
	released := validBatch()
	released.Status = constants.BatchStatusReleased
	if ready, _ := released.ReadyForRelease(); ready {
		t.Fatal("released batch must not report ready")
	}
	if released.Mutable() {
		t.Fatal("released batch must not be mutable")
	}
}
