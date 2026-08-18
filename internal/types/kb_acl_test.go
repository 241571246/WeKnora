package types

import "testing"

func TestKBRoleCapabilityContract(t *testing.T) {
	all := KBRoleCapabilities(KBMemberRoleOwner)
	if len(all) != 17 {
		t.Fatalf("owner capability count = %d, want 17", len(all))
	}
	assertContains := func(role KBMemberRole, capability KBCapability, want bool) {
		t.Helper()
		got := false
		for _, item := range KBRoleCapabilities(role) {
			got = got || item == capability
		}
		if got != want {
			t.Fatalf("role %q capability %q = %v, want %v", role, capability, got, want)
		}
	}
	assertContains(KBMemberRoleAIUser, KBCapabilityAIQuery, true)
	assertContains(KBMemberRoleAIUser, KBCapabilityMetadataRead, true)
	assertContains(KBMemberRoleAIUser, KBCapabilityDocumentsList, false)
	assertContains(KBMemberRoleAIUser, KBCapabilityDocumentPreview, false)
	assertContains(KBMemberRoleAIUser, KBCapabilityChunkPreview, false)
	assertContains(KBMemberRoleEditor, KBCapabilityDocumentDelete, true)
	assertContains(KBMemberRoleEditor, KBCapabilityDelete, false)
	assertContains(KBMemberRoleDocumentViewer, KBCapabilityDocumentDownload, true)
	assertContains(KBMemberRoleDocumentViewer, KBCapabilityDocumentEdit, false)
}

func TestAllKBCapabilitiesAreValidAndUnique(t *testing.T) {
	seen := map[KBCapability]bool{}
	for _, capability := range AllKBCapabilities() {
		if !capability.IsValid() {
			t.Fatalf("invalid declared capability %q", capability)
		}
		if seen[capability] {
			t.Fatalf("duplicate capability %q", capability)
		}
		seen[capability] = true
	}
}
