package session

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/gin-gonic/gin"
)

type agentIntersectionKBService struct {
	interfaces.KnowledgeBaseService
	rows []*types.KnowledgeBase
}

func (s *agentIntersectionKBService) ListKnowledgeBasesByTenantID(context.Context, uint64) ([]*types.KnowledgeBase, error) {
	return s.rows, nil
}

type agentIntersectionAuthorizer struct {
	interfaces.KBAuthorizer
	allowed map[string]bool
}

func (a *agentIntersectionAuthorizer) Authorize(_ context.Context, req interfaces.KBPolicyRequest) (types.KBPolicyDecision, error) {
	return types.KBPolicyDecision{Allowed: a.allowed[req.KBID], Capability: req.Capability}, nil
}

func newAgentIntersectionContext() (*gin.Context, context.Context) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest("POST", "/api/v1/sessions/s-1/qa", nil)
	ctx := context.WithValue(req.Context(), types.TenantIDContextKey, uint64(8))
	ctx = context.WithValue(ctx, types.UserIDContextKey, "caller")
	ctx = context.WithValue(ctx, types.TenantRoleContextKey, types.TenantRoleViewer)
	c.Request = req.WithContext(ctx)
	c.Set(types.TenantIDContextKey.String(), uint64(8))
	return c, ctx
}

func TestAgentAllModeMaterializesCallerAuthorizedIntersection(t *testing.T) {
	c, ctx := newAgentIntersectionContext()
	h := &Handler{
		knowledgebaseService: &agentIntersectionKBService{rows: []*types.KnowledgeBase{
			{ID: "kb-allowed"}, {ID: "kb-denied"},
		}},
		kbAuthorizer: &agentIntersectionAuthorizer{allowed: map[string]bool{"kb-allowed": true}},
	}
	agent := &types.CustomAgent{ID: "shared-agent", TenantID: 99, Config: types.CustomAgentConfig{KBSelectionMode: "all"}}
	filtered, err := h.filterAgentKnowledgeBasesByAIQuery(ctx, c, agent, agent.ID)
	if err != nil {
		t.Fatal(err)
	}
	if filtered.Config.KBSelectionMode != "selected" || len(filtered.Config.KnowledgeBases) != 1 || filtered.Config.KnowledgeBases[0] != "kb-allowed" {
		t.Fatalf("filtered config=%+v", filtered.Config)
	}
	if agent.Config.KBSelectionMode != "all" || len(agent.Config.KnowledgeBases) != 0 {
		t.Fatalf("input agent was mutated: %+v", agent.Config)
	}
}

func TestAgentSelectedModeDropsUnauthorizedConfiguredKB(t *testing.T) {
	c, ctx := newAgentIntersectionContext()
	h := &Handler{kbAuthorizer: &agentIntersectionAuthorizer{allowed: map[string]bool{"kb-allowed": true}}}
	agent := &types.CustomAgent{ID: "agent-1", TenantID: 8, Config: types.CustomAgentConfig{
		KBSelectionMode: "selected", KnowledgeBases: []string{"kb-denied", "kb-allowed", "kb-allowed"},
	}}
	filtered, err := h.filterAgentKnowledgeBasesByAIQuery(ctx, c, agent, agent.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered.Config.KnowledgeBases) != 1 || filtered.Config.KnowledgeBases[0] != "kb-allowed" {
		t.Fatalf("filtered KBs=%v", filtered.Config.KnowledgeBases)
	}
}
