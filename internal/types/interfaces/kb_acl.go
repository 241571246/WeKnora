package interfaces

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
)

type KBMembershipRepository interface {
	Create(ctx context.Context, membership *types.KBMembership, capabilities []types.KBCapability) error
	Get(ctx context.Context, tenantID uint64, kbID, userID string) (*types.KBMembership, error)
	ListByKB(ctx context.Context, tenantID uint64, kbID string) ([]*types.KBMembership, error)
	ListAccessibleKBIDs(ctx context.Context, tenantID uint64, userID string) ([]string, error)
	Update(ctx context.Context, membership *types.KBMembership, capabilities []types.KBCapability) error
	Delete(ctx context.Context, tenantID uint64, kbID, userID string) error
	CountOwners(ctx context.Context, tenantID uint64, kbID string) (int64, error)
}

type KBMembershipService interface {
	Create(ctx context.Context, actorID string, membership *types.KBMembership, capabilities []types.KBCapability) error
	Get(ctx context.Context, tenantID uint64, kbID, userID string) (*types.KBMembership, error)
	ListByKB(ctx context.Context, tenantID uint64, kbID string) ([]*types.KBMembership, error)
	ListAccessibleKBIDs(ctx context.Context, tenantID uint64, userID string) ([]string, error)
	Update(ctx context.Context, actorID string, membership *types.KBMembership, capabilities []types.KBCapability) error
	Delete(ctx context.Context, actorID string, tenantID uint64, kbID, userID string) error
}

type KBPolicyRequest struct {
	TenantID       uint64
	KBID           string
	UserID         string
	TenantRole     types.TenantRole
	Capability     types.KBCapability
	OrganizationID string
	AgentID        string
	APIKeyID       string
}

// KBAuthorizer is the only supported decision point for same-workspace ACL,
// organization shares, shared agents and API keys.
type KBAuthorizer interface {
	Authorize(ctx context.Context, request KBPolicyRequest) (types.KBPolicyDecision, error)
	ListAccessibleKBIDs(ctx context.Context, tenantID uint64, userID string, role types.TenantRole) ([]string, error)
}

type KBCollectionRepository interface {
	Create(ctx context.Context, collection *types.KBCollection) error
	Get(ctx context.Context, tenantID uint64, id string) (*types.KBCollection, error)
	List(ctx context.Context, tenantID uint64) ([]*types.KBCollection, error)
	Update(ctx context.Context, collection *types.KBCollection) error
	Delete(ctx context.Context, tenantID uint64, id string) error
	CountChildrenAndBindings(ctx context.Context, tenantID uint64, id string) (int64, error)
	SetBinding(ctx context.Context, binding *types.KBCollectionBinding) error
	DeleteBinding(ctx context.Context, tenantID uint64, kbID string) error
	ListBindings(ctx context.Context, tenantID uint64) ([]*types.KBCollectionBinding, error)
}

type KBCollectionService interface {
	Create(ctx context.Context, collection *types.KBCollection) error
	GetTree(ctx context.Context, tenantID uint64, visibleKBIDs []string) ([]*types.KBCollection, []*types.KBCollectionBinding, error)
	Update(ctx context.Context, collection *types.KBCollection) error
	Delete(ctx context.Context, tenantID uint64, id string) error
	SetKnowledgeBaseCollection(ctx context.Context, binding *types.KBCollectionBinding) error
}
