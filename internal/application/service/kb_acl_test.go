package service

import (
	"context"
	"testing"
	"time"

	"github.com/Tencent/WeKnora/internal/application/repository"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fixedKBShareService struct {
	interfaces.KBShareService
	permission types.OrgMemberRole
	allowed    bool
}

func (s *fixedKBShareService) CheckTenantKBPermission(
	context.Context, string, uint64, types.TenantRole,
) (types.OrgMemberRole, bool, error) {
	return s.permission, s.allowed, nil
}

type fixedAgentShareService struct {
	interfaces.AgentShareService
	agent *types.CustomAgent
}

type activeTenantMemberService struct{ interfaces.TenantMemberService }

func (*activeTenantMemberService) GetMembership(_ context.Context, userID string, tenantID uint64) (*types.TenantMember, error) {
	return &types.TenantMember{UserID: userID, TenantID: tenantID, Status: types.TenantMemberStatusActive}, nil
}

type captureKBACLAudit struct {
	interfaces.AuditLogService
	entries []*types.AuditLog
}

func (a *captureKBACLAudit) Log(_ context.Context, entry *types.AuditLog) error {
	a.entries = append(a.entries, entry)
	return nil
}

func (s *fixedAgentShareService) GetSharedAgentForTenant(
	context.Context, uint64, types.TenantRole, string, ...uint64,
) (*types.CustomAgent, error) {
	return s.agent, nil
}

func TestKBAuthorizerAIUserCannotPreviewDocuments(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&types.KnowledgeBase{}, &types.KBMembership{}, &types.KBMemberCapability{}); err != nil {
		t.Fatal(err)
	}
	members := repository.NewKBMembershipRepository(db)
	kbs := repository.NewKnowledgeBaseRepository(db)
	ctx := context.Background()
	if err := kbs.CreateKnowledgeBase(ctx, &types.KnowledgeBase{ID: "kb-1", TenantID: 7, Name: "test"}); err != nil {
		t.Fatal(err)
	}
	if err := members.Create(ctx, &types.KBMembership{TenantID: 7, KBID: "kb-1", UserID: "u-1", Role: types.KBMemberRoleAIUser, GrantedBy: "owner"}, nil); err != nil {
		t.Fatal(err)
	}
	authorizer := NewKBAuthorizer(members, kbs, nil, nil)
	query, err := authorizer.Authorize(ctx, interfaces.KBPolicyRequest{TenantID: 7, KBID: "kb-1", UserID: "u-1", TenantRole: types.TenantRoleViewer, Capability: types.KBCapabilityAIQuery})
	if err != nil || !query.Allowed {
		t.Fatalf("AI query decision=%+v err=%v", query, err)
	}
	preview, err := authorizer.Authorize(ctx, interfaces.KBPolicyRequest{TenantID: 7, KBID: "kb-1", UserID: "u-1", TenantRole: types.TenantRoleViewer, Capability: types.KBCapabilityDocumentPreview})
	if err != nil || preview.Allowed {
		t.Fatalf("preview decision=%+v err=%v", preview, err)
	}
}

func TestKBMembershipMutationsWriteKBScopedAuditEntries(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&types.KBMembership{}, &types.KBMemberCapability{}); err != nil {
		t.Fatal(err)
	}
	audit := &captureKBACLAudit{}
	service := NewKBMembershipService(repository.NewKBMembershipRepository(db), &activeTenantMemberService{}, audit)
	ctx := context.Background()
	membership := &types.KBMembership{TenantID: 7, KBID: "kb-1", UserID: "employee", Role: types.KBMemberRoleCustom}
	if err := service.Create(ctx, "owner", membership, []types.KBCapability{types.KBCapabilityAIQuery}); err != nil {
		t.Fatal(err)
	}
	if err := service.Update(ctx, "owner", membership, []types.KBCapability{types.KBCapabilityMetadataRead}); err != nil {
		t.Fatal(err)
	}
	if err := service.Delete(ctx, "owner", 7, "kb-1", "employee"); err != nil {
		t.Fatal(err)
	}
	want := []types.AuditAction{
		types.AuditActionKBMemberGranted,
		types.AuditActionKBMemberChanged,
		types.AuditActionKBMemberRevoked,
	}
	if len(audit.entries) != len(want) {
		t.Fatalf("audit entries=%d want=%d", len(audit.entries), len(want))
	}
	for index, entry := range audit.entries {
		if entry.Action != want[index] || entry.ScopeType != "knowledge_base" || entry.ScopeID != "kb-1" ||
			entry.TargetType != "user" || entry.TargetID != "employee" || entry.Outcome != types.AuditOutcomeSuccess {
			t.Fatalf("audit[%d]=%+v", index, entry)
		}
	}
}

