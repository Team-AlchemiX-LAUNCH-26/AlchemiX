package resilience

import (
	"reflect"
	"testing"
)

func TestTowerFailureState(t *testing.T) {
	state := NewState(
		[]string{"Aegis", "Dawn"},
		2,
	)

	state.DisableTower("Aegis", 5)
	state.DisableTower("Aegis", 2)
	state.DisableTower("Dawn", 1)

	if !state.IsTowerDisabled("Aegis", 2) {
		t.Fatal("expected Aegis tower 2 to be disabled")
	}

	got := state.DisabledTowersSnapshot()

	want := map[string][]int{
		"Aegis": {2, 5},
		"Dawn":  {1},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"expected %#v, got %#v",
			want,
			got,
		)
	}

	state.EnableTower("Aegis", 2)

	if state.IsTowerDisabled("Aegis", 2) {
		t.Fatal("expected Aegis tower 2 to be restored")
	}

	state.Reset()

	if got := state.DisabledTowersSnapshot(); len(got) != 0 {
		t.Fatalf(
			"expected reset to restore all towers, got %#v",
			got,
		)
	}
}

func TestDisabledTowerSetSnapshotIsDeepCopy(t *testing.T) {
	state := NewState(
		[]string{"Aegis"},
		1,
	)

	state.DisableTower("Aegis", 3)

	snapshot := state.DisabledTowerSetSnapshot()

	// Change the returned copy.
	delete(snapshot["Aegis"], 3)

	// Internal state must remain unchanged.
	if !state.IsTowerDisabled("Aegis", 3) {
		t.Fatal(
			"editing the snapshot must not change internal state",
		)
	}
}
