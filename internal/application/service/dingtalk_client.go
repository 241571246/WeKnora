package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Tencent/WeKnora/internal/config"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

const (
	dingTalkAuthorizeEndpoint  = "https://login.dingtalk.com/oauth2/auth"
	dingTalkUserTokenEndpoint  = "https://api.dingtalk.com/v1.0/oauth2/userAccessToken"
	dingTalkUserInfoEndpoint   = "https://api.dingtalk.com/v1.0/contact/users/me"
	dingTalkAppTokenEndpoint   = "https://api.dingtalk.com/v1.0/oauth2/accessToken"
	dingTalkInsideUserEndpoint = "https://oapi.dingtalk.com/topapi/v2/user/getuserinfo"
	dingTalkUserDetailEndpoint = "https://oapi.dingtalk.com/topapi/v2/user/get"
)

type dingTalkClient struct {
	cfg  *config.Config
	http *http.Client
}

func NewDingTalkClient(cfg *config.Config) interfaces.DingTalkClient {
	timeout := 30 * time.Second
	if cfg != nil && cfg.DingTalkAuth != nil && cfg.DingTalkAuth.RequestTimeout > 0 {
		timeout = cfg.DingTalkAuth.RequestTimeout
	}
	return &dingTalkClient{cfg: cfg, http: newDingTalkHTTPClient(timeout)}
}

type dialContextFunc func(context.Context, string, string) (net.Conn, error)

func newDingTalkHTTPClient(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	transport.DialContext = forceIPv4DialContext(dialer.DialContext)
	return &http.Client{Timeout: timeout, Transport: transport}
}

func forceIPv4DialContext(dial dialContextFunc) dialContextFunc {
	return func(ctx context.Context, _ string, address string) (net.Conn, error) {
		return dial(ctx, "tcp4", address)
	}
}

func (c *dingTalkClient) BrowserIdentity(ctx context.Context, code string) (*types.DingTalkIdentityProfile, error) {
	cfg, err := c.requiredConfig()
	if err != nil {
		return nil, err
	}
	var token struct {
		AccessToken string `json:"accessToken"`
	}
	if err := c.postJSON(ctx, dingTalkUserTokenEndpoint, map[string]any{
		"clientId": cfg.ClientID, "clientSecret": cfg.ClientSecret,
		"code": strings.TrimSpace(code), "grantType": "authorization_code",
	}, nil, &token); err != nil {
		return nil, fmt.Errorf("exchange DingTalk user code: %w", err)
	}
	if strings.TrimSpace(token.AccessToken) == "" {
		return nil, errors.New("DingTalk did not return a user access token")
	}
	profile := map[string]any{}
	if err := c.getJSON(ctx, dingTalkUserInfoEndpoint, map[string]string{
		"x-acs-dingtalk-access-token": token.AccessToken,
	}, &profile); err != nil {
		return nil, fmt.Errorf("load DingTalk user profile: %w", err)
	}
	return dingTalkProfile(profile), nil
}

func (c *dingTalkClient) InsideIdentity(ctx context.Context, authCode string) (*types.DingTalkIdentityProfile, error) {
	cfg, err := c.requiredConfig()
	if err != nil {
		return nil, err
	}
	var token struct {
		AccessToken string `json:"accessToken"`
	}
	if err := c.postJSON(ctx, dingTalkAppTokenEndpoint, map[string]any{
		"appKey": cfg.ClientID, "appSecret": cfg.ClientSecret,
	}, nil, &token); err != nil {
		return nil, fmt.Errorf("load DingTalk app token: %w", err)
	}
	if strings.TrimSpace(token.AccessToken) == "" {
		return nil, errors.New("DingTalk did not return an app access token")
	}
	info, err := c.oldAPI(ctx, dingTalkInsideUserEndpoint, token.AccessToken, map[string]any{
		"code": strings.TrimSpace(authCode),
	})
	if err != nil {
		return nil, fmt.Errorf("resolve DingTalk inside auth code: %w", err)
	}
	userID := stringValue(info["userid"], info["userId"])
	if userID == "" {
		return nil, errors.New("DingTalk inside auth code did not return a user id")
	}
	detail, err := c.oldAPI(ctx, dingTalkUserDetailEndpoint, token.AccessToken, map[string]any{
		"userid": userID, "language": "zh_CN",
	})
	if err != nil {
		return nil, fmt.Errorf("load DingTalk user detail: %w", err)
	}
	detail["userid"] = userID
	return dingTalkProfile(detail), nil
}

