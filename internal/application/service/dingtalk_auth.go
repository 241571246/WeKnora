package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/WeKnora/internal/config"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	secutils "github.com/Tencent/WeKnora/internal/utils"
	"github.com/google/uuid"
)

type dingTalkAuthService struct {
	cfg      *config.Config
	identity interfaces.DingTalkExternalIdentityRepository
	users    interfaces.UserService
	tenants  interfaces.TenantService
	members  interfaces.TenantMemberService
	client   interfaces.DingTalkClient
}

func NewDingTalkAuthService(
	cfg *config.Config,
	identity interfaces.DingTalkExternalIdentityRepository,
	users interfaces.UserService,
	tenants interfaces.TenantService,
	members interfaces.TenantMemberService,
	client interfaces.DingTalkClient,
) interfaces.DingTalkAuthService {
	return &dingTalkAuthService{
		cfg: cfg, identity: identity, users: users, tenants: tenants, members: members, client: client,
	}
}

func (s *dingTalkAuthService) PublicConfig() *types.DingTalkAuthConfigResponse {
	cfg, enabled := s.enabledConfig()
	response := &types.DingTalkAuthConfigResponse{Success: true, Enabled: enabled, InsideLogin: enabled}
	if enabled {
		response.CorpID = cfg.CorpID
		response.ClientID = cfg.ClientID
	}
	return response
}

func (s *dingTalkAuthService) CreateAuthorizationURL(ctx context.Context) (*types.DingTalkAuthURLResponse, error) {
	cfg, enabled := s.enabledConfig()
	if !enabled {
		return nil, errors.New("DingTalk login is disabled")
	}
	nonceBytes := make([]byte, 24)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, fmt.Errorf("generate DingTalk state nonce: %w", err)
	}
	nonce := base64.RawURLEncoding.EncodeToString(nonceBytes)
	state, err := secutils.SignOIDCState(&secutils.OIDCStatePayload{
		Nonce: nonce, RedirectURI: strings.TrimSpace(cfg.RedirectURI),
	})
	if err != nil {
		return nil, fmt.Errorf("sign DingTalk state: %w", err)
	}
	authorizationURL, err := DingTalkAuthorizationURL(cfg, state)
	if err != nil {
		return nil, err
	}
	return &types.DingTalkAuthURLResponse{Success: true, AuthorizationURL: authorizationURL, Nonce: nonce}, nil
}

func (s *dingTalkAuthService) LoginWithBrowserCode(ctx context.Context, code string) (*types.OIDCCallbackResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, errors.New("DingTalk authorization code is required")
	}
	profile, err := s.client.BrowserIdentity(ctx, code)
	if err != nil {
		return nil, err
	}
	return s.login(ctx, profile)
}

func (s *dingTalkAuthService) LoginWithInsideCode(ctx context.Context, authCode string) (*types.OIDCCallbackResponse, error) {
	if strings.TrimSpace(authCode) == "" {
		return nil, errors.New("DingTalk inside auth code is required")
	}
	profile, err := s.client.InsideIdentity(ctx, authCode)
	if err != nil {
		return nil, err
	}
	return s.login(ctx, profile)
}

func (s *dingTalkAuthService) login(ctx context.Context, profile *types.DingTalkIdentityProfile) (*types.OIDCCallbackResponse, error) {
	cfg, enabled := s.enabledConfig()
	if !enabled {
		return nil, errors.New("DingTalk login is disabled")
	}
	if profile == nil || strings.TrimSpace(profile.ExternalUserID) == "" {
		return nil, errors.New("DingTalk did not return a stable user id")
	}
	if !profile.Active {
		return nil, errors.New("DingTalk account is inactive")
	}

	identity, err := s.identity.Find(ctx, cfg.CorpID, profile.ExternalUserID, profile.UnionID)
	if err != nil {
		return nil, fmt.Errorf("load DingTalk identity: %w", err)
	}
	user, bindingMethod, isNewUser, err := s.resolveUser(ctx, cfg, identity, profile)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, errors.New("local account is inactive")
	}
	if bindingMethod == "jit" {
		if err := s.syncJITUserProfile(ctx, cfg.CorpID, user, profile); err != nil {
			return nil, err
		}
		if err := s.ensureJITDefaultTenant(ctx, cfg, user); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	raw, _ := json.Marshal(profile.Raw)
	if identity == nil {
		identity = &types.DingTalkExternalIdentity{
			ID: uuid.NewString(), Provider: types.DingTalkAuthProvider, CorpID: cfg.CorpID,
			ExternalUserID: profile.ExternalUserID, CreatedAt: now,
		}
	}
	identity.UnionID = profile.UnionID
	identity.UserID = &user.ID
	identity.DisplayName = firstNonEmpty(profile.DisplayName, user.Username)
	identity.Email = profile.Email
	identity.Mobile = profile.Mobile
	identity.Status = "active"
	identity.BindingMethod = bindingMethod
	identity.RawProfile = types.JSON(raw)
	identity.LastAuthenticatedAt = &now
	identity.UpdatedAt = now
	if err := s.identity.Upsert(ctx, identity); err != nil {
		return nil, fmt.Errorf("save DingTalk identity: %w", err)
	}

	accessToken, refreshToken, err := s.users.GenerateTokens(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("generate WeKnora tokens: %w", err)
	}
	var tenant *types.Tenant
	if user.TenantID > 0 && s.tenants != nil {
		tenant, _ = s.tenants.GetTenantByID(ctx, user.TenantID)
	}
	return &types.OIDCCallbackResponse{
		Success: true, Message: "登录成功", User: user, Tenant: tenant,
		Memberships: s.users.BuildLoginMemberships(ctx, user, tenant),
		Token:       accessToken, RefreshToken: refreshToken, IsNewUser: isNewUser,
	}, nil
}

