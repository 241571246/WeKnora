package service

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/Tencent/WeKnora/internal/config"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	secutils "github.com/Tencent/WeKnora/internal/utils"
	"github.com/stretchr/testify/require"
)

type fakeDingTalkIdentityRepo struct {
	identity *types.DingTalkExternalIdentity
	saved    *types.DingTalkExternalIdentity
}

func (r *fakeDingTalkIdentityRepo) Find(context.Context, string, string, string) (*types.DingTalkExternalIdentity, error) {
	return r.identity, nil
}
func (r *fakeDingTalkIdentityRepo) Upsert(_ context.Context, identity *types.DingTalkExternalIdentity) error {
	copy := *identity
	r.saved = &copy
	return nil
}

type fakeDingTalkClient struct {
	profile *types.DingTalkIdentityProfile
}

func (c *fakeDingTalkClient) BrowserIdentity(context.Context, string) (*types.DingTalkIdentityProfile, error) {
	return c.profile, nil
}
func (c *fakeDingTalkClient) InsideIdentity(context.Context, string) (*types.DingTalkIdentityProfile, error) {
	return c.profile, nil
}

type fakeDingTalkUserService struct {
	interfaces.UserService
	user            *types.User
	registered      bool
	registerRequest *types.RegisterRequest
	updatedTenantID uint64
	updatedCount    int
	usersByEmail    map[string]*types.User
	usersByUsername map[string]*types.User
}

func (s *fakeDingTalkUserService) GetUserByID(context.Context, string) (*types.User, error) {
	return s.user, nil
}
func (s *fakeDingTalkUserService) GetUserByEmail(_ context.Context, email string) (*types.User, error) {
	if user := s.usersByEmail[email]; user != nil {
		return user, nil
	}
	if s.user != nil && s.user.Email == email {
		return s.user, nil
	}
	return nil, errors.New("user not found")
}
func (s *fakeDingTalkUserService) GetUserByUsername(_ context.Context, username string) (*types.User, error) {
	if user := s.usersByUsername[username]; user != nil {
		return user, nil
	}
	if s.user != nil && s.user.Username == username {
		return s.user, nil
	}
	return nil, errors.New("user not found")
}
func (s *fakeDingTalkUserService) Register(_ context.Context, request *types.RegisterRequest) (*types.User, error) {
	s.registered = true
	copy := *request
	s.registerRequest = &copy
	tenantID := uint64(0)
	if request.TenantProvisioning == types.TenantProvisioningCreatePersonal {
		tenantID = 7
	}
	s.user = &types.User{ID: "new-user", Username: request.Username, Email: request.Email, TenantID: tenantID, IsActive: true}
	return s.user, nil
}
func (s *fakeDingTalkUserService) UpdateUser(_ context.Context, user *types.User) error {
	s.updatedTenantID = user.TenantID
	s.updatedCount++
	return nil
}
func (s *fakeDingTalkUserService) GenerateTokens(context.Context, *types.User) (string, string, error) {
	return "weknora-access", "weknora-refresh", nil
}
func (s *fakeDingTalkUserService) BuildLoginMemberships(context.Context, *types.User, *types.Tenant) []types.Membership {
	return []types.Membership{{TenantID: 7, Role: types.TenantRoleOwner}}
}

type fakeDingTalkTenantService struct{ interfaces.TenantService }

func (s *fakeDingTalkTenantService) GetTenantByID(_ context.Context, tenantID uint64) (*types.Tenant, error) {
	return &types.Tenant{ID: tenantID, Name: "Local workspace", Status: "active"}, nil
}

type fakeDingTalkMemberService struct {
	interfaces.TenantMemberService
	membership *types.TenantMember
	added      *types.TenantMember
}

func (s *fakeDingTalkMemberService) GetMembership(context.Context, string, uint64) (*types.TenantMember, error) {
	return s.membership, nil
}
func (s *fakeDingTalkMemberService) AddMember(
	_ context.Context, userID string, tenantID uint64, role types.TenantRole, _ *string,
) (*types.TenantMember, error) {
	s.added = &types.TenantMember{UserID: userID, TenantID: tenantID, Role: role}
	s.membership = s.added
	return s.added, nil
}

