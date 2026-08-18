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

var ErrLastKBOwner = errors.New("repository: last active knowledge-base owner")

type kbMembershipRepository struct{ db *gorm.DB }

func NewKBMembershipRepository(db *gorm.DB) interfaces.KBMembershipRepository {
	return &kbMembershipRepository{db: db}
}

func (r *kbMembershipRepository) Create(
	ctx context.Context, membership *types.KBMembership, capabilities []types.KBCapability,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(membership).Error; err != nil {
			return err
		}
		return replaceKBMemberCapabilities(tx, membership, capabilities)
	})
}

func (r *kbMembershipRepository) Get(
	ctx context.Context, tenantID uint64, kbID, userID string,
) (*types.KBMembership, error) {
	var membership types.KBMembership
	err := r.db.WithContext(ctx).Preload("Capabilities").
		Where("tenant_id = ? AND kb_id = ? AND user_id = ?", tenantID, kbID, userID).
		First(&membership).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &membership, err
}

func (r *kbMembershipRepository) ListByKB(
	ctx context.Context, tenantID uint64, kbID string,
) ([]*types.KBMembership, error) {
	var memberships []*types.KBMembership
	err := r.db.WithContext(ctx).Preload("Capabilities").
		Where("tenant_id = ? AND kb_id = ?", tenantID, kbID).
		Order("created_at ASC, id ASC").Find(&memberships).Error
	return memberships, err
}

func (r *kbMembershipRepository) ListAccessibleKBIDs(
	ctx context.Context, tenantID uint64, userID string,
) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Model(&types.KBMembership{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Distinct().Pluck("kb_id", &ids).Error
	return ids, err
}

func (r *kbMembershipRepository) Update(
	ctx context.Context, membership *types.KBMembership, capabilities []types.KBCapability,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockKBOwnerSet(tx, membership.TenantID, membership.KBID); err != nil {
			return err
		}
		var current types.KBMembership
		query := tx.Where("tenant_id = ? AND kb_id = ? AND user_id = ?",
			membership.TenantID, membership.KBID, membership.UserID)
		if tx.Dialector.Name() == "postgres" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := query.First(&current).Error; err != nil {
			return err
		}
		if current.Role == types.KBMemberRoleOwner && membership.Role != types.KBMemberRoleOwner {
			if err := ensureAnotherKBOwner(tx, &current); err != nil {
				return err
			}
		}
		if err := tx.Model(&current).Updates(map[string]any{
			"role": membership.Role, "granted_by": membership.GrantedBy, "updated_at": time.Now().UTC(),
		}).Error; err != nil {
			return err
		}
		current.Role = membership.Role
		return replaceKBMemberCapabilities(tx, &current, capabilities)
	})
}

func (r *kbMembershipRepository) Delete(
	ctx context.Context, tenantID uint64, kbID, userID string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockKBOwnerSet(tx, tenantID, kbID); err != nil {
			return err
		}
		var current types.KBMembership
		query := tx.Where("tenant_id = ? AND kb_id = ? AND user_id = ?", tenantID, kbID, userID)
		if tx.Dialector.Name() == "postgres" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := query.First(&current).Error; err != nil {
			return err
		}
		if current.Role == types.KBMemberRoleOwner {
			if err := ensureAnotherKBOwner(tx, &current); err != nil {
				return err
			}
		}
		return tx.Delete(&current).Error
	})
}

// lockKBOwnerSet acquires every active owner row in deterministic order before
// a delete/demotion decision. PostgreSQL then serializes competing final-owner
// transitions instead of allowing each transaction to lock a different owner.
func lockKBOwnerSet(tx *gorm.DB, tenantID uint64, kbID string) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	var owners []types.KBMembership
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tenant_id = ? AND kb_id = ? AND role = ?", tenantID, kbID, types.KBMemberRoleOwner).
		Order("id ASC").Find(&owners).Error
}

func (r *kbMembershipRepository) CountOwners(
	ctx context.Context, tenantID uint64, kbID string,
) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.KBMembership{}).
		Where("tenant_id = ? AND kb_id = ? AND role = ?", tenantID, kbID, types.KBMemberRoleOwner).
		Count(&count).Error
	return count, err
}

func ensureAnotherKBOwner(tx *gorm.DB, current *types.KBMembership) error {
	var others []types.KBMembership
	query := tx.Where("tenant_id = ? AND kb_id = ? AND id <> ? AND role = ?",
		current.TenantID, current.KBID, current.ID, types.KBMemberRoleOwner)
	if tx.Dialector.Name() == "postgres" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Find(&others).Error; err != nil {
		return err
	}
	if len(others) == 0 {
		return ErrLastKBOwner
	}
	return nil
}

func replaceKBMemberCapabilities(
	tx *gorm.DB, membership *types.KBMembership, capabilities []types.KBCapability,
) error {
	if err := tx.Where("membership_id = ?", membership.ID).
		Delete(&types.KBMemberCapability{}).Error; err != nil {
		return err
	}
	if membership.Role != types.KBMemberRoleCustom || len(capabilities) == 0 {
		return nil
	}
	rows := make([]types.KBMemberCapability, 0, len(capabilities))
	for _, capability := range capabilities {
		rows = append(rows, types.KBMemberCapability{MembershipID: membership.ID, Capability: capability})
	}
	return tx.Create(&rows).Error
}
