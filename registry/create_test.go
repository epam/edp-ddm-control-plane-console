package registry

import "testing"

func TestPrepareRegistryCodebase(t *testing.T) {
	cb := prepareRegistryCodebase("host", &registry{
		RegistryGitBranch: "1.2.4-snap.130",
	})

	if cb.Spec.JobProvisioning == nil || *cb.Spec.JobProvisioning != "default-1-2-4-snap-130" {
		t.Fatalf("wrong job provisioning: %v", cb.Spec.JobProvisioning)
	}
}
