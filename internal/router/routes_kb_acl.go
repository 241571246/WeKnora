package router

import (
	"github.com/Tencent/WeKnora/internal/handler"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/gin-gonic/gin"
)

func RegisterKBACLRoutes(r *gin.RouterGroup, h *handler.KBACLHandler, g *rbacGuards) {
	if h == nil {
		return
	}
	// Access introspection is safe for every authenticated workspace member:
	// the Authorizer returns only the caller's own decision/default denial and
	// does not expose KB content. Gating this endpoint by metadata.read would
	// create a circular dependency for capability-aware clients.
	r.GET("/knowledge-bases/:id/access", g.Viewer(), h.GetAccess)
	r.GET("/knowledge-bases/:id/members", g.Viewer(), g.KBCapability("id", types.KBCapabilityMembersManage), h.ListMembers)
	r.POST("/knowledge-bases/:id/members", g.Viewer(), g.KBCapability("id", types.KBCapabilityMembersManage), h.AddMember)
	r.PATCH("/knowledge-bases/:id/members/:user_id", g.Viewer(), g.KBCapability("id", types.KBCapabilityMembersManage), h.UpdateMember)
	r.DELETE("/knowledge-bases/:id/members/:user_id", g.Viewer(), g.KBCapability("id", types.KBCapabilityMembersManage), h.DeleteMember)
}

func RegisterKBCollectionRoutes(r *gin.RouterGroup, h *handler.KBCollectionHandler, g *rbacGuards) {
	if h == nil {
		return
	}
	r.GET("/knowledge-base-collections/tree", g.Viewer(), h.GetTree)
	r.POST("/knowledge-base-collections", g.Owner(), h.Create)
	r.PATCH("/knowledge-base-collections/:id", g.Owner(), h.Update)
	r.DELETE("/knowledge-base-collections/:id", g.Owner(), h.Delete)
	r.PUT("/knowledge-bases/:id/collection", g.Viewer(), g.KBCapability("id", types.KBCapabilitySettingsEdit), h.SetBinding)
}
