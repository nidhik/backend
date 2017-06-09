package models

import (
	"testing"
)

func TestIncrement(t *testing.T) {
	task0 := NewTask("abcd")
	task0.Increment("Claimed", 2)

	if task0.Claimed != 0 {
		t.Fatal("Expected task claimed would equal 0 before saving task. Actual was:", task0.Claimed)
	}

}

func TestUnset(t *testing.T) {
	task0 := NewTask("abcd")
	task0.Set("Status", "ERROR")

	if task0.Status != "ERROR" {
		t.Fatal("Set did not work.")
	}

	task0.Unset("Status")

	if task0.Status != "" {
		t.Fatal("Unset did not work.")
	}
}
