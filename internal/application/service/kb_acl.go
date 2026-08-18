package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tencent/WeKnora/internal/application/repository"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

var (
	ErrInvalidKBMemberRole      = errors.New("invalid knowledge-base member role")
	ErrInvalidKBCapability      = errors.New("invalid knowledge-base capability")
	ErrKBMemberOutsideWorkspace = errors.New("knowledge-base member is not active in workspace")
	ErrLastKBOwner              = errors.New("cannot remove or demote the final knowledge-base owner")
)

type kbMembershipService struct {
	repo          interfaces.KBMembershipRepository
	tenantMembers interfaces.TenantMemberService
	audit         interfaces.AuditLogService
}

func NewKBMembershipService(
	repo interfaces.KBMembershipRepository,
	tenantMembers interfaces.TenantMemberService,
	audit interfaces.AuditLogService,
) interfaces.KBMembershipService {
	return &kbMembershipService{repo: repo, tenantMembers: tenantMembers, audit: audit}
}

func (s *kbMembershipService) Create(
	ctx context.Context, actorID string, membership *types.KBMembership, capabilities []types.KBCapability,
) error {
	if err := s.validate(ctx, membership, capabilities); err != nil {
		return err
	}
	membership.GrantedBy = actorID
	if err := s.repo.Create(ctx, membership, capabilities); err != nil {
		return err
	}
	s.auditChange(ctx, membership, actorID, types.AuditActionKBMemberGranted)
	return nil
}

func (s *kbMembershipService) Get(ctx context.Context, tenantID uint64, kbID, userID string) (*types.KBMembership, error) {
	return s.repo.Get(ctx, tenantID, kbID, userID)
}

func (s *kbMembershipService) ListByKB(ctx context.Context, tenantID uint64, kbID string) ([]*types.KBMembership, error) {
	return s.repo.ListByKB(ctx, tenantID, kbID)
}

func (s *kbMembershipService) ListAccessibleKBIDs(ctx context.Context, tenantID uint64, userID string) ([]string, error) {
	return s.repo.ListAccessibleKBIDs(ctx, tenantID, userID)
}

func (s *kbMembershipService) Update(
	ctx context.Context, actorID string, membership *types.KBMembership, capabilities []types.KBCapability,
) error {
	if err := s.validate(ctx, membership, capabilities); err != nil {
		return err
	}
	membership.GrantedBy = actorID
	if err := s.repo.Update(ctx, membership, capabilities); err != nil {
		if errors.Is(err, repository.ErrLastKBOwner) {
			return ErrLastKBOwner
		}
		return err
	}
	s.auditChange(ctx, membership, actorID, types.AuditActionKBMemberChanged)
	return nil
}

func (s *kbMembershipService) Delete(
	ctx context.Context, actorID string, tenantID uint64, kbID, userID string,
) error {
	current, err := s.repo.Get(ctx, tenantID, kbID, userID)
	if err != nil {
		return err
	}
	if current == nil {
		return gormRecordNotFound("knowledge-base member")
	}
	if err := s.repo.Delete(ctx, tenantID, kbID, userID); err != nil {
		if errors.Is(err, repository.ErrLastKBOwner) {
			return ErrLastKBOwner
		}
		return err
	}
	s.auditChange(ctx, current, actorID, types.AuditActionKBMemberRevoked)
	return nil
}

func (s *kbMembershipService) validate(
	ctx context.Context, membership *types.KBMembership, capabilities []types.KBCapability,
) error {
	if membership == nil || !membership.Role.IsValid() {
		return ErrInvalidKBMemberRole
	}
	member, err := s.tenantMembers.GetMembership(ctx, membership.UserID, membership.TenantID)
	if err != nil {
		return err
	}
	if member == nil || member.Status != types.TenantMemberStatusActive {
		return ErrKBMemberOutsideWorkspace
	}
	seen := map[types.KBCapability]bool{}
	for _, capability := range capabilities {
		if !capability.IsValid() || seen[capability] {
			return fmt.Errorf("%w: %s", ErrInvalidKBCapability, capability)
		}
		seen[capability] = true
	}
	if membership.Role != types.KBMemberRoleCustom && len(capabilities) != 0 {
		return fmt.Errorf("%w: presets cannot carry overrides", ErrInvalidKBCapability)
	}
	return nil
}