func dingTalkTestConfig() *config.Config {
	return &config.Config{
		Auth: &config.AuthConfig{DefaultTenantMode: config.AuthDefaultTenantModeCreatePersonal},
		DingTalkAuth: &config.DingTalkAuthConfig{
			Enable: true, CorpID: "corp-weknora", ClientID: "weknora-client",
			ClientSecret: "server-only-secret", RedirectURI: "https://brain.example.com/api/v1/auth/dingtalk/callback",
		},
	}
}

func TestDingTalkAuthorizationURLUsesIndependentPublicConfig(t *testing.T) {
	service := NewDingTalkAuthService(
		dingTalkTestConfig(), &fakeDingTalkIdentityRepo{}, &fakeDingTalkUserService{},
		&fakeDingTalkTenantService{}, &fakeDingTalkMemberService{}, &fakeDingTalkClient{},
	)
	result, err := service.CreateAuthorizationURL(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, result.Nonce)
	parsed, err := url.Parse(result.AuthorizationURL)
	require.NoError(t, err)
	require.Equal(t, "weknora-client", parsed.Query().Get("client_id"))
	require.Equal(t, "https://brain.example.com/api/v1/auth/dingtalk/callback", parsed.Query().Get("redirect_uri"))
	require.NotContains(t, result.AuthorizationURL, "server-only-secret")
	state, err := secutils.VerifyOIDCState(parsed.Query().Get("state"))
	require.NoError(t, err)
	require.Equal(t, result.Nonce, state.Nonce)
}

func TestDingTalkMappedLoginIssuesOnlyWeKnoraTokens(t *testing.T) {
	userID := "local-user"
	repo := &fakeDingTalkIdentityRepo{identity: &types.DingTalkExternalIdentity{
		ID: "identity", Provider: types.DingTalkAuthProvider, CorpID: "corp-weknora",
		ExternalUserID: "ding-user", UserID: &userID, BindingMethod: "admin",
	}}
	users := &fakeDingTalkUserService{user: &types.User{ID: userID, TenantID: 7, IsActive: true}}
	service := NewDingTalkAuthService(dingTalkTestConfig(), repo, users, &fakeDingTalkTenantService{}, &fakeDingTalkMemberService{}, &fakeDingTalkClient{
		profile: &types.DingTalkIdentityProfile{ExternalUserID: "ding-user", DisplayName: "Ding User", Active: true},
	})
	result, err := service.LoginWithBrowserCode(context.Background(), "one-time-code")
	require.NoError(t, err)
	require.Equal(t, "weknora-access", result.Token)
	require.Equal(t, "weknora-refresh", result.RefreshToken)
	require.False(t, users.registered)
	require.NotNil(t, repo.saved)
	require.Equal(t, userID, *repo.saved.UserID)
}

func TestDingTalkUnmappedLoginIsPendingWhenJITDisabled(t *testing.T) {
	repo := &fakeDingTalkIdentityRepo{}
	service := NewDingTalkAuthService(dingTalkTestConfig(), repo, &fakeDingTalkUserService{}, &fakeDingTalkTenantService{}, &fakeDingTalkMemberService{}, &fakeDingTalkClient{
		profile: &types.DingTalkIdentityProfile{ExternalUserID: "new-ding-user", DisplayName: "New User", Active: true},
	})
	_, err := service.LoginWithInsideCode(context.Background(), "one-time-code")
	require.ErrorContains(t, err, "not bound")
	require.NotNil(t, repo.saved)
	require.Equal(t, "pending", repo.saved.Status)
	require.Nil(t, repo.saved.UserID)
}

