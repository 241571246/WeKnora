package middleware

import (
	"context"
	"encoding/json"

	apperrors "github.com/Tencent/WeKnora/internal/errors"
	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/gin-gonic/gin"
)

const KBPolicyDecisionContextKey = "rbac.kb_policy_decision"

// RequireKBCapability applies the VONE fine-grained decision before the
// legacy KBAccess middleware performs source-tenant rewriting. It fails
// closed and therefore prevents same-workspace access from inheriting the old
// implicit "all members can read every KB" behaviour.
func RequireKBCapability(
	resolveKBID KBIDResolver,
	capability types.KBCapability,
	authorizer interfaces.KBAuthorizer,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		mode := types.CurrentKBACLMode()
		if mode == types.KBACLModeOff {
			c.Next()
			return
		}
		if authorizer == nil {
			types.RecordKBACLError()
			logger.Errorf(c.Request.Context(), "[kb_acl] authorizer unavailable; denying capability=%s path=%s", capability, c.Request.URL.Path)
			_ = c.Error(apperrors.NewForbiddenError("knowledge-base capability denied"))
			c.Abort()
			return
		}
		kbID, err := resolveKBID(c)
		if err != nil {
			types.RecordKBACLError()
			_ = c.Error(err)
			c.Abort()
			return
		}
		ctx := c.Request.Context()
		tenantID, ok := types.TenantIDFromContext(ctx)
		userID, userOK := types.UserIDFromContext(ctx)
		if !ok || tenantID == 0 || !userOK || userID == "" {
			_ = c.Error(apperrors.NewUnauthorizedError("Unauthorized"))
			c.Abort()
			return
		}
		decision, err := authorizer.Authorize(ctx, interfaces.KBPolicyRequest{
			TenantID: tenantID, KBID: kbID, UserID: userID,
			TenantRole: types.TenantRoleFromContext(ctx), Capability: capability,
			AgentID: c.Query("agent_id"),
		})
		if err != nil {
			logger.ErrorWithFields(ctx, err, map[string]interface{}{"kb_id": kbID, "capability": capability})
			_ = c.Error(apperrors.NewServiceUnavailableError("cannot verify KB capability right now"))
			c.Abort()
			return
		}
		if !decision.Allowed {
			logger.Warnf(ctx, "[kb_acl] denied tenant=%d user=%s kb=%s capability=%s source=%s reason=%s",
				tenantID, userID, kbID, capability, decision.Source, decision.Reason)
			if mode == types.KBACLModeShadow {
				types.RecordKBACLShadowDenied()
				auditKBCapabilityDenied(c, tenantID, userID, kbID, capability, decision)
				logger.Infof(ctx, "[kb_acl_metric] shadow_denied=1 tenant=%d kb=%s capability=%s", tenantID, kbID, capability)
				c.Next()
				return
			}
			types.RecordKBACLDenied()
			auditKBCapabilityDenied(c, tenantID, userID, kbID, capability, decision)
			_ = c.Error(apperrors.NewForbiddenError("knowledge-base capability denied"))
			c.Abort()
			return
		}
		types.RecordKBACLAllowed()
		c.Set(KBPolicyDecisionContextKey, decision)
		c.Request = c.Request.WithContext(context.WithValue(ctx, KBPolicyDecisionContextKey, decision))
		c.Next()
	}
}

func auditKBCapabilityDenied(
	c *gin.Context,
	tenantID uint64,
	userID, kbID string,
	capability types.KBCapability,
	decision types.KBPolicyDecision,
) {
	svc := AuditServiceFromContext(c)
	if svc == nil {
		return
	}
	details, _ := json.Marshal(map[string]string{
		"required_capability": string(capability),
		"policy_source":       string(decision.Source),
		"policy_reason":       decision.Reason,
	})
	path := c.FullPath()
	if path == "" && c.Request != nil {
		path = c.Request.URL.Path
	}
	method := ""
	if c.Request != nil {
		method = c.Request.Method
	}
	_ = svc.Log(c.Request.Context(), &types.AuditLog{
		TenantID: tenantID, ActorUserID: userID, ActorRole: string(types.TenantRoleFromContext(c.Request.Context())),
		Action: types.AuditActionAccessDenied, ScopeType: "knowledge_base", ScopeID: kbID,
		TargetType: "knowledge_base", TargetID: kbID,
		RequestPath: path, RequestMethod: method, Outcome: types.AuditOutcomeDenied,
		Details: types.JSON(details),
	})
}