func (s *dingTalkAuthService) resolveUser(
	ctx context.Context,
	cfg *config.DingTalkAuthConfig,
	identity *types.DingTalkExternalIdentity,
	profile *types.DingTalkIdentityProfile,
) (*types.User, string, bool, error) {
	if identity != nil && identity.UserID != nil && strings.TrimSpace(*identity.UserID) != "" {
		user, err := s.users.GetUserByID(ctx, *identity.UserID)
		return user, firstNonEmpty(identity.BindingMethod, "existing"), false, err
	}
	if cfg.EmailAutoBind && strings.TrimSpace(profile.Email) != "" {
		user, err := s.users.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(profile.Email)))
		if err == nil && user != nil {
			return user, "email", false, nil
		}
		if err != nil && !isNotFoundError(err) {
			return nil, "", false, fmt.Errorf("query local user by email: %w", err)
		}
	}
	if !cfg.JITProvisioning {
		if err := s.savePendingIdentity(ctx, cfg.CorpID, identity, profile); err != nil {
			return nil, "", false, err
		}
		return nil, "", false, errors.New("DingTalk account is not bound to a WeKnora user")
	}

	placeholderUsername, placeholderEmail, suffix := dingTalkPlaceholderAccount(cfg.CorpID, profile.ExternalUserID)
	username, err := s.availableDingTalkUsername(ctx, profile.DisplayName, placeholderUsername, suffix, "")
	if err != nil {
		return nil, "", false, err
	}
	provisioning := s.defaultTenantMode()
	if cfg.AutoJoinTenant {
		provisioning = types.TenantProvisioningTenantless
	}
	user, err := s.users.Register(ctx, &types.RegisterRequest{
		Username: username, Email: placeholderEmail,
		Password: uuid.NewString() + "Aa1!", TenantProvisioning: provisioning,
	})
	if errors.Is(err, ErrUserUsernameExists) && username != placeholderUsername {
		username = dingTalkSuffixedUsername(profile.DisplayName, suffix, placeholderUsername)
		user, err = s.users.Register(ctx, &types.RegisterRequest{
			Username: username, Email: placeholderEmail,
			Password: uuid.NewString() + "Aa1!", TenantProvisioning: provisioning,
		})
	}
	if err != nil {
		// A concurrent first login may have created the deterministic account.
		user, err = s.users.GetUserByEmail(ctx, placeholderEmail)
	}
	if err != nil || user == nil {
		return nil, "", false, fmt.Errorf("provision DingTalk user: %w", err)
	}
	federatedOnly := true
	user.Preferences.OidcOnlyLogin = &federatedOnly
	if err := s.users.UpdateUser(ctx, user); err != nil {
		return nil, "", false, fmt.Errorf("mark DingTalk user as federated-only: %w", err)
	}
	return user, "jit", true, nil
}

func (s *dingTalkAuthService) syncJITUserProfile(
	ctx context.Context,
	corpID string,
	user *types.User,
	profile *types.DingTalkIdentityProfile,
) error {
	if user == nil || profile == nil {
		return nil
	}
	placeholderUsername, placeholderEmail, suffix := dingTalkPlaceholderAccount(corpID, profile.ExternalUserID)
	changed := false
	if user.Username == placeholderUsername && strings.TrimSpace(profile.DisplayName) != "" {
		username, err := s.availableDingTalkUsername(ctx, profile.DisplayName, placeholderUsername, suffix, user.ID)
		if err != nil {
			return err
		}
		if username != user.Username {
			user.Username = username
			changed = true
		}
	}

	profileEmail := strings.ToLower(strings.TrimSpace(profile.Email))
	if strings.EqualFold(user.Email, placeholderEmail) && profileEmail != "" {
		existing, err := s.users.GetUserByEmail(ctx, profileEmail)
		if err != nil && !isNotFoundError(err) {
			return fmt.Errorf("query DingTalk profile email: %w", err)
		}
		if existing == nil || existing.ID == user.ID {
			user.Email = profileEmail
			changed = true
		}
	}

	if !changed {
		return nil
	}
	if err := s.users.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("sync DingTalk user profile: %w", err)
	}
	return nil
}