func (s *kbMembershipService) auditChange(
	ctx context.Context, membership *types.KBMembership, actorID string, action types.AuditAction,
) {
	if s.audit == nil {
		return
	}
	_ = s.audit.Log(ctx, &types.AuditLog{
		TenantID: membership.TenantID, ActorUserID: actorID, Action: action,
		ScopeType: "knowledge_base", ScopeID: membership.KBID,
		TargetType: "user", TargetID: membership.UserID, TargetUserID: membership.UserID,
		Outcome: types.AuditOutcomeSuccess,
	})
}

func gormRecordNotFound(resource string) error { return fmt.Errorf("%s not found", resource) }

type kbAuthorizer struct {
	members interfaces.KBMembershipRepository
	kbs     interfaces.KnowledgeBaseRepository
	shares  interfaces.KBShareService
	agents  interfaces.AgentShareService
}

func NewKBAuthorizer(
	members interfaces.KBMembershipRepository,
	kbs interfaces.KnowledgeBaseRepository,
	shares interfaces.KBShareService,
	agents interfaces.AgentShareService,
) interfaces.KBAuthorizer {
	return &kbAuthorizer{members: members, kbs: kbs, shares: shares, agents: agents}
}

func (a *kbAuthorizer) Authorize(
	ctx context.Context, request interfaces.KBPolicyRequest,
) (types.KBPolicyDecision, error) {
	denied := types.KBPolicyDecision{Allowed: false, Capability: request.Capability, Source: types.KBPolicySourceNone, Reason: "default_deny"}
	if request.TenantID == 0 || request.KBID == "" || !request.Capability.IsValid() {
		return denied, nil
	}
	// An Agent is a scope constraint, never an independent KB grant. The
	// effective retrieval set must be:
	//
	//     agent configured KBs ∩ caller kb.ai.query KBs
	//
	// Resolve the Agent before every other grant source so Workspace Owner,
	// membership, organization-share and API-key decisions cannot accidentally
	// query a KB outside the Agent configuration. Conversely, a configured KB
	// still falls through to the caller's ordinary authorization below; sharing
	// an Agent alone cannot make source documents queryable.
	if request.AgentID != "" && request.Capability == types.KBCapabilityAIQuery {
		if a.agents == nil {
			return denied, nil
		}
		agent, err := a.agents.GetSharedAgentForTenant(ctx, request.TenantID, request.TenantRole, request.AgentID)
		if err != nil {
			return denied, err
		}
		if agent == nil || !agentAllowsKB(agent, request.KBID) {
			denied.Reason = "agent_scope_missing"
			return denied, nil
		}
	}
	if scope, ok := types.TenantAPIKeyScopeFromContext(ctx); ok {
		capabilities := apiKeyKBCapabilities(scope)
		if !scope.AllowsKnowledgeBase(request.KBID) {
			return denied, nil
		}
		decision := decisionFromCapabilities(request.Capability, types.KBMemberRoleCustom, types.KBPolicySourceAPIKey, capabilities)
		return decision, nil
	}
	if request.TenantRole == types.TenantRoleOwner {
		return decisionForRole(request.Capability, types.KBMemberRoleOwner, types.KBPolicySourceWorkspaceOwner), nil
	}
	if request.UserID != "" {
		membership, err := a.members.Get(ctx, request.TenantID, request.KBID, request.UserID)
		if err != nil {
			return denied, err
		}
		if membership != nil {
			capabilities := effectiveMembershipCapabilities(membership)
			return decisionFromCapabilities(request.Capability, membership.Role, types.KBPolicySourceMembership, capabilities), nil
		}
	}
	if a.shares != nil {
		permission, ok, err := a.shares.CheckTenantKBPermission(ctx, request.KBID, request.TenantID, request.TenantRole)
		if err != nil {
			return denied, err
		}
		if ok {
			role := types.KBMemberRoleDocumentViewer
			if permission == types.OrgRoleAdmin || permission == types.OrgRoleEditor {
				role = types.KBMemberRoleEditor
				return decisionForRole(request.Capability, role, types.KBPolicySourceOrganization), nil
			}
			// Preserve the existing organization-viewer contract: parsed
			// content and citations are readable, but the original file is not
			// downloadable. The central decision must express this restriction
			// instead of allowing download here and relying on a later legacy
			// handler check to deny it.
			capabilities := types.KBRoleCapabilities(role)
			filtered := make([]types.KBCapability, 0, len(capabilities))
			for _, candidate := range capabilities {
				if candidate != types.KBCapabilityDocumentDownload {
					filtered = append(filtered, candidate)
				}
			}
			return decisionFromCapabilities(request.Capability, role, types.KBPolicySourceOrganization, filtered), nil
		}
	}
	return denied, nil
}

