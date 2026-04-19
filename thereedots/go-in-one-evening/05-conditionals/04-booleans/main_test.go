// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package conditionals

import "testing"

func TestIn20thCentury(t *testing.T) {
	for i := 1901; i <= 2000; i++ {
		ok := In20thCentury(i)
		if !ok {
			t.Errorf("Expected %v to be in 20th century", i)
		}
	}

	ok := In20thCentury(1900)
	if ok {
		t.Errorf("Didn't expect 1900 to be in 20th century")
	}

	ok = In20thCentury(2001)
	if ok {
		t.Errorf("Didn't expect 2001 to be in 20th century")
	}
}