func TestDingTalkJITCreatesOrdinaryLocalUser(t *testing.T) {
	cfg := dingTalkTestConfig()
	cfg.DingTalkAuth.JITProvisioning = true
	repo := &fakeDingTalkIdentityRepo{}
	users := &fakeDingTalkUserService{}
	service := NewDingTalkAuthService(cfg, repo, users, &fakeDingTalkTenantService{}, &fakeDingTalkMemberService{}, &fakeDingTalkClient{
		profile: &types.DingTalkIdentityProfile{
			ExternalUserID: "new-ding-user", DisplayName: "New User", Email: "new.user@example.com", Active: true,
		},
	})
	result, err := service.LoginWithInsideCode(context.Background(), "one-time-code")
	require.NoError(t, err)
	require.True(t, users.registered)
	require.False(t, result.User.IsSystemAdmin)
	require.False(t, result.User.CanAccessAllTenants)
	require.Equal(t, "New User", result.User.Username)
	require.Equal(t, "new.user@example.com", result.User.Email)
	require.Contains(t, users.registerRequest.Email, "@dingtalk.local")
	require.Equal(t, "jit", repo.saved.BindingMethod)
}

func TestDingTalkExistingJITUserReplacesOnlyPlaceholderProfile(t *testing.T) {
	cfg := dingTalkTestConfig()
	cfg.DingTalkAuth.JITProvisioning = true
	username, email, _ := dingTalkPlaceholderAccount(cfg.DingTalkAuth.CorpID, "existing-ding-user")
	userID := "existing-user"
	user := &types.User{ID: userID, Username: username, Email: email, TenantID: 7, IsActive: true}
	repo := &fakeDingTalkIdentityRepo{identity: &types.DingTalkExternalIdentity{
		ID: "identity", Provider: types.DingTalkAuthProvider, CorpID: cfg.DingTalkAuth.CorpID,
		ExternalUserID: "existing-ding-user", UserID: &userID, BindingMethod: "jit",
	}}
	users := &fakeDingTalkUserService{user: user}
	service := NewDingTalkAuthService(cfg, repo, users, &fakeDingTalkTenantService{}, &fakeDingTalkMemberService{}, &fakeDingTalkClient{
		profile: &types.DingTalkIdentityProfile{
			ExternalUserID: "existing-ding-user", DisplayName: "雷蛟", Email: "rocklei@vonechina.com", Active: true,
		},
	})

	result, err := service.LoginWithBrowserCode(context.Background(), "one-time-code")
	require.NoError(t, err)
	require.Equal(t, "雷蛟", result.User.Username)
	require.Equal(t, "rocklei@vonechina.com", result.User.Email)
	require.Equal(t, 1, users.updatedCount)
}

func TestDingTalkExistingJITUserPreservesManagedProfile(t *testing.T) {
	cfg := dingTalkTestConfig()
	cfg.DingTalkAuth.JITProvisioning = true
	userID := "managed-user"
	user := &types.User{ID: userID, Username: "管理员设置姓名", Email: "managed@vonechina.com", TenantID: 7, IsActive: true}
	repo := &fakeDingTalkIdentityRepo{identity: &types.DingTalkExternalIdentity{
		ID: "identity", Provider: types.DingTalkAuthProvider, CorpID: cfg.DingTalkAuth.CorpID,
		ExternalUserID: "managed-ding-user", UserID: &userID, BindingMethod: "jit",
	}}
	users := &fakeDingTalkUserService{user: user}
	service := NewDingTalkAuthService(cfg, repo, users, &fakeDingTalkTenantService{}, &fakeDingTalkMemberService{}, &fakeDingTalkClient{
		profile: &types.DingTalkIdentityProfile{
			ExternalUserID: "managed-ding-user", DisplayName: "钉钉姓名", Email: "dingtalk@vonechina.com", Active: true,
		},
	})

	result, err := service.LoginWithBrowserCode(context.Background(), "one-time-code")
	require.NoError(t, err)
	require.Equal(t, "管理员设置姓名", result.User.Username)
	require.Equal(t, "managed@vonechina.com", result.User.Email)
	require.Zero(t, users.updatedCount)
}

