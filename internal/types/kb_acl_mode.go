package types

import (
	"os"
	"strings"
)

type KBACLMode string

const (
	KBACLModeOff     KBACLMode = "off"
	KBACLModeShadow  KBACLMode = "shadow"
	KBACLModeEnforce KBACLMode = "enforce"
)

// CurrentKBACLMode is the rollout switch for VONE-0.7.2.1. Unknown values
// fail safe to enforce; operators must explicitly choose off or shadow.
func CurrentKBACLMode() KBACLMode {
	switch KBACLMode(strings.ToLower(strings.TrimSpace(os.Getenv("WEKNORA_VONE_KB_ACL_MODE")))) {
	case KBACLModeOff:
		return KBACLModeOff
	case KBACLModeShadow:
		return KBACLModeShadow
	default:
		return KBACLModeEnforce
	}
}
