package handler

import (
	"encoding/json"

	"github.com/Tencent/WeKnora/internal/middleware"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/gin-gonic/gin"
)

// auditKBHTTPAction records security-sensitive HTTP actions that do not pass
// through a domain service carrying AuditLogService (downloads and direct
// chunk deletes). Audit failures are deliberately non-fatal to the business
// response, matching the rest of the audit subsystem.
func auditKBHTTPAction(
	c *gin.Context,
	action types.AuditAction,
	kbID, targetType, targetID string,
	details map[string]any,
) {
	if c == nil || c.Request == nil {
		return
	}
	audit := middleware.AuditServiceFromContext(c)
	if audit == nil {
		return
	}
	ctx := c.Request.Context()
	tenantID, _ := types.TenantIDFromContext(ctx)
	actorID, _ := types.UserIDFromContext(ctx)
	detailJSON, _ := json.Marshal(details)
	_ = audit.Log(ctx, &types.AuditLog{
		TenantID: tenantID, ActorUserID: actorID, ActorRole: string(types.TenantRoleFromContext(ctx)),
		Action: action, ScopeType: "knowledge_base", ScopeID: kbID,
		TargetType: targetType, TargetID: targetID,
		RequestPath: c.FullPath(), RequestMethod: c.Request.Method,
		Outcome: types.AuditOutcomeSuccess, Details: types.JSON(detailJSON),
	})
}
