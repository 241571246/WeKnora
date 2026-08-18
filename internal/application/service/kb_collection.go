package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Tencent/WeKnora/internal/application/repository"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrKBCollectionNotFound = errors.New("knowledge-base collection not found")
	ErrKBCollectionCycle    = errors.New("knowledge-base collection cycle")
	ErrKBCollectionNotEmpty = errors.New("knowledge-base collection is not empty")
)

type kbCollectionService struct {
	repo  interfaces.KBCollectionRepository
	kbs   interfaces.KnowledgeBaseRepository
	audit interfaces.AuditLogService
}

func NewKBCollectionService(
	repo interfaces.KBCollectionRepository,
	kbs interfaces.KnowledgeBaseRepository,
	audit interfaces.AuditLogService,
) interfaces.KBCollectionService {
	return &kbCollectionService{repo: repo, kbs: kbs, audit: audit}
}

func (s *kbCollectionService) Create(ctx context.Context, collection *types.KBCollection) error {
	collection.Name = strings.TrimSpace(collection.Name)
	if collection.TenantID == 0 || collection.Name == "" || len(collection.Name) > 128 {
		return fmt.Errorf("invalid collection name or workspace")
	}
	if collection.ID == "" {
		collection.ID = uuid.NewString()
	}
	if err := s.validateParent(ctx, collection.TenantID, collection.ID, collection.ParentID); err != nil {
		return err
	}
	if err := s.repo.Create(ctx, collection); err != nil {
		return err
	}
	s.auditCollection(ctx, collection, types.AuditActionKBCollectionCreated)
	return nil
}

func (s *kbCollectionService) GetTree(
	ctx context.Context, tenantID uint64, visibleKBIDs []string,
) ([]*types.KBCollection, []*types.KBCollectionBinding, error) {
	collections, err := s.repo.List(ctx, tenantID)
	if err != nil {
		return nil, nil, err
	}
	bindings, err := s.repo.ListBindings(ctx, tenantID)
	if err != nil {
		return nil, nil, err
	}
	// nil is the explicit workspace-Owner view: include empty collections so
	// administrators can build the hierarchy before assigning KBs.
	if visibleKBIDs == nil {
		return collections, bindings, nil
	}
	visible := make(map[string]bool, len(visibleKBIDs))
	for _, id := range visibleKBIDs {
		visible[id] = true
	}
	filteredBindings := make([]*types.KBCollectionBinding, 0, len(bindings))
	keep := map[string]bool{}
	for _, binding := range bindings {
		if visible[binding.KBID] {
			filteredBindings = append(filteredBindings, binding)
			keep[binding.CollectionID] = true
		}
	}
	byID := make(map[string]*types.KBCollection, len(collections))
	for _, collection := range collections {
		byID[collection.ID] = collection
	}
	for id := range keep {
		current := byID[id]
		for depth := 0; current != nil && current.ParentID != nil && depth < 256; depth++ {
			keep[*current.ParentID] = true
			current = byID[*current.ParentID]
		}
	}
	filteredCollections := make([]*types.KBCollection, 0, len(keep))
	for _, collection := range collections {
		if keep[collection.ID] {
			filteredCollections = append(filteredCollections, collection)
		}
	}
	return filteredCollections, filteredBindings, nil
}

func (s *kbCollectionService) Update(ctx context.Context, collection *types.KBCollection) error {
	collection.Name = strings.TrimSpace(collection.Name)
	if collection.Name == "" || len(collection.Name) > 128 {
		return fmt.Errorf("invalid collection name")
	}
	if err := s.validateParent(ctx, collection.TenantID, collection.ID, collection.ParentID); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, collection); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrKBCollectionNotFound
		}
		return err
	}
	s.auditCollection(ctx, collection, types.AuditActionKBCollectionUpdated)
	return nil
}

func (s *kbCollectionService) Delete(ctx context.Context, tenantID uint64, id string) error {
	collection, err := s.repo.Get(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if collection == nil {
		return ErrKBCollectionNotFound
	}
	count, err := s.repo.CountChildrenAndBindings(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrKBCollectionNotEmpty
	}
	if err := s.repo.Delete(ctx, tenantID, id); err != nil {
		return err
	}
	s.auditCollection(ctx, collection, types.AuditActionKBCollectionDeleted)
	return nil
}

func (s *kbCollectionService) SetKnowledgeBaseCollection(ctx context.Context, binding *types.KBCollectionBinding) error {
	kb, err := s.kbs.GetKnowledgeBaseByIDAndTenant(ctx, binding.KBID, binding.TenantID)
	if err != nil || kb == nil {
		if errors.Is(err, repository.ErrKnowledgeBaseNotFound) || kb == nil {
			return repository.ErrKnowledgeBaseNotFound
		}
		return err
	}
	if binding.CollectionID == "" {
		if err := s.repo.DeleteBinding(ctx, binding.TenantID, binding.KBID); err != nil {
			return err
		}
		s.auditBinding(ctx, binding, "")
		return nil
	}
	collection, err := s.repo.Get(ctx, binding.TenantID, binding.CollectionID)
	if err != nil {
		return err
	}
	if collection == nil {
		return ErrKBCollectionNotFound
	}
	if err := s.repo.SetBinding(ctx, binding); err != nil {
		return err
	}
	s.auditBinding(ctx, binding, collection.ID)
	return nil
}

func (s *kbCollectionService) validateParent(ctx context.Context, tenantID uint64, id string, parentID *string) error {
	if parentID == nil || *parentID == "" {
		return nil
	}
	if *parentID == id {
		return ErrKBCollectionCycle
	}
	currentID := *parentID
	for depth := 0; depth < 256; depth++ {
		parent, err := s.repo.Get(ctx, tenantID, currentID)
		if err != nil {
			return err
		}
		if parent == nil {
			return ErrKBCollectionNotFound
		}
		if parent.ParentID == nil || *parent.ParentID == "" {
			return nil
		}
		if *parent.ParentID == id {
			return ErrKBCollectionCycle
		}
		currentID = *parent.ParentID
	}
	return ErrKBCollectionCycle
}

func (s *kbCollectionService) auditCollection(ctx context.Context, collection *types.KBCollection, action types.AuditAction) {
	if s.audit == nil {
		return
	}
	actor, _ := types.UserIDFromContext(ctx)
	_ = s.audit.Log(ctx, &types.AuditLog{TenantID: collection.TenantID, ActorUserID: actor,
		Action: action, ScopeType: "knowledge_base_collection", ScopeID: collection.ID,
		TargetType: "knowledge_base_collection", TargetID: collection.ID, Outcome: types.AuditOutcomeSuccess})
}

func (s *kbCollectionService) auditBinding(ctx context.Context, binding *types.KBCollectionBinding, collectionID string) {
	if s.audit == nil || binding == nil {
		return
	}
	actor, _ := types.UserIDFromContext(ctx)
	detailJSON, _ := json.Marshal(map[string]string{"collection_id": collectionID})
	_ = s.audit.Log(ctx, &types.AuditLog{TenantID: binding.TenantID, ActorUserID: actor,
		Action: types.AuditActionKBCollectionBound, ScopeType: "knowledge_base", ScopeID: binding.KBID,
		TargetType: "knowledge_base_collection", TargetID: collectionID,
		Outcome: types.AuditOutcomeSuccess, Details: types.JSON(detailJSON)})
}
