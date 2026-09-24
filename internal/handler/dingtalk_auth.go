package handler

import (
	"crypto/hmac"
	"net/http"
	"strings"

	"github.com/Tencent/WeKnora/internal/config"
	apperrors "github.com/Tencent/WeKnora/internal/errors"
	"github.com/Tencent/WeKnora/internal/handler/dto"
	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	secutils "github.com/Tencent/WeKnora/internal/utils"
	"github.com/gin-gonic/gin"
)

const dingTalkNonceCookieName = "weknora_dingtalk_nonce"
const dingTalkNonceCookieMaxAge = 600

type DingTalkAuthHandler struct {
	config  *config.Config
	service interfaces.DingTalkAuthService
}

func NewDingTalkAuthHandler(cfg *config.Config, service interfaces.DingTalkAuthService) *DingTalkAuthHandler {
	return &DingTalkAuthHandler{config: cfg, service: service}
}

func (h *DingTalkAuthHandler) Config(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, h.service.PublicConfig())
}

func (h *DingTalkAuthHandler) Start(c *gin.Context) {
	result, err := h.service.CreateAuthorizationURL(c.Request.Context())
	if err != nil {
		c.Error(apperrors.NewForbiddenError("DingTalk authorization unavailable").WithDetails(err.Error()))
		return
	}
	h.setNonceCookie(c, result.Nonce)
	c.Redirect(http.StatusFound, result.AuthorizationURL)
}

func (h *DingTalkAuthHandler) Callback(c *gin.Context) {
	ctx := c.Request.Context()
	state := strings.TrimSpace(c.Query("state"))
	payload, err := secutils.VerifyOIDCState(state)
	if err != nil || !h.validNonce(c, payload) || !h.validCallbackURI(payload.RedirectURI) {
		h.clearNonceCookie(c)
		c.Redirect(http.StatusFound, "/#dingtalk_error=invalid_state")
		return
	}
	h.clearNonceCookie(c)

	code := strings.TrimSpace(c.Query("authCode"))
	if code == "" {
		code = strings.TrimSpace(c.Query("code"))
	}
	if code == "" {
		c.Redirect(http.StatusFound, "/#dingtalk_error=missing_code")
		return
	}
	response, err := h.service.LoginWithBrowserCode(ctx, code)
	if err != nil {
		logger.Errorf(ctx, "DingTalk browser login failed: %v", err)
		c.Redirect(http.StatusFound, "/#dingtalk_error=login_failed&dingtalk_error_description="+urlQueryEscape(dingTalkPublicError(err)))
		return
	}
	encoded, err := encodeOIDCCallbackPayload(response)
	if err != nil {
		c.Redirect(http.StatusFound, "/#dingtalk_error=payload_encode_failed")
		return
	}
	c.Redirect(http.StatusFound, "/#dingtalk_result="+urlQueryEscape(encoded))
}

func (h *DingTalkAuthHandler) Inside(c *gin.Context) {
	var request struct {
		AuthCode string `json:"auth_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(apperrors.NewValidationError("auth_code is required"))
		return
	}
	response, err := h.service.LoginWithInsideCode(c.Request.Context(), request.AuthCode)
	if err != nil {
		logger.Errorf(c.Request.Context(), "DingTalk inside login failed: %v", err)
		c.Error(apperrors.NewUnauthorizedError("DingTalk login failed").WithDetails(dingTalkPublicError(err)))
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, dto.NewAuthOIDCCallbackResponse(response))
}

func dingTalkPublicError(err error) string {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "not bound"):
		return "DingTalk account is not bound to a WeKnora user"
	case strings.Contains(message, "inactive") || strings.Contains(message, "disabled"):
		return "DingTalk or WeKnora account is inactive"
	default:
		return "DingTalk login failed; contact the administrator"
	}
}

func (h *DingTalkAuthHandler) setNonceCookie(c *gin.Context, nonce string) {
	secure := c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(dingTalkNonceCookieName, nonce, dingTalkNonceCookieMaxAge, "/api/v1/auth/dingtalk", "", secure, true)
}

func (h *DingTalkAuthHandler) clearNonceCookie(c *gin.Context) {
	secure := c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(dingTalkNonceCookieName, "", -1, "/api/v1/auth/dingtalk", "", secure, true)
}

func (h *DingTalkAuthHandler) validNonce(c *gin.Context, payload *secutils.OIDCStatePayload) bool {
	if payload == nil || strings.TrimSpace(payload.Nonce) == "" {
		return false
	}
	cookie, err := c.Cookie(dingTalkNonceCookieName)
	return err == nil && hmac.Equal([]byte(cookie), []byte(payload.Nonce))
}

func (h *DingTalkAuthHandler) validCallbackURI(uri string) bool {
	return h.config != nil && h.config.DingTalkAuth != nil &&
		strings.TrimSpace(uri) != "" && strings.TrimSpace(uri) == strings.TrimSpace(h.config.DingTalkAuth.RedirectURI)
}