func TestKBAuthorizerAPIKeyRequiresCapabilityAndAllowList(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&types.KnowledgeBase{}, &types.KBMembership{}, &types.KBMemberCapability{}); err != nil {
		t.Fatal(err)
	}
	authorizer := NewKBAuthorizer(repository.NewKBMembershipRepository(db), repository.NewKnowledgeBaseRepository(db), nil, nil)
	ctx := types.WithTenantAPIKeyScope(context.Background(), types.TenantAPIKeyScope{
		KeyID: 9, Capabilities: types.StringArray{string(types.APIKeyCapabilityRetrieve)}, KnowledgeBaseIDs: types.StringArray{"kb-1"},
	})
	allowed, err := authorizer.Authorize(ctx, interfaces.KBPolicyRequest{TenantID: 7, KBID: "kb-1", Capability: types.KBCapabilityAIQuery})
	if err != nil || !allowed.Allowed || allowed.Source != types.KBPolicySourceAPIKey {
		t.Fatalf("allowed=%+v err=%v", allowed, err)
	}
	blocked, err := authorizer.Authorize(ctx, interfaces.KBPolicyRequest{TenantID: 7, KBID: "kb-2", Capability: types.KBCapabilityAIQuery})
	if err != nil || blocked.Allowed {
		t.Fatalf("blocked=%+v err=%v", blocked, err)
	}
	write, err := authorizer.Authorize(ctx, interfaces.KBPolicyRequest{TenantID: 7, KBID: "kb-1", Capability: types.KBCapabilityDocumentUpload})
	if err != nil || write.Allowed {
		t.Fatalf("write=%+v err=%v", write, err)
	}
}

func TestKBAuthorizerKeepsThreeDeleteCapabilitiesIndependent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&types.KnowledgeBase{}, &types.KBMembership{}, &types.KBMemberCapability{}); err != nil {
		t.Fatal(err)
	}
	members := repository.NewKBMembershipRepository(db)
	kbs := repository.NewKnowledgeBaseRepository(db)
	ctx := context.Background()
	if err := kbs.CreateKnowledgeBase(ctx, &types.KnowledgeBase{ID: "kb-1", TenantID: 7, Name: "test"}); err != nil {
		t.Fatal(err)
	}
	if err := members.Create(ctx, &types.KBMembership{TenantID: 7, KBID: "kb-1", UserID: "u-1", Role: types.KBMemberRoleCustom, GrantedBy: "owner"}, []types.KBCapability{types.KBCapabilityDocumentDelete}); err != nil {
		t.Fatal(err)
	}
	authorizer := NewKBAuthorizer(members, kbs, nil, nil)
	for capability, want := range map[types.KBCapability]bool{
		types.KBCapabilityDocumentDelete: true,
		types.KBCapabilityChunkDelete:    false,
		types.KBCapabilityDelete:         false,
	} {
		decision, err := authorizer.Authorize(ctx, interfaces.KBPolicyRequest{TenantID: 7, KBID: "kb-1", UserID: "u-1", TenantRole: types.TenantRoleViewer, Capability: capability})
		if err != nil || decision.Allowed != want {
			t.Fatalf("capability %s allowed=%v want=%v err=%v", capability, decision.Allowed, want, err)
		}
	}
}

func TestKBAuthorizerRevocationIsEffectiveImmediately(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&types.KnowledgeBase{}, &types.KBMembership{}, &types.KBMemberCapability{}); err != nil {
		t.Fatal(err)
	}
	members := repository.NewKBMembershipRepository(db)
	authorizer := NewKBAuthorizer(members, repository.NewKnowledgeBaseRepository(db), nil, nil)
	ctx := context.Background()
	membership := &types.KBMembership{TenantID: 7, KBID: "kb-1", UserID: "u-1", Role: types.KBMemberRoleCustom, GrantedBy: "owner"}
	if err := members.Create(ctx, membership, []types.KBCapability{types.KBCapabilityAIQuery}); err != nil {
		t.Fatal(err)
	}
	req := interfaces.KBPolicyRequest{TenantID: 7, KBID: "kb-1", UserID: "u-1", TenantRole: types.TenantRoleViewer, Capability: types.KBCapabilityAIQuery}
	before, err := authorizer.Authorize(ctx, req)
	if err != nil || !before.Allowed {
		t.Fatalf("before revocation=%+v err=%v", before, err)
	}
	started := time.Now()
	membership.Role = types.KBMemberRoleCustom
	if err := members.Update(ctx, membership, nil); err != nil {
		t.Fatal(err)
	}
	after, err := authorizer.Authorize(ctx, req)
	if err != nil || after.Allowed {
		t.Fatalf("after revocation=%+v err=%v", after, err)
	}
	if elapsed := time.Since(started); elapsed >= 5*time.Second {
		t.Fatalf("revocation took %s, must be under 5 seconds", elapsed)
	}
}

