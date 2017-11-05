package helpers

import "testing"

func TestIsVoxPopuli(t *testing.T) {
	if !IsVoxPopuli("cyberanalytics") {
		t.FailNow()
	}
	if IsVoxPopuli("chiliec") {
		t.FailNow()
	}
}
