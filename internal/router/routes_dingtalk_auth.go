package router

import (
	"github.com/Tencent/WeKnora/internal/handler"
	"github.com/Tencent/WeKnora/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterDingTalkAuthRoutes(r *gin.RouterGroup, h *handler.DingTalkAuthHandler) {
	limited := middleware.PublicAuthRateLimit()
	r.GET("/auth/dingtalk/config", h.Config)
	r.GET("/auth/dingtalk/start", limited, h.Start)
	r.GET("/auth/dingtalk/callback", limited, h.Callback)
	r.POST("/auth/dingtalk/inside", limited, h.Inside)
}