func TestDingTalkExistingJITUserKeepsPlaceholderEmailWhenRealEmailConflicts(t *testing.T) {
	cfg := dingTalkTestConfig()
	cfg.DingTalkAuth.JITProvisioning = true
	username, email, _ := dingTalkPlaceholderAccount(cfg.DingTalkAuth.CorpID, "conflicting-ding-user")
	userID := "conflicting-user"
	user := &types.User{ID: userID, Username: username, Email: email, TenantID: 7, IsActive: true}
	repo := &fakeDingTalkIdentityRepo{identity: &types.DingTalkExternalIdentity{
		ID: "identity", Provider: types.DingTalkAuthProvider, CorpID: cfg.DingTalkAuth.CorpID,
		ExternalUserID: "conflicting-ding-user", UserID: &userID, BindingMethod: "jit",
	}}
	users := &fakeDingTalkUserService{
		user: user,
		usersByEmail: map[string]*types.User{
			"shared@vonechina.com": {ID: "other-user", Email: "shared@vonechina.com"},
		},
	}
	service := NewDingTalkAuthService(cfg, repo, users, &fakeDingTalkTenantService{}, &fakeDingTalkMemberService{}, &fakeDingTalkClient{
		profile: &types.DingTalkIdentityProfile{
			ExternalUserID: "conflicting-ding-user", DisplayName: "同名用户", Email: "shared@vonechina.com", Active: true,
		},
	})

	result, err := service.LoginWithBrowserCode(context.Background(), "one-time-code")
	require.NoError(t, err)
	require.Equal(t, "同名用户", result.User.Username)
	require.Equal(t, email, result.User.Email)
	require.Equal(t, 1, users.updatedCount)
}

func TestDingTalkJITAddsStableSuffixWhenDisplayNameAlreadyExists(t *testing.T) {
	cfg := dingTalkTestConfig()
	cfg.DingTalkAuth.JITProvisioning = true
	_, _, suffix := dingTalkPlaceholderAccount(cfg.DingTalkAuth.CorpID, "same-name-ding-user")
	users := &fakeDingTalkUserService{
		usersByUsername: map[string]*types.User{
			"同名用户": {ID: "other-user", Username: "同名用户"},
		},
	}
	service := NewDingTalkAuthService(cfg, &fakeDingTalkIdentityRepo{}, users, &fakeDingTalkTenantService{}, &fakeDingTalkMemberService{}, &fakeDingTalkClient{
		profile: &types.DingTalkIdentityProfile{ExternalUserID: "same-name-ding-user", DisplayName: "同名用户", Active: true},
	})

	result, err := service.LoginWithInsideCode(context.Background(), "one-time-code")
	require.NoError(t, err)
	require.Equal(t, "同名用户_"+suffix[:6], result.User.Username)
}

func TestDingTalkJITJoinsConfiguredDefaultTenant(t *testing.T) {
	cfg := dingTalkTestConfig()
	cfg.DingTalkAuth.JITProvisioning = true
	cfg.DingTalkAuth.AutoJoinTenant = true
	cfg.DingTalkAuth.DefaultTenantID = 10000
	cfg.DingTalkAuth.DefaultTenantRole = "viewer"
	repo := &fakeDingTalkIdentityRepo{}
	users := &fakeDingTalkUserService{}
	members := &fakeDingTalkMemberService{}
	service := NewDingTalkAuthService(cfg, repo, users, &fakeDingTalkTenantService{}, members, &fakeDingTalkClient{
		profile: &types.DingTalkIdentityProfile{ExternalUserID: "default-tenant-user", DisplayName: "New User", Active: true},
	})

	result, err := service.LoginWithInsideCode(context.Background(), "one-time-code")
	require.NoError(t, err)
	require.NotNil(t, users.registerRequest)
	require.Equal(t, types.TenantProvisioningTenantless, users.registerRequest.TenantProvisioning)
	require.NotNil(t, members.added)
	require.Equal(t, uint64(10000), members.added.TenantID)
	require.Equal(t, types.TenantRoleViewer, members.added.Role)
	require.Equal(t, uint64(10000), users.updatedTenantID)
	require.Equal(t, uint64(10000), result.User.TenantID)
	require.Equal(t, uint64(10000), result.Tenant.ID)
}
