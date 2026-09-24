package interfaces

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
)

type DingTalkExternalIdentityRepository interface {
	Find(ctx context.Context, corpID, externalUserID, unionID string) (*types.DingTalkExternalIdentity, error)
	Upsert(ctx context.Context, identity *types.DingTalkExternalIdentity) error
}

type DingTalkClient interface {
	BrowserIdentity(ctx context.Context, code string) (*types.DingTalkIdentityProfile, error)
	InsideIdentity(ctx context.Context, authCode string) (*types.DingTalkIdentityProfile, error)
}

type DingTalkAuthService interface {
	PublicConfig() *types.DingTalkAuthConfigResponse
	CreateAuthorizationURL(ctx context.Context) (*types.DingTalkAuthURLResponse, error)
	LoginWithBrowserCode(ctx context.Context, code string) (*types.OIDCCallbackResponse, error)
	LoginWithInsideCode(ctx context.Context, authCode string) (*types.OIDCCallbackResponse, error)
}
