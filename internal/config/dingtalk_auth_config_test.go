package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyDingTalkAuthEnvOverridesDefaultTenant(t *testing.T) {
	t.Setenv("WEKNORA_DINGTALK_AUTH_ENABLE", "true")
	t.Setenv("WEKNORA_DINGTALK_AUTH_CORP_ID", "corp")
	t.Setenv("WEKNORA_DINGTALK_AUTH_CLIENT_ID", "client")
	t.Setenv("WEKNORA_DINGTALK_AUTH_CLIENT_SECRET", "secret")
	t.Setenv("WEKNORA_DINGTALK_AUTH_REDIRECT_URI", "https://brain.example.com/api/v1/auth/dingtalk/callback")
	t.Setenv("WEKNORA_DINGTALK_AUTH_JIT_PROVISIONING", "true")
	t.Setenv("WEKNORA_DINGTALK_AUTH_AUTO_JOIN_TENANT", "true")
	t.Setenv("WEKNORA_DINGTALK_AUTH_DEFAULT_TENANT_ID", "10000")
	t.Setenv("WEKNORA_DINGTALK_AUTH_DEFAULT_TENANT_ROLE", "viewer")

	cfg := &Config{}
	applyDingTalkAuthEnvOverrides(cfg)

	require.True(t, cfg.DingTalkAuth.AutoJoinTenant)
	require.Equal(t, uint64(10000), cfg.DingTalkAuth.DefaultTenantID)
	require.Equal(t, "viewer", cfg.DingTalkAuth.DefaultTenantRole)
	require.NoError(t, ValidateConfig(cfg))
}

func TestValidateDingTalkDefaultTenantRejectsMissingTenantID(t *testing.T) {
	cfg := &Config{DingTalkAuth: &DingTalkAuthConfig{
		Enable: true, CorpID: "corp", ClientID: "client", ClientSecret: "secret",
		RedirectURI:    "https://brain.example.com/api/v1/auth/dingtalk/callback",
		AutoJoinTenant: true, DefaultTenantRole: "viewer",
	}}

	err := ValidateConfig(cfg)
	require.ErrorContains(t, err, "default_tenant_id must be greater than 0")
}
