package game

import (
	"reflect"
	"testing"
)

func TestApplyShootMoon(t *testing.T) {
	round := []int{26, 0, 0, 0}
	got := ApplyShootMoon(round)
	want := []int{0, 26, 26, 26}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("shoot the moon scoring mismatch: got=%v want=%v", got, want)
	}
}

func TestApplyShootMoonNoop(t *testing.T) {
	round := []int{10, 5, 8, 3}
	got := ApplyShootMoon(round)

	if !reflect.DeepEqual(got, round) {
		t.Fatalf("expected unchanged scores, got=%v want=%v", got, round)
	}
}
