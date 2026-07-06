package idempotency

import "testing"

func TestHash_Deterministic(t *testing.T) {
	first := Hash("auction-1", "guild-1", "100")
	second := Hash("auction-1", "guild-1", "100")

	if first != second {
		t.Fatalf("Hash() is not deterministic: %q != %q", first, second)
	}
}

func TestHash_DifferentInput(t *testing.T) {
	first := Hash("auction-1", "guild-1", "100")
	second := Hash("auction-1", "guild-1", "101")

	if first == second {
		t.Fatal("different requests produced the same hash")
	}
}

func TestHash_PreservesPartBoundaries(t *testing.T) {
	first := Hash("ab", "c")
	second := Hash("a", "bc")

	if first == second {
		t.Fatal("different part boundaries produced the same hash")
	}
}
