package requirements

import (
	"errors"
	"testing"
)

func TestIsNotMet(t *testing.T) {
	t.Parallel()
	e := &ErrNotMet{Kind: "building", Key: "shipyard", Need: 5, Have: 2}
	wrapped := errors.Join(errors.New("outer"), e)
	if !IsNotMet(wrapped) {
		t.Fatalf("expected IsNotMet to find wrapped ErrNotMet")
	}
}

func TestErrNotMetMessage(t *testing.T) {
	t.Parallel()
	e := &ErrNotMet{Kind: "research", Key: "laser_tech", Need: 3, Have: 0}
	if got := e.Error(); got != "requirement not met: research laser_tech level 3 (have 0)" {
		t.Fatalf("unexpected: %q", got)
	}
}
