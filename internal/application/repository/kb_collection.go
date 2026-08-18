package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type kbCollectionRepository struct{ db *gorm.DB }

func NewKBCollectionRepository(db *gorm.DB) interfaces.KBCollectionRepository {
	return &kbCollectionRepository{db: db}
}

func (r *kbCollectionRepository) Create(ctx context.Context, collection *types.KBCollection) error {
	return r.db.WithContext(ctx).Create(collection).Error
}

func (r *kbCollectionRepository) Get(ctx context.Context, tenantID uint64, id string) (*types.KBCollection, error) {
	var collection types.KBCollection
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&collection).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &collection, err
}

func (r *kbCollectionRepository) List(ctx context.Context, tenantID uint64) ([]*types.KBCollection, error) {
	var rows []*types.KBCollection
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).
		Order("sort_order ASC, name ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *kbCollectionRepository) Update(ctx context.Context, collection *types.KBCollection) error {
	res := r.db.WithContext(ctx).Model(&types.KBCollection{}).
		Where("tenant_id = ? AND id = ?", collection.TenantID, collection.ID).
		Updates(map[string]any{"parent_id": collection.ParentID, "name": collection.Name,
			"sort_order": collection.SortOrder, "updated_at": time.Now().UTC()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *kbCollectionRepository) Delete(ctx context.Context, tenantID uint64, id string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&types.KBCollection{}).Error
}

func (r *kbCollectionRepository) CountChildrenAndBindings(ctx context.Context, tenantID uint64, id string) (int64, error) {
	var children, bindings int64
	if err := r.db.WithContext(ctx).Model(&types.KBCollection{}).
		Where("tenant_id = ? AND parent_id = ?", tenantID, id).Count(&children).Error; err != nil {
		return 0, err
	}
	if err := r.db.WithContext(ctx).Model(&types.KBCollectionBinding{}).
		Where("tenant_id = ? AND collection_id = ?", tenantID, id).Count(&bindings).Error; err != nil {
		return 0, err
	}
	return children + bindings, nil
}

func (r *kbCollectionRepository) SetBinding(ctx context.Context, binding *types.KBCollectionBinding) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "kb_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"tenant_id": binding.TenantID, "collection_id": binding.CollectionID,
			"updated_by": binding.UpdatedBy, "updated_at": time.Now().UTC(),
		}),
	}).Create(binding).Error
}

func (r *kbCollectionRepository) DeleteBinding(ctx context.Context, tenantID uint64, kbID string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND kb_id = ?", tenantID, kbID).
		Delete(&types.KBCollectionBinding{}).Error
}

func (r *kbCollectionRepository) ListBindings(ctx context.Context, tenantID uint64) ([]*types.KBCollectionBinding, error) {
	var rows []*types.KBCollectionBinding
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&rows).Error
	return rows, err
}