func DingTalkAuthorizationURL(cfg *config.DingTalkAuthConfig, state string) (string, error) {
	if cfg == nil || strings.TrimSpace(cfg.ClientID) == "" || strings.TrimSpace(cfg.RedirectURI) == "" {
		return "", errors.New("DingTalk auth is not configured")
	}
	endpoint, _ := url.Parse(dingTalkAuthorizeEndpoint)
	query := endpoint.Query()
	query.Set("redirect_uri", cfg.RedirectURI)
	query.Set("response_type", "code")
	query.Set("client_id", cfg.ClientID)
	query.Set("scope", "openid")
	query.Set("state", state)
	query.Set("prompt", "consent")
	endpoint.RawQuery = query.Encode()
	return endpoint.String(), nil
}

func (c *dingTalkClient) requiredConfig() (*config.DingTalkAuthConfig, error) {
	if c.cfg == nil || c.cfg.DingTalkAuth == nil || !c.cfg.DingTalkAuth.Enable {
		return nil, errors.New("DingTalk login is disabled")
	}
	if strings.TrimSpace(c.cfg.DingTalkAuth.ClientID) == "" || strings.TrimSpace(c.cfg.DingTalkAuth.ClientSecret) == "" {
		return nil, errors.New("DingTalk login is not configured")
	}
	return c.cfg.DingTalkAuth, nil
}

func (c *dingTalkClient) oldAPI(
	ctx context.Context, endpoint, accessToken string, body map[string]any,
) (map[string]any, error) {
	u, _ := url.Parse(endpoint)
	query := u.Query()
	query.Set("access_token", accessToken)
	u.RawQuery = query.Encode()
	payload := map[string]any{}
	if err := c.postJSON(ctx, u.String(), body, nil, &payload); err != nil {
		return nil, err
	}
	if code := intValue(payload["errcode"]); code != 0 {
		return nil, fmt.Errorf("DingTalk API error %d: %s", code, stringValue(payload["errmsg"]))
	}
	if result, ok := payload["result"].(map[string]any); ok {
		return result, nil
	}
	return payload, nil
}

func (c *dingTalkClient) getJSON(ctx context.Context, endpoint string, headers map[string]string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	return c.doJSON(req, headers, out)
}

func (c *dingTalkClient) postJSON(ctx context.Context, endpoint string, body any, headers map[string]string, out any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doJSON(req, headers, out)
}

func (c *dingTalkClient) doJSON(req *http.Request, headers map[string]string, out any) error {
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("DingTalk API returned HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out); err != nil {
		return fmt.Errorf("decode DingTalk API response: %w", err)
	}
	return nil
}

func dingTalkProfile(raw map[string]any) *types.DingTalkIdentityProfile {
	active := true
	if value, ok := raw["active"].(bool); ok {
		active = value
	}
	return &types.DingTalkIdentityProfile{
		ExternalUserID: stringValue(raw["userid"], raw["userId"], raw["openid"], raw["openId"], raw["unionid"], raw["unionId"]),
		UnionID:        stringValue(raw["unionid"], raw["unionId"]),
		DisplayName:    stringValue(raw["name"], raw["nick"]),
		Email:          strings.ToLower(stringValue(raw["email"])),
		Mobile:         stringValue(raw["mobile"]),
		Active:         active,
		Raw:            raw,
	}
}

func stringValue(values ...any) string {
	for _, value := range values {
		if normalized := strings.TrimSpace(fmt.Sprint(value)); value != nil && normalized != "" && normalized != "<nil>" {
			return normalized
		}
	}
	return ""
}

func intValue(value any) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case json.Number:
		parsed, _ := typed.Int64()
		return int(parsed)
	default:
		return 0
	}
}