func apiKeyKBCapabilities(scope types.TenantAPIKeyScope) []types.KBCapability {
	if scope.FullAccess {
		return types.AllKBCapabilities()
	}
	set := map[types.KBCapability]bool{}
	add := func(items ...types.KBCapability) {
		for _, item := range items {
			set[item] = true
		}
	}
	if scope.HasCapability(types.APIKeyCapabilityRetrieve) || scope.HasCapability(types.APIKeyCapabilityChat) {
		add(types.KBCapabilityAIQuery, types.KBCapabilityMetadataRead, types.KBCapabilityDocumentsList,
			types.KBCapabilityDocumentPreview, types.KBCapabilityChunkPreview, types.KBCapabilityDocumentDownload)
	}
	if scope.HasCapability(types.APIKeyCapabilityIngest) {
		add(types.KBCapabilityDocumentUpload, types.KBCapabilityDocumentEdit, types.KBCapabilityDocumentDelete,
			types.KBCapabilityDocumentReparse, types.KBCapabilityChunkEdit, types.KBCapabilityChunkDelete, types.KBCapabilityFolderManage)
	}
	if scope.HasCapability(types.APIKeyCapabilityManageKnowledgeBases) {
		add(types.KBCapabilityMetadataRead, types.KBCapabilitySettingsEdit, types.KBCapabilityDelete)
	}
	result := make([]types.KBCapability, 0, len(set))
	for item := range set {
		result = append(result, item)
	}
	return result
}

func agentAllowsKB(agent *types.CustomAgent, kbID string) bool {
	if agent == nil {
		return false
	}
	switch agent.Config.KBSelectionMode {
	case "all":
		return true
	case "selected", "":
		for _, allowed := range agent.Config.KnowledgeBases {
			if allowed == kbID {
				return true
			}
		}
	}
	return false
}

func (a *kbAuthorizer) ListAccessibleKBIDs(
	ctx context.Context, tenantID uint64, userID string, role types.TenantRole,
) ([]string, error) {
	if role == types.TenantRoleOwner {
		kbs, err := a.kbs.ListKnowledgeBasesByTenantID(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		ids := make([]string, 0, len(kbs))
		for _, kb := range kbs {
			ids = append(ids, kb.ID)
		}
		return ids, nil
	}
	return a.members.ListAccessibleKBIDs(ctx, tenantID, userID)
}

func effectiveMembershipCapabilities(membership *types.KBMembership) []types.KBCapability {
	if membership.Role != types.KBMemberRoleCustom {
		return types.KBRoleCapabilities(membership.Role)
	}
	capabilities := make([]types.KBCapability, 0, len(membership.Capabilities))
	for _, row := range membership.Capabilities {
		capabilities = append(capabilities, row.Capability)
	}
	return capabilities
}

func decisionForRole(capability types.KBCapability, role types.KBMemberRole, source types.KBPolicySource) types.KBPolicyDecision {
	return decisionFromCapabilities(capability, role, source, types.KBRoleCapabilities(role))
}

func decisionFromCapabilities(
	capability types.KBCapability, role types.KBMemberRole, source types.KBPolicySource, capabilities []types.KBCapability,
) types.KBPolicyDecision {
	allowed := false
	for _, candidate := range capabilities {
		allowed = allowed || candidate == capability
	}
	reason := "capability_missing"
	if allowed {
		reason = "capability_granted"
	}
	return types.KBPolicyDecision{Allowed: allowed, Capability: capability, Source: source, Role: role, Capabilities: capabilities, Reason: reason}
}