func TestKBAuthorizerOrganizationViewerCannotDownloadOriginal(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&types.KnowledgeBase{}, &types.KBMembership{}, &types.KBMemberCapability{}); err != nil {
		t.Fatal(err)
	}
	authorizer := NewKBAuthorizer(
		repository.NewKBMembershipRepository(db), repository.NewKnowledgeBaseRepository(db),
		&fixedKBShareService{permission: types.OrgRoleViewer, allowed: true}, nil,
	)
	base := interfaces.KBPolicyRequest{TenantID: 8, KBID: "foreign-kb", UserID: "u-1", TenantRole: types.TenantRoleViewer}
	base.Capability = types.KBCapabilityDocumentPreview
	preview, err := authorizer.Authorize(context.Background(), base)
	if err != nil || !preview.Allowed || preview.Source != types.KBPolicySourceOrganization {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
	base.Capability = types.KBCapabilityDocumentDownload
	download, err := authorizer.Authorize(context.Background(), base)
	if err != nil || download.Allowed {
		t.Fatalf("download=%+v err=%v", download, err)
	}
}

func TestKBAuthorizerAgentRequiresConfiguredKBAndCallerAIQuery(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&types.KnowledgeBase{}, &types.KBMembership{}, &types.KBMemberCapability{}); err != nil {
		t.Fatal(err)
	}
	agent := &types.CustomAgent{ID: "agent-1", Config: types.CustomAgentConfig{
		KBSelectionMode: "selected", KnowledgeBases: []string{"kb-1"},
	}}
	authorizer := NewKBAuthorizer(
		repository.NewKBMembershipRepository(db), repository.NewKnowledgeBaseRepository(db), nil,
		&fixedAgentShareService{agent: agent},
	)
	members := repository.NewKBMembershipRepository(db)
	for _, kbID := range []string{"kb-1", "kb-outside-agent"} {
		membership := &types.KBMembership{TenantID: 8, KBID: kbID, UserID: "u-1", Role: types.KBMemberRoleCustom, GrantedBy: "owner"}
		if err := members.Create(context.Background(), membership, []types.KBCapability{types.KBCapabilityAIQuery}); err != nil {
			t.Fatal(err)
		}
	}
	base := interfaces.KBPolicyRequest{TenantID: 8, KBID: "kb-1", UserID: "u-1", TenantRole: types.TenantRoleViewer, AgentID: "agent-1"}
	base.Capability = types.KBCapabilityAIQuery
	query, err := authorizer.Authorize(context.Background(), base)
	if err != nil || !query.Allowed || query.Source != types.KBPolicySourceMembership {
		t.Fatalf("query=%+v err=%v", query, err)
	}
	base.KBID = "kb-outside-agent"
	outsideAgent, err := authorizer.Authorize(context.Background(), base)
	if err != nil || outsideAgent.Allowed || outsideAgent.Reason != "agent_scope_missing" {
		t.Fatalf("outsideAgent=%+v err=%v", outsideAgent, err)
	}
	base.KBID = "kb-1"
	base.UserID = "u-without-kb-grant"
	withoutCallerGrant, err := authorizer.Authorize(context.Background(), base)
	if err != nil || withoutCallerGrant.Allowed {
		t.Fatalf("withoutCallerGrant=%+v err=%v", withoutCallerGrant, err)
	}
	base.UserID = "u-1"
	base.Capability = types.KBCapabilityDocumentPreview
	preview, err := authorizer.Authorize(context.Background(), base)
	if err != nil || preview.Allowed {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
	base.KBID = "kb-2"
	base.Capability = types.KBCapabilityAIQuery
	outside, err := authorizer.Authorize(context.Background(), base)
	if err != nil || outside.Allowed {
		t.Fatalf("outside=%+v err=%v", outside, err)
	}
}

func TestKBAuthorizerWorkspaceAdminDoesNotReceiveOwnerOverride(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&types.KnowledgeBase{}, &types.KBMembership{}, &types.KBMemberCapability{}); err != nil {
		t.Fatal(err)
	}
	authorizer := NewKBAuthorizer(repository.NewKBMembershipRepository(db), repository.NewKnowledgeBaseRepository(db), nil, nil)
	decision, err := authorizer.Authorize(context.Background(), interfaces.KBPolicyRequest{
		TenantID: 7, KBID: "kb-1", UserID: "admin", TenantRole: types.TenantRoleAdmin, Capability: types.KBCapabilityMetadataRead,
	})
	if err != nil || decision.Allowed {
		t.Fatalf("workspace admin decision=%+v err=%v", decision, err)
	}
}
