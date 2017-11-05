package helpers

import "testing"

func TestContains(t *testing.T) {
	if !Contains([]int{1, 2, 3}, 3) {
		t.FailNow()
	}
	if !Contains([]string{"123", "321"}, "321") {
		t.FailNow()
	}
	if Contains([]string{"123", "321"}, "456") {
		t.FailNow()
	}
}
