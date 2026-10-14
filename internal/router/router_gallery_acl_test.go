package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tencent/WeKnora/internal/handler"
	"github.com/Tencent/WeKnora/internal/middleware"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type galleryCapabilityAuthorizer struct {
	interfaces.KBAuthorizer
	grants map[types.KBCapability]bool
	denied types.KBCapability
}

func (a *galleryCapabilityAuthorizer) Authorize(_ context.Context, req interfaces.KBPolicyRequest) (types.KBPolicyDecision, error) {
	allowed := a.grants[req.Capability]
	if !allowed {
		a.denied = req.Capability
	}
	return types.KBPolicyDecision{Allowed: allowed, Capability: req.Capability}, nil
}

// Exercise registered routes: a generic Viewer must not gain source-image
// access through the new gallery when VONE grants only AI or file names.
func TestGalleryRoutesRejectMissingContentCapabilities(t *testing.T) {
	t.Setenv("WEKNORA_VONE_KB_ACL_MODE", "enforce")
	gin.SetMode(gin.TestMode)
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/images"},
		{http.MethodPost, "/images"},
		{http.MethodGet, "/gallery-config"},
	} {
		for _, missing := range []types.KBCapability{types.KBCapabilityDocumentsList, types.KBCapabilityDocumentPreview} {
			t.Run(route.method+route.path+string(missing), func(t *testing.T) {
				authz := &galleryCapabilityAuthorizer{grants: map[types.KBCapability]bool{
					types.KBCapabilityDocumentsList: true, types.KBCapabilityDocumentPreview: true,
				}}
				authz.grants[missing] = false
				r := gin.New()
				r.Use(middleware.ErrorHandler(), func(c *gin.Context) {
					ctx := context.WithValue(c.Request.Context(), types.TenantIDContextKey, uint64(1))
					ctx = context.WithValue(ctx, types.UserIDContextKey, "viewer")
					ctx = context.WithValue(ctx, types.TenantRoleContextKey, types.TenantRoleViewer)
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				RegisterKnowledgeRoutes(r.Group("/api/v1"), &handler.KnowledgeHandler{}, &rbacGuards{kbAuthorizer: authz})
				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest(route.method, "/api/v1/knowledge-bases/kb-1"+route.path, nil))
				require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
				require.Equal(t, missing, authz.denied)
			})
		}
	}
}
