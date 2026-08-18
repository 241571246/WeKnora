package types

import "testing"

func TestKBACLProcessMetricsExposeLowCardinalityDecisionCounters(t *testing.T) {
	before := CurrentKBACLMetricSnapshot()
	RecordKBACLAllowed()
	RecordKBACLDenied()
	RecordKBACLShadowDenied()
	RecordKBACLError()
	after := CurrentKBACLMetricSnapshot()
	if after.Allowed != before.Allowed+1 || after.Denied != before.Denied+1 ||
		after.ShadowDenied != before.ShadowDenied+1 || after.Errors != before.Errors+1 {
		t.Fatalf("before=%+v after=%+v", before, after)
	}
}
