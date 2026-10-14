package router

import (
	"bytes"
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

type wikiBatchAuthorizer struct {
	interfaces.KBAuthorizer
	denied string
	seen   []string
}

func (a *wikiBatchAuthorizer) Authorize(_ context.Context, req interfaces.KBPolicyRequest) (types.KBPolicyDecision, error) {
	a.seen = append(a.seen, req.KBID)
	return types.KBPolicyDecision{Allowed: req.KBID != a.denied, Capability: req.Capability}, nil
}

func TestWikiBatchSearchChecksEveryKBAndPreservesBody(t *testing.T) {
	t.Setenv("WEKNORA_VONE_KB_ACL_MODE", "enforce")
	gin.SetMode(gin.TestMode)
	for _, denied := range []string{"second", ""} {
		t.Run("denied="+denied, func(t *testing.T) {
			a := &wikiBatchAuthorizer{denied: denied}
			r := gin.New()
			r.Use(middleware.ErrorHandler(), func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), types.TenantIDContextKey, uint64(1))
				ctx = context.WithValue(ctx, types.UserIDContextKey, "viewer")
				ctx = context.WithValue(ctx, types.TenantRoleContextKey, types.TenantRoleViewer)
				c.Request = c.Request.WithContext(ctx)
				c.Next()
			})
			g := &rbacGuards{kbAuthorizer: a}
			r.POST("/wiki-search", g.WikiSearchKBCapabilities(), func(c *gin.Context) {
				var body handler.SearchWikiRequest
				require.NoError(t, c.ShouldBindJSON(&body))
				require.Equal(t, "query", body.Query)
				c.Status(http.StatusNoContent)
			})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/wiki-search", bytes.NewBufferString(`{"query":"query","knowledge_base_id":"second","knowledge_base_ids":[" first ","first"]}`)))
			want := http.StatusNoContent
			if denied != "" {
				want = http.StatusForbidden
			}
			require.Equal(t, want, w.Code, w.Body.String())
			require.Equal(t, []string{"first", "second"}, a.seen)
		})
	}
}
