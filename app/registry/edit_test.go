package registry

import "testing"

func TestBranchVersion(t *testing.T) {
	v := branchVersion("1.5.3-SNAPSHOT-157")
	if v != 153157 {
		t.Fatalf("wrong version: %d", v)
	}
}
