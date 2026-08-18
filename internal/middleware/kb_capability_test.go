package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/gin-gonic/gin"
)

type fixedKBAuthorizer struct {
	allowed bool
	calls   int
}

type captureKBDenyAudit struct {
	interfaces.AuditLogService
	entries []*types.AuditLog
}

func (a *captureKBDenyAudit) Log(_ context.Context, entry *types.AuditLog) error {
	a.entries = append(a.entries, entry)
	return nil
}

func (a *fixedKBAuthorizer) Authorize(_ context.Context, req interfaces.KBPolicyRequest) (types.KBPolicyDecision, error) {
	a.calls++
	return types.KBPolicyDecision{Allowed: a.allowed, Capability: req.Capability, Source: types.KBPolicySourceMembership}, nil
}

func (*fixedKBAuthorizer) ListAccessibleKBIDs(context.Context, uint64, string, types.TenantRole) ([]string, error) {
	return nil, nil
}

func runKBCapabilityRequest(t *testing.T, mode string, authorizer interfaces.KBAuthorizer) int {
	t.Helper()
	t.Setenv("WEKNORA_VONE_KB_ACL_MODE", mode)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorHandler())
	r.GET("/kb/:id", RequireKBCapability(func(c *gin.Context) (string, error) { return c.Param("id"), nil }, types.KBCapabilityDocumentPreview, authorizer), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	req := httptest.NewRequest(http.MethodGet, "/kb/kb-1", nil)
	ctx := context.WithValue(req.Context(), types.TenantIDContextKey, uint64(7))
	ctx = context.WithValue(ctx, types.UserIDContextKey, "u-1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestRequireKBCapabilityRolloutModes(t *testing.T) {
	denied := &fixedKBAuthorizer{}
	if got := runKBCapabilityRequest(t, "enforce", denied); got != http.StatusForbidden {
		t.Fatalf("enforce status=%d want=%d", got, http.StatusForbidden)
	}
	if got := runKBCapabilityRequest(t, "shadow", denied); got != http.StatusNoContent {
		t.Fatalf("shadow status=%d want=%d", got, http.StatusNoContent)
	}
	off := &fixedKBAuthorizer{}
	if got := runKBCapabilityRequest(t, "off", off); got != http.StatusNoContent || off.calls != 0 {
		t.Fatalf("off status=%d calls=%d", got, off.calls)
	}
}

func TestRequireKBCapabilityWritesQueryableDenialAudit(t *testing.T) {
	t.Setenv("WEKNORA_VONE_KB_ACL_MODE", "enforce")
	gin.SetMode(gin.TestMode)
	audit := &captureKBDenyAudit{}
	r := gin.New()
	r.Use(AuditServiceProvider(audit), ErrorHandler())
	r.GET("/kb/:id", RequireKBCapability(
		func(c *gin.Context) (string, error) { return c.Param("id"), nil },
		types.KBCapabilityDocumentDownload,
		&fixedKBAuthorizer{},
	), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	req := httptest.NewRequest(http.MethodGet, "/kb/kb-1", nil)
	ctx := context.WithValue(req.Context(), types.TenantIDContextKey, uint64(7))
	ctx = context.WithValue(ctx, types.UserIDContextKey, "u-1")
	ctx = context.WithValue(ctx, types.TenantRoleContextKey, types.TenantRoleViewer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req.WithContext(ctx))
	if w.Code != http.StatusForbidden || len(audit.entries) != 1 {
		t.Fatalf("status=%d audit entries=%d", w.Code, len(audit.entries))
	}
	entry := audit.entries[0]
	if entry.Action != types.AuditActionAccessDenied || entry.ScopeID != "kb-1" || entry.Outcome != types.AuditOutcomeDenied {
		t.Fatalf("unexpected audit entry: %+v", entry)
	}
	if !bytes.Contains(entry.Details, []byte(string(types.KBCapabilityDocumentDownload))) {
		t.Fatalf("audit details omit capability: %s", entry.Details)
	}
}
