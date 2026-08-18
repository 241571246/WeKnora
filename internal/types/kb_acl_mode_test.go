package types

import "testing"

func TestCurrentKBACLMode(t *testing.T) {
	tests := []struct {
		value string
		want  KBACLMode
	}{
		{"", KBACLModeEnforce},
		{"unknown", KBACLModeEnforce},
		{"OFF", KBACLModeOff},
		{" shadow ", KBACLModeShadow},
		{"enforce", KBACLModeEnforce},
	}
	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			t.Setenv("WEKNORA_VONE_KB_ACL_MODE", test.value)
			if got := CurrentKBACLMode(); got != test.want {
				t.Fatalf("CurrentKBACLMode() = %q, want %q", got, test.want)
			}
		})
	}
}
