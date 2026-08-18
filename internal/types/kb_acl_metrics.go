package types

import "sync/atomic"

// KBACLMetricSnapshot is an operator-facing, process-lifetime summary. It is
// intentionally low-cardinality: tenant/user/KB/capability dimensions remain
// in structured logs and audit rows, while these counters are safe to expose
// from /system/info without leaking resource identifiers.
type KBACLMetricSnapshot struct {
	Allowed      uint64 `json:"allowed"`
	Denied       uint64 `json:"denied"`
	ShadowDenied uint64 `json:"shadow_denied"`
	Errors       uint64 `json:"errors"`
}

var kbACLMetricCounters struct {
	allowed      atomic.Uint64
	denied       atomic.Uint64
	shadowDenied atomic.Uint64
	errors       atomic.Uint64
}

func RecordKBACLAllowed()      { kbACLMetricCounters.allowed.Add(1) }
func RecordKBACLDenied()       { kbACLMetricCounters.denied.Add(1) }
func RecordKBACLShadowDenied() { kbACLMetricCounters.shadowDenied.Add(1) }
func RecordKBACLError()        { kbACLMetricCounters.errors.Add(1) }

func CurrentKBACLMetricSnapshot() KBACLMetricSnapshot {
	return KBACLMetricSnapshot{
		Allowed:      kbACLMetricCounters.allowed.Load(),
		Denied:       kbACLMetricCounters.denied.Load(),
		ShadowDenied: kbACLMetricCounters.shadowDenied.Load(),
		Errors:       kbACLMetricCounters.errors.Load(),
	}
}
