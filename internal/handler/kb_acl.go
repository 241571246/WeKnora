package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Tencent/WeKnora/internal/application/service"
	apperrors "github.com/Tencent/WeKnora/internal/errors"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/gin-gonic/gin"
)

type KBACLHandler struct {
	members    interfaces.KBMembershipService
	authorizer interfaces.KBAuthorizer
}

func NewKBACLHandler(members interfaces.KBMembershipService, authorizer interfaces.KBAuthorizer) *KBACLHandler {
	return &KBACLHandler{members: members, authorizer: authorizer}
}

type kbMemberRequest struct {
	UserID       string               `json:"user_id"`
	Role         types.KBMemberRole   `json:"role" binding:"required"`
	Capabilities []types.KBCapability `json:"capabilities"`
}

func (h *KBACLHandler) GetAccess(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	userID, _ := types.UserIDFromContext(ctx)
	capability := types.KBCapability(strings.TrimSpace(c.Query("capability")))
	if capability == "" {
		capability = types.KBCapabilityMetadataRead
	}
	decision, err := h.authorizer.Authorize(ctx, interfaces.KBPolicyRequest{
		TenantID: tenantID, KBID: c.Param("id"), UserID: userID,
		TenantRole: types.TenantRoleFromContext(ctx), Capability: capability,
	})
	if err != nil {
		c.Error(apperrors.NewInternalServerError("failed to resolve access").WithDetails(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": decision})
}

func (h *KBACLHandler) ListMembers(c *gin.Context) {
	rows, err := h.members.ListByKB(c.Request.Context(), types.MustTenantIDFromContext(c.Request.Context()), c.Param("id"))
	if err != nil {
		c.Error(apperrors.NewInternalServerError("failed to list knowledge-base members").WithDetails(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"members": rows}})
}

func (h *KBACLHandler) AddMember(c *gin.Context) {
	var req kbMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.UserID) == "" {
		c.Error(apperrors.NewValidationError("user_id and role are required"))
		return
	}
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	if req.Role == types.KBMemberRoleOwner && !h.requireOwnersManage(c) {
		return
	}
	actor, _ := types.UserIDFromContext(ctx)
	membership := &types.KBMembership{TenantID: tenantID, KBID: c.Param("id"), UserID: req.UserID, Role: req.Role}
	if err := h.members.Create(ctx, actor, membership, req.Capabilities); err != nil {
		h.writeMemberError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": membership})
}

func (h *KBACLHandler) UpdateMember(c *gin.Context) {
	var req kbMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewValidationError("role is required"))
		return
	}
	if !h.canUpdateMemberRole(c, c.Param("user_id"), req.Role) {
		return
	}
	ctx := c.Request.Context()
	actor, _ := types.UserIDFromContext(ctx)
	membership := &types.KBMembership{TenantID: types.MustTenantIDFromContext(ctx), KBID: c.Param("id"), UserID: c.Param("user_id"), Role: req.Role}
	if err := h.members.Update(ctx, actor, membership, req.Capabilities); err != nil {
		h.writeMemberError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": membership})
}

func (h *KBACLHandler) DeleteMember(c *gin.Context) {
	ctx := c.Request.Context()
	current, err := h.members.Get(ctx, types.MustTenantIDFromContext(ctx), c.Param("id"), c.Param("user_id"))
	if err != nil {
		h.writeMemberError(c, err)
		return
	}
	if current != nil && current.Role == types.KBMemberRoleOwner && !h.requireOwnersManage(c) {
		return
	}
	actor, _ := types.UserIDFromContext(ctx)
	if err := h.members.Delete(ctx, actor, types.MustTenantIDFromContext(ctx), c.Param("id"), c.Param("user_id")); err != nil {
		h.writeMemberError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *KBACLHandler) canUpdateMemberRole(c *gin.Context, targetUserID string, nextRole types.KBMemberRole) bool {
	ctx := c.Request.Context()
	current, err := h.members.Get(ctx, types.MustTenantIDFromContext(ctx), c.Param("id"), targetUserID)
	if err != nil {
		h.writeMemberError(c, err)
		return false
	}
	// kb.members.manage controls ordinary member changes. Any transition
	// into or out of Owner (and any mutation of an existing Owner row)
	// additionally requires kb.owners.manage. Otherwise a custom role with
	// only members.manage could demote or remove a co-owner as long as another
	// owner remained, bypassing the separately frozen ownership capability.
	if nextRole != types.KBMemberRoleOwner && (current == nil || current.Role != types.KBMemberRoleOwner) {
		return true
	}
	return h.requireOwnersManage(c)
}

func (h *KBACLHandler) requireOwnersManage(c *gin.Context) bool {
	ctx := c.Request.Context()
	actor, _ := types.UserIDFromContext(ctx)
	decision, err := h.authorizer.Authorize(ctx, interfaces.KBPolicyRequest{
		TenantID: types.MustTenantIDFromContext(ctx), KBID: c.Param("id"), UserID: actor,
		TenantRole: types.TenantRoleFromContext(ctx), Capability: types.KBCapabilityOwnersManage,
	})
	if err != nil || !decision.Allowed {
		c.Error(apperrors.NewForbiddenError("kb.owners.manage capability required"))
		return false
	}
	return true
}

func (h *KBACLHandler) writeMemberError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidKBMemberRole), errors.Is(err, service.ErrInvalidKBCapability):
		c.Error(apperrors.NewValidationError(err.Error()))
	case errors.Is(err, service.ErrKBMemberOutsideWorkspace):
		c.Error(apperrors.NewBadRequestError(err.Error()))
	case errors.Is(err, service.ErrLastKBOwner):
		c.Error(apperrors.NewConflictError(err.Error()))
	default:
		c.Error(apperrors.NewInternalServerError("knowledge-base member operation failed").WithDetails(err.Error()))
	}
}