func (s *dingTalkAuthService) availableDingTalkUsername(
	ctx context.Context,
	displayName string,
	placeholderUsername string,
	suffix string,
	currentUserID string,
) (string, error) {
	candidate := truncateDingTalkRunes(strings.TrimSpace(displayName), 50)
	if candidate == "" {
		return placeholderUsername, nil
	}
	existing, err := s.users.GetUserByUsername(ctx, candidate)
	if err != nil && !isNotFoundError(err) {
		return "", fmt.Errorf("query DingTalk display name: %w", err)
	}
	if existing == nil || existing.ID == currentUserID {
		return candidate, nil
	}
	return dingTalkSuffixedUsername(displayName, suffix, placeholderUsername), nil
}

func dingTalkPlaceholderAccount(corpID, externalUserID string) (string, string, string) {
	digest := sha256.Sum256([]byte(corpID + ":" + externalUserID))
	suffix := hex.EncodeToString(digest[:])
	return "dt_" + suffix[:20], "dt_" + suffix[:32] + "@dingtalk.local", suffix
}

func dingTalkSuffixedUsername(displayName, suffix, fallback string) string {
	base := truncateDingTalkRunes(strings.TrimSpace(displayName), 43)
	if base == "" {
		return fallback
	}
	return base + "_" + suffix[:6]
}

func truncateDingTalkRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func (s *dingTalkAuthService) ensureJITDefaultTenant(
	ctx context.Context, cfg *config.DingTalkAuthConfig, user *types.User,
) error {
	if cfg == nil || !cfg.AutoJoinTenant {
		return nil
	}
	if s.tenants == nil || s.members == nil {
		return errors.New("DingTalk default tenant services are unavailable")
	}
	tenant, err := s.tenants.GetTenantByID(ctx, cfg.DefaultTenantID)
	if err != nil {
		return fmt.Errorf("load DingTalk default tenant %d: %w", cfg.DefaultTenantID, err)
	}
	if tenant == nil {
		return fmt.Errorf("DingTalk default tenant %d not found", cfg.DefaultTenantID)
	}
	if !strings.EqualFold(strings.TrimSpace(tenant.Status), "active") {
		return fmt.Errorf("DingTalk default tenant %d is inactive", cfg.DefaultTenantID)
	}
	membership, err := s.members.GetMembership(ctx, user.ID, cfg.DefaultTenantID)
	if err != nil {
		return fmt.Errorf("load DingTalk default tenant membership: %w", err)
	}
	if membership == nil {
		role := types.TenantRole(strings.ToLower(strings.TrimSpace(cfg.DefaultTenantRole)))
		if _, err := s.members.AddMember(ctx, user.ID, cfg.DefaultTenantID, role, nil); err != nil &&
			!errors.Is(err, ErrMembershipAlreadyExists) {
			return fmt.Errorf("join DingTalk default tenant: %w", err)
		}
	}
	if user.TenantID != cfg.DefaultTenantID {
		user.TenantID = cfg.DefaultTenantID
		if err := s.users.UpdateUser(ctx, user); err != nil {
			return fmt.Errorf("set DingTalk default tenant: %w", err)
		}
	}
	return nil
}

func (s *dingTalkAuthService) savePendingIdentity(
	ctx context.Context, corpID string, identity *types.DingTalkExternalIdentity, profile *types.DingTalkIdentityProfile,
) error {
	now := time.Now()
	raw, _ := json.Marshal(profile.Raw)
	if identity == nil {
		identity = &types.DingTalkExternalIdentity{
			ID: uuid.NewString(), Provider: types.DingTalkAuthProvider, CorpID: corpID,
			ExternalUserID: profile.ExternalUserID, CreatedAt: now,
		}
	}
	identity.UnionID = profile.UnionID
	identity.DisplayName = firstNonEmpty(profile.DisplayName, "钉钉用户")
	identity.Email = profile.Email
	identity.Mobile = profile.Mobile
	identity.Status = "pending"
	identity.BindingMethod = "pending"
	identity.RawProfile = types.JSON(raw)
	identity.UpdatedAt = now
	return s.identity.Upsert(ctx, identity)
}

func (s *dingTalkAuthService) enabledConfig() (*config.DingTalkAuthConfig, bool) {
	if s.cfg == nil || s.cfg.DingTalkAuth == nil || !s.cfg.DingTalkAuth.Enable {
		return nil, false
	}
	return s.cfg.DingTalkAuth, true
}

func (s *dingTalkAuthService) defaultTenantMode() types.TenantProvisioningMode {
	if s.cfg != nil && s.cfg.Auth != nil && s.cfg.Auth.DefaultTenantMode == config.AuthDefaultTenantModeTenantless {
		return types.TenantProvisioningTenantless
	}
	return types.TenantProvisioningCreatePersonal
}

func isNotFoundError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "not found")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
