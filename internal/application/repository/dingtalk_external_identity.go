package repository

import (
	"context"
	"errors"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type dingTalkExternalIdentityRepository struct {
	db *gorm.DB
}

func NewDingTalkExternalIdentityRepository(db *gorm.DB) interfaces.DingTalkExternalIdentityRepository {
	return &dingTalkExternalIdentityRepository{db: db}
}

func (r *dingTalkExternalIdentityRepository) Find(
	ctx context.Context, corpID, externalUserID, unionID string,
) (*types.DingTalkExternalIdentity, error) {
	query := r.db.WithContext(ctx).
		Where("provider = ? AND corp_id = ?", types.DingTalkAuthProvider, corpID).
		Where("external_user_id = ?", externalUserID)
	if unionID != "" {
		query = r.db.WithContext(ctx).
			Where("provider = ? AND corp_id = ?", types.DingTalkAuthProvider, corpID).
			Where("external_user_id = ? OR union_id = ?", externalUserID, unionID)
	}
	var identity types.DingTalkExternalIdentity
	if err := query.First(&identity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &identity, nil
}

func (r *dingTalkExternalIdentityRepository) Upsert(
	ctx context.Context, identity *types.DingTalkExternalIdentity,
) error {
	if identity == nil {
		return errors.New("dingtalk identity is required")
	}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "provider"}, {Name: "corp_id"}, {Name: "external_user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"union_id", "user_id", "display_name", "email", "mobile", "status",
			"binding_method", "raw_profile", "last_authenticated_at", "updated_at",
		}),
	}).Create(identity).Error
	if err != nil {
		return err
	}
	stored, err := r.Find(ctx, identity.CorpID, identity.ExternalUserID, identity.UnionID)
	if err == nil && stored != nil {
		*identity = *stored
	}
	return err
}