type KBCollectionHandler struct {
	service    interfaces.KBCollectionService
	authorizer interfaces.KBAuthorizer
}

func NewKBCollectionHandler(service interfaces.KBCollectionService, authorizer interfaces.KBAuthorizer) *KBCollectionHandler {
	return &KBCollectionHandler{service: service, authorizer: authorizer}
}

type kbCollectionRequest struct {
	ParentID  *string `json:"parent_id"`
	Name      string  `json:"name" binding:"required"`
	SortOrder int     `json:"sort_order"`
}

func (h *KBCollectionHandler) GetTree(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	userID, _ := types.UserIDFromContext(ctx)
	visible, err := h.authorizer.ListAccessibleKBIDs(ctx, tenantID, userID, types.TenantRoleFromContext(ctx))
	if err != nil {
		c.Error(apperrors.NewInternalServerError("failed to resolve visible knowledge bases").WithDetails(err.Error()))
		return
	}
	if types.TenantRoleFromContext(ctx) == types.TenantRoleOwner {
		visible = nil
	}
	collections, bindings, err := h.service.GetTree(ctx, tenantID, visible)
	if err != nil {
		c.Error(apperrors.NewInternalServerError("failed to load collection tree").WithDetails(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"collections": collections, "bindings": bindings, "visible_kb_ids": visible}})
}

func (h *KBCollectionHandler) Create(c *gin.Context) {
	var req kbCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewValidationError("name is required"))
		return
	}
	ctx := c.Request.Context()
	actor, _ := types.UserIDFromContext(ctx)
	collection := &types.KBCollection{TenantID: types.MustTenantIDFromContext(ctx), ParentID: req.ParentID, Name: req.Name, SortOrder: req.SortOrder, CreatedBy: actor}
	if err := h.service.Create(ctx, collection); err != nil {
		h.writeCollectionError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": collection})
}

func (h *KBCollectionHandler) Update(c *gin.Context) {
	var req kbCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewValidationError("name is required"))
		return
	}
	ctx := c.Request.Context()
	collection := &types.KBCollection{ID: c.Param("id"), TenantID: types.MustTenantIDFromContext(ctx), ParentID: req.ParentID, Name: req.Name, SortOrder: req.SortOrder}
	if err := h.service.Update(ctx, collection); err != nil {
		h.writeCollectionError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": collection})
}

func (h *KBCollectionHandler) Delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), types.MustTenantIDFromContext(c.Request.Context()), c.Param("id"))
	if err != nil {
		h.writeCollectionError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *KBCollectionHandler) SetBinding(c *gin.Context) {
	var req struct {
		CollectionID string `json:"collection_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewValidationError("invalid request body"))
		return
	}
	ctx := c.Request.Context()
	actor, _ := types.UserIDFromContext(ctx)
	binding := &types.KBCollectionBinding{KBID: c.Param("id"), TenantID: types.MustTenantIDFromContext(ctx), CollectionID: req.CollectionID, UpdatedBy: actor}
	if err := h.service.SetKnowledgeBaseCollection(ctx, binding); err != nil {
		h.writeCollectionError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": binding})
}

func (h *KBCollectionHandler) writeCollectionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrKBCollectionNotFound):
		c.Error(apperrors.NewNotFoundError(err.Error()))
	case errors.Is(err, service.ErrKBCollectionCycle):
		c.Error(apperrors.NewValidationError(err.Error()))
	case errors.Is(err, service.ErrKBCollectionNotEmpty):
		c.Error(apperrors.NewConflictError(err.Error()))
	default:
		c.Error(apperrors.NewBadRequestError(err.Error()))
	}
}
